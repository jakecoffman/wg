package wg

import (
	"encoding/json"
	"log"
	"math/rand"
)

const letterBytes = "1234567890"

func GenId() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

const (
	cmdDisconnect = "disconnect"
	cmdJoin       = "join"
	cmdLeave      = "leave"
	cmdStop       = "stop"
)

type Command struct {
	PlayerId string
	Ws       Connector
	Type     string
	Version  int
	Data     json.RawMessage
}

func ProcessPlayerCommands(NewGame func(string) *Game) func(Connector, string) {
	return func(ws Connector, playerId string) {
		cmd := &Command{}
		var game *Game

		defer func() {
			if game != nil {
				game.Cmd <- &Command{Type: cmdDisconnect, PlayerId: playerId}
			}
		}()

		var id string
		for {
			if err := ws.Recv(cmd); err != nil {
				return
			}
			cmd.Ws = ws
			cmd.PlayerId = playerId
			switch cmd.Type {
			case cmdJoin:
				if game != nil {
					game.Cmd <- &Command{Type: cmdLeave, PlayerId: playerId}
					game = nil
				}

				if err := json.Unmarshal(cmd.Data, &id); err != nil {
					log.Println("Couldn't decode join code", err)
					continue
				}

				// new
				if id == "" {
					id = GenId()
					AllGames.Set(NewGame(id))
				}

				if game = AllGames.Get(id); game == nil {
					id = GenId()
					game = NewGame(id)
					AllGames.Set(game)
				}
				game.Cmd <- cmd
			case cmdStop:
				// players can't stop the game goroutine
			default:
				if game != nil {
					game.Cmd <- cmd
				}
			}
		}
	}
}
