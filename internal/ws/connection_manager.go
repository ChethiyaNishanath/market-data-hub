package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Client struct {
	ID     uuid.UUID
	Conn   *websocket.Conn
	SendCh chan []byte
	Topics map[string]bool
}

type ConnectionManager struct {
	ctx           context.Context
	mu            sync.RWMutex
	clients       map[string]*Client
	subscriptions map[string]map[string]*Client
}

func NewConnectionManager(ctx context.Context) *ConnectionManager {
	return &ConnectionManager{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[string]map[string]*Client),
	}
}

func (m *ConnectionManager) Register(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client.ID.String()] = client
	slog.Info(fmt.Sprintf("Client registered, total: %d", len(m.clients)))
}

func (m *ConnectionManager) Unregister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.clients, client.ID.String())

	for topic, subs := range m.subscriptions {
		delete(subs, client.ID.String())
		if len(subs) == 0 {
			delete(m.subscriptions, topic)
		}
	}

	client.Conn.Close(websocket.StatusNormalClosure, "client disconnected")
	slog.Info(fmt.Sprintf("Client unregistered, total: %d", len(m.clients)))
}

func (m *ConnectionManager) Subscribe(client *Client, topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.subscriptions[topic]; !ok {
		m.subscriptions[topic] = make(map[string]*Client)
	}

	m.subscriptions[topic][client.ID.String()] = client
	client.Topics[topic] = true
}

func (m *ConnectionManager) Unsubscribe(client *Client, topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if subs, ok := m.subscriptions[topic]; ok {
		delete(subs, client.ID.String())
		if len(subs) == 0 {
			delete(m.subscriptions, topic)
		}
	}
	delete(client.Topics, topic)
}

func (m *ConnectionManager) Broadcast(ctx context.Context, topic string, msg any) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := m.subscriptions[topic]
	if len(clients) == 0 {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		slog.Warn("Failed to marshal broadcast", "warning", err)
		return
	}

	for _, client := range clients {
		select {
		case client.SendCh <- data:
		default:
			slog.Warn("send buffer full, dropping message for client", "client_id", client.ID)
		}
	}
}

func (m *ConnectionManager) GetClient(conn *websocket.Conn) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		if client.Conn == conn {
			return client
		}
	}
	return nil
}

func (m *ConnectionManager) GetClientByID(id string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[id]
}

func (c *Client) WritePump(ctx context.Context) {
	for {
		select {
		case msg, ok := <-c.SendCh:
			if !ok {
				return
			}

			if err := c.Conn.Write(ctx, websocket.MessageText, msg); err != nil {
				slog.Error("write error:", "error", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) ReadPump(ctx context.Context, m *ConnectionManager) {
	defer func() {
		m.Unregister(c)
		slog.Info("Client connection closed")
	}()

	<-ctx.Done()

	c.Conn.Close(websocket.StatusNormalClosure, "context cancelled")
}
