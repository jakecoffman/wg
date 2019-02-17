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
               "sha256_cert_fingerprints": ["F1:68:01:70:C3:51:03:78:53:95:4C:24:FA:AC:A8:2B:79:65:F4:BD:B2:AF:A5:34:85:62:7E:FB:46:4F:A5:84"] }
}]`))
	})
	port := "8112"
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
