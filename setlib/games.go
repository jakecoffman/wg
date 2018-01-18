package setlib

import (
	"github.com/jakecoffman/wg"
	"time"
)

func init() {
	// check if games are abandoned, and if so remove them
	// TODO this is a race condition
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			for _, id := range AllGames.Ids() {
				game := AllGames.Get(id).(*Set)
				if time.Now().Sub(game.Updated).Hours() > 24 && game.NumConns() == 0 {
					game.Cmd(&wg.Command{Type: cmdStop})
					AllGames.Delete(id)
				}
			}
		}
	}()
}

var AllGames = wg.NewGames()
