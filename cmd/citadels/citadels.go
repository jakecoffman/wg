package main

import (
	"github.com/jakecoffman/wg"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
	"math/rand"
	"github.com/jakecoffman/wg/citadels"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/ws", websocket.Handler(wg.WsHandler(wg.ProcessPlayerCommands(citadels.NewGame))))
	port := "8113"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
