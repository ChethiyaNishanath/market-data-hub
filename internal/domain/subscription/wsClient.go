package subscription

import (
	"context"

	"github.com/coder/websocket"
)

type Client interface {
	ID() string
	Conn() *websocket.Conn

	AddTopic(topic string)
	RemoveTopic(topic string)

	Send(data []byte)
	Close(reason string) error

	ReadPump(ctx context.Context, m ClientConnectionManager)
	WritePump(ctx context.Context)
}
