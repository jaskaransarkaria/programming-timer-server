package session

import (
	"log"
	"errors"
	"time"
	"github.com/gorilla/websocket"
	"github.com/jaskaransarkaria/programming-timer-server/utils"
)

// User is ... each connected user
type User struct {
	UUID string
	Conn *websocket.Conn
}

// Session is ... each active session
type Session struct {
	SessionID string
	CurrentDriver User
	Duration int64
	StartTime int64
	EndTime int64
	PreviousDrivers []User
	Users []User
}

// InitSessionResponse is ... on inital connection
type InitSessionResponse struct {
	Session Session
	User User
}

// StartTimerReq ... JSON request from the client
type StartTimerReq struct {
	Duration int64 `json:"duration"`
	StartTime int64 `json:"startTime"`
}

// ExistingSessionReq ... join an existing session http request
type ExistingSessionReq struct {
	JoinSessionID string `json:"joinSession"`
}

// Sessions is a collection of all current sessions
var Sessions []Session

// UpdateTimerChannel reads updates as they come in via updateSessionEndpoint
var UpdateTimerChannel = make(chan Session)

// CreateNewUserAndSession creates new users and sessions
func CreateNewUserAndSession(
	newSessionData StartTimerReq,
	newUser User,
	generateIDFunc utils.RandomGenerator,
	) Session {
	var newSession = Session{
		SessionID: generateIDFunc("session"),
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
func AddUserConnToSession(uuid string, conn *websocket.Conn) error {
	sessionIdx, sessionErr := findSession(uuid)
		if sessionErr != nil {
		return sessionErr
	}
	userIdx, userErr := Sessions[sessionIdx].findUser(uuid)
		if userErr != nil {
		return userErr
	}
	Sessions[sessionIdx].Users[userIdx].Conn = conn
	return nil
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
	updatedSessionIdx, updateErr := sessionToUpdate.handleTimerEnd()
	if updateErr != nil {
		log.Println("updateError", updateErr)
		return
	}
	Sessions[updatedSessionIdx].broadcastToSessionUsers()
}

// HandleRemoveUser ... of a disconneted user from the relevent session
func HandleRemoveUser(conn *websocket.Conn) (error) {
	sessionIdx, userIdx, findConnErr := findUserByConn(conn)
	if findConnErr != nil {
		return findConnErr
	}
	Sessions[sessionIdx].removeUser(userIdx)
	if len(Sessions[sessionIdx].Users) == 0 {
		RemoveSession(Sessions[sessionIdx].SessionID)
	}
	return nil
}

func (session *Session) broadcastToSessionUsers() {
		for _, user := range session.Users {
			user.Conn.WriteJSON(session)
	}
}

// RemoveSession ... for a abandoned session
func RemoveSession(sessionID string) error {
	// find session by sessionID
	sessionIndex, sessionErr := findSession(sessionID)
	if sessionErr != nil {
		return sessionErr
	}
	// remove the session from slice
	Sessions = append(Sessions[:sessionIndex], Sessions[sessionIndex+1:]...)
	return nil
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
	if len(session.PreviousDrivers) >= len(session.Users) {
		session.PreviousDrivers = nil
	}
	session.selectNewDriver()
}

func (session *Session) selectNewDriver() {
	for _, user := range session.Users {
		if user.UUID != session.CurrentDriver.UUID {
			beenDriver := session.hasUserBeenDriver(user.UUID)
			if beenDriver == false {
				session.CurrentDriver = user
				session.PreviousDrivers = append(
					session.PreviousDrivers,
					session.CurrentDriver,
				)
				log.Println("new driver selected")
				break
			}
		}
	}
}
	func (session *Session) hasUserBeenDriver(uuid string) bool {
		if len(session.PreviousDrivers) > 0 {
			for _, prevDriver := range session.PreviousDrivers {
				if uuid == prevDriver.UUID {
					return true
				}
			}
		}
	return false
}

func (session *Session) resetTimer() {
	var nowMsec = time.Now().UnixNano() / int64(time.Millisecond)
	session.StartTime = nowMsec
	session.EndTime = nowMsec + session.Duration
}

func (session *Session) addUser(user User) {
		session.Users = append(session.Users, user)
	}



func findSession(keyToFind interface{}) (int, error) {
	switch keyToFind.(type) {
	case string:
		for idx, session := range Sessions {
			if session.SessionID == keyToFind {
				return idx, nil
			}
			for _, user := range session.Users {
				if user.UUID == keyToFind {
					return idx, nil
				}
			}
		}
	case *websocket.Conn:
		for idx, session := range Sessions {
			for _, user := range session.Users {
				if user.Conn == keyToFind {
					return idx, nil
				}
			}
		}
	}
	return -1, errors.New("Cannot find Session")
}

func (session *Session) findUser(keyToFind interface{}) (int, error) {
	switch keyToFind.(type) {
	case string:
		for idx, user := range session.Users {
			if user.UUID == keyToFind {
				return idx, nil
			}
		}
	case *websocket.Conn:
		for idx, user := range session.Users {
			if user.Conn == keyToFind {
				return idx, nil
			}
		}
	}
	return -1, errors.New("Cannot find user")
}

func (session *Session) removeUser(userIdx int) {
	session.resetCurrentDriver(session.Users[userIdx])
	// Copy last element to index userIdx
	session.Users[userIdx] = session.Users[len(session.Users)-1]
	// Erase last element (write zero value).
	session.Users[len(session.Users)-1] = User{}
	// Truncate slice.
	session.Users = session.Users[:len(session.Users)-1]
}

func (session *Session) resetCurrentDriver(userToBeRemoved User) {
	if userToBeRemoved == session.CurrentDriver {
		session.changeDriver()
		session.broadcastToSessionUsers()
	}
}

func findUserByConn(conn *websocket.Conn) (int, int, error) {
	sessionIdx, sessionErr := findSession(conn)
	if sessionErr != nil {
		return -1, -1, sessionErr
	}
	userIdx, userErr := Sessions[sessionIdx].findUser(conn)
	if userErr != nil {
		return -1, -1, userErr
	}
	return sessionIdx, userIdx, nil
}
