package resistance

import (
	"github.com/jakecoffman/set-game/gamelib"
	"log"
	"time"
	"github.com/google/uuid"
	"math/rand"
	"fmt"
	"encoding/json"
	"strconv"
)

type Resist struct {
	cmd chan *gamelib.Command `json:"-"`

	Id string

	Players      []*Player
	playerCursor int
	Leader       int // Leader is the position of the leader in the Players list

	State          string
	Missions       []*Mission
	CurrentMission int
	History        []*History
	NumFailed      int
	Version        int
	Created        time.Time `json:"-"`
	Updated        time.Time `json:"-"`
}

type Player struct {
	ws        gamelib.Connector
	Uuid      string `json:"-"`
	Id        int
	Name      string
	Connected bool
	Ip        string `json:"-"`
	IsSpy     bool   `json:"-"`
	IsBot     bool
	IsReady   bool
	IsLeader  bool
	OnMission bool
}

// Find returns the player object and the position they are in
func Find(players []*Player, uuid string) (*Player, int) {
	for i, player := range players {
		if player.Uuid == uuid {
			return player, i
		}
	}
	return nil, -1
}

type Mission struct {
	Slots        int          // number of people that need to go
	Assignments  []int        // the players that will go on the mission
	Votes        map[int]bool // true is accept
	successVotes map[int]bool // true is pass (this is kept secret)
	Success      bool         // success/fail result
	NumFails     int          // number of fail votes on mission

	Complete bool // just a flag to tell if the mission has finished
}

type History struct {
	Mission     int
	Assignments []int
	Votes       map[int]bool
}

func NewMissions(slots []int) []*Mission {
	var missions []*Mission
	for _, slot := range slots {
		missions = append(missions, &Mission{
			Slots: slot, Assignments: []int{}, Votes: map[int]bool{}, successVotes: map[int]bool{}})
	}
	return missions
}

func NewGame(id string) gamelib.Game {
	g := &Resist{
		cmd: make(chan *gamelib.Command),

		Players:      []*Player{},
		playerCursor: 1,
		Id:           id,
		Created:      time.Now(),
	}
	go g.run()
	g.reset()
	return g
}

func (g *Resist) reset() {
	g.State = stateLobby
	g.History = []*History{}
	g.Missions = []*Mission{}
	g.NumFailed = 0
	g.CurrentMission = 0
	g.Leader = 0
	for _, p := range g.Players {
		p.IsSpy = false
		p.IsLeader = false
		p.IsReady = false
		p.OnMission = false
		p.IsReady = false
	}
}

func (g *Resist) Cmd(c *gamelib.Command) {
	g.cmd <- c
}

// states
const (
	stateLobby         = "lobby"
	stateTeambuilding  = "building"
	stateTeamvoting    = "voting"
	stateMission       = "mission"
	stateSpywin        = "spywin"
	stateResistanceWin = "resistancewin"
)

// message types
const (
	cmdJoin       = "join"
	cmdLeave      = "leave"
	cmdDisconnect = "disconnect"
	cmdStop       = "stop"
	cmdName       = "name"

	// anyone can do these things
	cmdAddBot    = "addbot"
	cmdRemoveBot = "removebot"
	cmdStart     = "start"

	// leader assigns a team
	cmdAssign = "assign"

	cmdVoteTeam    = "voteteam"
	cmdVoteMission = "votemission"
	cmdReady       = "ready" // make a new game, or start current game
)

