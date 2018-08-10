package citadels

import (
	"github.com/jakecoffman/wg"
	"log"
	"time"
	"runtime/debug"
	"math/rand"
	"encoding/json"
	"sort"
	"fmt"
)

type Citadels struct {
	*wg.Game

	Players      []*Player
	playerCursor int

	Turn         Circular // used to tell whose turn it is
	CharCur      int
	State        State
	characters   []*ChoosableCharacter // the characters in this game (since characters can be substituted)
	districtDeck []*District
	crown        Circular // tracks who practically has the crown (doesn't move until next turn)
	FirstToEight int

	Kill int // assassin chose to kill this player
}

type ChoosableCharacter struct {
	*Character
	Chosen bool    // lets the frontend know that character has been chosen or discarded
	player *Player // point back to the player that is assigned this round
}

type Circular struct {
	Value int
	Max   int
}

func (c *Circular) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

// Inc acts like ++ in other languages: returns the current value and then increments, however if the
// value is more than the max, the value resets to 0.
func (c *Circular) Inc() int {
	rv := c.Value
	c.Value++
	if c.Value >= c.Max {
		c.Value = 0
	}
	return rv
}

type Player struct {
	ws        wg.Connector
	Uuid      string `json:"-"`
	Id        int
	Name      string
	Connected bool
	Ip        string `json:"-"`

	IsBot     bool // TODO implement bots for this game
	HasCrown  bool
	Gold      int
	hand      []*District
	Districts []*District

	score int
	IsReady bool `json:",omitempty"`
}

func NewGame(id string) *wg.Game {
	c := &Citadels{
		Players:      []*Player{},
		playerCursor: 1,
	}
	c.Game = wg.NewGame(c, id)
	go c.run()
	c.reset()
	return c.Game
}

func (c *Citadels) reset() {
	c.districtDeck = make([]*District, 0, len(Districts))
	c.crown = Circular{Value: 0, Max: len(c.Players)}
	c.State = lobby
	for _, p := range c.Players {
		p.Gold = 2
		p.HasCrown = false
		p.Districts = []*District{}
		p.IsReady = false
	}
}

type State int

func (s State) String() string {
	switch int(s) {
	case 0:
		return "Lobby"
	case 1:
		return "Choose"
	case 2:
		return "GoldOrDraw"
	case 3:
		return "PutCardBack"
	case 4:
		return "Build"
	case 5:
		return "EndTurn"
	case 6:
		return "GameOver"
	default:
		panic(s)
	}
}

const (
	lobby = State(iota)
	choose
	// player turns
	goldOrDraw
	putCardBack
	build
	endTurn
	// game over man
	gameOver
)

// message types
const (
	cmdJoin       = "join"
	cmdLeave      = "leave"
	cmdDisconnect = "disconnect"
	cmdStop       = "stop"
	cmdName       = "name"
	cmdReady      = "ready"

	// anyone can do these things
	cmdAddBot    = "addbot"
	cmdRemoveBot = "removebot"
	cmdStart     = "start"

	cmdChoose = "choose"

	// turn actions
	cmdAction = "action"
	cmdBuild  = "build"
	cmdPowers = "powers"
	cmdEnd    = "end"
)

func (c *Citadels) run() {
	var cmd *wg.Command

	defer func() {
		if r := recover(); r != nil {
			log.Println("Game crashed", r)
			log.Printf("State: %#v\n", c)
			log.Println("Last command received:", cmd)
			debug.PrintStack()
		}
	}()

	var update bool
	for {
		cmd = <-c.Cmd

		if c.Version != cmd.Version {
			continue
		}

		if cmd.Type == cmdStop {
			return
		} else {
			update = c.handler(cmd)
		}

		if update {
			c.sendEveryoneEverything()
			c.Updated = time.Now()
		}
	}

}

