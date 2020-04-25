package session

import (
	"github.com/google/uuid"
	"log"
	"errors"
	"github.com/gorilla/websocket"
	"time"
)

// User is ...
type User struct {
	UUID string
	Conn *websocket.Conn
}

// Session is ...
type Session struct {
	SessionID string
	CurrentDriver User
	Duration int64
	StartTime int64
	EndTime int64
	PreviousDrivers []User
	Users []User
}

// InitSessionResponse is ... 
type InitSessionResponse struct {
	Session Session
	User User
}

// StartTimerReq ... JSON request from the client
type StartTimerReq struct {
	Duration int64 `json:"duration"`
	StartTime int64 `json:"startTime"`
}

// ExistingSessionReq ...
type ExistingSessionReq struct {
	JoinSessionID string `json:"joinSession"`
}

// Sessions is a collection of all current sessions
var Sessions []Session

// UpdateTimerChannel is the channel which reads updates
var UpdateTimerChannel = make(chan Session)

// GenerateRandomID generates session & user ids
func GenerateRandomID(typeOfID string) string {
	length, err := getIDLength(typeOfID)
		if err != nil {
			log.Println("err generating ID", err)
		}
	uuid := uuid.New().String()
	
	// b := make([]byte, length)
	// rand.Read(b)
	// s := hex.EncodeToString(b)
	return uuid[:length]
}

// CreateNewUserAndSession creates new users and sessions
func CreateNewUserAndSession(newSessionData StartTimerReq, newUser User) Session {
	var newSession = Session{
		SessionID: GenerateRandomID("session"),
		CurrentDriver: newUser,
		Duration: newSessionData.Duration,
		StartTime: newSessionData.StartTime,
		EndTime: newSessionData.Duration + newSessionData.StartTime,
	}
	newSession.addUser(newUser)
	Sessions = append(Sessions, newSession)
	return newSession
}

// AddUserConnToSession adds the ws connection to the relevant session
func AddUserConnToSession(uuid string, conn *websocket.Conn) {
	sessionIdx := findSession(uuid)
	userIdx := Sessions[sessionIdx].findUser(uuid)
	Sessions[sessionIdx].Users[userIdx].Conn = conn
}

// JoinExistingSession adds a user to an existing session
func JoinExistingSession(joinExistingSessionData ExistingSessionReq, newUser User) (Session, error) {
	matchedSessionIdx, err := getExistingSession(joinExistingSessionData.JoinSessionID)
	if err != nil {
		return Sessions[matchedSessionIdx], err
	}
	Sessions[matchedSessionIdx].addUser(newUser)
	return Sessions[matchedSessionIdx], nil
}

// HandleUpdateSession when a timer finishes
func HandleUpdateSession(sessionToUpdate Session) {
	updatedSessionidx, updateErr := sessionToUpdate.handleTimerEnd()
	if updateErr != nil {
		log.Println("updateError", updateErr)
		return
	}
	Sessions[updatedSessionidx].broadcastToSessionUsers()
}


func (session *Session) broadcastToSessionUsers() {
		for _, user := range session.Users {
		user.Conn.WriteJSON(session)
	}
}


func (session *Session) handleTimerEnd() (int, error) {
	updatedSessionIdx, err := getExistingSession(session.SessionID)
	if err != nil  {
		return -1, err
	}
	Sessions[updatedSessionIdx].changeDriver()
	Sessions[updatedSessionIdx].resetTimer()
	return updatedSessionIdx, nil
}

func getExistingSession(desiredSessionID string) (int, error) {
	for idx, session :=  range Sessions {
		if session.SessionID == desiredSessionID {
			return idx, nil
		}
	}
	return -1, errors.New("There are no sessions with the id:" + desiredSessionID)
}

func (session *Session) changeDriver() {
	if len(session.PreviousDrivers) == len(session.Users) {
		session.PreviousDrivers = nil
		session.selectNewDriver()

	} else {
		session.PreviousDrivers = append(
			session.PreviousDrivers,
			session.CurrentDriver,
		)
		session.selectNewDriver()
	}
}

func (session *Session) selectNewDriver() {
	for _, user := range session.Users {
		if user.UUID != session.CurrentDriver.UUID {
			session.CurrentDriver = user
			break
		}
	}
}

func (session *Session) resetTimer() {
	var nowMsec = time.Now().UnixNano() / int64(time.Millisecond)
	session.StartTime = nowMsec
	session.EndTime = nowMsec + session.Duration
}

func (session *Session) addUser(user User) {
		session.Users = append(session.Users, user)
	}

func getIDLength(typeOfID string) (int8, error) {
		if (typeOfID == "session") {
		return 4, nil
	}
	if (typeOfID == "user") {
		return 8, nil
	}
	return -1, errors.New("Invalid typeofID")
}


func findSession(uuid string) int {
	for idx, session := range Sessions {
		for _, user := range session.Users {
			if user.UUID == uuid {
				return idx
			}
		}
	}
	return -1
}

func (session *Session) findUser(uuid string) int {
	for idx, user := range session.Users {
		if user.UUID == uuid {
			return idx
		}
	}
	return -1
}
