package session

import (
	"fmt"
	"testing"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/websocket"
)

type mockConnection struct {
}

func (u *mockConnection) Upgrade() (*websocket.Conn, error) {
	return &websocket.Conn{}, nil
}

func mockGenerateRandomID(expectedID string) string {
	return fmt.Sprintf("mocked-id-%d", len(Sessions))
}

func setup() (User, StartTimerReq, int) {
	var connToAdd = mockConnection{}
	mockUpgradeConn, _ := connToAdd.Upgrade()

	var newSessionData = StartTimerReq{
		Duration: 60000,
		StartTime: 1000,
	}

	var newUser = User{
		UUID: "test-uuid",
		Conn: mockUpgradeConn,
	}

	var sessionsLengthBeforeSessionCreated = len(Sessions)

	CreateNewUserAndSession(
		newSessionData,
		newUser,
		mockGenerateRandomID,
	)

	return newUser, newSessionData, sessionsLengthBeforeSessionCreated
}

func cleanup(sessionID string) {
	// remove sessions correctly here and then include them in cleanup
	RemoveSession(sessionID)
}

func TestCreateNewUserAndSession(t *testing.T) {
	var newSessionData = StartTimerReq{
		Duration: 60000,
		StartTime: 1000,
	}
	var newUser = User{
		UUID: "test-uuid",
	}

	var expected = Session{
		SessionID: "mocked-id-0",
		CurrentDriver: newUser,
		Duration: newSessionData.Duration,
		StartTime: newSessionData.StartTime,
		EndTime: newSessionData.Duration + newSessionData.StartTime,
		Users: []User{newUser},
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
	_, _, sessionsLengthBeforeSessionCreated := setup()
	var sessionID = fmt.Sprintf("mocked-id-%d", sessionsLengthBeforeSessionCreated)
	var connToAdd = mockConnection{}
	mockUpgradeConn, err := connToAdd.Upgrade()
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
	existingUser, existingSessionData, sessionsLengthBeforeSessionCreated := setup()

	sessionID := fmt.Sprintf("mocked-id-%d", sessionsLengthBeforeSessionCreated)
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
		SessionID: sessionID,
		CurrentDriver: existingUser,
		Duration: existingSessionData.Duration,
		StartTime: existingSessionData.StartTime,
		EndTime: existingSessionData.Duration + existingSessionData.StartTime,
		Users: []User{existingUser, newUser},
	}

	if !cmp.Equal(actual, expected, cmpopts.IgnoreFields(User{}, "Conn")) {
		t.Errorf("Expected: %+v but recieved: %+v", expected, actual)
	}
	cleanup(sessionID)
}

func TestRemoveSession(t *testing.T) {
	_, _, sessionsLengthBeforeSessionCreated := setup()
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
	
	// func TestHandleUpdateSession(t *testing.T) {}
	




}
