package wg

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

// Connector wraps connections so tests are easier
type Connector interface {
	Send(v interface{})
	Recv(v interface{}) error

	SendRaw(v []byte)
	RecvRaw(v []byte) error

	Close() error

	Request() *http.Request
}

// WsConn is a websocket connection that implements Connector
type wsConn struct {
	conn *websocket.Conn
}

func NewWsConn(ws *websocket.Conn) *wsConn {
	conn := &wsConn{
		conn: ws,
	}
	return conn
}

func (c *wsConn) Send(v interface{}) {
	if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Println(err)
		return
	}
	if err := websocket.JSON.Send(c.conn, v); err != nil {
		log.Println(err)
		return
	}
}

func (c *wsConn) SendRaw(v []byte) {
	if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Println(err)
		return
	}
	if err := websocket.Message.Send(c.conn, v); err != nil {
		log.Println(err)
		return
	}
}

func (c *wsConn) Recv(v interface{}) error {
	if err := c.conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
		return err
	}
	return websocket.JSON.Receive(c.conn, v)
}

func (c *wsConn) RecvRaw(v []byte) error {
	if err := c.conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
		return err
	}
	return websocket.Message.Receive(c.conn, v)
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) Request() *http.Request {
	return c.conn.Request()
}
