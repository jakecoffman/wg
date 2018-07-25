package citadels

type Color int

const (
	Green = Color(iota) // Trade
	Blue                // Religious
	Red                 // Military
	Yellow              // Noble
	Purple              // Special
)

type District struct {
	Name string
	Value int
	Color Color
}

var Districts []*District

func init() {
	Districts = append(Districts, &District{"Tavern", 1, Green})
	Districts = append(Districts, &District{"Tavern", 1, Green})
	Districts = append(Districts, &District{"Tavern", 1, Green})
	Districts = append(Districts, &District{"Tavern", 1, Green})
	Districts = append(Districts, &District{"Tavern", 1, Green})

	Districts = append(Districts, &District{"Market", 2, Green})
	Districts = append(Districts, &District{"Market", 2, Green})
	Districts = append(Districts, &District{"Market", 2, Green})
	Districts = append(Districts, &District{"Market", 2, Green})

	Districts = append(Districts, &District{"Trading Post", 2, Green})
	Districts = append(Districts, &District{"Trading Post", 2, Green})
	Districts = append(Districts, &District{"Trading Post", 2, Green})

	Districts = append(Districts, &District{"Docks", 3, Green})
	Districts = append(Districts, &District{"Docks", 3, Green})
	Districts = append(Districts, &District{"Docks", 3, Green})

	Districts = append(Districts, &District{"Harbor", 4, Green})
	Districts = append(Districts, &District{"Harbor", 4, Green})
	Districts = append(Districts, &District{"Harbor", 4, Green})

	Districts = append(Districts, &District{"Town Hall", 5, Green})
	Districts = append(Districts, &District{"Town Hall", 5, Green})

	Districts = append(Districts, &District{"Temple", 1, Blue})
	Districts = append(Districts, &District{"Temple", 1, Blue})
	Districts = append(Districts, &District{"Temple", 1, Blue})

	Districts = append(Districts, &District{"Church", 2, Blue})
	Districts = append(Districts, &District{"Church", 2, Blue})
	Districts = append(Districts, &District{"Church", 2, Blue})

	Districts = append(Districts, &District{"Monastery", 3, Blue})
	Districts = append(Districts, &District{"Monastery", 3, Blue})
	Districts = append(Districts, &District{"Monastery", 3, Blue})

	Districts = append(Districts, &District{"Cathedral", 5, Blue})
	Districts = append(Districts, &District{"Cathedral", 5, Blue})

	Districts = append(Districts, &District{"Watchtower", 1, Red})
	Districts = append(Districts, &District{"Watchtower", 1, Red})
	Districts = append(Districts, &District{"Watchtower", 1, Red})

	Districts = append(Districts, &District{"Prison", 2, Red})
	Districts = append(Districts, &District{"Prison", 2, Red})
	Districts = append(Districts, &District{"Prison", 2, Red})

	Districts = append(Districts, &District{"Battlefield", 3, Red})
	Districts = append(Districts, &District{"Battlefield", 3, Red})
	Districts = append(Districts, &District{"Battlefield", 3, Red})

	Districts = append(Districts, &District{"Fortress", 5, Red})
	Districts = append(Districts, &District{"Fortress", 5, Red})

	Districts = append(Districts, &District{"Manor", 3, Yellow})
	Districts = append(Districts, &District{"Manor", 3, Yellow})
	Districts = append(Districts, &District{"Manor", 3, Yellow})
	Districts = append(Districts, &District{"Manor", 3, Yellow})
	Districts = append(Districts, &District{"Manor", 3, Yellow})

	Districts = append(Districts, &District{"Castle", 4, Yellow})
	Districts = append(Districts, &District{"Castle", 4, Yellow})
	Districts = append(Districts, &District{"Castle", 4, Yellow})
	Districts = append(Districts, &District{"Castle", 4, Yellow})

	Districts = append(Districts, &District{"Palace", 5, Yellow})
	Districts = append(Districts, &District{"Palace", 5, Yellow})
	Districts = append(Districts, &District{"Palace", 5, Yellow})

	Districts = append(Districts, &District{"Haunted City", 2, Purple})

	Districts = append(Districts, &District{"Keep", 3, Purple})
	Districts = append(Districts, &District{"Keep", 3, Purple})
	Districts = append(Districts, &District{"Keep", 3, Purple})

	Districts = append(Districts, &District{"Laboratory", 5, Purple})
	Districts = append(Districts, &District{"Smithy", 5, Purple})
	Districts = append(Districts, &District{"Graveyard", 5, Purple})
	Districts = append(Districts, &District{"Observatory", 5, Purple})
	Districts = append(Districts, &District{"School of Magic", 6, Purple})
	Districts = append(Districts, &District{"Library", 6, Purple})
	Districts = append(Districts, &District{"Great Wall", 6, Purple})
	Districts = append(Districts, &District{"University", 8, Purple})
	Districts = append(Districts, &District{"Dragon Gate", 8, Purple})
}