func (c *Citadels) handler(cmd *wg.Command) bool {
	switch cmd.Type {
	case cmdJoin:
		return c.handleJoin(cmd)
	case cmdLeave:
		return c.handleLeave(cmd)
	case cmdDisconnect:
		return c.handleDisconnect(cmd)
	case cmdName:
		return c.handleName(cmd)
	case cmdStart:
		return c.handleStart(cmd)
	case cmdChoose:
		return c.handleChoose(cmd)
	case cmdAction:
		return c.handleAction(cmd)
	case cmdBuild:
		return c.handleBuild(cmd)
	case cmdPowers:
		return c.handleSpecial(cmd)
	case cmdEnd:
		return c.handleEndTurn(cmd)
	case cmdReady:
		return c.handleReady(cmd)
	default:
		log.Println("Unknown message:", cmd.Type)
		return false
	}
}

// Find returns the player object and the position they are in
func Find(players []*Player, uuid string) (*Player, int) {
	for i, player := range players {
		if player.Uuid == uuid {
			return player, i
		}
	}
	return nil, -1
}

type MsgMsg struct {
	Type string
	Msg  string
}

func sendMsg(c wg.Connector, msg string) {
	c.Send(&MsgMsg{Type: "msg", Msg: msg})
}

func (c *Citadels) handleJoin(cmd *wg.Command) bool {
	player, i := Find(c.Players, cmd.PlayerId)
	if i == -1 {
		// player was not here before
		if c.State != lobby {
			sendMsg(cmd.Ws, "Can't join game in progress")
			return false
		}
		if len(c.Players) >= 10 {
			// can't have more than 10 players
			sendMsg(cmd.Ws, "Can't have more than 10 players")
			return false
		}
		player = &Player{Uuid: cmd.PlayerId, Id: c.playerCursor}
		c.Players = append(c.Players, player)
		c.playerCursor += 1
	}
	player.ws = cmd.Ws
	player.Connected = true
	player.Ip = player.ws.Ip()

	if c.State != lobby {
		c.sendGameInfo(cmd.Ws)
	}

	log.Println("Player", player.Id, "joined")
	return true
}

func (c *Citadels) handleLeave(cmd *wg.Command) bool {
	for i, player := range c.Players {
		if player.Uuid == cmd.PlayerId {
			c.Players = append(c.Players[0:i], c.Players[i+1:]...)
			log.Println("Player", i, "left")
			return true
		}
	}
	return false
}

func (c *Citadels) handleDisconnect(cmd *wg.Command) bool {
	player, i := Find(c.Players, cmd.PlayerId)
	if i == -1 {
		log.Println("Couldn't find player", cmd.PlayerId)
		return false
	}
	player.ws = nil
	player.Connected = false
	return true
}

func (c *Citadels) handleStart(cmd *wg.Command) bool {
	if c.Version != cmd.Version {
		sendMsg(cmd.Ws, "Someone else started the game first")
		return false
	}

	if c.State != lobby {
		sendMsg(cmd.Ws, "Illegal state")
		return false
	}

	if len(c.Players) < 2 || len(c.Players) > 7 {
		sendMsg(cmd.Ws, "Need 2-7 players to start the game")
		return false
	}

	// TODO handle game variants, for now this is just the 2 player game

	c.State = choose
	c.Kill = -1
	c.FirstToEight = -1

	// remove unconnected players and shuffle them, leader always starts in position 1
	{
		var newPlayers []*Player
		walk := rand.Perm(len(c.Players))
		for _, i := range walk {
			if !c.Players[i].IsBot && !c.Players[i].Connected {
				continue
			}
			newPlayers = append(newPlayers, c.Players[i])
		}
		c.Players = newPlayers
		c.crown = Circular{Value: 0, Max: len(c.Players)}
		c.Players[0].HasCrown = true
		c.Turn.Value = 0
		c.Turn.Max = len(c.Players)
	}

	// TODO: add the expansion characters and allow users to choose
	c.characters = []*ChoosableCharacter{
		{Character: Assassin},
		{Character: Thief},
		{Character: Magician},
		{Character: King},
		{Character: Bishop},
		{Character: Merchant},
		{Character: Architect},
		{Character: Warlord},
	}
	// shuffle the deck
	for _, i := range rand.Perm(len(Districts)) {
		c.districtDeck = append(c.districtDeck, Districts[i])
	}

	// 2 player variant only: discard 1 without anyone seeing
	// TODO other player variants
	if len(c.Players) == 2 {
		c.characters[rand.Intn(8)].Chosen = true
	}

	// deal 4 districts to each player, and give starting gold
	for _, p := range c.Players {
		p.hand = c.districtDeck[0:4]
		c.districtDeck = c.districtDeck[4:]
		p.Gold = 2
		p.Districts = []*District{}
	}

	// tell all players the characters in this game
	for _, p := range c.Players {
		if p.ws != nil {
			c.sendGameInfo(p.ws)
		}
	}

	log.Println("Game started")

	return true
}

