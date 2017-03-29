package resistance

import (
	"github.com/jakecoffman/set-game/gamelib"
)

// Player is a user in-game
type Player struct {
	ws        gamelib.Connector
	Uuid      string `json:"-"`
	Id        int
	Connected bool
	Ip        string `json:"-"`
	IsBot bool
}

func Find(players []*Player, uuid string) (*Player, bool) {
	for _, player := range players {
		if player.Uuid == uuid {
			return player, true
		}
	}
	return nil, false
}

func Remove(players []*Player, uuid string) bool {
	for i, player := range players {
		if player.Uuid == uuid {
			players = append(players[0:i], players[i:]...)
			return true
		}
	}
	return false
}

type ResistCmd struct {
	Type     string
	PlayerId string
	Ws       gamelib.Connector
	Version  int
}

func (c *ResistCmd) IsValid() bool {
	return c.PlayerId != ""
}

type userInput struct {
	Type    string
	Join    string
	Version int
}

func ProcessPlayerCommands(ws gamelib.Connector, playerId string) {
	input := &userInput{}
	var game gamelib.Game

	defer func() {
		if game != nil {
			game.Cmd(&ResistCmd{Type: disconnect, PlayerId: playerId})
		}
	}()

	for {
		if err := ws.Recv(input); err != nil {
			return
		}
		switch input.Type {
		case join:
			if game != nil {
				game.Cmd(&ResistCmd{Type: leave, PlayerId: playerId})
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
			game.Cmd(&ResistCmd{Type: join, Ws: ws, PlayerId: playerId})
		default:
			game.Cmd(&ResistCmd{Type: input.Type, Ws: ws, PlayerId: playerId, Version: input.Version})
		}
	}
}
