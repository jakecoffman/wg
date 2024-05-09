package main

import (
	"github.com/jakecoffman/wg"
	"github.com/jakecoffman/wg/citadels"
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
	games := wg.NewGames[*citadels.Citadels]()
	http.Handle("/ws", wg.WsHandler(wg.ProcessPlayerCommands(games, citadels.NewGame)))
	port := "8113"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
