package main

import (
	"log"
	"net/http"
	"github.com/jakecoffman/set-game/gamelib"
	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/resistance"
)

const (
	files = "./www/resistance"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(files))))
	http.Handle("/ws", websocket.Handler(gamelib.WsHandler(gamelib.ProcessPlayerCommands(resistance.NewGame))))
	http.HandleFunc("/admin", resistance.HandleAdmin)
	port := "8112"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:" + port, nil))
}
