package setlib

import (
	"math/rand"
	"log"
	"golang.org/x/net/websocket"
	"fmt"
	"sort"
	"time"
)

type Game struct {
	Join         chan *websocket.Conn `json:"-"`
	Leave        chan *websocket.Conn `json:"-"`
	Play         chan *Move `json:"-"`
	NoSets       chan *websocket.Conn `json:"-"`
	Stop         chan struct{} `json:"-"`

	Id           string
	board        map[int]Card
	rands        []int
	cursor       int

	players      map[*websocket.Conn]*Player
	playerCursor int

	Created      time.Time
	Updated      time.Time
}

type Player struct {
	Id    int
	Score int
}

type Move struct {
	Ws   *websocket.Conn
	Locs []int
}

func NewGame(id string) *Game {
	g := &Game{
		Join: make(chan *websocket.Conn),
		Leave: make(chan *websocket.Conn),
		Play: make(chan *Move),
		NoSets: make(chan *websocket.Conn),
		Stop: make(chan struct{}),

		players: map[*websocket.Conn]*Player{},
		playerCursor: 1,
		board: map[int]Card{},
		Id: id,
		Created: time.Now(),
	}
	go g.run()
	g.reset()
	return g
}

func (g *Game) run() {
	for {
		select {
		case c := <-g.Join:
			g.players[c] = &Player{Id: g.playerCursor}
			g.playerCursor += 1
			g.sendEverything(c)
			g.sendMetaToEveryone()
		case c := <-g.Leave:
			if _, ok := g.players[c]; ok {
				delete(g.players, c)
			}
			g.sendMetaToEveryone()
		case move := <-g.Play:
			g.playone(move)
		case c := <-g.NoSets:
			g.dealmore(c)
		case <-g.Stop:
			return
		}
		g.Updated = time.Now()
	}
}

func (g *Game) sendEverything(ws *websocket.Conn) {
	update := UpdateMsg{Type: "all", Updates: []Update{}, Players: g.SlicePlayers(), GameId: g.Id}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i < len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	if err := websocket.JSON.Send(ws, update); err != nil {
		log.Println(err)
		return
	}
}

func (g *Game) sendEveryoneEverything() {
	update := UpdateMsg{Type: "all", Updates: []Update{}, Players: g.SlicePlayers(), GameId: g.Id}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i < len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	for ws := range g.players {
		if err := websocket.JSON.Send(ws, update); err != nil {
			log.Println(err)
			return
		}
	}
}

func (g *Game) sendMetaToEveryone() {
	update := UpdateMsg{Type: "meta", Players: g.SlicePlayers(), GameId: g.Id}

	for ws := range g.players {
		if err := websocket.JSON.Send(ws, update); err != nil {
			log.Println(err)
			return
		}
	}
}

func (g *Game) reset() {
	g.rands = rand.Perm(len(deck))
	g.board = map[int]Card{}
	for g.cursor = 0; g.cursor < 12; g.cursor++ {
		g.board[g.cursor] = deck[g.rands[g.cursor]]
	}
	log.Println(g.Sets())
}

func (g *Game) dealmore(ws *websocket.Conn) {
	sets := g.FindSets()
	if len(sets) > 0 {
		g.players[ws].Score -= len(sets)
	} else {
		g.players[ws].Score += 1
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
	update := UpdateMsg{Type: "update", Players: g.SlicePlayers(), GameId: g.Id}
	update.Updates = []Update{
		{Location: len(g.board) - 3, Card: g.board[len(g.board) - 3]},
		{Location: len(g.board) - 2, Card: g.board[len(g.board) - 2]},
		{Location: len(g.board) - 1, Card: g.board[len(g.board) - 1]},
	}
	for ws := range g.players {
		if err := websocket.JSON.Send(ws, update); err != nil {
			log.Println(err)
		}
	}
	g.Sets()
}

func (g *Game) playone(move *Move) {
	update := UpdateMsg{Type: "update", Players: g.SlicePlayers(), GameId: g.Id}

	if isSet(g.board[move.Locs[0]], g.board[move.Locs[1]], g.board[move.Locs[2]]) {
		g.players[move.Ws].Score += 1
		if (g.cursor == len(g.rands)) {
			// out of cards
			g.board[move.Locs[0]] = Card{Amount: -1}
			g.board[move.Locs[1]] = Card{Amount: -1}
			g.board[move.Locs[2]] = Card{Amount: -1}
		} else if (len(g.board) > 12) {
			delete(g.board, move.Locs[0])
			delete(g.board, move.Locs[1])
			delete(g.board, move.Locs[2])
			newBoard := map[int]Card{}
			i := 0
			for _, card := range g.board {
				newBoard[i] = card
				i++
			}
			g.board = newBoard
			log.Println(g.Sets())
			g.sendEveryoneEverything()
			return
		} else {
			g.board[move.Locs[0]] = deck[g.rands[g.cursor + 0]]
			g.board[move.Locs[1]] = deck[g.rands[g.cursor + 1]]
			g.board[move.Locs[2]] = deck[g.rands[g.cursor + 2]]
			g.cursor += 3
		}
		log.Println(g.Sets())
		update.Updates = []Update{
			{Location: move.Locs[0], Card: g.board[move.Locs[0]]},
			{Location: move.Locs[1], Card: g.board[move.Locs[1]]},
			{Location: move.Locs[2], Card: g.board[move.Locs[2]]},
		}
		for ws := range g.players {
			if err := websocket.JSON.Send(ws, update); err != nil {
				log.Println(err)
			}
		}
	} else {
		log.Println("Not a set...")
		g.players[move.Ws].Score -= 1
		g.sendMetaToEveryone()
	}
}

func (g Game) Sets() string {
	sets := g.FindSets()
	str := fmt.Sprint(len(g.rands) - g.cursor, " left, ", len(sets), " sets:")
	for _, set := range sets {
		str += fmt.Sprint(set[0] + 1, "-", g.board[set[0]], set[1] + 1, "-", g.board[set[1]], set[2] + 1, "-", g.board[set[2]])
	}
	return str
}

func (g Game) FindSets() [][]int {
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

func (g *Game) SlicePlayers() []*Player {
	players := []*Player{}
	for _, p := range g.players {
		players = append(players, p)
	}
	sort.Slice(players, func (i, j int) bool {
		return players[i].Score >= players[j].Score
	})
	return players
}

func (g Game) NumConns() int {
	return len(g.players)
}

type UpdateMsg struct {
	Type    string   `json:"type"`
	Updates []Update `json:"updates"`
	GameId  string
	Players []*Player
}

type Update struct {
	Location int `json:"location"`
	Card     Card   `json:"card"`
}