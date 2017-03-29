package example

import (
	"github.com/jakecoffman/set-game/gamelib"
)

// Player is a user in-game
type Player struct {
	ws        gamelib.Connector
	Uuid      string `json:"-"`
	Id        int
	Connected bool
	ip        string
}

type Players []*Player

func (p Players) Find(uuid string) (*Player, bool) {
	for _, player := range []*Player(p) {
		if player.Uuid == uuid {
			return player, true
		}
	}
	return nil, false
}

func (p Players) Remove(uuid string) bool {
	players := []*Player(p)
	for i, player := range players {
		if player.Uuid == uuid {
			players = append(p[0:i], p[i:]...)
			return true
		}
	}
	return false
}

type ExampleCommand struct {
	Type     string
	PlayerId string
	Ws       gamelib.Connector
}

func (c *ExampleCommand) IsValid() bool {
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
			game.Cmd(&ExampleCommand{Type: "Disconnect", PlayerId: playerId})
		}
	}()

	for {
		if err := ws.Recv(input); err != nil {
			return
		}
		switch input.Type {
		case "join":
			if game != nil {
				game.Cmd(&ExampleCommand{Type: "Leave", PlayerId: playerId})
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
			game.Cmd(&ExampleCommand{Type: "Join", Ws: ws, PlayerId: playerId})
		}
	}
}
