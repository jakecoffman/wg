package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/set-game/", http.StripPrefix("/set-game", http.FileServer(http.Dir("./set-game"))))
	http.Handle("/set-game/ws", websocket.Handler(wsHandler))
	log.Fatal(http.ListenAndServe("0.0.0.0:8222", nil))
}

type Card struct {
	Shape   string `json:"shape"`
	Pattern string `json:"pattern"`
	Color   string `json:"color"`
	Amount  int    `json:"amount"`
}

var deck []Card

func init() {
	rand.Seed(time.Now().Unix())

	shapes := []string{"pill", "nut", "diamond"}
	patterns := []string{"hollow", "striped", "solid"}
	colors := []string{"red", "purple", "green"}
	amount := []int{1, 2, 3}

	deck = []Card{}
	for _, s := range shapes {
		for _, p := range patterns {
			for _, c := range colors {
				for _, a := range amount {
					deck = append(deck, Card{Shape: s, Pattern: p, Color: c, Amount: a})
				}
			}
		}
	}
}

type UpdateMsg struct {
	Type    string   `json:"type"`
	Updates []Update `json:"updates"`
	GameId  string
	Players int
}

type Update struct {
	Location string `json:"location"`
	Card     Card   `json:"card"`
}

type JoinOrNew struct {
	Join string
}

func wsHandler(ws *websocket.Conn) {
	defer ws.Close()

	// is this a new game or a join?
	var joinOrNew JoinOrNew
	if err := websocket.JSON.Receive(ws, &joinOrNew); err != nil {
		log.Println(err)
		return
	}

	id := joinOrNew.Join

	if id == "" {
		id = genId()
		games[id] = NewGame(id)
	}

	// game not found
	if _, ok := games[id]; !ok {
		id = genId()
		games[id] = NewGame(id)
	}

	games[id].join <- ws
	defer func() {
		games[id].leave <- ws
	}()

	userInput := struct{
		Type string `json:"type"`
		Play []string `json:"play"`
	}{}
	for {
		if err := websocket.JSON.Receive(ws, &userInput); err != nil {
			return
		}
		if userInput.Type == "play" {
			games[id].play <- userInput.Play
		}
		if userInput.Type == "nosets" {
			games[id].nosets <- struct{}{}
		}
	}
}

func isSet(card1, card2, card3 Card) bool {
	if !(same(card1.Shape, card2.Shape, card3.Shape) || different(card1.Shape, card2.Shape, card3.Shape)) {
		return false
	}
	if !(same(card1.Pattern, card2.Pattern, card3.Pattern) || different(card1.Pattern, card2.Pattern, card3.Pattern)) {
		return false
	}
	if !(sameInt(card1.Amount, card2.Amount, card3.Amount) || differentInt(card1.Amount, card2.Amount, card3.Amount)) {
		return false
	}
	if !(same(card1.Color, card2.Color, card3.Color) || different(card1.Color, card2.Color, card3.Color)) {
		return false
	}
	return true
}

func same(s1, s2, s3 string) bool {
	if s3 == "" {
		return false
	}
	return s1 == s2 && s2 == s3 && s3 != ""
}

func different(s1, s2, s3 string) bool {
	if s3 == "" {
		return false
	}
	return s1 != s2 && s2 != s3 && s3 != s1
}

func sameInt(s1, s2, s3 int) bool {
	if s3 == -1 {
		return false
	}
	return s1 == s2 && s2 == s3
}

func differentInt(s1, s2, s3 int) bool {
	if s3 == -1 {
		return false
	}
	return s1 != s2 && s2 != s3 && s3 != s1
}

func findSets(board map[string]Card) [][]string {
	sets := [][]string{}
	size := len(board)
	var card1, card2, card3 Card

	boardIndex := map[int]string{}
	index := 0
	for key := range board {
		boardIndex[index] = key
		index++
	}

	for i := 0; i < size - 2; i++ {
		card1 = board[boardIndex[i]]
		for j := i + 1; j < size - 1; j++ {
			card2 = board[boardIndex[j]]
			for k := j + 1; k < size; k++ {
				card3 = board[boardIndex[k]]
				if isSet(card1, card2, card3) {
					sets = append(sets, []string{boardIndex[i], boardIndex[j], boardIndex[k]})
				}
			}
		}
	}

	return sets
}

