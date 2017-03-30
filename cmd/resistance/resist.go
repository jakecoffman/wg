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
	http.Handle("/ws", websocket.Handler(gamelib.WsHandler(resistance.ProcessPlayerCommands)))
	http.HandleFunc("/admin", resistance.HandleAdmin)
	host := "0.0.0.0:8112"
	log.Println("Serving http://localhost:8112")
	log.Fatal(http.ListenAndServe(host, nil))
}
