package main

import (
	"log"
	"net/http"
	"github.com/jakecoffman/set-game/gamelib"
	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/resistance"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/ws", websocket.Handler(gamelib.WsHandler(gamelib.ProcessPlayerCommands(resistance.AllGames, resistance.NewGame))))
	port := "8112"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:" + port, nil))
}
