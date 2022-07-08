package justone

import (
	"encoding/json"
	"fmt"
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
	guessMe string
	// the round number
	Round int
	// how many rounds won
	Score int
	// list of clues to give to the guesser
	clues []Clue
}

type Clue struct {
	Text string
	Dupe bool
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
	stateLobby     = "lobby"
	stateWrite     = "writing"
	stateReconcile = "reconcile"
	stateGuess     = "guessing"
	stateEnd       = "end"
)

// message types
const (
	cmdJoin       = "join"
	cmdLeave      = "leave"
	cmdDisconnect = "disconnect"
	cmdStop       = "stop"
	cmdName       = "name"

	cmdReady     = "ready"
	cmdWrite     = "write"
	cmdReconcile = "reconcile"
	cmdGuess     = "guess"
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
		case cmdReconcile:
			update = g.handleReconcile(cmd)
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
	You    *You
}

type You struct {
	Id        int
	IsGuesser bool   `json:",omitempty"`
	IsReady   bool   `json:",omitempty"`
	Clue      string `json:",omitempty"`

	// hidden from guesser
	Clues   []Clue `json:",omitempty"`
	GuessMe string `json:",omitempty"`
}

func (g *JustOne) sendEveryoneEverything() {
	for _, p := range g.Players {
		if p.ws != nil {
			msg := &UpdateMsg{Type: "all", Update: g}
			msg.You = &You{
				Id:        p.Id,
				IsGuesser: p.IsGuesser,
				IsReady:   p.Ready,
				Clue:      p.Clue,
				Clues:     g.clues,
				GuessMe:   g.guessMe,
			}
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
	if g.State != stateLobby && g.State != stateEnd {
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

	if g.State != stateLobby && g.State != stateEnd && g.State != stateReconcile {
		sendMsg(p.ws, "Already ready already")
		return false
	}

	if len(g.Players) < 2 {
		sendMsg(p.ws, "You need at least 2 players")
		return false
	}

	err := json.Unmarshal(cmd.Data, &p.Ready)
	if err != nil {
		sendMsg(p.ws, err.Error())
		return false
	}

	for _, player := range g.Players {
		if g.State == stateReconcile && player.IsGuesser {
			continue
		}
		if !player.Ready {
			return true
		}
	}
	if g.State == stateReconcile {
		g.State = stateGuess
		for i := 0; i < len(g.clues); {
			if g.clues[i].Dupe {
				g.clues = append(g.clues[:i], g.clues[i+1:]...)
			} else {
				i++
			}
		}
	} else {
		g.State = stateWrite
		g.Score = 0
		g.clues = nil
		g.guessMe = wordlist[rand.Intn(len(wordlist))]
		g.Players[rand.Intn(len(g.Players))].IsGuesser = true
	}

	for _, p := range g.Players {
		p.Ready = false
	}

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

	if Equalish(p.Clue, g.guessMe) {
		sendMsg(p.ws, "That's cheating")
		p.Clue = ""
		return false
	}

	g.clues = g.clues[0:0]
	for _, player := range g.Players {
		if player.IsGuesser {
			continue
		}
		player.Ready = false
		if player.Clue == "" {
			sendMsg(p.ws, "Waiting for other players")
			return true
		}
		g.clues = append(g.clues, Clue{player.Clue, false})
	}
	g.State = stateReconcile

	return true
}

func (g *JustOne) handleReconcile(cmd *wg.Command) bool {
	p, _ := Find(g.Players, cmd.PlayerId)

	if g.State != stateReconcile {
		return false
	}

	var clue Clue
	err := json.Unmarshal(cmd.Data, &clue)
	if err != nil {
		sendMsg(p.ws, err.Error())
		return false
	}

	for i := range g.clues {
		c := &g.clues[i]
		if c.Text == clue.Text {
			c.Dupe = clue.Dupe
		}
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

	var guess string
	err := json.Unmarshal(cmd.Data, &guess)
	if err != nil {
		sendMsg(p.ws, err.Error())
		return false
	}
	if Equalish(guess, g.guessMe) {
		g.Score++
		g.sendMsgAll(fmt.Sprintf("Guess '%v' is correct! 😀", guess))
	} else {
		g.sendMsgAll(fmt.Sprintf("Guess '%v' 'is incorrect! The word was %v 😢", guess, g.guessMe))
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
	g.clues = nil

	g.Round++
	if g.Round >= 13 {
		g.State = stateEnd
	} else {
		// pick another word from the list
		g.guessMe = wordlist[rand.Intn(len(wordlist))]
		g.State = stateWrite
	}

	return true
}

func (g *JustOne) String() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func Equalish(word, guess string) bool {
	word = strings.ToLower(strings.TrimSpace(word))
	guess = strings.ToLower(strings.TrimSpace(guess))

	if word == guess {
		return true
	}

	word = strings.TrimSuffix(word, "s")
	guess = strings.TrimSuffix(guess, "s")

	if word == guess {
		return true
	}

	word = strings.TrimSuffix(word, "e")
	guess = strings.TrimSuffix(guess, "e")

	if word == guess {
		return true
	}

	return false
}
