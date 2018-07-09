package setlib

import (
	"testing"
	"github.com/jakecoffman/wg"
	"log"
	"math/rand"
	"time"
)

func TestSet(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	const gameId = "0"
	const player1 = "1"
	p1Conn := wg.NewFakeConn(player1)

	game := NewGame(gameId)
	set := game.Class.(*Set)

	game.Cmd <- &wg.Command{player1, p1Conn, cmdJoin, set.Version, nil}

	for i := 0; i < 1000; i++ {
	drain:
		for {
			select {
			case <-p1Conn.Msgs:
			default:
				break drain
			}
		}

		// let the resistance goroutine go, probably should improve this with locking
		time.Sleep(1 * time.Millisecond)

		// Try doing every board possibility, then press deal more?
	}
}
