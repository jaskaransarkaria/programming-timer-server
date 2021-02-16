package session

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/websocket"

	// "github.com/stretchr/testify/assert"
	"github.com/jaskaransarkaria/programming-timer-server/mocks"
)

type mockConnection struct {
}

// problem here is that im actually still using the gorilla websock sooo when I call broadcast user i'm still trying to make calls with the Conns

func (c *mockConnection) Upgrade() (*websocket.Conn, error) {
	return &websocket.Conn{}, nil
}

func mockGenerateRandomID(expectedID string) string {
	return fmt.Sprintf("mocked-id-%d", len(Sessions))
}

func setup() (User, StartTimerReq, int, *mocks.Connector) {
	// create a new session with a user
	// var connToAdd = mockConnection{}
	// mockUpgradeConn, _ := connToAdd.Upgrade()
	mockConn := &mocks.Connector{}
	var newSessionData = StartTimerReq{
		Duration:  60000,
		StartTime: 1000,
	}
	mockConn.On("ReadMessage").Return(1, []byte("test byte array"), nil)
	var newUser = User{
		UUID: "test-uuid",
		// Conn: mockUpgradeConn, // add the mock connector here
		Conn: mockConn, // add the mock connector here
	}

	var sessionsLengthBeforeSessionCreated = len(Sessions)

	CreateNewUserAndSession(
		newSessionData,
		newUser,
		mockGenerateRandomID,
	)

	return newUser, newSessionData, sessionsLengthBeforeSessionCreated, mockConn
}

func cleanup(sessionID string) {
	// remove sessions correctly here and then include them in cleanup
	RemoveSession(sessionID)
}

func TestCreateNewUserAndSession(t *testing.T) {
	var newSessionData = StartTimerReq{
		Duration:  60000,
		StartTime: 1000,
	}
	var newUser = User{
		UUID: "test-uuid",
	}

	var expected = Session{
		SessionID:     "mocked-id-0",
		CurrentDriver: newUser,
		Duration:      newSessionData.Duration,
		StartTime:     newSessionData.StartTime,
		EndTime:       newSessionData.Duration + newSessionData.StartTime,
		Users:         []User{newUser},
	}

	actual := CreateNewUserAndSession(
		newSessionData,
		newUser,
		mockGenerateRandomID,
	)
	if !cmp.Equal(expected, actual) {
		t.Errorf("Expected: %+v but recieved: %+v", expected, actual)
	}
	if len(Sessions) != 1 {
		t.Errorf("Expected: %+v but recieved: %+v", Sessions, actual)
	}
	if !cmp.Equal(Sessions[0], actual) {
		t.Errorf("Expected: %+v but recieved: %+v", Sessions[0], actual)
	}
	cleanup(actual.SessionID)
}

func TestAddUserConnToSession(t *testing.T) {
	_, _, sessionsLengthBeforeSessionCreated, _ := setup()
	var sessionID = fmt.Sprintf("mocked-id-%d", sessionsLengthBeforeSessionCreated)
	var connToAdd = mockConnection{}
	mockUpgradeConn, err := connToAdd.Upgrade() // add mock conn here
	if err != nil {
		t.Errorf("Expected: nil but recieved: %+v", err)
	}
	actual := AddUserConnToSession("test-uuid", mockUpgradeConn)
	if actual != nil {
		t.Errorf("Expected: nil but recieved: %+v", actual)
	}
	var expectedConn = Sessions[sessionsLengthBeforeSessionCreated].Users[0].Conn
	if mockUpgradeConn != expectedConn {
		t.Errorf("Expected: %+v but recieved: %+v", mockUpgradeConn, expectedConn)
	}
	cleanup(sessionID)
}

func TestJoinExistingSession(t *testing.T) {
	existingUser, existingSessionData, sessionIndex, _ := setup()

	sessionID := fmt.Sprintf("mocked-id-%d", sessionIndex)
	var newUser = User{
		UUID: "test-uuid2",
	}

	var sessionToJoin = ExistingSessionReq{
		JoinSessionID: sessionID,
	}
	actual, err := JoinExistingSession(sessionToJoin, newUser)

	if err != nil {
		t.Errorf("Expected: %+v but recieved: %+v", nil, err)
	}

	var expected = Session{
		SessionID:     sessionID,
		CurrentDriver: existingUser,
		Duration:      existingSessionData.Duration,
		StartTime:     existingSessionData.StartTime,
		EndTime:       existingSessionData.Duration + existingSessionData.StartTime,
		Users:         []User{existingUser, newUser},
	}

	if !cmp.Equal(actual, expected, cmpopts.IgnoreFields(User{}, "Conn")) {
		t.Errorf("Expected: %+v but recieved: %+v", expected, actual)
	}
	cleanup(sessionID)
}

func TestRemoveSession(t *testing.T) {
	_, _, sessionsLengthBeforeSessionCreated, _ := setup()
	sessionID := fmt.Sprintf("mocked-id-%d", sessionsLengthBeforeSessionCreated)
	removeSessionErr := RemoveSession(sessionID)

	if removeSessionErr != nil {
		t.Errorf("Expected nil but received %+v", removeSessionErr)
	}

	if len(Sessions) != sessionsLengthBeforeSessionCreated {
		t.Errorf("Expected %d sessions in Sessions slice but received %d\n with %+v",
			sessionsLengthBeforeSessionCreated,
			len(Sessions),
			Sessions,
		)

		for _, session := range Sessions {
			if session.SessionID == sessionID {
				t.Errorf("session not removed: %+v", session)
			}
		}
	}
}

func TestHandleUpdateSession(t *testing.T) {
	// take an existing session
	_, _, sessionIndex, mockConnInitUser := setup()
	// add another user (so we can verify that the function is switching driver correctly
	sessionID := fmt.Sprintf("mocked-id-%d", sessionIndex)
	mockConnJoiningUser := &mocks.Connector{}

	var newUser = User{
		UUID: "test-uuid2",
		Conn: mockConnJoiningUser,
	}

	var sessionToJoin = ExistingSessionReq{
		JoinSessionID: sessionID,
	}
	testSession, _ := JoinExistingSession(sessionToJoin, newUser)
	mockUpdateRequest := UpdateRequest{
		SessionID:       testSession.SessionID,
		UpdatedDuration: testSession.Duration,
	}
	// mock broadcast to all sessionUsers
	mockConnInitUser.On("WriteJSON", &Sessions[sessionIndex]).Return(nil)
	mockConnJoiningUser.On("WriteJSON", &Sessions[sessionIndex]).Return(nil)
	// fire handle time end  (changes driver and resets the timer)
	HandleUpdateSession(mockUpdateRequest)

	if Sessions[sessionIndex].CurrentDriver.UUID != newUser.UUID {
		t.Errorf("The Driver has not been correctly changed")
	}
}
