package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Msg struct {
	MsgText string `json:"msg_text"`
}

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(w, r, "foo")
	})
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(w, r, "bar")
	})
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func serveWebsocket(w http.ResponseWriter, r *http.Request, answer string) {
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
		msg := Msg{MsgText: "Answer from " + answer}
		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
	}

}
