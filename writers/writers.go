package writers

import (
	"github.com/gorilla/websocket"
	"log"
)

func NewConnWriter(conn *websocket.Conn, messageType int, message []byte) {
	// message the client
	if err := conn.WriteMessage(messageType, message); err != nil {
		log.Println(err)
		}
}
