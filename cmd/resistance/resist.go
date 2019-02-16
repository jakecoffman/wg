package main

import (
	"github.com/jakecoffman/wg"
	"github.com/jakecoffman/wg/resistance"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
	"math/rand"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.Handle("/ws", websocket.Handler(wg.WsHandler(wg.ProcessPlayerCommands(resistance.NewGame))))
	http.HandleFunc("/.well-known/assetlinks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
  "relation": ["delegate_permission/common.handle_all_urls"],
  "target" : { "namespace": "android_app", "package_name": "com.jakecoffman.spytown",
               "sha256_cert_fingerprints": ["B4:9D:C1:38:30:3E:32:5E:1E:25:72:5E:0A:87:B9:D4:F7:49:14:BC:DA:C3:E9:E4:AC:1F:15:A6:20:4C:3E:A7"] }
}]`))
	})
	port := "8112"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