func (c *Citadels) handleName(cmd *wg.Command) bool {
	p, _ := Find(c.Players, cmd.PlayerId)
	if c.State != lobby && p.Name != "" {
		sendMsg(p.ws, "Wait for the lobby to change your name again")
		return false
	}

	var name string
	err := json.Unmarshal(cmd.Data, &name)
	if err != nil {
		log.Println(err)
		sendMsg(p.ws, "Got invalid data for name")
		return false
	}

	if len(name) > 8 {
		p.Name = name[0:8]
	} else {
		p.Name = name
	}

	return true
}

func (c *Citadels) handleChoose(cmd *wg.Command) bool {
	p, i := Find(c.Players, cmd.PlayerId)
	if c.Turn.Value != i {
		log.Println("Not your turn")
		return false
	}

	if c.State != choose {
		log.Println("Wrong state")
		return false
	}

	var choice int
	if err := json.Unmarshal(cmd.Data, &choice); err != nil {
		log.Println(err)
		sendMsg(p.ws, "Couldn't decode choice")
		return false
	}

	if choice > 8 || choice < 0 {
		sendMsg(p.ws, "Invalid choice")
		return false
	}

	if c.characters[choice].Chosen {
		sendMsg(p.ws, "Character already chosen")
		return false
	}

	choosed := 0
	for _, char := range c.characters {
		if char.Chosen {
			choosed++
		}
	}

	// TODO: other player variants
	if len(c.Players) == 2 {
		switch choosed {
		case 1:
			// player 1 chose their first character, turn over
			c.characters[choice].Chosen = true
			c.characters[choice].player = p
			c.Turn.Inc()
		case 2:
			// player 2 chose their first character and gets to go again
			c.characters[choice].Chosen = true
			c.characters[choice].player = p
		case 3:
			// player 2 is discarding this character
			c.characters[choice].Chosen = true
			c.Turn.Inc()
		case 4:
			// player 1 chose their second character and gets to go again
			c.characters[choice].Chosen = true
			c.characters[choice].player = p
		case 5:
			// player 1 is discarding this character
			c.characters[choice].Chosen = true
			c.Turn.Inc()
		case 6:
			// player 2 chose their second character and this phase is over
			c.characters[choice].Chosen = true
			c.characters[choice].player = p
			c.State = goldOrDraw

			// figure out whose turn it is
			c.CharCur = 0
			for i, char := range c.characters {
				if char.player != nil {
					_, c.Turn.Value = Find(c.Players, char.player.Uuid)
					if i == 4 && c.characters[4].Character == King {
						c.crown.Value = c.Turn.Value
					}
					c.CharCur = i
					log.Println("Round starting: Player", c.Turn.Value)
					return true
				}
			}
		default:
			log.Println("Unexpected state:", choosed, choice, p, c)
			return false
		}
	}

	log.Println("Player made a character choice")
	return true
}

