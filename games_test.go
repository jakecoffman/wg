package wg

import (
	"testing"
)

type FakeGame struct{}

func TestGames(t *testing.T) {
	games := NewGames[*FakeGame]()

	if len(games.Ids()) != 0 {
		t.Error("There shouldn't be any games yet")
	}

	game := NewGame[*FakeGame](nil, "1")
	games.Set(game, "1")
	maybe := games.Get("1")

	if maybe.Id != game.Id {
		t.Error("Game I put in was not game I took out", game, maybe)
	}

	if len(games.Ids()) != 1 {
		t.Error("There should be 1 game")
	}

	if games.Ids()[0] != "1" {
		t.Error("Unexpected game ID:", games.Ids(), "expected 1")
	}

	games.Delete("1")

	if len(games.Ids()) != 0 {
		t.Error("There should be 0 games", games.Ids())
	}
}
