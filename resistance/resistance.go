package resistance

import (
	"github.com/jakecoffman/set-game/gamelib"
	"log"
	"time"
	"github.com/google/uuid"
)

type Resist struct {
	cmd          chan *ResistCmd `json:"-"`

	Id           string

	Players      []*Player
	playerCursor int

	State       string
	Version      int
	Created      time.Time `json:"-"`
	Updated      time.Time `json:"-"`
}

func NewGame(id string) *Resist {
	g := &Resist{
		cmd: make(chan *ResistCmd),

		Players: []*Player{},
		playerCursor: 1,
		Id: id,
		Created: time.Now(),
	}
	go g.run()
	g.reset()
	return g
}

func (g *Resist) reset() {
	g.State = lobby
}

func (g *Resist) Cmd(c gamelib.Command) {
	g.cmd <- c.(*ResistCmd)
}

// states
const (
	lobby = "lobby"
	teambuilding = "team"
	mission = "mission"
)

// message types
const (
	join = "join"
	leave = "leave"
	disconnect = "disconnect"
	stop = "stop"

	addbot = "addbot"
	removebot = "removebot"
	start = "start"
	end = "end"
)

func (g *Resist) run() {
	var cmd *ResistCmd
	for {
		cmd = <-g.cmd
		if !cmd.IsValid() {
			log.Printf("Invalid command sent: %#v\n", cmd)
			continue
		}
		log.Println("Processing", cmd.Type)
		switch cmd.Type {
		case join:
			var player *Player
			var ok bool
			if player, ok = Find(g.Players, cmd.PlayerId); !ok {
				// player was not here before
				if len(g.Players) >= 10 {
					// can't have more than 10 players
					continue
				}
				player = &Player{Uuid: cmd.PlayerId, Id: g.playerCursor}
				g.Players = append(g.Players, player)
				g.playerCursor += 1
			}
			player.ws = cmd.Ws
			player.Connected = true
			player.Ip = player.ws.Request().Header.Get("X-Forwarded-For")
		case leave:
			Remove(g.Players, cmd.PlayerId)
		case disconnect:
			player, found := Find(g.Players, cmd.PlayerId)
			if !found {
				log.Println("Couldn't find player", cmd.PlayerId)
				continue
			}
			player.ws = nil
			player.Connected = false
		case stop:
			return
		case addbot:
			if len(g.Players) >= 10 {
				continue
			}
			player := &Player{Uuid: uuid.New().String(), Id: g.playerCursor, IsBot: true}
			g.Players = append(g.Players, player)
			g.playerCursor += 1
		case removebot:
			for _, p := range g.Players {
				if p.IsBot {
					Remove(g.Players, p.Uuid)
					break
				}
			}
		case start:
			if g.Version != cmd.Version || g.State != lobby || len(g.Players) < 5 {
				continue
			}
			g.State = teambuilding
			// remove unconnected players
			// mix up where players are "seated"
			// assign secret roles to players (based on # of players)
			// assign the initial leader
		case end:
			g.reset()
		default:
			log.Println("Unknown message:", cmd.Type)
			continue
		}
		g.sendEveryoneEverything()
		g.Updated = time.Now()
	}
}

type UpdateMsg struct {
	Type   string
	Update interface{}
}

func (g *Resist) sendEverythingTo(ws gamelib.Connector) {
	if ws == nil {
		return
	}
	ws.Send(UpdateMsg{Type: "all", Update: g})
}

func (g *Resist) sendEveryoneEverything() {
	g.sendAll(&UpdateMsg{Type: "all", Update: g})
}

func (g *Resist) sendMetaToEveryone() {
	g.sendAll(&UpdateMsg{Type: "meta", Update: g})
}

func (g *Resist) sendAll(msg interface{}) {
	for _, player := range g.Players {
		if player.ws != nil {
			log.Println("Sending to", player.Id)
			player.ws.Send(msg)
		}
	}
}
