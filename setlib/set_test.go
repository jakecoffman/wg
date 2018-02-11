package setlib

import (
	"testing"
	"github.com/jakecoffman/wg"
	"golang.org/x/net/websocket"
	"net/http/httptest"
	"log"
)

type tester struct {
	ws wg.Connector
	cookie string
}

func TestSet(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	handler := websocket.Handler(wg.WsHandler(wg.ProcessPlayerCommands(NewGame)))
	server := httptest.NewServer(handler)
	defer server.Close()
	addr := server.Listener.Addr().String()
	conn, err := websocket.Dial("ws://" + addr, "", "http://"+addr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer conn.Close()

	ws1 := wg.NewWsConn(conn)
	{
		var cookie map[string]string
		recv(ws1, &cookie, t)
		if cookie["Type"] != "cookie" {
			t.Fatal("Type not cookie", cookie["Type"])
		}
	}

	ws1.Send(map[string]interface{}{"Type": "join", "Data": ""})

	{
		var u UpdateMsg
		recv(ws1, &u, t)
		if u.Type != "all" {
			t.Fatal("Expected all got", u.Type)
		}
	}
	{
		var meta MetaMsg
		recv(ws1, &meta, t)
		if meta.Type != "meta" {
			t.Fatal("Expected meta got", meta.Type)
		}
	}
}

func recv(ws wg.Connector, v interface{}, t *testing.T) {
	if err := ws.Recv(v); err != nil {
		t.Fatal(err.Error())
	}
}
