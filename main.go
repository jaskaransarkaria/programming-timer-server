package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"log"
	"flag"
	"github.com/gorilla/websocket"
	"math/rand"
	"encoding/hex"
	"errors"
)

// User is ...
type user struct {
	UUID string
}

// Session is ...
type session struct {
	SessionID string
	Duration int64
	StartTime int64
	EndTime int64
	Users []user
}

// StartTimer ... JSON response from the client
type startTimerReq struct {
	Duration int64 `json:"duration"`
	StartTime int64 `json:"startTime"`
}

var sessions []session

// flag allows you to create cli flags and assign a default
var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{
	// empty struct means use defaults
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func (session *session) AddUser(user user) []user {
    session.Users = append(session.Users, user)
    return session.Users
	}

func getIDLength(typeOfID string) (int8, error) {
		if (typeOfID == "session") {
		return 2, nil // equals 4 characters long
	} 
	if (typeOfID == "user") {
		return 4, nil // equals 8 characters long
	}
	return -1, errors.New("Invalid typeofID as parameter")
}

func generateRandomID(typeOfID string) string {
	length, err := getIDLength(typeOfID)
		if err != nil {
			log.Println("err generating ID", err)
		}
	b := make([]byte, length)
	rand.Read(b) 
	s := hex.EncodeToString(b)
	return s
}

func createNewUserAndSession(newSessionData startTimerReq) session {
	var newUser = user{ UUID: generateRandomID("user") }
	var newSession = session{
				SessionID: generateRandomID("session"),
				Duration: newSessionData.Duration,
				StartTime: newSessionData.StartTime,
				EndTime: newSessionData.Duration + newSessionData.StartTime,
			}
	newSession.AddUser(newUser)
	sessions = append(sessions, newSession)
	return newSession
}

func writer(conn *websocket.Conn, messageType int, message []byte) {
	// message the client
	if err := conn.WriteMessage(messageType, message); err != nil {
		log.Println(err)
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
			var startTimerData startTimerReq
			err = json.Unmarshal(p, &startTimerData)
			if err != nil {
				writer(conn, messageType, []byte("well done you've connected via web sockets to a go server"))
			}

		if (startTimerData.Duration != 0 && startTimerData.StartTime != 0) {
			newSession := createNewUserAndSession(startTimerData)
			newSessionRes, _ := json.Marshal(newSession)
			writer(conn, messageType, newSessionRes)
			log.Println("SESSION RESPONSE", string(newSessionRes))
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
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Golang WebSockets running...")
	setupRoutes()
	flag.Parse()
	fmt.Println("Listening on:", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
