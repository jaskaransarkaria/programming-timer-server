package readers

import (
	"github.com/gorilla/websocket"
	"log"
	"github.com/jaskaransarkaria/programming-timer-server/session"
	"github.com/jaskaransarkaria/programming-timer-server/writers"
)

func NewConnReader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		log.Println(string(p))
		session.AddUserConnToSession(string(p), conn)
		if err != nil {
			log.Println("Connection closing:", err)
			// hear we are actually listening for close connections shown in err
			conn.Close()
		}
		writers.NewConnWriter(conn, messageType, []byte("well done you've connected via web sockets to a go server"))
	}
}

func UpdateChannelReader() {
	for {
		recievedUpdate := <- session.UpdateTimerChannel
		session.HandleUpdateSession(recievedUpdate)
	}
}