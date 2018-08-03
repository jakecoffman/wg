package citadels

import "encoding/json"

type Character struct {
	Name string
	CanTax Color
	special func(c *Citadels, player *Player, data json.RawMessage) bool
}

var Assassin = &Character{
	"Assassin",
	None,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		var choice int
		if err := json.Unmarshal(data, &choice); err != nil {
			sendMsg(player.ws, "Couldn't unmarshal choice")
			return false
		}
		if choice < 0 || choice > 8 {
			sendMsg(player.ws, "Invalid assassination")
			return false
		}
		c.Kill = choice
		return true
	},
}

var Thief = &Character{
	"Thief",
	None,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		var choice int
		if err := json.Unmarshal(data, &choice); err != nil {
			sendMsg(player.ws, "Couldn't unmarshal choice")
			return false
		}
		if choice < 2 || choice > 8 || choice == c.Kill {
			sendMsg(player.ws, "Cannot steal from assassin or assassin's target")
			return false
		}
		return true
	},
}

var Magician = &Character{
	"Magician",
	None,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		var choice struct {
			Swap   *int
			Redraw []int
		}
		if err := json.Unmarshal(data, &choice); err != nil {
			sendMsg(player.ws, "Couldn't unmarshal choice")
			return false
		}
		if choice.Swap != nil {
			value := *choice.Swap
			if value < 0 || value == 2 || value > 8 {
				sendMsg(player.ws, "Invalid card swap target")
				return false
			}
			c.Players[value].hand, c.Players[2].hand = c.Players[2].hand, c.Players[value].hand
			return true
		}
		if len(choice.Redraw) > 0 {
			var validIndices []*District
			for i := 0; i < len(choice.Redraw); i++ {
				if choice.Redraw[i] > 0 && choice.Redraw[i] < len(c.Players[2].hand) {
					validIndices = append(validIndices, player.hand[choice.Redraw[i]])
				} else {
					sendMsg(player.ws, "Invalid redraw target")
					return false
				}
			}
			// discard and redraw these indices
			for i := 0; i < len(validIndices); i++ {
				validIndices[i] = c.districtDeck[0]
				c.districtDeck = c.districtDeck[1:]
			}
		}
		sendMsg(player.ws, "Magician didn't do special power?")
		return false
	},
}

var King = &Character{
	"King",
	Yellow,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		// King automatically receives crown
		return false
	},
}

var Bishop = &Character{
	"Bishop",
	Blue,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		// Bishop is immune to warlord
		return false
	},
}

var Merchant = &Character{
	"Merchant",
	Green,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		// merchant's additional gold is added in the action phase
		return false
	},
}

var Architect = &Character{
	"Architect",
	None,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		if c.State != build {
			sendMsg(player.ws, "Must use power after action phase")
			return false
		}
		player.hand = append(player.hand, c.districtDeck[:2]...)
		c.districtDeck = c.districtDeck[2:]
		return true
	},
}

var Warlord = &Character{
	"Warlord",
	Red,
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		var choice struct {
			Swap   *int
			Redraw []int
		}
		if err := json.Unmarshal(data, &choice); err != nil {
			sendMsg(player.ws, "Couldn't unmarshal choice")
			return false
		}
		// TODO destroy
		return true
	},
}
