package setlib

import (
	"math/rand"
	"log"
	"golang.org/x/net/websocket"
	"fmt"
)

type Game struct {
	conns map[*websocket.Conn]bool

	Join  chan *websocket.Conn
	Leave chan *websocket.Conn
	Play  chan []int
	NoSets chan struct{}
	Stop chan struct{}

	Id string
	board map[int]Card
	rands []int
	cursor int
}

func NewGame(id string) *Game {
	g := &Game{
		Join: make(chan *websocket.Conn),
		Leave: make(chan *websocket.Conn),
		Play: make(chan []int),
		NoSets: make(chan struct{}),
		Stop: make(chan struct{}),

		conns: make(map[*websocket.Conn]bool),
		board: map[int]Card{},
		Id: id,
	}
	go g.run()
	g.reset()
	return g
}

func (g *Game) run() {
	for {
		select {
		case c := <-g.Join:
			log.Println("Player joined")
			g.conns[c] = true
			g.sendEverything(c)
			g.sendMetaToEveryone()
		case c := <-g.Leave:
			if _, ok := g.conns[c]; ok {
				delete(g.conns, c)
			}
			g.sendMetaToEveryone()
		case read := <- g.Play:
			g.playone(read)
		case <- g.NoSets:
			g.dealmore()
		case <- g.Stop:
			return
		}
	}
}

func (g *Game) sendEverything(ws *websocket.Conn) {
	update := UpdateMsg{Type: "all", Updates: []Update{}, Players: len(g.conns), GameId: g.Id}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i< len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	if err := websocket.JSON.Send(ws, update); err != nil {
		log.Println(err)
		return
	}
}

func (g *Game) sendEveryoneEverything() {
	update := UpdateMsg{Type: "all", Updates: []Update{}, Players: len(g.conns), GameId: g.Id}

	// order is important when sending all because javascript is rebuilding DOM
	for i := 0; i< len(g.board); i++ {
		update.Updates = append(update.Updates, Update{Location: i, Card: g.board[i]})
	}

	for ws := range g.conns {
		if err := websocket.JSON.Send(ws, update); err != nil {
			log.Println(err)
			return
		}
	}
}

func (g *Game) sendMetaToEveryone() {
	update := UpdateMsg{Type: "meta", Players: len(g.conns), GameId: g.Id}

	for ws := range g.conns {
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

func (g *Game) dealmore() {
	//if len(findSets(g.board)) != 0 {
	//	return
	//}

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
	update := UpdateMsg{Type: "update", Players: len(g.conns), GameId: g.Id}
	update.Updates = []Update{
		{Location: len(g.board)-3, Card: g.board[len(g.board)-3]},
		{Location: len(g.board)-2, Card: g.board[len(g.board)-2]},
		{Location: len(g.board)-1, Card: g.board[len(g.board)-1]},
	}
	for ws := range g.conns {
		if err := websocket.JSON.Send(ws, update); err != nil {
			log.Println(err)
		}
	}
	g.Sets()
}

func (g *Game) playone(read []int) {
	update := UpdateMsg{Type: "update", Players: len(g.conns), GameId: g.Id}

	if isSet(g.board[read[0]], g.board[read[1]], g.board[read[2]]) {
		if (g.cursor == len(g.rands)) {
			// out of cards
			g.board[read[0]] = Card{Amount: -1}
			g.board[read[1]] = Card{Amount: -1}
			g.board[read[2]] = Card{Amount: -1}
		} else if (len(g.board) > 12) {
			delete(g.board, read[0])
			delete(g.board, read[1])
			delete(g.board, read[2])
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
			g.board[read[0]] = deck[g.rands[g.cursor + 0]]
			g.board[read[1]] = deck[g.rands[g.cursor + 1]]
			g.board[read[2]] = deck[g.rands[g.cursor + 2]]
			g.cursor += 3
		}
		log.Println(g.Sets())
		update.Updates = []Update{
			{Location: read[0], Card: g.board[read[0]]},
			{Location: read[1], Card: g.board[read[1]]},
			{Location: read[2], Card: g.board[read[2]]},
		}
		for ws := range g.conns {
			if err := websocket.JSON.Send(ws, update); err != nil {
				log.Println(err)
			}
		}
	} else {
		log.Println("Not a set...")
	}
}

func (g Game) Sets() string {
	sets := g.FindSets()
	str := fmt.Sprint(len(g.rands) - g.cursor, " left, ", len(sets), " sets:")
	for _, set := range sets {
		str += fmt.Sprint(set[0]+1, "-", g.board[set[0]], set[1]+1, "-",  g.board[set[1]], set[2]+1, "-", g.board[set[2]])
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

func (g Game) NumConns() int {
	return len(g.conns)
}

type UpdateMsg struct {
	Type    string   `json:"type"`
	Updates []Update `json:"updates"`
	GameId  string
	Players int
}

type Update struct {
	Location int `json:"location"`
	Card     Card   `json:"card"`
}