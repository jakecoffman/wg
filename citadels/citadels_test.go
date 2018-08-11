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

	var you, p1You, p2You secret

	var games int
	for games < 10 {
		if time.Now().Sub(start) > 30 * time.Second {
			t.Fatal("Stuck", citadels.State)
		}

		time.Sleep(10 * time.Millisecond)

	drain:
		for {
			select {
			case m := <-p1Conn.Msgs:
				switch m.(type) {
				case *UpdateMsg:
					u := m.(*UpdateMsg)
					p1You = *u.You
				}
			case m := <-p2Conn.Msgs:
				switch m.(type) {
				case *UpdateMsg:
					u := m.(*UpdateMsg)
					p2You = *u.You
				}
			default:
				break drain
			}
		}

		var player string
		var conn *wg.FakeConn

		if p1You.Turn {
			conn = p1Conn
			player = player1
			you = p1You
		} else if p2You.Turn {
			conn = p2Conn
			player = player2
			you = p2You
		} else {
			continue
		}

		switch citadels.State {
		case choose:
			b, _ := json.Marshal(rand.Intn(8))
			game.Cmd <- &wg.Command{player, conn, cmdChoose, game.Version, b}
		case goldOrDraw:
			p := citadels.Players[citadels.Turn.Value]
			var b json.RawMessage
			if len(p.hand) < 4 {
				b, _ = json.Marshal(1)
			} else {
				b, _ = json.Marshal(0)
			}
			game.Cmd <- &wg.Command{player, conn, cmdAction, game.Version, b}
		case putCardBack:
			length := len(citadels.Players[citadels.Turn.Value].hand)
			b, _ := json.Marshal([]int{length - (1+rand.Intn(2))})
			game.Cmd <- &wg.Command{player, conn, cmdAction, game.Version, b}
		case build:
			switch you.Character.Character {
			case King:
				fallthrough
			case Bishop:
				fallthrough
			case Merchant:
				fallthrough
			case Warlord:
				game.Cmd <- &wg.Command{player, conn, cmdTax, game.Version, nil}
			}
			p := citadels.Players[citadels.Turn.Value]
			for i := range p.hand {
				b, _ := json.Marshal([]int{i})
				game.Cmd <- &wg.Command{player, conn, cmdBuild, game.Version, b}
			}
			b, _ := json.Marshal([]int{})
			game.Cmd <- &wg.Command{player, conn, cmdBuild, game.Version, b}
		case endTurn:
			switch you.Character.Character {
			case Assassin:
				b, _ := json.Marshal(rand.Intn(7)+1)
				game.Cmd <- &wg.Command{player, conn, cmdSpecial, game.Version, b}
			case Thief:
				b, _ := json.Marshal(rand.Intn(6)+2)
				game.Cmd <- &wg.Command{player, conn, cmdSpecial, game.Version, b}
			case Magician:
			case Warlord:
				for i, p := range citadels.Players {
					for j, d := range p.Districts {
						if d.Value - 1 < p.Gold {
							b, _ := json.Marshal(warlordAction{Player: i, District: j})
							game.Cmd <- &wg.Command{player, conn, cmdSpecial, game.Version, b}
						}
					}
				}
			}
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
	}
}
