package citadels

import "encoding/json"

type Character struct {
	Name string
	Play func(c *Citadels, player *Player, data json.RawMessage) bool
}

func (c Character) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

var Assassin = &Character{
	"Assassin",
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
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		_, c.crown.Value = Find(c.Players, player.Uuid)
		for _, card := range player.hand {
			if card.Color == Yellow {
				player.Gold++
			}
		}
		return true
	},
}

var Bishop = &Character{
	"Bishop",
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		for _, card := range player.hand {
			if card.Color == Blue {
				player.Gold++
			}
		}
		return true
	},
}

var Merchant = &Character{
	"Merchant",
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		// merchant's additional gold is added in the action phase
		for _, card := range player.hand {
			if card.Color == Green {
				player.Gold++
			}
		}
		return true
	},
}

var Architect = &Character{
	"Architect",
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		// architect gets additional district cards in the action phase
		// and can build up to three districts
		return true
	},
}

var Warlord = &Character{
	"Warlord",
	func(c *Citadels, player *Player, data json.RawMessage) bool {
		var choice struct {
			Swap   *int
			Redraw []int
		}
		if err := json.Unmarshal(data, &choice); err != nil {
			sendMsg(player.ws, "Couldn't unmarshal choice")
			return false
		}

		for _, card := range player.hand {
			if card.Color == Red {
				player.Gold++
			}
		}
		return true
	},
}