func (c *Citadels) handleAction(cmd *wg.Command) bool {
	p, _ := Find(c.Players, cmd.PlayerId)
	if p != c.Players[c.Turn.Value] {
		sendMsg(p.ws, "Not your turn yet")
		return false
	}
	if c.State < goldOrDraw || c.State > putCardBack {
		sendMsg(p.ws, "It's not time for actions")
		return false
	}

	// player can either get 2 gold or draw 2 cards and put one card back
	if c.State == goldOrDraw {
		var choice int
		if err := json.Unmarshal(cmd.Data, &choice); err != nil {
			log.Println(err)
			sendMsg(p.ws, "couldn't unmarshal choice")
			return false
		}
		// merchant draws an additional gold
		if c.CharCur == 5 {
			p.Gold++
		}
		// architect draws two additional district cards
		if c.CharCur == 6 && c.characters[6].Character == Architect {
			p.hand = append(p.hand, c.districtDeck[:2]...)
			c.districtDeck = c.districtDeck[2:]
		}
		if choice == 0 {
			p.Gold += 2
			c.State = build
			log.Println("Player chose gold")
			return true
		}
		// give two cards, they will return one next
		c.State = putCardBack
		p.hand = append(p.hand, c.districtDeck[:2]...)
		c.districtDeck = c.districtDeck[2:]
		log.Println("Player chose districts")
		return true
	}

	if c.State != putCardBack {
		panic(fmt.Sprintln("Wrong state:", c.State))
		return false
	}

	var choices []int
	if err := json.Unmarshal(cmd.Data, &choices); err != nil {
		log.Println(err)
		sendMsg(p.ws, "couldn't unmarshal choice")
		return false
	}

	if len(choices) != 1 {
		sendMsg(p.ws, "select one card to put back")
		return false
	}

	choice := choices[0]

	if choice < len(p.hand)-2 || choice > len(p.hand)-1 {
		sendMsg(p.ws, "discard one of the cards you drew (last 2)")
		return false
	}

	// if they chose second to last, swap the last two cards
	if choice == len(p.hand)-2 {
		p.hand[len(p.hand)-2] = p.hand[len(p.hand)-1]
	}
	// drop the last card off their deck
	p.hand = p.hand[:len(p.hand)-1]
	c.State = build
	log.Println("Player returned a district")
	return true
}

func (c *Citadels) handleBuild(cmd *wg.Command) bool {
	p, _ := Find(c.Players, cmd.PlayerId)
	if p != c.Players[c.Turn.Value] {
		log.Println("Not your turn yet")
		sendMsg(p.ws, "Not your turn yet")
		return false
	}
	if c.State != build {
		log.Println("Not time to build")
		sendMsg(p.ws, "It's not time to build")
		return false
	}

	var choices []int
	if err := json.Unmarshal(cmd.Data, &choices); err != nil {
		log.Println(err)
		sendMsg(p.ws, "Couldn't unmarshal choice")
		return false
	}

	if len(choices) == 0 {
		c.State = endTurn
		return true
	}

	if c.CharCur == 6 && c.characters[6].Character == Architect && len(choices) > 3 {
		sendMsg(p.ws, "Architect can only build up to three times per round")
		return false
	} else {
		if len(choices) > 1 {
			log.Println("Player tried to build too many times")
			sendMsg(p.ws, "Only architect can build more than once per round")
			return false
		}
	}

	// highest to lowest so we can remove from hand as we go later
	sort.Slice(choices, func(i, j int) bool {
		return choices[i] > choices[j]
	})

	var sum int
	for _, choice := range choices {
		chosenDistrict := p.hand[choice]
		if p.Gold < chosenDistrict.Value {
			sendMsg(p.ws, "You can't afford a district")
			return false
		}
		for _, district := range p.Districts {
			if district.Name == chosenDistrict.Name {
				log.Println("Duplicates!", district.Name, chosenDistrict.Name)
				sendMsg(p.ws, "Can't have duplicate districts")
				return false
			}
		}
		sum += chosenDistrict.Value
	}

	if sum > p.Gold {
		sendMsg(p.ws, "You can't afford all of these")
		return false
	}

	for _, choice := range choices {
		chosenDistrict := p.hand[choice]
		log.Printf("Player built %v (%v %v)", chosenDistrict.Name, chosenDistrict.Value, chosenDistrict.Color)
		p.Gold -= chosenDistrict.Value
		p.Districts = append(p.Districts, p.hand[choice])
		p.hand = append(p.hand[:choice], p.hand[choice+1:]...)
	}

	if len(p.Districts) >= 8 && c.FirstToEight == -1 {
		log.Println("Player", c.Turn.Value, "is first to 8 districts")
		c.FirstToEight = c.Turn.Value
	}

	c.State = endTurn
	return true
}

