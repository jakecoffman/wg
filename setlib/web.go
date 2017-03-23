package setlib

import (
	"net/http"
	"fmt"
	"sort"
	"text/template"
	"log"
)

type info struct {
	Game    *Set
	Players interface{}
	Sets    []string
}

func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	response := []info{}
	for _, id := range Games.Ids() {
		game := Games.Get(id).(*Set)
		sets := game.FindSets()
		compactSets := []string{}
		for _, set := range sets {
			compactSets = append(compactSets, fmt.Sprint(set[0] + 1, " ", set[1] + 1, " ", set[2] + 1))
		}
		n4 := info{Game: game, Players: game.SlicePlayersAdmin(), Sets: compactSets}
		response = append(response, n4)
	}
	sort.Slice(response, func(i, j int) bool {
		return response[i].Game.Updated.After(response[j].Game.Updated)
	})
	t, err := template.ParseFiles("set-game/admin.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if err = t.Execute(w, response); err != nil {
		log.Println(err)
	}
}
