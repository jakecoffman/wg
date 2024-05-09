package main

import (
	"github.com/jakecoffman/wg"
	"github.com/jakecoffman/wg/setlib"
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

	games := wg.NewGames[*setlib.Set]()

	http.Handle("/ws", wg.WsHandler(wg.ProcessPlayerCommands(games, setlib.NewGame)))
	port := "8222"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