func (c *Citadels) handleEndTurn(cmd *wg.Command) bool {
	p, _ := Find(c.Players, cmd.PlayerId)
	if p != c.Players[c.Turn.Value] {
		sendMsg(p.ws, "Not your turn yet")
		return false
	}

	log.Println("Player ended turn")

	// next player's turn?
	for c.CharCur += 1; c.CharCur < 8; c.CharCur++ {
		if c.characters[c.CharCur].player != nil {
			_, c.Turn.Value = Find(c.Players, c.characters[c.CharCur].player.Uuid)
			if c.CharCur == 4 && c.characters[4].Character == King {
				c.crown.Value = c.Turn.Value
			}
			if c.Kill == c.CharCur {
				continue
			}
			c.State = goldOrDraw
			return true
		}
	}

	// end of round, check for win condition and winner

	for _, p := range c.Players {
		p.score = 0
		colors := [5]int{}
		for _, card := range p.Districts {
			p.score += card.Value
			colors[int(card.Color)]++
		}
		allColors := true
		for _, c := range colors {
			if c == 0 {
				allColors = false
				break
			}
		}
		if allColors {
			p.score += 3
		}
		if c.FirstToEight == p.Id {
			p.score += 4
		}
		if len(p.Districts) >= 8 {
			c.State = gameOver
		}
	}

	// no one won yet
	if c.State != gameOver {
		c.State = choose
		// if King moved the crown this turn, don't do this
		if !(c.characters[3].Character == King && c.characters[3].Chosen) {
			c.crown.Inc()
			c.Turn.Value = c.crown.Value
		}
		c.Turn.Max = len(c.Players)
		c.Kill = -1
		for _, char := range c.characters {
			char.Chosen = false
			char.player = nil
		}
		// TODO: this is two player variant only
		if len(c.Players) == 2 {
			c.characters[rand.Intn(8)].Chosen = true
		}
	}

	return true
}

func (c *Citadels) handleSpecial(cmd *wg.Command) bool {
	p, _ := Find(c.Players, cmd.PlayerId)
	if p != c.Players[c.Turn.Value] {
		sendMsg(p.ws, "Not your turn yet")
		return false
	}

	var choice int
	if err := json.Unmarshal(cmd.Data, &choice); err != nil {
		log.Println(err)
		sendMsg(p.ws, "Couldn't decode choice")
		return false
	}

	character := c.characters[c.CharCur]

	if choice == 0 {
		return character.special(c, p, cmd.Data)
	}

	if character.CanTax != None {
		for _, district := range p.hand {
			if district.Color == character.CanTax {
				p.Gold++
			}
		}
		return true
	}

	return false
}

func (c *Citadels) handleReady(cmd *wg.Command) bool {
	if c.State != gameOver {
		return false
	}
	allReady := true
	for _, p := range c.Players {
		if p.Uuid == cmd.PlayerId {
			p.IsReady = !p.IsReady
		}
		if !p.IsBot && !p.IsReady {
			allReady = false
		}
	}
	if allReady {
		c.Version += 1
		c.reset()
	}
	return true
}

// This info is sent once at the beginning of each game (and when users reconnect)
type GameMsg struct {
	Type       string
	Characters []*Character
}

func (c *Citadels) sendGameInfo(ws wg.Connector) {
	var characters []*Character

	for _, char := range c.characters {
		characters = append(characters, char.Character)
	}

	msg := &GameMsg{
		Type:       "info",
		Characters: characters,
	}

	ws.Send(msg)
}

type UpdateMsg struct {
	Type   string
	Update *Citadels
	You    *secret
}

type secret struct {
	Id        int
	Turn      bool                  `json:",omitempty"`
	HasCrown  bool                  `json:",omitempty"`
	Roles     []*ChoosableCharacter `json:",omitempty"`
	Character *ChoosableCharacter   `json:",omitempty"`
	Hand      []*District
}

func (c *Citadels) sendEveryoneEverything() {
	for i, p := range c.Players {
		if p.ws != nil {
			msg := &UpdateMsg{Type: "all", Update: c}
			msg.You = &secret{
				Id:       p.Id,
				HasCrown: p.HasCrown,
				Hand:     p.hand,
			}
			if c.Turn.Value == i {
				msg.You.Turn = true
				if c.State > choose && c.State < gameOver {
					msg.You.Character = c.characters[c.CharCur]
				} else if c.State == choose {
					msg.You.Roles = c.characters
				}
			}
			p.ws.Send(msg)
		}
	}
}
