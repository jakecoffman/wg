package justone

import (
	"encoding/json"
	"github.com/jakecoffman/wg"
	"log"
	"math/rand"
	"runtime/debug"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type JustOne struct {
	*wg.Game

	Players      []*Player
	playerCursor int
	State        string

	// the word to be guessed
	GuessMe string
	// the round number
	Round int
	// how many rounds won
	Score int
	// list of clues to give to the guesser
	Clues []string
}

type Player struct {
	ws        wg.Connector
	Uuid      string `json:"-"`
	Id        int
	Name      string
	Connected bool
	Ip        string `json:"-"`
	Ready     bool
	// the word the player wants to give to the guesser
	Clue string `json:",omitempty"`
	// is the player the guesser this round?
	IsGuesser bool `json:",omitempty"`
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

func NewGame(id string) *wg.Game {
	g := &JustOne{
		Players:      []*Player{},
		playerCursor: 1,
	}
	g.Game = wg.NewGame(g, id)
	go g.run()
	g.reset()
	return g.Game
}

func (g *JustOne) reset() {
	g.State = stateLobby
}

// states
const (
	stateLobby = "lobby"
	stateWrite = "writing"
	stateGuess = "guessing"
	stateEnd   = "end"
)

// message types
const (
	cmdJoin       = "join"
	cmdLeave      = "leave"
	cmdDisconnect = "disconnect"
	cmdStop       = "stop"
	cmdName       = "name"

	cmdReady = "ready"
	cmdWrite = "write"
	cmdGuess = "guess"
)

func (g *JustOne) run() {
	var cmd *wg.Command

	defer func() {
		if r := recover(); r != nil {
			log.Println("Game crashed", r)
			log.Printf("State: %#v\n", g)
			log.Println("Last command received:", cmd)
			debug.PrintStack()
		}
	}()

	var update bool
	for {
		cmd = <-g.Cmd

		if g.Version != cmd.Version {
			continue
		}

		switch cmd.Type {
		case cmdJoin:
			update = g.handleJoin(cmd)
		case cmdLeave:
			update = g.handleLeave(cmd)
		case cmdDisconnect:
			update = g.handleDisconnect(cmd)
		case cmdStop:
			return
		case cmdReady:
			update = g.handleReady(cmd)
		case cmdName:
			update = g.handleName(cmd)
		case cmdWrite:
			update = g.handleWrite(cmd)
		case cmdGuess:
			update = g.handleGuess(cmd)
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
	Update *JustOne
}

func (g *JustOne) sendEveryoneEverything() {
	for _, p := range g.Players {
		if p.ws != nil {
			msg := &UpdateMsg{Type: "all", Update: g}
			p.ws.Send(msg)
		}
	}
}

type MsgMsg struct {
	Type string
	Msg  string
}

func sendMsg(c wg.Connector, msg string) {
	c.Send(&MsgMsg{Type: "msg", Msg: msg})
}

func (g *JustOne) sendMsgAll(msg string) {
	for _, p := range g.Players {
		if p.ws != nil {
			p.ws.Send(&MsgMsg{Type: "msg", Msg: msg})
		}
	}
}

func (g *JustOne) handleJoin(cmd *wg.Command) bool {
	player, i := Find(g.Players, cmd.PlayerId)
	if i == -1 {
		// player was not here before
		if g.State != stateLobby {
			sendMsg(cmd.Ws, "Can't join game in progress")
			return false
		}
		if len(g.Players) >= 7 {
			sendMsg(cmd.Ws, "Can't have more than 7 players")
			return false
		}
		player = &Player{Uuid: cmd.PlayerId, Id: g.playerCursor}
		g.Players = append(g.Players, player)
		g.playerCursor += 1
	}
	player.ws = cmd.Ws
	player.Connected = true
	player.Ip = player.ws.Ip()
	return true
}

func (g *JustOne) handleLeave(cmd *wg.Command) bool {
	for i, player := range g.Players {
		if player.Uuid == cmd.PlayerId {
			g.Players = append(g.Players[0:i], g.Players[i+1:]...)
			return true
		}
	}
	return false
}

func (g *JustOne) handleDisconnect(cmd *wg.Command) bool {
	player, i := Find(g.Players, cmd.PlayerId)
	if i == -1 {
		log.Println("Couldn't find player", cmd.PlayerId)
		return false
	}
	player.ws = nil
	player.Connected = false
	return true
}

func (g *JustOne) handleName(cmd *wg.Command) bool {
	p, _ := Find(g.Players, cmd.PlayerId)
	if g.State != stateLobby && p.Name != "" {
		sendMsg(p.ws, "Can only change name in the lobby")
		return false
	}

	var name string
	err := json.Unmarshal(cmd.Data, &name)
	if err != nil {
		log.Println(err)
		sendMsg(p.ws, "Got invalid data for name")
		return false
	}

	if len(name) > 8 {
		p.Name = name[0:8]
	} else {
		p.Name = name
	}

	return true
}

func (g *JustOne) handleReady(cmd *wg.Command) bool {
	p, _ := Find(g.Players, cmd.PlayerId)

	if g.State != stateLobby && g.State != stateEnd {
		sendMsg(p.ws, "Already ready already")
		return false
	}

	p.Ready = true
	for _, player := range g.Players {
		player.Clue = ""
		player.IsGuesser = false
		if !player.Ready {
			return true
		}
	}
	g.State = stateGuess
	g.GuessMe = wordlist[rand.Intn(len(wordlist))]
	g.Players[rand.Intn(len(g.Players))].IsGuesser = true

	return true
}

func (g *JustOne) handleWrite(cmd *wg.Command) bool {
	p, _ := Find(g.Players, cmd.PlayerId)

	if g.State != stateWrite {
		sendMsg(p.ws, "Not in write state")
		return false
	}

	if p.IsGuesser {
		sendMsg(p.ws, "Guesser doesn't write...")
		return false
	}

	err := json.Unmarshal(cmd.Data, &p.Clue)
	if err != nil {
		sendMsg(p.ws, err.Error())
		return false
	}

	clues := map[string]struct{}{}
	for _, player := range g.Players {
		player.Ready = false
		if player.Clue == "" {
			return false
		}
		clues[player.Clue] = struct{}{}
	}
	// all clues are in
	g.Clues = []string{}
	for clue := range clues {
		g.Clues = append(g.Clues, clue)
	}

	return true
}

func (g *JustOne) handleGuess(cmd *wg.Command) bool {
	p, _ := Find(g.Players, cmd.PlayerId)

	if g.State != stateGuess {
		sendMsg(p.ws, "Not in guess state")
		return false
	}

	if !p.IsGuesser {
		sendMsg(p.ws, "Not the guesser")
		return false
	}

	g.State = stateEnd
	var guess string
	err := json.Unmarshal(cmd.Data, &guess)
	if err != nil {
		sendMsg(p.ws, err.Error())
		return false
	}
	if strings.ToUpper(guess) == g.GuessMe {
		g.Score++
	}

	// reset game state
	var guesserIndex int
	for i, player := range g.Players {
		if player.IsGuesser {
			guesserIndex = i
		}
		player.Clue = ""
		player.IsGuesser = false
		player.Ready = false
	}
	guesserIndex++
	g.Players[guesserIndex%len(g.Players)].IsGuesser = true
	g.Clues = nil

	g.Round++
	if g.Round >= 13 {
		g.State = stateEnd
	} else {
		// pick another word from the list
		g.GuessMe = wordlist[rand.Intn(len(wordlist))]
		g.State = stateWrite
	}

	return true
}

func (g *JustOne) String() string {
	b, _ := json.Marshal(g)
	return string(b)
}
