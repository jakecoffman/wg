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
	IsSpy     bool `json:"-"`
	IsBot     bool
	IsReady   bool
	IsLeader  bool
	OnMission bool
}

func Find(players []*Player, uuid string) (*Player, int) {
	for i, player := range players {
		if player.Uuid == uuid {
			return player, i
		}
	}
	return nil, -1
}

type ResistCmd struct {
	*UserInput
	PlayerId string
	Ws       gamelib.Connector
}

func (c *ResistCmd) IsValid() bool {
	return c.PlayerId != ""
}

type UserInput struct {
	Type       string
	Join       string
	Version    int
	Assignment []int // leader's team assignment (player locations in array)
	Vote       bool  // used for team accept and voting on missions
}

func ProcessPlayerCommands(ws gamelib.Connector, playerId string) {
	input := &UserInput{}
	var game gamelib.Game

	defer func() {
		if game != nil {
			game.Cmd(&ResistCmd{UserInput: &UserInput{Type: msg_disconnect}, PlayerId: playerId})
		}
	}()

	for {
		if err := ws.Recv(input); err != nil {
			return
		}
		switch input.Type {
		case msg_join:
			if game != nil {
				game.Cmd(&ResistCmd{UserInput: &UserInput{Type: msg_leave}, PlayerId: playerId})
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
			game.Cmd(&ResistCmd{Ws: ws, PlayerId: playerId, UserInput: input})
		case msg_stop:
			// players can't delete the game
		default:
			game.Cmd(&ResistCmd{Ws: ws, PlayerId: playerId, UserInput: input})
		}
	}
}
