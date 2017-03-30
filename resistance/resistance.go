package resistance

import (
	"github.com/jakecoffman/set-game/gamelib"
	"log"
	"time"
	"github.com/google/uuid"
	"math/rand"
)

type Resist struct {
	cmd            chan *ResistCmd `json:"-"`

	Id             string

	Players        []*Player
	playerCursor   int
	Leader         int

	State          string
	Missions       []*Mission
	CurrentMission int
	NumFailed      int
	Version        int
	Created        time.Time `json:"-"`
	Updated        time.Time `json:"-"`
}

type Mission struct {
	Slots        int          // number of people that need to go
	Assignments  []int        // the players that will go on the mission
	Votes        map[int]bool // true is accept
	successVotes map[int]bool // true is pass (this is kept secret)
	Success      bool         // success/fail result

	Complete     bool         // just a flag to tell if the mission has finished
}

func NewMissions(slots []int) []*Mission {
	missions := []*Mission{}
	for _, slot := range slots {
		missions = append(missions, &Mission{
			Slots: slot, Assignments: []int{}, Votes: map[int]bool{}, successVotes: map[int]bool{}})
	}
	return missions
}

func NewGame(id string) *Resist {
	g := &Resist{
		cmd: make(chan *ResistCmd),

		Players: []*Player{},
		playerCursor: 1,
		Id: id,
		Created: time.Now(),
	}
	go g.run()
	g.reset()
	return g
}

func (g *Resist) reset() {
	g.State = state_lobby
}

func (g *Resist) Cmd(c gamelib.Command) {
	g.cmd <- c.(*ResistCmd)
}

// states
const (
	state_lobby = "lobby"
	state_teambuilding = "building"
	state_teamvoting = "voting"
	state_mission = "mission"
	state_spywin = "spywin"
	state_resistance_win = "resistancewin"
)

// message types
const (
	msg_join = "join"
	msg_leave = "leave"
	msg_disconnect = "disconnect"
	msg_stop = "stop"

	// anyone can do these things
	msg_addbot = "addbot"
	msg_removebot = "removebot"
	msg_start = "start"
	msg_end = "end"

	// leader assigns a team
	msg_assign = "assign"

	msg_vote_team = "voteteam"
	msg_vote_mission = "votemission"
	msg_ready = "ready" // make a new game, or start current game
)

