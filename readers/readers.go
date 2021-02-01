package readers

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/jaskaransarkaria/programming-timer-server/session"
)

// ConnReader ... add/ remove client connections
func ConnReader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Connection closing:", err)
			removeUserErr := session.HandleRemoveUser(conn)
			if removeUserErr != nil {
				log.Println(removeUserErr)
			}
			conn.Close()
			break
		} else {
			log.Println(string(p))
			addToSessionErr := session.AddUserConnToSession(string(p), conn)
			if addToSessionErr != nil {
				log.Println(addToSessionErr)
			}
		}
	}
}

// UpdateChannelReader handle updates to sessions
func UpdateChannelReader() {
	for {
		recievedUpdate := <-session.UpdateTimerChannel
		session.HandleUpdateSession(recievedUpdate)
	}
}

//PauseChannelReader handles pause requests
func PauseChannelReader() {
	for {
		pauseRequest := <-session.PauseTimerChannel
		session.HandlePauseSession(pauseRequest)
	}
}

//UnpauseChannelReader handles restart requests
func UnpauseChannelReader() {
	for {
		unpauseRequest := <-session.UnpauseTimerChannel
		session.HandleUnpauseSession(unpauseRequest)
	}
}
