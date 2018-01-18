package wg

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

const letterBytes = "1234567890"

func init() {
	rand.Seed(time.Now().Unix())
}

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

func ProcessPlayerCommands(AllGames *Games, NewGame func(string) Game) func(Connector, string) {
	return func(ws Connector, playerId string) {
		input := &Command{}
		var game Game

		defer func() {
			if game != nil {
				game.Cmd(&Command{Type: cmdDisconnect, PlayerId: playerId})
			}
		}()

		var id string
		for {
			if err := ws.Recv(input); err != nil {
				return
			}
			input.Ws = ws
			input.PlayerId = playerId
			switch input.Type {
			case cmdJoin:
				if game != nil {
					game.Cmd(&Command{Type: cmdLeave, PlayerId: playerId})
					game = nil
				}

				if err := json.Unmarshal(input.Data, &id); err != nil {
					log.Println("Couldn't decode join code", err)
					continue
				}

				// new
				if id == "" {
					id = GenId()
					AllGames.Set(id, NewGame(id))
				}

				if game = AllGames.Get(id); game == nil {
					id = GenId()
					game = NewGame(id)
					AllGames.Set(id, game)
				}
				game.Cmd(input)
			case cmdStop:
			// players can't stop the game goroutine
			default:
				if game != nil {
					game.Cmd(input)
				}
			}
		}
	}
}
