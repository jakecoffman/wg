package gamelib

import (
	"testing"
	"net/http"
	"net/http/httptest"
)

type testHandler struct {}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

type fakeConn struct {
	Connector
	req *http.Request
	wasClosed bool
	sentMsg interface{}
}

func (c *fakeConn) Send(msg interface{}) {
	c.sentMsg = msg
}

func (c *fakeConn) Close() error {
	c.wasClosed = true
	return nil
}

func (c *fakeConn) Request() *http.Request {
	return c.req
}

func TestWsHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: COOKIE_NAME, Value: "BOOP"})

	var calledId string
	conn := &fakeConn{req: r}
	connHandler(func(c Connector, id string) {
		calledId = id
	}, conn)

	if calledId != "BOOP" {
		t.Error("Expected BOOP got", calledId)
	}
	if !conn.wasClosed {
		t.Error("Connection should have been closed")
	}
}

func TestWsHandler_NoCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)

	var calledId string
	conn := &fakeConn{req: r}
	connHandler(func(c Connector, id string) {
		calledId = id
	}, conn)

	if len(calledId) == 50 {
		t.Error("Expected a UUID got", calledId)
	}
	if !conn.wasClosed {
		t.Error("Connection should have been closed")
	}
	c := conn.sentMsg.(*cookieMsg)
	if c.Type != "cookie" && len(c.Cookie) != 8 {
		t.Error("Cookie not sent", c.Type, c.Cookie)
	}
}
