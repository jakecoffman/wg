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

func TestCookieMiddleware_NoCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	CookieMiddleware(&testHandler{})(w, r)

	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("cookie should have been set", w.Header())
	}
	if len(cookie) != 50 {
		t.Error(cookie)
	}
}

func TestCookieMiddleware_AlreadyCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: COOKIE_NAME, Value: "BOOP"})

	CookieMiddleware(&testHandler{})(w, r)

	cookie := w.Header().Get("Set-Cookie")
	if cookie != "" {
		t.Fatal("new cookie should not have been set", w.Header())
	}
}

type fakeConn struct {
	Connector
	req *http.Request
	wasClosed bool
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
}
