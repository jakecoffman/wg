package gamelib

import (
	"golang.org/x/net/websocket"
	"net/http"
	"time"
)

// Connector wraps connections so tests are easier
type Connector interface {
	Send(v interface{}) error
	Recv(v interface{}) error
	Close() error

	Request() *http.Request
}

// WsConn is a websocket connection that implements Connector
type WsConn struct {
	conn *websocket.Conn
}

func (c *WsConn) Send(v interface{}) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}
	return websocket.JSON.Send(c.conn, v)
}

func (c *WsConn) Recv(v interface{}) error {
	if err := c.conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
		return err
	}
	return websocket.JSON.Receive(c.conn, v)
}

func (c *WsConn) Close() error {
	return c.conn.Close()
}

func (c *WsConn) Request() *http.Request {
	return c.conn.Request()
}
