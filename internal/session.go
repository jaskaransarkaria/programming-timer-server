package internal
import (
	"fmt"
	"log"
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
	CurrentDriver User
	Duration int64
	StartTime int64
	EndTime int64
	PreviousDrivers []User
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

func selectNewDriver(session Session) Session {
	// choose a uuid from the users that's not the currentDriver
	for _, user := range session.Users {
		if user != session.CurrentDriver {
			session.CurrentDriver = user
		} 
	}
	return session
}

func changeDriver(session Session) Session {
	if len(session.PreviousDrivers) == len(session.Users) {
		session.PreviousDrivers = session.PreviousDrivers[:0]
		return selectNewDriver(session)
	}
	session.PreviousDrivers = append(session.PreviousDrivers, session.CurrentDriver)
  return selectNewDriver(session)
}

func HandleTimerEnd(session Session) Session {
	updatedSession := changeDriver(session)
	fmt.Printf("%+v\n", updatedSession)
	return updatedSession
	// send new session state down to all the users via websocket
}


func GetExistingSession(desiredSessionID string) (Session, error) {
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

func CreateNewUserAndSession(newSessionData StartTimerReq) Session {
	var newUser = User{ UUID: GenerateRandomID("user") }
	var newSession = Session{
				SessionID: GenerateRandomID("session"),
				CurrentDriver: newUser,
				Duration: newSessionData.Duration,
				StartTime: newSessionData.StartTime,
				EndTime: newSessionData.Duration + newSessionData.StartTime,
			}
	newSession.AddUser(newUser)
	sessions = append(sessions, newSession)
	return newSession
}