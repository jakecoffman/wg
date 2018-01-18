package main

import (
	"github.com/jakecoffman/wg"
	"github.com/jakecoffman/wg/setlib"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

const (
	live_ws     = "wss://set.jakecoffman.com/ws"
	live_origin = "https://set.jakecoffman.com"
	test_ws     = "ws://127.0.0.1:8111/ws"
	test_origin = "http://127.0.0.1:8111"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.Handle("/ws", websocket.Handler(wg.WsHandler(wg.ProcessPlayerCommands(setlib.NewGame))))
	go func() {
		log.Fatal(http.ListenAndServe(":8111", nil))
	}()

	log.Println("Dialing")
	//ws, err := websocket.Dial(test_ws, "", test_origin)
	ws, err := websocket.Dial(live_ws, "", live_origin)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	log.Println("Connected")
	incoming := make(chan *setlib.UpdateMsg)
	go read(ws, incoming)

	if err = websocket.JSON.Send(ws, map[string]interface{}{"type": "join", "Join": ""}); err != nil {
		log.Fatal(err)
	}

	log.Println("Sent message")

	msg := <-incoming

	log.Printf("%#v\n", msg)

	msg = <-incoming

	log.Printf("%#v\n", msg)

}

func read(ws *websocket.Conn, c chan *setlib.UpdateMsg) {
	for {
		var message *setlib.UpdateMsg
		err := websocket.JSON.Receive(ws, &message)
		if err != nil {
			log.Fatal(err)
		}
		c <- message
	}
}
