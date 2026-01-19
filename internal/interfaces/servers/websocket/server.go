package websocket

import (
	"context"
	"log/slog"

	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/subscription"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type WSClient struct {
	id     uuid.UUID
	conn   *websocket.Conn
	sendCh chan []byte
	topics map[string]bool
}

func NewWsClient(conn *websocket.Conn) *WSClient {
	return &WSClient{
		id:     uuid.New(),
		conn:   conn,
		sendCh: make(chan []byte, 256),
		topics: make(map[string]bool),
	}
}

func (s *WSClient) ID() string {
	return s.id.String()
}

func (s *WSClient) Conn() *websocket.Conn {
	return s.conn
}

func (s *WSClient) AddTopic(topic string) {
	s.topics[topic] = true
}

func (s *WSClient) RemoveTopic(topic string) {
	delete(s.topics, topic)
}

func (s *WSClient) Send(data []byte) {
	select {
	case s.sendCh <- data:
	default:
		slog.Warn("send buffer full, dropping message", "client_id", s.ID())
	}
}

func (s *WSClient) Close(reason string) error {
	return s.conn.Close(websocket.StatusNormalClosure, reason)
}

func (s *WSClient) WritePump(ctx context.Context) {
	for {
		select {
		case msg, ok := <-s.sendCh:
			if !ok {
				return
			}

			if err := s.conn.Write(ctx, websocket.MessageText, msg); err != nil {
				slog.Error("write error:", "error", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *WSClient) ReadPump(ctx context.Context, m subscription.ClientConnectionManager) {
	defer func() {
		m.Unregister(s)
		slog.Info("Client connection closed")
	}()

	<-ctx.Done()

	if err := s.conn.Close(websocket.StatusNormalClosure, "context cancelled"); err != nil {
		return
	}
}
