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
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(files))))
	http.Handle("/ws", websocket.Handler(gamelib.WsHandler(example.ProcessPlayerCommands)))
	http.HandleFunc("/admin", example.HandleAdmin)
	host := "0.0.0.0:8111"
	log.Println("Serving http://localhost:8111")
	log.Fatal(http.ListenAndServe(host, nil))
}
