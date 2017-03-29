package example

import (
	"net/http"
	"sort"
	"log"
	"html/template"
)

type info struct {
	Game    *Example
	Players interface{}
	Sets    []string
}

func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	response := []info{}
	for _, id := range Games.Ids() {
		game := Games.Get(id).(*Example)
		n4 := info{Game: game}
		response = append(response, n4)
	}
	sort.Slice(response, func(i, j int) bool {
		return response[i].Game.Updated.After(response[j].Game.Updated)
	})
	t, err := template.ParseFiles("www/example/admin.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if err = t.Execute(w, response); err != nil {
		log.Println(err)
	}
}
