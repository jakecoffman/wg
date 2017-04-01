package gamelib

import (
	"golang.org/x/net/websocket"
	"net/http"
	"github.com/google/uuid"
	"time"
	"log"
)

type PlayerCommandHandler func(Connector, string)

const COOKIE_NAME = "PLAYER_COOKIE"

// WsHandler handles player web connections
func WsHandler(cmdHandler PlayerCommandHandler) websocket.Handler {
	return func(ws *websocket.Conn) {
		connHandler(cmdHandler, NewWsConn(ws))
	}
}

type cookieMsg struct {
	Type, Cookie string
}

// testable!
func connHandler(cmdHandler PlayerCommandHandler, ws Connector) {
	defer ws.Close()

	var playerId string
	cookie, err := ws.Request().Cookie(COOKIE_NAME)
	if err == http.ErrNoCookie {
		playerId = uuid.New().String()
		c := http.Cookie{Name: COOKIE_NAME, Value: playerId, Expires: time.Now().Add(24 * 365 * 20 * time.Hour)}
		ws.Send(&cookieMsg{Type: "cookie", Cookie: c.String()})
		log.Println("New player connected", c.String())
	} else {
		playerId = cookie.Value
		log.Println("Player returned", cookie.Value)
	}

	cmdHandler(ws, playerId)
}
