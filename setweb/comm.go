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
	"github.com/google/uuid"
	"sort"
)

// technically not thread-safe
var games = map[string]*setlib.Game{}

func init() {
	// check if games are abandoned, and if so remove them
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			for id, game := range games {
				if time.Now().Sub(game.Updated).Hours() > 24 && game.NumConns() == 0 {
					game.Cmd <- &setlib.Command{Type: "Stop"}
					delete(games, id)
				}
			}
		}
	}()
}

type userInput struct {
	Type string `json:"type"`
	Play []int `json:"play"`
	Join string `json:"join"`
}

func WsHandler(ws *websocket.Conn) {
	defer ws.Close()

	var playerId string
	cookie, err := ws.Request().Cookie(COOKIE_NAME)
	if err == http.ErrNoCookie {
		// use has cookies turned off, not going to remember their game as long
		playerId = uuid.New().String()
	} else {
		playerId = cookie.Value
	}

	input := &userInput{}
	var game *setlib.Game

	defer func() {
		if game != nil {
			game.Cmd <- &setlib.Command{Type: "Disconnect", PlayerId: playerId}
		}
	}()

	for {
		if err := websocket.JSON.Receive(ws, input); err != nil {
			return
		}
		switch input.Type {
		case "join":
			if game != nil {
				game.Cmd <- &setlib.Command{Type: "Leave", PlayerId: playerId}
				game = nil
			}

			id := input.Join

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
			game.Cmd <- &setlib.Command{Type: "Join", Ws: ws, PlayerId: playerId}
		case "play":
			if game != nil {
				game.Cmd <- &setlib.Command{Type: "Play", Locs: input.Play, PlayerId: playerId}
			}
		case "nosets":
			if game != nil {
				game.Cmd <- &setlib.Command{Type: "NoSets", PlayerId: playerId}
			}
		}
	}
}

type info struct {
	Game    *setlib.Game
	Players []*setlib.Player
	Sets    []string
}

func Admin(w http.ResponseWriter, r *http.Request) {
	response := []*info{}
	for _, game := range games {
		sets := game.FindSets()
		compactSets := []string{}
		for _, set := range sets {
			compactSets = append(compactSets, fmt.Sprint(set[0] + 1, " ", set[1] + 1, " ", set[2] + 1))
		}
		n4 := &info{Game: game, Players: game.SlicePlayers(), Sets: compactSets}
		response = append(response, n4)
	}
	sort.Slice(&response, func(i, j int) bool {
		return response[i].Game.Updated.Before(response[j].Game.Updated)
	})
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

const COOKIE_NAME = "PLAYER_COOKIE"

// CookieMiddleware just makes sure every user gets a cookie
func CookieMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(COOKIE_NAME)
		if err == http.ErrNoCookie {
			cookie = &http.Cookie{Name: COOKIE_NAME, Value: uuid.New().String()}
			http.SetCookie(w, cookie)
		}
		handler.ServeHTTP(w, r)
	}
}
