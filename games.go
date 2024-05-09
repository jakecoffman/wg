package wg

import (
	"sync"
	"time"
)

const gameCleanup = 48 * time.Hour

type Game[T any] struct {
	Class T             `json:"-"`
	Cmd   chan *Command `json:"-"`

	Id      string
	Version int
	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`
}

func NewGame[T any](class T, id string) *Game[T] {
	return &Game[T]{
		Cmd:     make(chan *Command),
		Class:   class,
		Id:      id,
		Created: time.Now(),
		Updated: time.Now(),
	}
}

type Games[T any] struct {
	sync.RWMutex
	games   map[string]*Game[T]
	players map[string]*Game[T]
}

func NewGames[T any]() *Games[T] {
	games := &Games[T]{
		games:   map[string]*Game[T]{},
		players: map[string]*Game[T]{},
	}
	// check if games are abandoned, and if so remove them
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			for _, id := range games.Ids() {
				game := games.Get(id)
				if time.Now().Sub(game.Created) > gameCleanup && time.Now().Sub(game.Updated) > gameCleanup {
					game.Cmd <- &Command{Type: cmdStop}
					games.Delete(id)
				}
			}
		}
	}()
	return games
}

func (g *Games[T]) Ids() []string {
	g.RLock()
	defer g.RUnlock()
	var ids []string
	for id := range g.games {
		ids = append(ids, id)
	}
	return ids
}

func (g *Games[T]) Get(id string) *Game[T] {
	g.RLock()
	defer g.RUnlock()
	return g.games[id]
}

func (g *Games[T]) Set(game *Game[T], pid string) {
	if game.Id == "" {
		// this is programmer error, ok with panic
		panic("game needs an ID")
	}
	g.Lock()
	g.games[game.Id] = game
	g.players[pid] = game
	g.Unlock()
}

func (g *Games[T]) Delete(id string) {
	g.Lock()
	delete(g.games, id)
	for pid, game := range g.players {
		if game.Id == id {
			defer delete(g.players, pid)
		}
	}
	g.Unlock()
}

func (g *Games[T]) Find(pid string) *Game[T] {
	g.RLock()
	defer g.RUnlock()
	return g.players[pid]
}
