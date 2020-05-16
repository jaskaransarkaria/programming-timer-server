package session

import (
	"testing"
	"github.com/google/go-cmp/cmp"
)

func mockGenerateRandomID(expectedID string) string {
	return expectedID
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
		SessionID: "session",
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
		t.Errorf("Expected: %+v but recieved: %+v", expected, actual) // need to mock out the call to utils.GenerateRandomID()
	}
}
func TestAddUserConnToSession(t *testing.T) {}
func TestJoinExistingSession(t *testing.T) {}
func TestHandleUpdateSession(t *testing.T) {}
func TestHandleRemoveSession(t *testing.T) {}
