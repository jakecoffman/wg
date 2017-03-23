package gamelib

import "sync"

type Game interface {
	Cmd(Command)
}

type Command interface {
	IsValid() bool
}

type Games struct {
	sync.RWMutex
	games map[string]Game
}

func NewGames() *Games {
	return &Games{games: map[string]Game{}}
}

func (g *Games) Ids() []string {
	g.RLock()
	defer g.RUnlock()
	ids := []string{}
	for id := range g.games {
		ids = append(ids, id)
	}
	return ids
}

func (g *Games) Get(id string) Game {
	g.RLock()
	defer g.RUnlock()
	return g.games[id]
}

func (g *Games) Set(id string, game Game) {
	g.Lock()
	g.games[id] = game
	g.Unlock()
}

func (g *Games) Delete(id string) {
	g.Lock()
	delete(g.games, id)
	g.Unlock()
}
