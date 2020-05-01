package readers

import (
	"github.com/gorilla/websocket"
	"log"
	"github.com/jaskaransarkaria/programming-timer-server/session"
	// "github.com/jaskaransarkaria/programming-timer-server/writers"
)

func NewConnReader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Connection closing:", err)
			session.RemoveUser(conn)
			conn.Close()
			break
			} else {
			log.Println(string(p))
			session.AddUserConnToSession(string(p), conn)
		}
		// writers.NewConnWriter(conn, messageType, []byte("well done you've connected via web sockets to a go server"))
	}
}

func UpdateChannelReader() {
	for {
		recievedUpdate := <- session.UpdateTimerChannel
		session.HandleUpdateSession(recievedUpdate)
	}
}
