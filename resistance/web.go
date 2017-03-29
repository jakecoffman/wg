package resistance

import (
	"net/http"
	"sort"
	"log"
	"html/template"
)

type info struct {
	Game    *Resist
	Players []*Player
	Sets    []string
}

func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	response := []info{}
	for _, id := range Games.Ids() {
		game := Games.Get(id).(*Resist)
		n4 := info{Game: game, Players: []*Player(game.Players)}
		response = append(response, n4)
	}
	sort.Slice(response, func(i, j int) bool {
		return response[i].Game.Updated.After(response[j].Game.Updated)
	})
	t, err := template.ParseFiles("www/resistance/admin.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if err = t.Execute(w, response); err != nil {
		log.Println(err)
	}
}
