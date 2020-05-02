package readers

import (
	"github.com/gorilla/websocket"
	"log"
	"github.com/jaskaransarkaria/programming-timer-server/session"
)

func NewConnReader(conn *websocket.Conn) {
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
			session.AddUserConnToSession(string(p), conn)
		}
	}
}

func UpdateChannelReader() {
	for {
		recievedUpdate := <- session.UpdateTimerChannel
		session.HandleUpdateSession(recievedUpdate)
	}
}