func (g *Resist) run() {
	var cmd *gamelib.Command
	var update bool
	for {
		// handle the case where a bot is leader
		if g.State == stateTeambuilding && g.Players[g.Leader].IsBot {
			thisMission := g.Missions[g.CurrentMission]
			thisMission.Assignments = rand.Perm(len(g.Players))[0:thisMission.Slots]
			for _, i := range thisMission.Assignments {
				g.Players[i].OnMission = true
			}
			g.State = stateTeamvoting
			g.sendEveryoneEverything()
		}
		cmd = <-g.cmd

		switch cmd.Type {
		case cmdJoin:
			update = g.handleJoin(cmd)
		case cmdLeave:
			update = g.handleLeave(cmd)
		case cmdDisconnect:
			update = g.handleDisconnect(cmd)
		case cmdStop:
			return
		case cmdAddBot:
			update = g.handleAddBot(cmd)
		case cmdRemoveBot:
			update = g.handleRemoveBot(cmd)
		case cmdStart:
			update = g.handleStart(cmd)
		case cmdAssign: // leader sent his chosen assignment
			update = g.handleAssignTeam(cmd)
		case cmdVoteTeam:
			update = g.handleVote(cmd)
		case cmdVoteMission:
			update = g.handleMission(cmd)
		case cmdReady:
			update = g.handleReady(cmd)
		case cmdName:
			update = g.handleName(cmd)
		default:
			log.Println("Unknown message:", cmd.Type)
			continue
		}
		if update {
			g.sendEveryoneEverything()
			g.Updated = time.Now()
		}
	}
}

type UpdateMsg struct {
	Type   string
	Update *Resist
	You    *secret
}

type secret struct {
	Id                           int
	Spies                        []int
	IsReady, IsLeader, OnMission bool
}

func (g *Resist) sendEveryoneEverything() {
	var spies []int
	for i, p := range g.Players {
		if p.IsSpy {
			spies = append(spies, i)
		}
	}
	for _, p := range g.Players {
		if p.ws != nil {
			msg := &UpdateMsg{Type: "all", Update: g}
			msg.You = &secret{Id: p.Id, IsReady: p.IsReady, IsLeader: p.IsLeader, OnMission: p.OnMission}
			if p.IsSpy {
				msg.You.Spies = spies
			}
			p.ws.Send(msg)
		}
	}
}

type MsgMsg struct {
	Type string
	Msg  string
}

func sendMsg(c gamelib.Connector, msg string) {
	c.Send(&MsgMsg{Type: "msg", Msg: msg})
}

func (g *Resist) resetReadies() {
	for _, p := range g.Players {
		p.IsReady = false
	}
}

func (g *Resist) handleJoin(cmd *gamelib.Command) bool {
	player, i := Find(g.Players, cmd.PlayerId)
	if i == -1 {
		// player was not here before
		if g.State != stateLobby {
			sendMsg(cmd.Ws, "Can't join game in progress")
			return false
		}
		if len(g.Players) >= 10 {
			// can't have more than 10 players
			sendMsg(cmd.Ws, "Can't have more than 10 players")
			return false
		}
		player = &Player{Uuid: cmd.PlayerId, Id: g.playerCursor}
		g.Players = append(g.Players, player)
		g.playerCursor += 1
	}
	player.ws = cmd.Ws
	player.Connected = true
	player.Ip = player.ws.Request().Header.Get("X-Forwarded-For")
	return true
}

func (g *Resist) handleLeave(cmd *gamelib.Command) bool {
	for i, player := range g.Players {
		if player.Uuid == cmd.PlayerId {
			g.Players = append(g.Players[0:i], g.Players[i+1:]...)
			return true
		}
	}
	return false
}

func (g *Resist) handleDisconnect(cmd *gamelib.Command) bool {
	player, i := Find(g.Players, cmd.PlayerId)
	if i == -1 {
		log.Println("Couldn't find player", cmd.PlayerId)
		return false
	}
	player.ws = nil
	player.Connected = false
	return true
}

func (g *Resist) handleReady(cmd *gamelib.Command) bool {
	allReady := true
	for _, p := range g.Players {
		if p.Uuid == cmd.PlayerId {
			p.IsReady = !p.IsReady
		}
		if !p.IsBot && !p.IsReady {
			allReady = false
		}
	}
	if allReady {
		switch g.State {
		case stateLobby:
			// TODO move start logic down here
			g.State = stateTeambuilding
			g.resetReadies()
		case stateSpywin:
			fallthrough
		case stateResistanceWin:
			g.reset()
		default:
			log.Println("Error: everyone voted ready but I am in state", g.State)
			g.resetReadies()
		}
	}
	return true
}

