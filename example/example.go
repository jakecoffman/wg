package example

import (
	"github.com/jakecoffman/set-game/gamelib"
	"log"
	"time"
)

type Example struct {
	cmd          chan *ExampleCommand `json:"-"`

	Id           string

	Players      Players
	playerCursor int

	InGame       bool // false means in lobby
	Version      int
	Created      time.Time `json:"-"`
	Updated      time.Time `json:"-"`
}

func NewGame(id string) *Example {
	g := &Example{
		cmd: make(chan *ExampleCommand),

		Players: []*Player{},
		playerCursor: 1,
		Id: id,
		Created: time.Now(),
	}
	go g.run()
	g.reset()
	return g
}

func (g *Example) reset() {

}

func (g *Example) Cmd(c gamelib.Command) {
	g.cmd <- c.(*ExampleCommand)
}

func (g *Example) run() {
	var cmd *ExampleCommand
	for {
		cmd = <-g.cmd
		if !cmd.IsValid() {
			log.Printf("Invalid command sent: %#v\n", cmd)
			continue
		}
		switch cmd.Type {
		case "Join":
			var player *Player
			var ok bool
			if player, ok = g.Players.Find(cmd.PlayerId); !ok {
				// player was not here before, create
				player = &Player{Uuid: cmd.PlayerId, Id: g.playerCursor}
				g.Players = append(g.Players, player)
				g.playerCursor += 1
			}
			player.ws = cmd.Ws
			player.Connected = true
			player.ip = player.ws.Request().Header.Get("X-Forwarded-For")
			g.sendEverythingTo(cmd.Ws)
			g.sendMetaToEveryone()
		case "Leave":
			g.Players.Remove(cmd.PlayerId)
			g.sendMetaToEveryone()
		case "Disconnect":
			p, _ := g.Players.Find(cmd.PlayerId)
			p.ws = nil
			p.Connected = false
			g.sendMetaToEveryone()
		case "Stop":
			return
		}
		g.Updated = time.Now()
	}
}

type UpdateMsg struct {
	Type    string
	Update interface{}
}

func (g *Example) sendEverythingTo(ws gamelib.Connector) {
	if ws == nil {
		return
	}
	ws.Send(UpdateMsg{Type: "all", Update: g})
}

func (g *Example) sendEveryoneEverything() {
	g.sendAll(&UpdateMsg{Type: "all", Update: g})
}

func (g *Example) sendMetaToEveryone() {
	g.sendAll(&UpdateMsg{Type: "meta", Update: g})
}

func (g *Example) sendAll(msg interface{}) {
	for _, player := range g.Players {
		if player.ws != nil {
			player.ws.Send(msg)
		}
	}
}
