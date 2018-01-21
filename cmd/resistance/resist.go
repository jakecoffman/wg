package main

import (
	"github.com/jakecoffman/wg"
	"github.com/jakecoffman/wg/resistance"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/ws", websocket.Handler(wg.WsHandler(wg.ProcessPlayerCommands(resistance.NewGame))))
	port := "8112"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
