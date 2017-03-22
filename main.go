package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/setweb"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/", setweb.CookieMiddleware(http.StripPrefix("/", http.FileServer(http.Dir("./set-game")))))
	http.Handle("/ws", websocket.Handler(setweb.WsHandler))
	http.HandleFunc("/admin", setweb.Admin)
	log.Fatal(http.ListenAndServe("0.0.0.0:8222", nil))
}
