package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"log"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// StartTimer ... JSON response from the client
type StartTimer struct {
	UUID string `json:"uuid"`
	Duration int64 `json:"duration"`
	StartTime int64 `json:"startTime"`
}

// flag allows you to create cli flags and assign a default
var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{
	// empty struct means use defaults
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}


func homeRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "home route")
	
}

func writer(conn *websocket.Conn, messageType int, message []byte) {
	// message the client
	if err := conn.WriteMessage(messageType, message); err != nil {
		log.Println(err)
			return
		}
}

func reader(conn *websocket.Conn) {
	// listen on this connection for new messages and send messages down that connection
	for {
			messageType, p, err := conn.ReadMessage()
			log.Println(string(p))
			if err != nil {
				log.Println(err)
			}

			var startTimerData StartTimer
			err = json.Unmarshal(p, &startTimerData)
			if err != nil {
				id, err := uuid.NewUUID()
				if err != nil {
					log.Println("error from new uuid")
				}
				log.Println(id.String())
				writer(conn, messageType, []byte("well done you've connected via web sockets to a go server"))
				writer(conn, messageType, []byte(id.String()))
			}
			log.Println("JSON recieved")
			log.Println(startTimerData.UUID)
			log.Println(startTimerData.Duration)
			log.Println(startTimerData.StartTime)
		}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// this is for CORS -  allow all origin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade http connection to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client successfully connected to Golang Websocket!")
	// either read json or read message
	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/", homeRoute)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Golang WebSockets running...")
	setupRoutes()
	flag.Parse()
	fmt.Println("Listening on:", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
