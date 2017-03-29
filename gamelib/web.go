package gamelib

import (
	"golang.org/x/net/websocket"
	"net/http"
	"github.com/google/uuid"
)

type PlayerCommandHandler func(Connector, string)

const COOKIE_NAME = "PLAYER_COOKIE"

// WsHandler handles player web connections
func WsHandler(cmdHandler PlayerCommandHandler) websocket.Handler {
	return func(ws *websocket.Conn) {
		connHandler(cmdHandler, NewWsConn(ws))
	}
}

// testable!
func connHandler(cmdHandler PlayerCommandHandler, ws Connector) {
	defer ws.Close()

	var playerId string
	cookie, err := ws.Request().Cookie(COOKIE_NAME)
	if err == http.ErrNoCookie {
		playerId = uuid.New().String()
		ws.Send(struct{Type, Cookie string}{Type: "cookie", Cookie: COOKIE_NAME+"="+playerId})
	} else {
		playerId = cookie.Value
	}

	cmdHandler(ws, playerId)
}

// CookieMiddleware just makes sure every user gets a cookie
func CookieMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(COOKIE_NAME)
		if err == http.ErrNoCookie {
			cookie = &http.Cookie{Name: COOKIE_NAME, Value: uuid.New().String()}
			http.SetCookie(w, cookie)
		}
		handler.ServeHTTP(w, r)
	}
}
