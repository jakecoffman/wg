package resistance

import (
	"testing"
	"github.com/jakecoffman/wg"
	"github.com/gin-gonic/gin/json"
	"math/rand"
	"log"
	"time"
)

func TestResistance(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	const gameId = "0"
	const player1 = "1"
	p1Conn := wg.NewFakeConn(player1)

	// new game
	game := &Resist{
		Players:      []*Player{},
		playerCursor: 1,
	}
	game.Game = wg.NewGame(game, gameId)
	go game.run()
	game.reset()

	game.Cmd <- &wg.Command{player1, p1Conn, cmdJoin, game.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, game.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, game.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, game.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdAddBot, game.Version, nil}
	game.Cmd <- &wg.Command{player1, p1Conn, cmdStart, game.Version, nil}

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

		// let the game goroutine go, probably should improve this with locking
		time.Sleep(1*time.Millisecond)

		switch game.State {
		case stateTeambuilding:
			assignment := rand.Perm(5)[:game.Missions[game.CurrentMission].Slots]
			b, _ := json.Marshal(assignment)
			game.Cmd <- &wg.Command{player1, p1Conn, cmdAssign, game.Version, b}
		case stateTeamvoting:
			if rand.Intn(2) == 0 {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteTeam, game.Version, []byte("false")}
			} else {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteTeam, game.Version, []byte("true")}
			}
		case stateMission:
			if rand.Intn(2) == 0 {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteMission, game.Version, []byte("false")}
			} else {
				game.Cmd <- &wg.Command{player1, p1Conn, cmdVoteMission, game.Version, []byte("true")}
			}
		case stateSpywin:
			fallthrough
		case stateResistanceWin:
			wins++
			game.Cmd <- &wg.Command{player1, p1Conn, cmdReady, game.Version, nil}
		case stateLobby:
			game.Cmd <- &wg.Command{player1, p1Conn, cmdStart, game.Version, nil}
		default:
			log.Fatal("ERROR:", game.State)
		}
	}
}
