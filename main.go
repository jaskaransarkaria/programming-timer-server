package main

import (
	"fmt"
	"errors"
	"math"
	"net/http"
	"log"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// empty struct means use defaults
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}


func homeRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "home route")
	
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

		// echo message back to the client
		if err := conn.WriteMessage(messageType, []byte("This is a response through the websocket connection from the server")); err != nil {
			log.Println(err)
			return
		}
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
	log.Fatal(http.ListenAndServe(":8080", nil))

	result, err := sqrt(16)
	if err != nil {
		fmt.Println(nil)
	} else {
		fmt.Println(result)
	}
}

func sqrt(x float64) (float64, error) {
	if x < 0 {
		return  0, errors.New("Undefined for negative numbers")
	}

	return math.Sqrt(x), nil
}