func (g *Resist) run() {
	var cmd *ResistCmd
	for {
		if g.State == state_teambuilding && g.Players[g.Leader].IsBot {
			// assign a random team
			log.Println("Bot is assigning a random team")
			thisMission := g.Missions[g.CurrentMission]
			thisMission.Assignments = rand.Perm(len(g.Players))[0:thisMission.Slots]
			for _, i := range thisMission.Assignments {
				g.Players[i].OnMission = true
			}
			g.State = state_teamvoting
			g.sendEveryoneEverything()
		}
		cmd = <-g.cmd
		if !cmd.IsValid() {
			log.Printf("Invalid command sent: %#v\n", cmd)
			continue
		}
		log.Println("Processing", cmd.Type)
		switch cmd.Type {
		case msg_join:
			player, i := Find(g.Players, cmd.PlayerId)
			if i == -1 {
				// player was not here before
				if len(g.Players) >= 10 {
					// can't have more than 10 players
					continue
				}
				player = &Player{Uuid: cmd.PlayerId, Id: g.playerCursor}
				g.Players = append(g.Players, player)
				g.playerCursor += 1
			}
			player.ws = cmd.Ws
			player.Connected = true
			player.Ip = player.ws.Request().Header.Get("X-Forwarded-For")
		case msg_leave:
			Remove(g.Players, cmd.PlayerId)
		case msg_disconnect:
			player, i := Find(g.Players, cmd.PlayerId)
			if i == -1 {
				log.Println("Couldn't find player", cmd.PlayerId)
				continue
			}
			player.ws = nil
			player.Connected = false
		case msg_stop:
			return
		case msg_addbot:
			if len(g.Players) >= 10 {
				continue
			}
			player := &Player{Uuid: uuid.New().String(), Id: g.playerCursor, IsBot: true}
			g.Players = append(g.Players, player)
			g.playerCursor += 1
		case msg_removebot:
			for _, p := range g.Players {
				if p.IsBot {
					Remove(g.Players, p.Uuid)
					break
				}
			}
		case msg_start:
			if g.Version != cmd.Version || g.State != state_lobby || len(g.Players) < 5 {
				continue
			}
			g.State = state_teambuilding
			g.CurrentMission = 0
			g.NumFailed = 0

			// remove unconnected players and reorder them, leader always starts in position 1
			{
				newPlayers := []*Player{}
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
				numResistance := map[int]int{5: 3, 6: 4, 7: 4, 8: 5, 9: 6, 10: 6}[len(g.Players)]
				for i := range rand.Perm(numResistance) {
					g.Players[i].IsSpy = true
				}
			}
			// init missions based on amount of players
			{
				slots := map[int][]int{
					5: []int{2, 3, 2, 3, 3},
					6: []int{2, 3, 4, 3, 4},
					7: []int{2, 3, 3, 4, 4},
					8: []int{3, 4, 4, 5, 5},
					9: []int{3, 4, 4, 5, 5},
					10: []int{3, 4, 4, 5, 5},
				}[len(g.Players)]
				g.Missions = NewMissions(slots)
			}
		case msg_assign: // leader sent his chosen assignment
			_, i := Find(g.Players, cmd.PlayerId)
			if g.Version != cmd.Version || g.State != state_teambuilding || g.Leader != i {
				continue
			}
			thisMission := g.Missions[g.CurrentMission]
			if len(cmd.UserInput.Assignment) != thisMission.Slots {
				log.Println("Number of assignments needs to be", thisMission.Slots, "but got", len(cmd.UserInput.Assignment))
				continue
			}
			thisMission.Assignments = cmd.UserInput.Assignment
			for _, i := range thisMission.Assignments {
				g.Players[i].OnMission = true
			}
			g.State = state_teamvoting
			g.Version += 1
		case msg_vote_team:
			_, i := Find(g.Players, cmd.PlayerId)
			if g.Version != cmd.Version || g.State != state_teamvoting {
				log.Println(g.Version, cmd.Version, g.State)
				continue
			}
			thisMission := g.Missions[g.CurrentMission]
			thisMission.Votes[i] = cmd.UserInput.Vote

			// this makes bots vote every time, not efficient but who cares
			for i, player := range g.Players {
				if player.IsBot {
					// TODO: get bot personalities to decide on how to vote
					thisMission.Votes[i] = true // bots always vote true, for now
				}
			}

			if len(thisMission.Votes) == len(g.Players) {
				g.Version += 1
				yeas := 0
				for _, vote := range thisMission.Votes {
					if vote {
						yeas += 1
					}
				}
				if yeas > (len(g.Players) / 2) {
					g.State = state_mission
					// make the bots pre-vote on the mission
					for _, i := range g.Missions[g.CurrentMission].Assignments {
						if g.Players[i].IsBot {
							p := g.Players[i]
							go func(bot *Player, v int) {
								g.cmd<-&ResistCmd{
									PlayerId: bot.Uuid,
									UserInput: &UserInput{
										Type: msg_vote_mission, Vote: !bot.IsSpy, Version: v,
									},
								}
							}(p, g.Version)
						}
					}
				} else {
					g.NumFailed += 1
					if g.NumFailed == 5 {
						g.State = state_spywin
					} else {
						g.State = state_teambuilding
						thisMission.Assignments = []int{}
						thisMission.Votes = map[int]bool{}
					}
					for _, p := range g.Players {
						p.OnMission = false
					}
				}
			} else {
				log.Println("Not enough votes yet")
			}
		case msg_vote_mission:
			p, i := Find(g.Players, cmd.PlayerId)

			if g.Version != cmd.Version || g.State != state_mission || !p.OnMission {
				log.Println(g.Version, cmd.Version, g.State, p.OnMission)
				continue
			}

			thisMission := g.Missions[g.CurrentMission]
			thisMission.successVotes[i] = cmd.Vote

			// voting is done
			if len(thisMission.successVotes) == len(thisMission.Assignments) {
				log.Println("Voting is done")
				g.Version += 1
				g.CurrentMission += 1
				g.Players[g.Leader].IsLeader = false
				if g.Leader >= len(g.Players) {
					g.Leader = 0
				}
				g.State = state_teambuilding // go back to team-building by default, unless game ends
				for _, p := range g.Players {
					p.OnMission = false
				}
				thisMission.Complete = true
				thisMission.Success = true
				for _, vote := range thisMission.successVotes {
					if vote == false {
						thisMission.Success = false
					}
				}
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
					g.State = state_resistance_win
					g.resetReadies()
				} else if fails >= 3 {
					g.State = state_spywin
					g.resetReadies()
				} else {
					// game is not over, assign the next leader
					g.Leader += 1
					g.Players[g.Leader].IsLeader = true
				}
			}
		case msg_ready:
			allReady := true
			for _, p := range g.Players {
				if p.Uuid == cmd.PlayerId {
					p.IsReady = true
				}
				if !p.IsBot && !p.IsReady {
					allReady = false
				}
			}
			if allReady {
				switch g.State {
				case state_lobby:
					// TODO move start logic down here
					g.State = state_teambuilding
					g.resetReadies()
				case state_spywin:
					fallthrough
				case state_resistance_win:
					g.State = state_lobby
					g.resetReadies()
				default:
					log.Println("Error: everyone voted ready but I am in state", g.State)
				}
			}
		case msg_end:
			g.reset()
		default:
			log.Println("Unknown message:", cmd.Type)
			continue
		}
		g.sendEveryoneEverything()
		g.Updated = time.Now()
	}
}

type UpdateMsg struct {
	Type   string
	Update *Resist
	You    you
}

func (g *Resist) sendEveryoneEverything() {
	msg := &UpdateMsg{Type: "all", Update: g}
	for _, player := range g.Players {
		if player.ws != nil {
			msg.You = you(*player)
			player.ws.Send(msg)
		}
	}
}

func (g *Resist) resetReadies() {
	for _, p := range g.Players {
		p.IsReady = false
	}
}

func Contains(list []int, i int) bool {
	for _, item := range list {
		if item == i {
			return true
		}
	}
	return false
}