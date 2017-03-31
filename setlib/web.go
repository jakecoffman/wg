package setlib

import (
	"net/http"
	"fmt"
	"sort"
	"html/template"
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
			compactSets = append(compactSets, fmt.Sprint(set[0] + 1, ",", set[1] + 1, ",", set[2] + 1))
		}
		n4 := info{Game: game, Players: game.SlicePlayersAdmin(), Sets: compactSets}
		response = append(response, n4)
	}
	sort.Slice(response, func(i, j int) bool {
		return response[i].Game.Updated.After(response[j].Game.Updated)
	})
	t, err := template.New("admin").Parse(adminPage)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if err = t.Execute(w, response); err != nil {
		log.Println(err)
	}
}

const adminPage = `
<html>
<head>
    <style>
        table {
            width: 100%;
            border: 1px solid black;
        }
        td {
            text-align: center;
            border: 1px solid black;
        }
    </style>
</head>
<body>
<table>
    <thead>
    <tr>
        <th>Game ID (# Plays)</th>
        <th>Players (Score)</th>
        <th>Sets</th>
        <th>Created</th>
        <th>Updated</th>
    </tr>
    </thead>
    <tbody>
    {{range $i, $data := .}}
    <tr>
        <td>{{$data.Game.Id}} ({{$data.Game.Version}})</td>
        <td>
            {{ range $_, $p := $data.Players}}
            {{$p.Id}} ({{$p.Score}}) {{$p.Addr}} {{if $p.Connected}}✅{{end}}<br/>
            {{end}}
        </td>
        <td>
            {{ range $_, $s := $data.Sets}}
            {{$s}} <br/>
            {{end}}
        </td>
        <td>{{$data.Game.Created.Format "01-02 15:04:05"}}</td>
        <td>{{$data.Game.Updated.Format "01-02 15:04:05"}}</td>
    </tr>
    {{end}}
    </tbody>
</table>
</body>
</html>
`