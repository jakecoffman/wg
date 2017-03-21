package setweb

import (
	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/setlib"
	"math/rand"
	"net/http"
	"encoding/json"
	"time"
	"fmt"
	"log"
)

// technically not thread-safe
var games = map[string]*setlib.Game{}

func init() {
	// check if games are abandoned, and if so remove them
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			for id, game := range games {
				if game.NumConns() == 0 {
					game.Stop <- struct{}{}
					delete(games, id)
				}
			}
		}
	}()
}

func WsHandler(ws *websocket.Conn) {
	defer ws.Close()

	userInput := struct{
		Type string `json:"type"`
		Play []int `json:"play"`
		Join string `json:"join"`
	}{}
	var game *setlib.Game
	defer func() {
		if game != nil {
			game.Leave <- ws
		}
	}()
	for {
		if err := websocket.JSON.Receive(ws, &userInput); err != nil {
			return
		}
		if userInput.Type == "join" {
			if game != nil {
				game.Leave <- ws
				game = nil
			}

			id := userInput.Join

			// new
			if id == "" {
				id = genId()
				games[id] = setlib.NewGame(id)
			}

			// game not found, start new
			if _, ok := games[id]; !ok {
				id = genId()
				games[id] = setlib.NewGame(id)
			}
			game = games[id]
			game.Join <- ws
		}
		if userInput.Type == "play" {
			if game != nil {
				game.Play <- &setlib.Move{Ws: ws, Locs: userInput.Play}
			}
		}
		if userInput.Type == "nosets" {
			if game != nil {
				game.NoSets <- ws
			}
		}
	}
}

type info struct {
	Game *setlib.Game
	Players []*setlib.Player
	Sets []string
}

func Admin(w http.ResponseWriter, r *http.Request) {
	response := []*info{}
	for _, game := range games {
		sets := game.FindSets()
		compactSets := []string{}
		for _, set := range sets {
			compactSets = append(compactSets, fmt.Sprint(set[0]+1, " ", set[1]+1, " ", set[2]+1))
		}
		n4 := &info{Game: game, Players: game.SlicePlayers(), Sets: compactSets}
		response = append(response, n4)
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
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