func (g *Resist) handleStart(cmd *gamelib.Command) bool {
	if g.Version != cmd.Version {
		sendMsg(cmd.Ws, "Someone else started the game first")
		return false
	}

	if g.State != stateLobby {
		sendMsg(cmd.Ws, "Illegal state")
		return false
	}

	if len(g.Players) < 5 || len(g.Players) > 10 {
		sendMsg(cmd.Ws, "Need 5-10 players to start the game")
		return false
	}

	g.State = stateTeambuilding
	g.CurrentMission = 0
	g.NumFailed = 0

	// remove unconnected players and reorder them, leader always starts in position 1
	{
		var newPlayers []*Player
		walk := rand.Perm(len(g.Players))
		for _, i := range walk {
			if !g.Players[i].IsBot && !g.Players[i].Connected {
				continue
			}
			newPlayers = append(newPlayers, g.Players[i])
		}
		g.Players = newPlayers
		g.Leader = 0
		g.Players[0].IsLeader = true
	}
	// assign secret roles to players (based on # of players)
	{
		numSpies := map[int]int{5: 2, 6: 2, 7: 3, 8: 3, 9: 3, 10: 4}[len(g.Players)]
		walk := rand.Perm(len(g.Players))
		for i, j := range walk {
			if i >= numSpies {
				break
			}
			g.Players[j].IsSpy = true
		}
	}
	// init missions based on amount of players
	{
		slots := map[int][]int{
			5:  {2, 3, 2, 3, 3},
			6:  {2, 3, 4, 3, 4},
			7:  {2, 3, 3, 4, 4},
			8:  {3, 4, 4, 5, 5},
			9:  {3, 4, 4, 5, 5},
			10: {3, 4, 4, 5, 5},
		}[len(g.Players)]
		g.Missions = NewMissions(slots)
	}
	return true
}

func (g *Resist) handleAddBot(cmd *gamelib.Command) bool {
	if len(g.Players) >= 10 {
		sendMsg(cmd.Ws, "Can't have more than 10 players")
		return false
	}
	player := &Player{Uuid: uuid.New().String(), Id: g.playerCursor, IsBot: true}
	g.Players = append(g.Players, player)
	g.playerCursor += 1
	return true
}

func (g *Resist) handleRemoveBot(cmd *gamelib.Command) bool {
	for i, p := range g.Players {
		if p.IsBot {
			g.Players = append(g.Players[0:i], g.Players[i+1:]...)
			return true
		}
	}
	sendMsg(cmd.Ws, "These aren't the bot's you're looking for...")
	return false
}

func (g *Resist) handleAssignTeam(cmd *gamelib.Command) bool {
	_, i := Find(g.Players, cmd.PlayerId)
	if g.Version != cmd.Version || g.State != stateTeambuilding || g.Leader != i {
		return false
	}
	var assignment []int
	err := json.Unmarshal(cmd.Data, &assignment)
	if err != nil {
		log.Println(err)
		sendMsg(cmd.Ws, "Got invalid data for team assignment")
		return false
	}
	thisMission := g.Missions[g.CurrentMission]
	if len(assignment) != thisMission.Slots {
		sendMsg(cmd.Ws, fmt.Sprint("Number of assignments needs to be ", thisMission.Slots, " but got ", len(assignment)))
		return false
	}
	thisMission.Assignments = assignment
	for _, i := range thisMission.Assignments {
		g.Players[i].OnMission = true
	}
	g.State = stateTeamvoting
	g.Version += 1
	return true
}

