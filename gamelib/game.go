package gamelib

type Game interface {
	Cmd(Command)
}

type Command interface {
	IsValid() bool
}

