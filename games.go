package wg

import (
	"sync"
	"time"
)

var AllGames = NewGames()

const gameCleanup = 48 * time.Hour

func init() {
	// check if games are abandoned, and if so remove them
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			for _, id := range AllGames.Ids() {
				game := AllGames.Get(id)
				if time.Now().Sub(game.Created) > gameCleanup && time.Now().Sub(game.Updated) > gameCleanup {
					game.Cmd <- &Command{Type: cmdStop}
					AllGames.Delete(id)
				}
			}
		}
	}()
}

type Game struct {
	Class interface{}   `json:"-"`
	Cmd   chan *Command `json:"-"`

	Id      string
	Version int
	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`
}

func NewGame(class interface{}, id string) *Game {
	return &Game{
		Cmd:     make(chan *Command),
		Class:   class,
		Id:      id,
		Created: time.Now(),
		Updated: time.Now(),
	}
}

type Games struct {
	sync.RWMutex
	games map[string]*Game
	players map[string]*Game
}

func NewGames() *Games {
	return &Games{
		games: map[string]*Game{},
		players: map[string]*Game{},
	}
}

func (g *Games) Ids() []string {
	g.RLock()
	defer g.RUnlock()
	var ids []string
	for id := range g.games {
		ids = append(ids, id)
	}
	return ids
}

func (g *Games) Get(id string) *Game {
	g.RLock()
	defer g.RUnlock()
	return g.games[id]
}

func (g *Games) Set(game *Game, pid string) {
	if game.Id == "" {
		// this is programmer error, ok with panic
		panic("game needs an ID")
	}
	g.Lock()
	g.games[game.Id] = game
	g.players[pid] = game
	g.Unlock()
}

func (g *Games) Delete(id string) {
	g.Lock()
	delete(g.games, id)
	for pid, game := range g.players {
		if game.Id == id {
			defer delete(g.players, pid)
		}
	}
	g.Unlock()
}

func (g *Games) Find(pid string) *Game {
	g.RLock()
	defer g.RUnlock()
	return g.players[pid]
}
