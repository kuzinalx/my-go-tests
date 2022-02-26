package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var interrupt chan os.Signal

type MessageInfo struct {
	Action string `json:"action"`
}

type FrameData struct {
	ImgOrig string `json:"img_orig"`
	ImgProj string `json:"img_proj"`
	Stamp   int    `json:"stamp"`
}

type MessageFrameData struct {
	MessageInfo
	FrameData FrameData `json:"body"`
}

type FrameDataDecoded struct {
	ImgOrig []byte
	ImgProj []byte
	Stamp   int
	mutex   sync.Mutex
}

var frameDataDecoded FrameDataDecoded

func main() {
	interrupt = make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	go startWebsocketClient()
	//<-interrupt
	//log.Println("Received SIGINT interrupt signal.")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mpw := multipart.NewWriter(w)
		w.Header().Add("Content-Type",
			fmt.Sprintf("multipart/x-mixed-replace; boundary=\"%v\"", mpw.Boundary()))
		//w.Header().Add("Content-Type", mpw.FormDataContentType())
		w.WriteHeader(http.StatusOK)

		partHdr := textproto.MIMEHeader(map[string][]string{})
		partHdr.Add("Content-Disposition", "inline")

		func() {
			frameDataDecoded.mutex.Lock()
			defer frameDataDecoded.mutex.Unlock()
			if frameDataDecoded.ImgOrig == nil {
				return
			}
			partHdr.Add("Content-Type", "image/jpeg")
			part, err := mpw.CreatePart(partHdr)
			if err != nil {
				fmt.Println(err)
				return
			}
			part.Write(frameDataDecoded.ImgOrig)
		}()

		mpw.Close()
	})
	http.ListenAndServe("localhost:8000", nil)
}

func startWebsocketClient() {
	socketUrl := "ws://localhost:8554" + "/frames"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go readWebsocketAnswers(conn)
	counter := 1

	for {
		<-time.After(time.Duration(1) * time.Millisecond * 1000)
		messageInfo := MessageInfo{Action: "frame"}
		err = conn.WriteJSON(messageInfo)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("%v : Sent message to %v\n", counter, socketUrl)
		counter++
	}
}

func readWebsocketAnswers(conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Print(err)
			continue
		}
		var mi MessageInfo
		err = json.Unmarshal(data, &mi)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("Received answer: '%v'\n", mi.Action)
		if mi.Action != "frame" {
			continue
		}
		var mfd MessageFrameData
		err = json.Unmarshal(data, &mfd)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Println("Message data was marshalled.")
		func() {
			frameDataDecoded.mutex.Lock()
			defer frameDataDecoded.mutex.Unlock()
			frameDataDecoded.Stamp = mfd.FrameData.Stamp
			frameDataDecoded.ImgOrig, err =
				base64.StdEncoding.DecodeString(mfd.FrameData.ImgOrig)
			if err != nil {
				log.Println(err)
			}
			frameDataDecoded.ImgProj, err =
				base64.StdEncoding.DecodeString(mfd.FrameData.ImgProj)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}
