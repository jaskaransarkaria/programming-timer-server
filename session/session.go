package session

import (
	"log"
	"math/rand"
	"encoding/hex"
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

type InitSessionResponse struct {
	Session Session
	User User
}

// StartTimer ... JSON request from the client
type StartTimerReq struct {
	Duration int64 `json:"duration"`
	StartTime int64 `json:"startTime"`
}

type ExistingSessionReq struct {
	JoinSessionID string `json:"joinSession"`
}

var Sessions []Session


func (session *Session) AddUser(user User) {
		session.Users = append(session.Users, user)
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

func GenerateRandomID(typeOfID string) string {
	length, err := getIDLength(typeOfID)
		if err != nil {
			log.Println("err generating ID", err)
		}
	b := make([]byte, length)
	rand.Read(b)
	s := hex.EncodeToString(b)
	return s
}

func (session *Session) selectNewDriver() Session {
	// choose a uuid from the users that's not the currentDriver
	for _, user := range session.Users {
		if user != session.CurrentDriver {
			session.CurrentDriver = user
		} 
	}
	return *session
}

func (session *Session) changeDriver() Session {
	if len(session.PreviousDrivers) == len(session.Users) {
		session.PreviousDrivers = session.PreviousDrivers[:0]
		return session.selectNewDriver()
	}
	session.PreviousDrivers = append(session.PreviousDrivers, session.CurrentDriver)
  return session.selectNewDriver()
}

func (session *Session) resetTimer() {
	var nowMsec = time.Now().UnixNano() / int64(time.Millisecond)
	session.StartTime = nowMsec
	session.EndTime = nowMsec + session.Duration
}

func (session *Session) HandleTimerEnd() (Session, error) {
	// update the session so that it has the most recent number of users
	updatedSessionIdx, err := GetExistingSession(session.SessionID)
	if err != nil  {
		return Session{}, err
	}
	Sessions[updatedSessionIdx].changeDriver()
	Sessions[updatedSessionIdx].resetTimer()
	return Sessions[updatedSessionIdx], nil
}

func AddUserConnToSession(uuid string, conn *websocket.Conn) {
	sessionIdx := findSession(uuid)
	userIdx := findUser(sessionIdx, uuid)
	// add conn to user
	Sessions[sessionIdx].Users[userIdx].Conn = conn
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

func findUser(sessionIdx int, uuid string) int {
	for idx, user := range Sessions[sessionIdx].Users {
		if user.UUID == uuid {
			return idx
		}
	}
	return -1
}

func GetExistingSession(desiredSessionID string) (int, error) {
	// iterate through sessions
	for idx, session :=  range Sessions {
		if session.SessionID == desiredSessionID {
			return idx, nil
		}
	}
	return -1, errors.New("There are no sessions with the id:" + desiredSessionID)
}

func CreateNewUserAndSession(newSessionData StartTimerReq, newUser User) Session {
	var newSession = Session{
				SessionID: GenerateRandomID("session"),
				CurrentDriver: newUser,
				Duration: newSessionData.Duration,
				StartTime: newSessionData.StartTime,
				EndTime: newSessionData.Duration + newSessionData.StartTime,
			}
	newSession.AddUser(newUser)
	Sessions = append(Sessions, newSession)
	return newSession
}

