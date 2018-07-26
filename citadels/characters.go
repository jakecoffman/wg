package citadels

import "encoding/json"

func Assassin(c *Citadels, player *Player, data json.RawMessage) bool {
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
}

func Thief(c *Citadels, player *Player, data json.RawMessage) bool {
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
}

func Magician(c *Citadels, player *Player, data json.RawMessage) bool {
	var choice struct {
		Swap *int
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
}

func King(c *Citadels, player *Player, data json.RawMessage) bool {
	_, c.crown.Value = Find(c.Players, player.Uuid)
	for _, card := range player.hand {
		if card.Color == Yellow {
			player.Gold++
		}
	}
	return true
}

func Bishop(c *Citadels, player *Player, data json.RawMessage) bool {
	for _, card := range player.hand {
		if card.Color == Blue {
			player.Gold++
		}
	}
	return true
}

func Merchant(c *Citadels, player *Player, data json.RawMessage) bool {
	// merchant's additional gold is added in the action phase
	for _, card := range player.hand {
		if card.Color == Green {
			player.Gold++
		}
	}
	return true
}

func Architect(c *Citadels, player *Player, data json.RawMessage) bool {
	// architect gets additional district cards in the action phase
	// and can build up to three districts
	return true
}

func Warlord(c *Citadels, player *Player, data json.RawMessage) bool {
	var choice struct {
		Swap *int
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
}

var Characters []func(*Citadels, *Player, json.RawMessage) bool

func init() {
	Characters = append(Characters, Assassin)
	Characters = append(Characters, Thief)
	Characters = append(Characters, Magician)
	Characters = append(Characters, King)
	Characters = append(Characters, Bishop)
	Characters = append(Characters, Merchant)
	Characters = append(Characters, Architect)
	Characters = append(Characters, Warlord)
}