func (g *Resist) handleVote(cmd *gamelib.Command) bool {
	_, i := Find(g.Players, cmd.PlayerId)
	if g.Version != cmd.Version || g.State != stateTeamvoting {
		log.Println(g.Version, cmd.Version, g.State)
		return false
	}
	thisMission := g.Missions[g.CurrentMission]
	var vote bool
	err := json.Unmarshal(cmd.Data, &vote)
	if err != nil {
		log.Println(err)
		sendMsg(cmd.Ws, "Got invalid data for team assignment")
		return false
	}
	thisMission.Votes[i] = vote

	// this makes bots vote every time, not efficient but who cares
	for i, player := range g.Players {
		if player.IsBot {
			if g.NumFailed < 4 {
				thisMission.Votes[i] = rand.Intn(2) == 1
			} else {
				thisMission.Votes[i] = true
			}
		}
	}

	if len(thisMission.Votes) == len(g.Players) {
		g.History = append(g.History, &History{Mission: g.CurrentMission, Assignments: thisMission.Assignments, Votes: thisMission.Votes})
		g.Version += 1
		yeas := 0
		for _, vote := range thisMission.Votes {
			if vote {
				yeas += 1
			}
		}
		if yeas > (len(g.Players) / 2) {
			g.State = stateMission
			g.NumFailed = 0
			// make the bots pre-vote on the mission
			for _, i := range g.Missions[g.CurrentMission].Assignments {
				if g.Players[i].IsBot {
					p := g.Players[i]
					go func(bot *Player, v int) {
						g.cmd <- &gamelib.Command{
							PlayerId: bot.Uuid,
							Type:     cmdVoteMission,
							Data:     []byte(strconv.FormatBool(!bot.IsSpy)),
							Version:  v,
						}
					}(p, g.Version)
				}
			}
		} else {
			g.NumFailed += 1
			if g.NumFailed == 5 {
				g.State = stateSpywin
			} else {
				g.State = stateTeambuilding
				g.Players[g.Leader].IsLeader = false
				g.Leader += 1
				if g.Leader >= len(g.Players) {
					g.Leader = 0
				}
				g.Players[g.Leader].IsLeader = true
				thisMission.Assignments = []int{}
				thisMission.Votes = map[int]bool{}
			}
			for _, p := range g.Players {
				p.OnMission = false
			}
		}
	}
	return true
}

func (g *Resist) handleMission(cmd *gamelib.Command) bool {
	p, i := Find(g.Players, cmd.PlayerId)

	if g.Version != cmd.Version || g.State != stateMission || !p.OnMission {
		log.Println(g.Version, cmd.Version, g.State, p.OnMission)
		return false
	}

	var vote bool
	err := json.Unmarshal(cmd.Data, &vote)
	if err != nil {
		log.Println(err)
		sendMsg(p.ws, "Got invalid data for team assignment")
		return false
	}

	thisMission := g.Missions[g.CurrentMission]
	if !p.IsSpy && vote == false {
		sendMsg(p.ws, "Resistance cannot vote to fail missions")
		return false
	}
	thisMission.successVotes[i] = vote

	// voting is done
	if len(thisMission.successVotes) == len(thisMission.Assignments) {
		g.Version += 1
		g.CurrentMission += 1
		g.Players[g.Leader].IsLeader = false
		if g.Leader >= len(g.Players) {
			g.Leader = 0
		}
		for _, p := range g.Players {
			p.OnMission = false
		}
		thisMission.Complete = true
		// figure out if successful or not
		thisMission.Success = true
		for _, vote := range thisMission.successVotes {
			if vote == false {
				thisMission.Success = false
				thisMission.NumFails += 1
			}
		}
		// check end game by counting up number of successful/failed missions
		succeeds := 0
		fails := 0
		for _, m := range g.Missions {
			if m.Complete {
				if m.Success {
					succeeds += 1
				} else {
					fails += 1
				}
			}
		}
		if succeeds >= 3 {
			g.State = stateResistanceWin
			g.resetReadies()
		} else if fails >= 3 {
			g.State = stateSpywin
			g.resetReadies()
		} else {
			// game is not over, assign the next leader
			g.State = stateTeambuilding
			g.Leader += 1
			g.Players[g.Leader].IsLeader = true
		}
	}
	return true
}

func (g *Resist) handleName(cmd *gamelib.Command) bool {
	p, _ := Find(g.Players, cmd.PlayerId)
	if g.State != stateLobby && p.Name != "" {
		sendMsg(p.ws, "Wait for the lobby to change your name again")
		return false
	}

	var name string
	err := json.Unmarshal(cmd.Data, &name)
	if err != nil {
		log.Println(err)
		sendMsg(p.ws, "Got invalid data for team assignment")
		return false
	}

	if len(name) > 8 {
		p.Name = name[0:8]
	} else {
		p.Name = name
	}

	return true
}
