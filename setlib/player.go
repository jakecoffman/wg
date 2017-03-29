package setlib

import (
	"github.com/jakecoffman/set-game/gamelib"
)

// Player is a user in-game
type Player struct {
	ws    gamelib.Connector
	Id    int
	Score int
	Connected bool
	ip string
}

type SetCommand struct {
	Type     string
	Ws       gamelib.Connector
	PlayerId string
	Locs     []int
	Version int
}

func (c *SetCommand) IsValid() bool {
	return c.PlayerId != "" && c.Type != ""
}

type userInput struct {
	Type string `json:"type"`
	Play []int `json:"play"`
	Join string `json:"join"`
	Version int `json:"version"`
}

// ProcessPlayerCommands is the main entry-point for players
func ProcessPlayerCommands(ws gamelib.Connector, playerId string) {
	input := &userInput{}
	var game gamelib.Game

	defer func() {
		if game != nil {
			game.Cmd(&SetCommand{Type: "Disconnect", PlayerId: playerId})
		}
	}()

	for {
		if err := ws.Recv(input); err != nil {
			return
		}
		switch input.Type {
		case "join":
			if game != nil {
				game.Cmd(&SetCommand{Type: "Leave", PlayerId: playerId})
				game = nil
			}

			id := input.Join

			// new
			if id == "" {
				id = gamelib.GenId()
				Games.Set(id, NewGame(id))
			}

			if game = Games.Get(id); game == nil {
				id = gamelib.GenId()
				Games.Set(id, NewGame(id))
			}
			game = Games.Get(id)
			game.Cmd(&SetCommand{Type: "Join", Ws: ws, PlayerId: playerId})
		case "play":
			if game != nil {
				game.Cmd(&SetCommand{Type: "Play", Locs: input.Play, PlayerId: playerId, Version: input.Version})
			}
		case "nosets":
			if game != nil {
				game.Cmd(&SetCommand{Type: "NoSets", PlayerId: playerId, Version: input.Version})
			}
		}
	}
}
