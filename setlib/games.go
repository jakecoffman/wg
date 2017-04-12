package setlib

import (
	"time"
	"github.com/jakecoffman/set-game/gamelib"
)

func init() {
	// check if games are abandoned, and if so remove them
	// TODO this is a race condition
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			for _, id := range gamelib.AllGames.Ids() {
				game := gamelib.AllGames.Get(id).(*Set)
				if time.Now().Sub(game.Updated).Hours() > 24 && game.NumConns() == 0 {
					game.Cmd(&gamelib.Command{Type: cmd_stop})
					gamelib.AllGames.Delete(id)
				}
			}
		}
	}()
}
