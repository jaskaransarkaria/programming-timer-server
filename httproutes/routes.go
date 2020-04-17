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

func joinExistingSession(joinExistingSessionData session.ExistingSessionReq) (session.Session, error) {
		var newUser = session.User{ UUID: session.GenerateRandomID("user") }
		matchedSessionIdx, err := session.GetExistingSession(joinExistingSessionData.JoinSessionID)
		if err != nil {
			return session.Sessions[matchedSessionIdx], err
		}
		session.Sessions[matchedSessionIdx].AddUser(newUser)
		return session.Sessions[matchedSessionIdx], nil
}

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
	reader(ws)
}

func enableCors(w *http.ResponseWriter) {(*w).Header().Set("Access-Control-Allow-Origin", "*")}

func newSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var timerRequest session.StartTimerReq
	var requestBody = r.Body
	enableCors(&w)
	
	err := json.NewDecoder(requestBody).Decode(&timerRequest)

	if err != nil {
		log.Println(err)
	}

	newSession := session.CreateNewUserAndSession(timerRequest)
	newSessionRes, _ := json.Marshal(newSession)
	w.Write(newSessionRes)
}

func joinSessionEndpoint(w http.ResponseWriter, r *http.Request) {
	var sessionRequest session.ExistingSessionReq
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

func updateSessionEndpoint(w http.ResponseWriter,r *http.Request) {
	var sessionRequest session.Session
	var requestBody = r.Body
	enableCors(&w)

	json.NewDecoder(requestBody).Decode(&sessionRequest)

	matchedSessionIdx, err := session.GetExistingSession(sessionRequest.SessionID)
	if err != nil {
		log.Println(err)
	}
	session.Sessions[matchedSessionIdx].HandleTimerEnd()
	log.Printf("%+v\n", session.Sessions[matchedSessionIdx])
}

func SetupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/session/new", newSessionEndpoint)
	http.HandleFunc("/session/join", joinSessionEndpoint)
	http.HandleFunc("/session/update", updateSessionEndpoint)
}