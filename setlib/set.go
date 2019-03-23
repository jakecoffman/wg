package setlib

import (
	"encoding/json"
	"github.com/jakecoffman/wg"
	"log"
	"math/rand"
	"sort"
	"time"
)

const DEV = false

type Set struct {
	*wg.Game

	board  []Card
	rands  []int
	cursor int

	players      map[string]*Player
	playerCursor int
}

type Player struct {
	ws        wg.Connector
	Id        int
	Score     int
	Connected bool `json:",omitempty"`
	ip        string
	Ready     bool `json:",omitempty"`
	Name      string `json:",omitempty"`
}

func NewGame(id string) *wg.Game {
	g := &Set{
		players:      map[string]*Player{},
		playerCursor: 1,
		board:        []Card{},
	}
	g.Game = wg.NewGame(g, id)
	go g.run()
	g.reset()
	return g.Game
}

const (
	cmdJoin       = "join"
	cmdLeave      = "leave"
	cmdDisconnect = "disconnect"
	cmdReady      = "ready"
	cmdRename     = "rename"
	cmdPlay       = "play"
	cmdNoSets     = "nosets"
	cmdStop       = "stop"
)

func (g *Set) run() {
	var cmd *wg.Command
	for {
		cmd = <-g.Cmd
		switch cmd.Type {
		case cmdJoin:
			g.join(cmd)
		case cmdLeave:
			g.leave(cmd)
		case cmdDisconnect:
			g.disconnect(cmd)
		case cmdReady:
			g.ready(cmd)
		case cmdRename:
			g.rename(cmd)
		case cmdNoSets:
			g.noSets(cmd)
		case cmdPlay:
			g.play(cmd)
		case cmdStop:
			log.Println("Stopping set game", g.Id)
			return
		}
		g.Updated = time.Now()
		if DEV {
			g.sendEveryoneCheats()
		}
	}
}

func (g *Set) disconnect(cmd *wg.Command) {
	p := g.players[cmd.PlayerId]
	if p != nil {
		p.ws = nil
		p.Connected = false
		g.sendMetaToEveryone()
	}
}

func (g *Set) ready(cmd *wg.Command) {
	p := g.players[cmd.PlayerId]

	var ready bool
	err := json.Unmarshal(cmd.Data, &ready)
	if err != nil {
		log.Println(err)
		return
	}

	if p != nil {
		p.Ready = ready
		g.sendMetaToEveryone()
	}
}

func (g *Set) rename(cmd *wg.Command) {
	p := g.players[cmd.PlayerId]
	if p == nil {
		return
	}

	var name string
	err := json.Unmarshal(cmd.Data, &name)
	if err != nil {
		log.Println(err)
		return
	}

	if len(name) > 8 {
		p.Name = name[0:8]
	} else {
		p.Name = name
	}
	g.sendMetaToEveryone()
}

func (g *Set) leave(cmd *wg.Command) {
	delete(g.players, cmd.PlayerId)
	g.sendMetaToEveryone()
}

func (g *Set) join(cmd *wg.Command) {
	var player *Player
	var ok bool
	if player, ok = g.players[cmd.PlayerId]; !ok {
		// player was not here before, create
		player = &Player{Id: g.playerCursor}
		g.players[cmd.PlayerId] = player
		g.playerCursor += 1
	}
	player.ws = cmd.Ws
	player.Connected = true
	player.ip = player.ws.Ip()
	g.sendEverythingTo(cmd.Ws)
	g.sendMetaToEveryone()
}

func (g *Set) sendEverythingTo(ws wg.Connector) {
	if ws == nil {
		return
	}

	update := UpdateMsg{
		Type:    "all",
		Updates: []Update{},
		Players: g.SlicePlayers(),
		Version: g.Version,
	}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i < len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	ws.Send(update)
}

func (g *Set) sendEveryoneEverything() {
	update := UpdateMsg{
		Type:    "all",
		Updates: []Update{},
		Players: g.SlicePlayers(),
		Version: g.Version,
	}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i < len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	g.sendAll(update)
}

func (g *Set) sendEveryoneCheats() {
	sets := g.FindSets()

	msg := map[string]interface{}{"Type": "cheat"}
	if len(sets) > 0 {
		msg["Sets"] = sets[0]
	} else {
		return
	}

	g.sendAll(msg)
}

func (g *Set) sendMetaToEveryone() {
	playing := true
	for _, p := range g.players {
		if p.Ready != true {
			playing = false
			break
		}
	}
	msg := MetaMsg{
		Type:    "meta",
		Players: g.SlicePlayers(),
		GameId:  g.Id,
		Playing: playing,
		Version: g.Version,
	}
	for _, player := range g.players {
		if player.ws != nil {
			msg.You = player.Id
			player.ws.Send(msg)
		}
	}
}

func (g *Set) sendAll(msg interface{}) {
	for _, player := range g.players {
		if player.ws != nil {
			player.ws.Send(msg)
		}
	}
}

func (g *Set) reset() {
	g.rands = rand.Perm(len(deck))
	g.board = []Card{}
	for g.cursor = 0; g.cursor < 12; g.cursor++ {
		g.board = append(g.board, deck[g.rands[g.cursor]])
	}
}

