package justone

import (
	_ "embed"
	"strings"
)

//go:embed words.txt
var words string
var wordlist []string

func init() {
	wordlist = strings.Split(words, "\n")
}
