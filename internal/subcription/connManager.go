package subcription

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/subscription"
	"github.com/coder/websocket"
)

type ConnectionManager struct {
	Ctx           context.Context
	mu            sync.RWMutex
	clients       map[string]subscription.Client
	subscriptions map[string]map[string]subscription.Client
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		Ctx:           context.Background(),
		clients:       make(map[string]subscription.Client),
		subscriptions: make(map[string]map[string]subscription.Client),
	}
}

func (m *ConnectionManager) Register(client subscription.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client.ID()] = client
	slog.Info(fmt.Sprintf("Client registered, total: %d", len(m.clients)))
}

func (m *ConnectionManager) Unregister(client subscription.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.clients, client.ID())

	for topic, subs := range m.subscriptions {
		delete(subs, client.ID())
		if len(subs) == 0 {
			delete(m.subscriptions, topic)
		}
	}

	if err := client.Conn().Close(websocket.StatusNormalClosure, "clientManager disconnected"); err != nil {
		slog.Error("Failed to close websocket connection", "error", err)
		return
	}
	slog.Info(fmt.Sprintf("Client unregistered, total: %d", len(m.clients)))
}

func (m *ConnectionManager) Subscribe(client subscription.Client, topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.subscriptions[topic]; !ok {
		m.subscriptions[topic] = make(map[string]subscription.Client)
	}

	m.subscriptions[topic][client.ID()] = client
	client.AddTopic(topic)
}

func (m *ConnectionManager) Unsubscribe(client subscription.Client, topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if subs, ok := m.subscriptions[topic]; ok {
		delete(subs, client.ID())
		if len(subs) == 0 {
			delete(m.subscriptions, topic)
		}
	}
	client.RemoveTopic(topic)
}

func (m *ConnectionManager) Broadcast(topic string, msg any) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subs, ok := m.subscriptions[topic]
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		slog.Warn("Failed to marshal broadcast", "warning", err)
		return
	}

	for _, client := range subs {
		client.Send(data)
	}
}

func (m *ConnectionManager) GetClient(conn *websocket.Conn) subscription.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		if client.Conn() == conn {
			return client
		}
	}
	return nil
}

func (m *ConnectionManager) GetClientByID(id string) subscription.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[id]
}
