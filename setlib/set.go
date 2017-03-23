package setlib

import (
	"math/rand"
	"log"
	"fmt"
	"sort"
	"time"
	"github.com/jakecoffman/set-game/gamelib"
	"sync"
)

type Set struct {
	cmd          chan *SetCommand `json:"-"`

	Id           string
	board        map[int]Card
	rands        []int
	cursor       int

	players      map[string]*Player
	playerCursor int

	Version      int
	Created      time.Time
	Updated      time.Time
}

func NewGame(id string) *Set {
	g := &Set{
		cmd: make(chan *SetCommand),

		players: map[string]*Player{},
		playerCursor: 1,
		board: map[int]Card{},
		Id: id,
		Created: time.Now(),
	}
	go g.run()
	g.reset()
	return g
}

func (g *Set) Cmd(c gamelib.Command) {
	g.cmd <- c.(*SetCommand)
}

func (g *Set) run() {
	var cmd *SetCommand
	for {
		cmd = <-g.cmd
		if !cmd.IsValid() {
			log.Printf("Invalid command sent: %#v\n", cmd)
			continue
		}
		switch cmd.Type {
		case "Join":
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
			player.ip = player.ws.Request().Header.Get("X-Forwarded-For")
			g.sendEverythingTo(cmd.Ws)
			g.sendMetaToEveryone()
		case "Leave":
			delete(g.players, cmd.PlayerId)
			g.sendMetaToEveryone()
		case "Disconnect":
			g.players[cmd.PlayerId].ws = nil
			g.players[cmd.PlayerId].Connected = false
			g.sendMetaToEveryone()
		case "NoSets":
			if cmd.Version != g.Version {
				// prevent losing points due to race
				log.Println("Race condition averted")
				g.sendMetaToEveryone()
				continue
			}
			g.dealmore(cmd.PlayerId)
		case "Play":
			if cmd.Version != g.Version {
				// prevent losing points due to race
				log.Println("Race condition averted")
				g.sendMetaToEveryone()
				continue
			}
			g.playone(cmd)
		case "Stop":
			return
		}
		g.Updated = time.Now()
	}
}

func (g *Set) sendEverythingTo(ws gamelib.Connector) {
	if ws == nil {
		return
	}

	update := UpdateMsg{Type: "all", Updates: []Update{}, Players: g.SlicePlayers(), GameId: g.Id, Version: g.Version}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i < len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	if err := ws.Send(update); err != nil {
		log.Println(err)
		return
	}
}

func (g *Set) sendEveryoneEverything() {
	update := &UpdateMsg{Type: "all", Updates: []Update{}, Players: g.SlicePlayers(), GameId: g.Id, Version: g.Version}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i < len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	g.sendAll(update)
}

func (g *Set) sendMetaToEveryone() {
	g.sendAll(&UpdateMsg{Type: "meta", Players: g.SlicePlayers(), GameId: g.Id, Version: g.Version})
}

func (g *Set) sendAll(msg interface{}) {
	wg := sync.WaitGroup{}
	for _, player := range g.players {
		if player.ws != nil {
			wg.Add(1)
			go func (p *Player) {
				if err := p.ws.Send(msg); err != nil {
					log.Println(err)
				}
				wg.Done()
			}(player)
		}
	}
	wg.Wait()
}

func (g *Set) reset() {
	g.rands = rand.Perm(len(deck))
	g.board = map[int]Card{}
	for g.cursor = 0; g.cursor < 12; g.cursor++ {
		g.board[g.cursor] = deck[g.rands[g.cursor]]
	}
}

