package wg

import "testing"

type fakeGame struct {
	Id       string
	Commands []*Command
}

func (f *fakeGame) Cmd(c *Command) {
	f.Commands = append(f.Commands, c)
}

func TestGames(t *testing.T) {
	games := NewGames()

	if len(games.Ids()) != 0 {
		t.Error("There shouldn't be any games yet")
	}

	fake := &fakeGame{Id: "f"}

	gameId := "1"
	games.Set(gameId, fake)
	maybe := games.Get(gameId).(*fakeGame)

	if maybe.Id != fake.Id {
		t.Error("Game I put in was not game I took out", fake, maybe)
	}

	if len(games.Ids()) != 1 {
		t.Error("There should be 1 game")
	}

	if games.Ids()[0] != gameId {
		t.Error("Unexpected game ID:", games.Ids(), "expected", gameId)
	}

	games.Delete(gameId)

	if len(games.Ids()) != 0 {
		t.Error("There should be 0 games", games.Ids())
	}
}
