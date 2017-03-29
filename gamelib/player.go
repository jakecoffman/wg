package gamelib

import (
	"math/rand"
	"time"
)

const letterBytes = "1234567890"

func init() {
	rand.Seed(time.Now().Unix())
}

func GenId() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
