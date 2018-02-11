package setlib

type Card struct {
	Shape   string `json:"s"`
	Pattern string `json:"p"`
	Color   string `json:"c"`
	Amount  int    `json:"a"`
}

var deck []Card
var shapes = []string{"p", "n", "d"}
var patterns = []string{"h", "s", "z"}
var colors = []string{"r", "p", "g"}
var amount = []int{1, 2, 3}

func init() {
	for _, s := range shapes {
		for _, p := range patterns {
			for _, c := range colors {
				for _, a := range amount {
					deck = append(deck, Card{Shape: s, Pattern: p, Color: c, Amount: a})
				}
			}
		}
	}
}

func isSet(card1, card2, card3 Card) bool {
	if any(-1, card1.Amount, card2.Amount, card3.Amount) {
		return false
	}
	if !(same(card1.Shape, card2.Shape, card3.Shape) || different(card1.Shape, card2.Shape, card3.Shape)) {
		return false
	}
	if !(same(card1.Pattern, card2.Pattern, card3.Pattern) || different(card1.Pattern, card2.Pattern, card3.Pattern)) {
		return false
	}
	if !(sameInt(card1.Amount, card2.Amount, card3.Amount) || differentInt(card1.Amount, card2.Amount, card3.Amount)) {
		return false
	}
	if !(same(card1.Color, card2.Color, card3.Color) || different(card1.Color, card2.Color, card3.Color)) {
		return false
	}
	return true
}

func same(s1, s2, s3 string) bool {
	return s1 == s2 && s2 == s3 && s3 != ""
}

func different(s1, s2, s3 string) bool {
	return s1 != s2 && s2 != s3 && s3 != s1
}

func sameInt(s1, s2, s3 int) bool {
	return s1 == s2 && s2 == s3
}

func differentInt(s1, s2, s3 int) bool {
	return s1 != s2 && s2 != s3 && s3 != s1
}

func any(are int, stuff... int) bool {
	for i := 0; i < len(stuff); i++ {
		if stuff[i] == are {
			return true
		}
	}
	return false
}
