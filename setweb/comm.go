package setweb

import (
	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/setlib"
	"math/rand"
)

// technically not thread-safe
var games = map[string]*setlib.Game{}

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
				game.Play <- userInput.Play
			}
		}
		if userInput.Type == "nosets" {
			if game != nil {
				game.NoSets <- struct{}{}
			}
		}
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