package main

import (
	"log"
	"net/http"
	"github.com/jakecoffman/set-game/gamelib"
	"golang.org/x/net/websocket"
	"github.com/jakecoffman/set-game/example"
)

const (
	files = "./www/example"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// this route is a dev convenience, should be served by a reverse proxy in prod
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(files))))
	http.Handle("/ws", websocket.Handler(gamelib.WsHandler(example.ProcessPlayerCommands)))
	http.HandleFunc("/admin", example.HandleAdmin)
	port := "8111"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:" + port, nil))
}