func (g *Set) noSets(cmd *wg.Command) {
	if cmd.Version != g.Version {
		// prevent losing points due to race
		log.Println("Race condition averted")
		g.sendMetaToEveryone()
		return
	}
	playerId := cmd.PlayerId
	sets := g.FindSets()
	if len(sets) > 0 {
		g.players[playerId].Score -= len(sets)
		g.sendAll(&PlayMsg{
			Type:   "play",
			Player: g.players[cmd.PlayerId].Id,
			Words:  "missed some",
			Score:  -len(sets),
		})
	} else {
		g.players[playerId].Score += 1
		g.sendAll(&PlayMsg{
			Type:   "play",
			Player: g.players[cmd.PlayerId].Id,
			Words: "no sets",
			Score:  1,
		})
	}

	if g.cursor == len(g.rands) {
		log.Println("Restarting game")
		g.reset()
		g.sendEveryoneEverything()
		return
	}

	g.board = append(g.board, deck[g.rands[g.cursor+0]])
	g.board = append(g.board, deck[g.rands[g.cursor+1]])
	g.board = append(g.board, deck[g.rands[g.cursor+2]])
	g.cursor += 3
	g.Version += 1
	update := UpdateMsg{
		Type:    "update",
		Players: g.SlicePlayers(),
		Version: g.Version,
		Updates: []Update{
			{Location: len(g.board) - 3, Card: g.board[len(g.board)-3]},
			{Location: len(g.board) - 2, Card: g.board[len(g.board)-2]},
			{Location: len(g.board) - 1, Card: g.board[len(g.board)-1]},
		},
	}
	g.sendAll(update)
}

func (g *Set) play(cmd *wg.Command) {
	if cmd.Version != g.Version {
		// prevent losing points due to race
		log.Println("Race condition averted")
		g.sendMetaToEveryone()
		return
	}
	var play []int
	err := json.Unmarshal(cmd.Data, &play)
	if err != nil {
		log.Println("error reading play data", err)
		return
	}
	if !isSet(g.board[play[0]], g.board[play[1]], g.board[play[2]]) {
		log.Println("Not a set...")
		g.players[cmd.PlayerId].Score -= 1
		g.sendMetaToEveryone()
		g.sendAll(&PlayMsg{
			Type:   "play",
			Player: g.players[cmd.PlayerId].Id,
			Cards:  []Card{g.board[play[0]], g.board[play[1]], g.board[play[2]]},
			Words: "not a set",
			Score:  -1,
		})
		return
	}
	// it's a set
	g.players[cmd.PlayerId].Score += 1
	g.Version += 1

	g.sendAll(&PlayMsg{
		Type:   "play",
		Player: g.players[cmd.PlayerId].Id,
		Cards:  []Card{g.board[play[0]], g.board[play[1]], g.board[play[2]]},
		Score:  1,
	})

	// just remove, don't deal
	if g.cursor == len(g.rands) || len(g.board) > 12 {
		sort.Ints(play)
		g.board = append(g.board[:play[2]], g.board[play[2]+1:]...)
		g.board = append(g.board[:play[1]], g.board[play[1]+1:]...)
		g.board = append(g.board[:play[0]], g.board[play[0]+1:]...)
		g.sendEveryoneEverything()
		return
	}

	// normal: replace cards with cards from the deck
	g.board[play[0]] = deck[g.rands[g.cursor]]
	g.board[play[1]] = deck[g.rands[g.cursor+1]]
	g.board[play[2]] = deck[g.rands[g.cursor+2]]
	g.cursor += 3
	update := &UpdateMsg{
		Type:    "update",
		Players: g.SlicePlayers(),
		Version: g.Version,
		Updates: []Update{
			{Location: play[0], Card: g.board[play[0]]},
			{Location: play[1], Card: g.board[play[1]]},
			{Location: play[2], Card: g.board[play[2]]},
		}}
	g.sendAll(update)
}

func (g Set) FindSets() [][]int {
	var sets [][]int
	size := len(g.board)
	var card1, card2, card3 Card

	boardIndex := map[int]int{}
	index := 0
	for key := range g.board {
		boardIndex[index] = key
		index++
	}

	for i := 0; i < size-2; i++ {
		card1 = g.board[boardIndex[i]]
		for j := i + 1; j < size-1; j++ {
			card2 = g.board[boardIndex[j]]
			for k := j + 1; k < size; k++ {
				card3 = g.board[boardIndex[k]]
				if isSet(card1, card2, card3) {
					sets = append(sets, []int{boardIndex[i], boardIndex[j], boardIndex[k]})
				}
			}
		}
	}

	return sets
}

func (g *Set) SlicePlayers() []*Player {
	var players []*Player
	for _, p := range g.players {
		players = append(players, p)
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].Score >= players[j].Score
	})
	return players
}

// SlicePlayersAdmin is like SlicePlayers but it exposes sensitive info, so careful!
func (g *Set) SlicePlayersAdmin() interface{} {
	type playa struct {
		*Player
		Addr string
	}
	var players []*playa
	for _, p := range g.players {
		players = append(players, &playa{Player: p, Addr: p.ip})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].Score >= players[j].Score
	})
	return players
}

type UpdateMsg struct {
	Type    string
	Updates []Update
	Players []*Player
	Version int
}

type Update struct {
	Location int
	Card     Card
}

type MetaMsg struct {
	Type    string
	GameId  string
	Players []*Player
	Version int
	You     int
	Playing bool
}

type PlayMsg struct {
	Type   string
	Player int
	Cards  []Card
	Score  int
	Words  string
}
