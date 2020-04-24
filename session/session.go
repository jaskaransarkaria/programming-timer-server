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

var UpdateTimerChannel = make(chan Session)


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
func AddUserConnToSession(uuid string, conn *websocket.Conn) {
	sessionIdx := findSession(uuid)
	userIdx := findUser(sessionIdx, uuid)
	// add conn to user
	Sessions[sessionIdx].Users[userIdx].Conn = conn
}

func JoinExistingSession(joinExistingSessionData ExistingSessionReq, newUser User) (Session, error) {
	matchedSessionIdx, err := getExistingSession(joinExistingSessionData.JoinSessionID)
	if err != nil {
		return Sessions[matchedSessionIdx], err
	}
	Sessions[matchedSessionIdx].addUser(newUser)
	return Sessions[matchedSessionIdx], nil
}

func HandleUpdateSession(sessionToUpdate Session) {
	updatedSession, updateErr := sessionToUpdate.handleTimerEnd()
	if updateErr != nil {
		log.Println("updateError", updateErr)
		return
	}
	for _, user := range Sessions[updatedSession].Users {
		user.Conn.WriteJSON(Sessions[updatedSession])
	}
}

func (session *Session) handleTimerEnd() (int, error) {
	// update the session so that it has the most recent number of users
	updatedSessionIdx, err := getExistingSession(session.SessionID)
	if err != nil  {
		return -1, err
	}
	changeDriver(updatedSessionIdx)
	resetTimer(updatedSessionIdx)
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

func changeDriver(sessionIndex int) {
	if len(Sessions[sessionIndex].PreviousDrivers) == len(Sessions[sessionIndex].Users) {
		Sessions[sessionIndex].PreviousDrivers = nil
		selectNewDriver(sessionIndex)
	}
	Sessions[sessionIndex].PreviousDrivers = append(
		Sessions[sessionIndex].PreviousDrivers,
		Sessions[sessionIndex].CurrentDriver,
	)
	selectNewDriver(sessionIndex)
}

func selectNewDriver(sessionIndex int) {
	for _, user := range Sessions[sessionIndex].Users {
		if user != Sessions[sessionIndex].CurrentDriver {
			Sessions[sessionIndex].CurrentDriver = user
			break
		}
	}
}

func resetTimer(sessionIndex int) {
	var nowMsec = time.Now().UnixNano() / int64(time.Millisecond)
	Sessions[sessionIndex].StartTime = nowMsec
	Sessions[sessionIndex].EndTime = nowMsec + Sessions[sessionIndex].Duration
}

func (session *Session) addUser(user User) {
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
