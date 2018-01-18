package wg

import (
	"github.com/google/uuid"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
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
		log.Println("New player connected", playerId, ws.Request().Header.Get("X-Forwarded-For"))
	} else {
		playerId = cookie.Value
		log.Println("Player returned", playerId, ws.Request().Header.Get("X-Forwarded-For"))
	}
	c := http.Cookie{Name: COOKIE_NAME, Value: playerId, Expires: time.Now().Add(24 * 365 * time.Hour)}
	ws.Send(&cookieMsg{Type: "cookie", Cookie: c.String()})

	cmdHandler(ws, playerId)
}
