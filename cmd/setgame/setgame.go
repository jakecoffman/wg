package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/setlib"
	"github.com/jakecoffman/set-game/gamelib"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.Handle("/ws", websocket.Handler(gamelib.WsHandler(gamelib.ProcessPlayerCommands(setlib.AllGames, setlib.NewGame))))
	port := "8222"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:" + port, nil))
}
