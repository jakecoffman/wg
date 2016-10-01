package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/setgame/", http.StripPrefix("/setgame", http.FileServer(http.Dir("./setgame"))))
	http.Handle("/setgame/ws", websocket.Handler(wsHandler))
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
}

type Update struct {
	Location string `json:"location"`
	Card     Card   `json:"card"`
}

func wsHandler(ws *websocket.Conn) {
	defer ws.Close()

	board := map[string]Card{}
	k := 0
	// rands is a random walk across the deck
	rands := rand.Perm(len(deck))
	for x := 0; x < 3; x++ {
		for y := 0; y < 4; y++ {
			board[fmt.Sprint(x, ",", y)] = deck[rands[k]]
			k++
		}
	}

	log.Println("Current sets:", findSets(board))

	// pre-declare variables for maximum performance as it it mattered
	read := []string{}
	var err error
	update := UpdateMsg{Type: "update", Updates: []Update{}}

	for l, u := range board {
		update.Updates = append(update.Updates, Update{Location: l, Card: u})
	}

	if err = json.NewEncoder(ws).Encode(update); err != nil {
		log.Println(err)
		return
	}

	for {
		if err = json.NewDecoder(ws).Decode(&read); err != nil {
			if err == io.EOF {
				return
			}
			log.Println(err)
			return
		}
		log.Println(read)
		if isSet(board[read[0]], board[read[1]], board[read[2]]) {
			// TODO: award points
			board[read[0]] = deck[rands[k + 0]]
			board[read[1]] = deck[rands[k + 1]]
			board[read[2]] = deck[rands[k + 2]]
			k += 3
			log.Println("Current sets:", findSets(board))
			update.Updates = []Update{
				{Location: read[0], Card: board[read[0]]},
				{Location: read[1], Card: board[read[1]]},
				{Location: read[2], Card: board[read[2]]},
			}
			if err = json.NewEncoder(ws).Encode(update); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println("Not a set...")
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
	return s1 == s2 && s2 == s3
}

func different(s1, s2, s3 string) bool {
	return s1 != s2 && s2 != s3 && s3 != s1
}

func sameInt(s1, s2, s3 int) bool {
	return s1 == s2 && s2 == s3
}

func differentInt(s1, s2, s3 int) bool {
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