package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/setweb"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/set-game/", http.StripPrefix("/set-game", http.FileServer(http.Dir("./set-game"))))
	http.Handle("/set-game/ws", websocket.Handler(setweb.WsHandler))
	log.Fatal(http.ListenAndServe("0.0.0.0:8222", nil))
}
