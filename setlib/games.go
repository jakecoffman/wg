package setlib

import (
	"time"
	"github.com/jakecoffman/set-game/gamelib"
)

var Games = gamelib.NewGames()

func init() {
	// check if games are abandoned, and if so remove them
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			for _, id := range Games.Ids() {
				game := Games.Get(id).(*Set)
				if time.Now().Sub(game.Updated).Hours() > 24 && game.NumConns() == 0 {
					game.Cmd(&SetCommand{Type: "Stop"})
					Games.Delete(id)
				}
			}
		}
	}()
}
