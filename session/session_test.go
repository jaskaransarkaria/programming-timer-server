package session

import (
	"testing"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
)

type mockConnection struct {
}

func (u *mockConnection) Upgrade() (*websocket.Conn, error) {
	return &websocket.Conn{}, nil
}

func mockGenerateRandomID(expectedID string) string {
	return "mocked-id"
}



func setup() (User, StartTimerReq) {
	var newSessionData = StartTimerReq{
		Duration: 60000,
		StartTime: 1000,
	}
	var newUser = User{
		UUID: "test-uuid",
	}
	CreateNewUserAndSession(
		newSessionData,
		newUser,
		mockGenerateRandomID,
	)

	return newUser, newSessionData
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
		SessionID: "mocked-id",
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
}

func TestAddUserConnToSession(t *testing.T) {
	setup()
	var connToAdd = mockConnection{}
	mockUpgradeConn, err := connToAdd.Upgrade()
	if err != nil {
		t.Errorf("Expected: nil but recieved: %+v", err)
	}
	actual := AddUserConnToSession("test-uuid", mockUpgradeConn)
	if actual != nil {
		t.Errorf("Expected: nil but recieved: %+v", actual)
	}
	var expectedConn = Sessions[0].Users[0].Conn
	if mockUpgradeConn != expectedConn {
		t.Errorf("Expected: %+v but recieved: %+v", mockUpgradeConn, expectedConn)
	}
}

func TestJoinExistingSession(t *testing.T) {
	existingUser, existingSessionData := setup()
	var newUser = User{
		UUID: "test-uuid2",
	}
	var sessionToJoin = ExistingSessionReq{
		JoinSessionID: "mocked-id",
	}
	actual, err := JoinExistingSession(sessionToJoin, newUser)
	if err != nil {
		t.Errorf("Expected: %+v but recieved: %+v", nil, err)
	}
	var expected = Session{
		SessionID: "mocked-id",
		CurrentDriver: existingUser,
		Duration: existingSessionData.Duration,
		StartTime: existingSessionData.StartTime,
		EndTime: existingSessionData.Duration + existingSessionData.StartTime,
		Users: []User{existingUser, newUser},
	}
	if !cmp.Equal(expected, actual) {
		t.Errorf("Expected: %+v but recieved: %+v", expected, actual)
	}
}

func TestHandleUpdateSession(t *testing.T) {}
func TestHandleRemoveSession(t *testing.T) {}
