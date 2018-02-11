package main

import (
	"log"
	"net/http"
	"github.com/jakecoffman/wg"
	"github.com/jakecoffman/wg/setlib"
	"golang.org/x/net/websocket"
	"time"
	"math/rand"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.Handle("/ws", websocket.Handler(wg.WsHandler(wg.ProcessPlayerCommands(setlib.NewGame))))
	port := "8222"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