func (g *Set) dealmore(playerId string) {
	sets := g.FindSets()
	if len(sets) > 0 {
		g.players[playerId].Score -= len(sets)
	} else {
		g.players[playerId].Score += 1
	}

	if g.cursor == len(g.rands) {
		log.Println("Restarting game")
		g.reset()
		g.sendEveryoneEverything()
		return
	}

	g.board[len(g.board)] = deck[g.rands[g.cursor + 0]]
	g.board[len(g.board)] = deck[g.rands[g.cursor + 1]]
	g.board[len(g.board)] = deck[g.rands[g.cursor + 2]]
	g.cursor += 3
	g.Version += 1
	update := &UpdateMsg{
		Type: "update",
		Players: g.SlicePlayers(),
		GameId: g.Id,
		Version: g.Version,
		Updates: []Update{
			{Location: len(g.board) - 3, Card: g.board[len(g.board) - 3]},
			{Location: len(g.board) - 2, Card: g.board[len(g.board) - 2]},
			{Location: len(g.board) - 1, Card: g.board[len(g.board) - 1]},
		},
	}
	g.sendAll(update)
}

func (g *Set) playone(cmd *SetCommand) {
	if isSet(g.board[cmd.Locs[0]], g.board[cmd.Locs[1]], g.board[cmd.Locs[2]]) {
		g.players[cmd.PlayerId].Score += 1
		if (g.cursor == len(g.rands)) {
			// out of cards
			g.board[cmd.Locs[0]] = Card{Amount: -1}
			g.board[cmd.Locs[1]] = Card{Amount: -1}
			g.board[cmd.Locs[2]] = Card{Amount: -1}
		} else if (len(g.board) > 12) {
			delete(g.board, cmd.Locs[0])
			delete(g.board, cmd.Locs[1])
			delete(g.board, cmd.Locs[2])
			newBoard := map[int]Card{}
			i := 0
			for _, card := range g.board {
				newBoard[i] = card
				i++
			}
			g.board = newBoard
			g.sendEveryoneEverything()
			return
		} else {
			g.board[cmd.Locs[0]] = deck[g.rands[g.cursor + 0]]
			g.board[cmd.Locs[1]] = deck[g.rands[g.cursor + 1]]
			g.board[cmd.Locs[2]] = deck[g.rands[g.cursor + 2]]
			g.cursor += 3
		}
		g.Version += 1
		update := &UpdateMsg{
			Type: "update",
			Players: g.SlicePlayers(),
			GameId: g.Id,
			Version: g.Version,
			Updates: []Update{
				{Location: cmd.Locs[0], Card: g.board[cmd.Locs[0]]},
				{Location: cmd.Locs[1], Card: g.board[cmd.Locs[1]]},
				{Location: cmd.Locs[2], Card: g.board[cmd.Locs[2]]},
			}}
		g.sendAll(update)
	} else {
		log.Println("Not a set...")
		g.players[cmd.PlayerId].Score -= 1
		g.sendMetaToEveryone()
	}
}

func (g Set) Sets() string {
	sets := g.FindSets()
	str := fmt.Sprint(len(g.rands) - g.cursor, " left, ", len(sets), " sets:")
	for _, set := range sets {
		str += fmt.Sprint(set[0] + 1, "-", g.board[set[0]], set[1] + 1, "-", g.board[set[1]], set[2] + 1, "-", g.board[set[2]])
	}
	return str
}

func (g Set) FindSets() [][]int {
	sets := [][]int{}
	size := len(g.board)
	var card1, card2, card3 Card

	boardIndex := map[int]int{}
	index := 0
	for key := range g.board {
		boardIndex[index] = key
		index++
	}

	for i := 0; i < size - 2; i++ {
		card1 = g.board[boardIndex[i]]
		for j := i + 1; j < size - 1; j++ {
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
	players := []*Player{}
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
	players := []*playa{}
	for _, p := range g.players {
		players = append(players, &playa{Player: p, Addr: p.ip})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].Score >= players[j].Score
	})
	return players
}

func (g Set) NumConns() int {
	sum := 0
	for _, p := range g.players {
		if p.ws != nil {
			sum += 1
		}
	}
	return sum
}

type UpdateMsg struct {
	Type    string   `json:"type"`
	Updates []Update `json:"updates"`
	GameId  string
	Players []*Player
	Version int
}

type Update struct {
	Location int `json:"location"`
	Card     Card   `json:"card"`
}
