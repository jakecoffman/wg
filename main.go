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
	http.Handle("/", gamelib.CookieMiddleware(http.StripPrefix("/", http.FileServer(http.Dir("./set-game")))))
	http.Handle("/ws", websocket.Handler(gamelib.WsHandler(setlib.ProcessPlayerCommands)))
	http.HandleFunc("/admin", setlib.HandleAdmin)
	host := "0.0.0.0:8222"
	log.Println("Serving on", host)
	log.Fatal(http.ListenAndServe(host, nil))
}
