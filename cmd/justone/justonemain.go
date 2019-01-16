package main

import (
	"github.com/jakecoffman/wg"
	"github.com/jakecoffman/wg/justone"
	"golang.org/x/net/websocket"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/ws", websocket.Handler(wg.WsHandler(wg.ProcessPlayerCommands(justone.NewGame))))
	port := "8112"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
