package citadels

import (
	"testing"
	"github.com/jakecoffman/wg"
			"log"
	"time"
	"math/rand"
	"encoding/json"
)

func TestCitadels(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	rand.Seed(time.Now().UnixNano())

	const gameId = "0"
	const player1 = "1"
	const player2 = "2"
	p1Conn := wg.NewFakeConn(player1)
	p2Conn := wg.NewFakeConn(player2)

	game := NewGame(gameId)
	citadels := game.Class.(*Citadels)

	game.Cmd <- &wg.Command{player1, p1Conn, cmdJoin, citadels.Version, nil}
	game.Cmd <- &wg.Command{player2, p2Conn, cmdJoin, citadels.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdStart, citadels.Version, nil}

	start := time.Now()

	time.Sleep(10 * time.Millisecond)

	var games int
	for games < 10 {
		if time.Now().Sub(start) > 30 * time.Second {
			t.Fatal("Stuck", citadels.State)
		}

	drain:
		for {
			select {
			case <-p1Conn.Msgs:
			case <-p2Conn.Msgs:
			default:
				break drain
			}
		}

		var player string
		var conn *wg.FakeConn

		if citadels.Players[citadels.Turn.Value].Uuid == player1 {
			conn = p1Conn
			player = player1
		} else {
			conn = p2Conn
			player = player2
		}

		switch citadels.State {
		case choose:
			b, _ := json.Marshal(rand.Intn(8))
			game.Cmd <- &wg.Command{player, conn, cmdChoose, game.Version, b}
		case goldOrDraw:
			log.Println("TURN:", player)
			p := citadels.Players[citadels.Turn.Value]
			log.Println("I have", len(p.Districts), "districts and", p.Gold, "gold")
			var b json.RawMessage
			if len(p.hand) < 4 {
				log.Println("I took districts")
				b, _ = json.Marshal(1)
			} else {
				log.Println("I took gold")
				b, _ = json.Marshal(0)
			}
			game.Cmd <- &wg.Command{player, conn, cmdAction, game.Version, b}
		case putCardBack:
			length := len(citadels.Players[citadels.Turn.Value].hand)
			b, _ := json.Marshal([]int{length - (1+rand.Intn(2))})
			game.Cmd <- &wg.Command{player, conn, cmdAction, game.Version, b}
		case build:
			p := citadels.Players[citadels.Turn.Value]
			for i := range p.hand {
				b, _ := json.Marshal([]int{i})
				game.Cmd <- &wg.Command{player, conn, cmdBuild, game.Version, b}
			}
			b, _ := json.Marshal([]int{})
			game.Cmd <- &wg.Command{player, conn, cmdBuild, game.Version, b}
		case endTurn:
			game.Cmd <- &wg.Command{player, conn, cmdEnd, game.Version, nil}
		case gameOver:
			games++
			game.Cmd <- &wg.Command{player1, p1Conn, cmdReady, game.Version, nil}
			game.Cmd <- &wg.Command{player2, p2Conn, cmdReady, game.Version, nil}
		case lobby:
			game.Cmd <- &wg.Command{player, conn, cmdStart, game.Version, nil}
		default:
			log.Fatal("ERROR:", citadels.State)
		}

		time.Sleep(10 * time.Millisecond)
	}
}
