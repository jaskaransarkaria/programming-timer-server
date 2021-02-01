package httproutes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jaskaransarkaria/programming-timer-server/readers"
	"github.com/jaskaransarkaria/programming-timer-server/session"
	"github.com/jaskaransarkaria/programming-timer-server/utils"
)

var upgrader = websocket.Upgrader{
	// empty struct means use defaults
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func enableCors(w *http.ResponseWriter) { (*w).Header().Set("Access-Control-Allow-Origin", "*") }

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// this is for CORS -  allow all origin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// upgrade http connection to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client successfully connected to Golang Websocket!")
	go readers.ConnReader(ws)
}

func updateSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var sessionToUpdate session.UpdateRequest
	var requestBody = r.Body
	enableCors(&w)
	err := json.NewDecoder(requestBody).Decode(&sessionToUpdate)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	session.UpdateTimerChannel <- sessionToUpdate
}

func pauseSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var sessionToPause session.PauseRequest
	var requestBody = r.Body
	log.Println("request body", requestBody)
	enableCors(&w)
	err := json.NewDecoder(requestBody).Decode(&sessionToPause)
	log.Println("pause session endpoint reached", sessionToPause)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	session.PauseTimerChannel <- sessionToPause
}

func unpauseSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var sessionToUnpause session.UnpauseRequest
	var requestBody = r.Body
	log.Println("request body", requestBody)
	enableCors(&w)
	err := json.NewDecoder(requestBody).Decode(&sessionToUnpause)
	log.Println("unpause session endpoint reached", sessionToUnpause)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	session.UnpauseTimerChannel <- sessionToUnpause
}

func newSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var timerRequest session.StartTimerReq
	var requestBody = r.Body
	log.Println(requestBody)
	enableCors(&w)
	err := json.NewDecoder(requestBody).Decode(&timerRequest)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	newUser := session.User{UUID: utils.GenerateRandomID("user")}
	newSession := session.CreateNewUserAndSession(
		timerRequest,
		newUser,
		utils.GenerateRandomID,
	)
	resp := session.InitSessionResponse{
		Session: newSession,
		User:    newUser,
	}
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
	var newUser = session.User{UUID: utils.GenerateRandomID("user")}
	matchedSession, err := session.JoinExistingSession(sessionRequest, newUser)
	if err != nil {
		bufferedErr, _ := json.Marshal(err)
		w.Write(bufferedErr)
	}
	resp := session.InitSessionResponse{
		Session: matchedSession,
		User:    newUser,
	}
	bufferedExistingSession, _ := json.Marshal(resp)
	w.Write(bufferedExistingSession)
}

func SetupRoutes() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/session/new", newSessionEndpoint)
	http.HandleFunc("/session/join", joinSessionEndpoint)
	http.HandleFunc("/session/update", updateSessionEndpoint)
	go readers.UpdateChannelReader()
	http.HandleFunc("/session/pause", pauseSessionEndpoint)
	go readers.PauseChannelReader()
	http.HandleFunc("/session/unpause", unpauseSessionEndpoint)
	go readers.UnpauseChannelReader()
}
