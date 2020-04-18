package httproutes

import (
	"github.com/gorilla/websocket"
	"log"
	"encoding/json"
	"net/http"
	"github.com/jaskaransarkaria/programming-timer-server/session"
)

var upgrader = websocket.Upgrader{
	// empty struct means use defaults
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func joinExistingSession(joinExistingSessionData session.ExistingSessionReq, newUser session.User) (session.Session, error) {
		matchedSessionIdx, err := session.GetExistingSession(joinExistingSessionData.JoinSessionID)
		if err != nil {
			return session.Sessions[matchedSessionIdx], err
		}
		session.Sessions[matchedSessionIdx].AddUser(newUser)
		return session.Sessions[matchedSessionIdx], nil
}

func enableCors(w *http.ResponseWriter) {(*w).Header().Set("Access-Control-Allow-Origin", "*")}

func writer(conn *websocket.Conn, messageType int, message []byte) {
	// message the client
	if err := conn.WriteMessage(messageType, message); err != nil {
		log.Println(err)
		}
}

func reader(conn *websocket.Conn) { // need to make each connection a go routine
	// listen on this connection for new messages and send messages down that connection
	for {
			messageType, p, err := conn.ReadMessage()
			log.Println(string(p))
			// find session
			sessionIdx := session.FindSession(string(p))
			// find the user 
			userIdx := session.FindUser(sessionIdx, string(p))
			// add conn to user
			session.Sessions[sessionIdx].Users[userIdx] = *conn
			if err != nil {
				log.Println(err)
				// hear we are actually listening for close connections shown in err
				conn.Close()
			}
			writer(conn, messageType, []byte("well done you've connected via web sockets to a go server"))

			var sessionToUpdate session.Session
			jsonErr := conn.ReadJSON(&sessionToUpdate)
			if jsonErr != nil {
				log.Println(jsonErr)
			}
			sessionToUpdate.HandleTimerEnd()
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
	// reader ran as a goroute from main? But it will be reading the message from the channel
	// write as a go routine
	reader(ws)
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
	
	newUser := session.User{ UUID: session.GenerateRandomID("user") }
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

	var newUser = session.User{ UUID: session.GenerateRandomID("user") }
	matchedSession, err := joinExistingSession(sessionRequest, newUser)
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
}


// think about if I need to re-architect the way I read and write messages? 
// Using goroutines and Channels?