package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Msg struct {
	MsgText string `json:"msg_text"`
	Url     string `json:"url"`
}

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(w, r)
	})
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(w, r)
	})
	log.Fatal(http.ListenAndServe("localhost:9000", nil))
}

func serveWebsocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received: %s", message)
		msg := Msg{MsgText: string(message), Url: conn.RemoteAddr().String()}
		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
	}

}
