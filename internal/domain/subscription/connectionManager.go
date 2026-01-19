package subscription

import (
	"github.com/coder/websocket"
)

type ClientConnectionManager interface {
	Register(client Client)
	Unregister(client Client)
	Subscribe(client Client, topic string)
	Unsubscribe(client Client, topic string)
	Broadcast(topic string, msg any)
	GetClient(conn *websocket.Conn) Client
	GetClientByID(id string) Client
}