// technically not thread-safe
var games = map[string]*Game{}

type Game struct {
	conns map[*websocket.Conn]bool

	join  chan *websocket.Conn
	leave chan *websocket.Conn
	play  chan []string
	nosets chan struct{}

	Id string
	board map[string]Card
	rands []int
	cursor int
}

func NewGame(id string) *Game {
	g := &Game{
		join: make(chan *websocket.Conn),
		leave: make(chan *websocket.Conn),
		play: make(chan []string),
		nosets: make(chan struct{}),

		conns: make(map[*websocket.Conn]bool),
		board: map[string]Card{},
		Id: id,
	}
	go g.run()
	g.reset()
	return g
}

func (g *Game) run() {
	for {
		select {
		case c := <-g.join:
			log.Println("Player joined")
			g.conns[c] = true
			g.sendEveryoneEverything()
		case c := <-g.leave:
			if _, ok := g.conns[c]; ok {
				delete(g.conns, c)
			}
			if len(g.conns) == 0 {
				log.Println("Game", g.Id, "abandoned")
				return
			}
		case read := <- g.play:
			g.playone(read)
		case <- g.nosets:
			g.dealmore()
		}
	}
}

func (g *Game) sendEveryoneEverything() {
	update := UpdateMsg{Type: "all", Updates: []Update{}, Players: len(g.conns), GameId: g.Id}

	for l, u := range g.board {
		update.Updates = append(update.Updates, Update{Location: l, Card: u})
	}

	for ws := range g.conns {
		if err := websocket.JSON.Send(ws, update); err != nil {
			log.Println(err)
			return
		}
	}
}

func (g *Game) reset() {
	g.cursor = 0
	g.rands = rand.Perm(len(deck))
	for x := 0; x < 3; x++ {
		for y := 0; y < 4; y++ {
			g.board[fmt.Sprint(x, ",", y)] = deck[g.rands[g.cursor]]
			g.cursor++
		}
	}
	log.Println("Current sets:", findSets(g.board))
}

func (g *Game) dealmore() {
	//if len(findSets(g.board)) != 0 {
	//	return
	//}

	if g.cursor == len(g.rands) {
		log.Println("Restarting game")
		g.reset()
		return
	}

	g.board["0,4"] = deck[g.rands[g.cursor + 0]]
	g.board["1,4"] = deck[g.rands[g.cursor + 1]]
	g.board["2,4"] = deck[g.rands[g.cursor + 2]]
	g.cursor += 3
	update := UpdateMsg{Type: "update", Players: len(g.conns), GameId: g.Id}
	update.Updates = []Update{
		{Location: "0,4", Card: g.board["0,4"]},
		{Location: "1,4", Card: g.board["1,4"]},
		{Location: "2,4", Card: g.board["2,4"]},
	}
	for ws := range g.conns {
		if err := websocket.JSON.Send(ws, update); err != nil {
			log.Println(err)
		}
	}
	log.Println(len(g.rands) - g.cursor, "left, sets:", findSets(g.board))
}

func (g *Game) playone(read []string) {
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
			values := []string{}
			for x := 0; x < 3; x++ {
				for y := 0; y < 4; y++ {
					values = append(values, fmt.Sprint(x, ",", y))
				}
			}
			newBoard := map[string]Card{}
			i := 0
			for _, card := range g.board {
				newBoard[values[i]] = card
				i++
			}
			g.board = newBoard
			log.Println(len(g.rands) - g.cursor, "left, sets:", findSets(g.board))
			g.sendEveryoneEverything()
			return
		} else {
			g.board[read[0]] = deck[g.rands[g.cursor + 0]]
			g.board[read[1]] = deck[g.rands[g.cursor + 1]]
			g.board[read[2]] = deck[g.rands[g.cursor + 2]]
			g.cursor += 3
		}
		log.Println(len(g.rands) - g.cursor, "left, sets:", findSets(g.board))
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

const letterBytes = "1234567890"

func genId() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
