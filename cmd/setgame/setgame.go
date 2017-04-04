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
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./www/set"))))
	http.Handle("/api/ws", websocket.Handler(gamelib.WsHandler(setlib.ProcessPlayerCommands)))
	http.HandleFunc("/api/admin", setlib.HandleAdmin)
	port := "8222"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:" + port, nil))
}
