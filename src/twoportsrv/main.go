package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Msg struct {
	MsgText string `json:"msg_text"`
	Url     string `json:"url"`
}

const ServerAddress = "127.0.0.1"
const PortAdmin = 8081
const PortUser = 8082

func main() {
	go startWebServer(PortAdmin)
	startWebServer(PortUser)
}

func startWebServer(port int) {
	mux := http.NewServeMux()
	host := ServerAddress + ":" + fmt.Sprint(port)
	mux.HandleFunc("/", serveWebsocket)
	log.Print("Start listening at " + host)
	http.ListenAndServe(host, mux)
}

func serveWebsocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgrade:", err)
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
