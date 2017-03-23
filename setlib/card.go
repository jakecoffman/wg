package setlib

import (
	"time"
	"math/rand"
)

type Card struct {
	Shape   string `json:"s"`
	Pattern string `json:"p"`
	Color   string `json:"c"`
	Amount  int    `json:"a"`
}

var deck []Card

func init() {
	rand.Seed(time.Now().Unix())

	shapes := []string{"p", "n", "d"}
	patterns := []string{"h", "s", "z"}
	colors := []string{"r", "p", "g"}
	amount := []int{1, 2, 3}

	deck = []Card{}
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
	if card1.Amount == -1 || card2.Amount == -1 || card3.Amount == -1 {
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
