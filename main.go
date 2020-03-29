package main

import (
	"fmt"
	"net/http"
	"log"
	"flag"
	"github.com/gorilla/websocket"
)

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
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(p))

		//message back to the client
		writer(conn, messageType, []byte("This is a response through the websocket connection from the server"))
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
