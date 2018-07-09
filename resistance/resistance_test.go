package resistance

import (
	"testing"
	"github.com/jakecoffman/wg"
	"github.com/gin-gonic/gin/json"
	"math/rand"
	"log"
	"time"
	"fmt"
)

func TestResistance(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	const gameId = "0"
	const player1 = "1"
	p1Conn := wg.NewFakeConn(player1)

	game := NewGame(gameId)
	resistance := game.Class.(*Resist)

	game.Cmd <- &wg.Command{player1, p1Conn, cmdJoin, resistance.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, resistance.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, resistance.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, resistance.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, resistance.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdStart, resistance.Version, nil}

	false := []byte("false")
	true := []byte("true")

	var wins int
	for wins < 1000 {
	drain:
		for {
			select {
			case <-p1Conn.Msgs:
			default:
				break drain
			}
		}

		// let the resistance goroutine go, probably should improve this with locking
		time.Sleep(1*time.Millisecond)

		switch resistance.State {
		case stateTeambuilding:
			assignment := rand.Perm(5)[:resistance.Missions[resistance.CurrentMission].Slots]
			b, _ := json.Marshal(assignment)
			game.Cmd <- &wg.Command{player1, p1Conn, cmdAssign, game.Version, b}
		case stateTeamvoting:
			if rand.Intn(2) == 0 {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteTeam, game.Version, false}
			} else {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteTeam, game.Version, true}
			}
		case stateMission:
			if rand.Intn(2) == 0 {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteMission, game.Version, false}
			} else {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteMission, game.Version, true}
			}
		case stateSpywin:
			fallthrough
		case stateResistanceWin:
			wins++
			fmt.Println(wins)
			game.Cmd <- &wg.Command{player1, p1Conn, cmdReady, game.Version, nil}
		case stateLobby:
			game.Cmd <- &wg.Command{player1, p1Conn, cmdStart, game.Version, nil}
		default:
			log.Fatal("ERROR:", resistance.State)
		}
	}
}
