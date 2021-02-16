package session

import (
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jaskaransarkaria/programming-timer-server/utils"
)

// Connector is .. the User's current connection
type Connector interface {
	WriteJSON(v interface{}) error
	ReadMessage() (int, []byte, error)
}

// User is ... each connected user
type User struct {
	UUID string
	Conn Connector
}

// Session is ... each active session
type Session struct {
	SessionID       string
	CurrentDriver   User
	Duration        int64
	StartTime       int64
	EndTime         int64
	PreviousDrivers []User
	Users           []User
}

// PauseSessionResponse ... the time when a user pauses the timer
type PauseSessionResponse struct {
	PauseTime int64
}

// UnpauseSessionResponse ... the time when a user restarts the timer
type UnpauseSessionResponse struct {
	UnpauseTime int64
}

// InitSessionResponse is ... on inital connection
type InitSessionResponse struct {
	Session Session
	User    User
}

// StartTimerReq ... JSON request from the client
type StartTimerReq struct {
	Duration  int64 `json:"duration"`
	StartTime int64 `json:"startTime"`
}

// ExistingSessionReq ... join an existing session http request
type ExistingSessionReq struct {
	JoinSessionID string `json:"joinSession"`
}

// UpdateRequest .. Incoming timer update from client (current driver)
type UpdateRequest struct {
	SessionID       string `json:"sessionId"`
	UpdatedDuration int64  `json:"updatedDuration,omitempty"`
}

// PauseRequest ... incoming pause time and session ID from client
type PauseRequest struct {
	SessionID string `json:"sessionId"`
	PauseTime int64  `json:"pauseTime"`
}

// UnpauseRequest ... incoming pause time and session ID from client
type UnpauseRequest struct {
	SessionID   string `json:"sessionId"`
	UnpauseTime int64  `json:"unpauseTime"`
}

// Sessions is a collection of all current sessions
var Sessions []Session

// UpdateTimerChannel reads updates as they come in via updateSessionEndpoint
var UpdateTimerChannel = make(chan UpdateRequest)

// PauseTimerChannel reads pause requests as they come in via pauseSessionEndpoint
var PauseTimerChannel = make(chan PauseRequest)

// UnpauseTimerChannel reads restart requests as they come in via unpauseSessionEndpoint
var UnpauseTimerChannel = make(chan UnpauseRequest)

// CreateNewUserAndSession creates new users and sessions
func CreateNewUserAndSession(
	newSessionData StartTimerReq,
	newUser User,
	generateIDFunc utils.RandomGenerator,
) Session {
	var newSession = Session{
		SessionID:     generateIDFunc("session"),
		CurrentDriver: newUser,
		Duration:      newSessionData.Duration,
		StartTime:     newSessionData.StartTime,
		EndTime:       newSessionData.Duration + newSessionData.StartTime,
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
func HandleUpdateSession(sessionToUpdate UpdateRequest) {
	updatedSessionIdx, updateErr := handleTimerEnd(sessionToUpdate)
	if updateErr != nil {
		log.Println("updateError", updateErr)
		return
	}
	Sessions[updatedSessionIdx].broadcast(Sessions[updatedSessionIdx])
}

// HandlePauseSession when the driver pauses the timer
func HandlePauseSession(sessionToPause PauseRequest) {
	pauseTime := PauseSessionResponse{PauseTime: sessionToPause.PauseTime}
	log.Println(sessionToPause)
	pausedSessionIdx, pauseErr := getExistingSession(sessionToPause.SessionID)
	if pauseErr != nil {
		log.Println("pauseError", pauseErr)
		return
	}
	Sessions[pausedSessionIdx].broadcast(pauseTime)
}

// HandleUnpauseSession when the driver pauses the timer
func HandleUnpauseSession(sessionToUnpause UnpauseRequest) {
	unpauseTime := UnpauseSessionResponse{UnpauseTime: sessionToUnpause.UnpauseTime}
	unpausedSessionIdx, unpauseErr := getExistingSession(sessionToUnpause.SessionID)
	if unpauseErr != nil {
		log.Println("pauseError", unpauseErr)
		return
	}
	Sessions[unpausedSessionIdx].broadcast(unpauseTime)
}

// HandleRemoveUser ... of a disconneted user from the relevent session
func HandleRemoveUser(conn *websocket.Conn) error {
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

func (session *Session) broadcast(payload interface{}) {
	for _, user := range session.Users {
		log.Println("broadcast", payload)
		user.Conn.WriteJSON(payload)
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

// Map the incoming session request to an in-memory session
func handleTimerEnd(session UpdateRequest) (int, error) {
	mappedSessionIdx, err := getExistingSession(session.SessionID)
	if err != nil {
		return -1, err
	}
	// update duration
	if session.UpdatedDuration > 0 {
		Sessions[mappedSessionIdx].Duration = session.UpdatedDuration
	}
	Sessions[mappedSessionIdx].changeDriver()
	Sessions[mappedSessionIdx].resetTimer()
	return mappedSessionIdx, nil
}

func getExistingSession(desiredSessionID string) (int, error) {
	for idx, session := range Sessions {
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
		session.broadcast(session)
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
