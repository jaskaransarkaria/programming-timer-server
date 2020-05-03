package httproutes

import (
	"github.com/gorilla/websocket"
	"log"
	"encoding/json"
	"net/http"
	"github.com/jaskaransarkaria/programming-timer-server/session"
	"github.com/jaskaransarkaria/programming-timer-server/readers"
	"github.com/jaskaransarkaria/programming-timer-server/utils"
)

var upgrader = websocket.Upgrader{
	// empty struct means use defaults
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func enableCors(w *http.ResponseWriter) {(*w).Header().Set("Access-Control-Allow-Origin", "*")}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// this is for CORS -  allow all origin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// upgrade http connection to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client successfully connected to Golang Websocket!")
	go readers.NewConnReader(ws)
}

func updateSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var sessionToUpdate session.Session
	var requestBody = r.Body
	enableCors(&w)
	err := json.NewDecoder(requestBody).Decode(&sessionToUpdate)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	session.UpdateTimerChannel <- sessionToUpdate
}

func newSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var timerRequest session.StartTimerReq
	var requestBody = r.Body
	enableCors(&w)
	err := json.NewDecoder(requestBody).Decode(&timerRequest)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	newUser := session.User{ UUID: utils.GenerateRandomID("user") }
	newSession := session.CreateNewUserAndSession(timerRequest, newUser)
	resp := session.InitSessionResponse{newSession, newUser}
	newSessionRes, _ := json.Marshal(resp)
	w.Write(newSessionRes)
}

func joinSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var sessionRequest session.ExistingSessionReq
	var requestBody = r.Body
	enableCors(&w)
	err := json.NewDecoder(requestBody).Decode(&sessionRequest)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	var newUser = session.User{ UUID: utils.GenerateRandomID("user") }
	matchedSession, err := session.JoinExistingSession(sessionRequest, newUser)
	if err != nil {
		bufferedErr, _ := json.Marshal(err)
		w.Write(bufferedErr)
	}
	resp := session.InitSessionResponse{matchedSession, newUser}
	bufferedExistingSession, _ := json.Marshal(resp)
	w.Write(bufferedExistingSession)
}

func SetupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/session/new", newSessionEndpoint)
	http.HandleFunc("/session/join", joinSessionEndpoint)
	http.HandleFunc("/session/update", updateSessionEndpoint)
	go readers.UpdateChannelReader()
}