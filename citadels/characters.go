package citadels

func Assassin(c *Citadels, player *Player, choice int) bool {
	if choice < 0 || choice > 8 {
		sendMsg(player.ws, "Invalid assasination")
		return false
	}
	c.Kill = choice
	return true
}

func Thief(c *Citadels, player *Player, choice int) bool {
	return true
}

func Magician(c *Citadels, player *Player, choice int) bool {
	return true
}

func King(c *Citadels, player *Player, choice int) bool {
	c.Crown.Value = player.Id
	return true
}

func Bishop(c *Citadels, player *Player, choice int) bool {
	return true
}

func Merchant(c *Citadels, player *Player, choice int) bool {
	return true
}

func Architect(c *Citadels, player *Player, choice int) bool {
	return true
}

func Warlord(c *Citadels, player *Player, choice int) bool {
	return true
}

var Characters []func(*Citadels, *Player, int) bool

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
