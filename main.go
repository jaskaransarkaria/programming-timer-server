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
type User struct {
	UUID string
}

// Session is ...
type Session struct {
	SessionID string
	Duration int64
	StartTime int64
	EndTime int64
	Users []User
}

// StartTimer ... JSON request from the client
type StartTimerReq struct {
	Duration int64 `json:"duration"`
	StartTime int64 `json:"startTime"`
}

type ExistingSessionReq struct {
	JoinSessionID string `json:"joinSession"`
}

var sessions []Session

// flag allows you to create cli flags and assign a default
var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{
	// empty struct means use defaults
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func (session *Session) AddUser(user User) []User {
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

func getExistingSession(desiredSessionID string) (Session, error) {
	// iterate through sessions
	var matchedSession Session
	for _, session :=  range sessions {
		if session.SessionID == desiredSessionID {
			matchedSession = session
			return matchedSession, nil
		}
	}
	return matchedSession, errors.New("There are no sessions with the id:" + desiredSessionID)
}

func createNewUserAndSession(newSessionData StartTimerReq) Session {
	var newUser = User{ UUID: generateRandomID("user") }
	var newSession = Session{
				SessionID: generateRandomID("session"),
				Duration: newSessionData.Duration,
				StartTime: newSessionData.StartTime,
				EndTime: newSessionData.Duration + newSessionData.StartTime,
			}
	newSession.AddUser(newUser)
	sessions = append(sessions, newSession)
	return newSession
}

func joinExistingSession(joinExistingSessionData ExistingSessionReq) (Session, error) {
		var newUser = User{ UUID: generateRandomID("user") }
		matchedSession, err := getExistingSession(joinExistingSessionData.JoinSessionID)
		if err != nil {
			return matchedSession, err
		}
		matchedSession.AddUser(newUser)
		return matchedSession, nil
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
			writer(conn, messageType, []byte("well done you've connected via web sockets to a go server"))
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

func enableCors(w *http.ResponseWriter) {(*w).Header().Set("Access-Control-Allow-Origin", "*")}

func newSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var timerRequest StartTimerReq
	var requestBody = r.Body
	enableCors(&w)
	
	err := json.NewDecoder(requestBody).Decode(&timerRequest)

	if err != nil {
		log.Println(err)
	}

	newSession := createNewUserAndSession(timerRequest)
	newSessionRes, _ := json.Marshal(newSession)
	w.Write(newSessionRes)
	// json.NewEncoder(w).Encode(newSessionRes)
}

func joinSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var sessionRequest ExistingSessionReq
	var requestBody = r.Body
	enableCors(&w)

	json.NewDecoder(requestBody).Decode(&sessionRequest)
	matchedSession, err := joinExistingSession(sessionRequest)
	if err != nil {
		bufferedErr, _ := json.Marshal(err)
		w.Write(bufferedErr)
	}
	bufferedExistingSession, _ := json.Marshal(matchedSession)
	w.Write(bufferedExistingSession)
}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/session/new", newSessionEndpoint)
	http.HandleFunc("/session/join", joinSessionEndpoint)
}

func main() {
	fmt.Println("Golang WebSockets running...")
	setupRoutes()
	flag.Parse()
	fmt.Println("Listening on:", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
