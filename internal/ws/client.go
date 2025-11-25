package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type WSClient struct {
	URL    string
	Conn   *websocket.Conn
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	OnMessage func(msgType websocket.MessageType, data []byte)

	PingInterval time.Duration
}

func NewWSClient(cont context.Context, url string) *WSClient {
	ctx, cancel := context.WithCancel(cont)
	return &WSClient{
		URL:          url,
		ctx:          ctx,
		cancel:       cancel,
		PingInterval: 20 * time.Second,
	}
}

func (c *WSClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn != nil {
		c.Conn.Close(websocket.StatusNormalClosure, "")
	}

	var err error
	c.Conn, _, err = websocket.Dial(c.ctx, c.URL, nil)
	if err != nil {
		return err
	}

	c.Conn.SetReadLimit(5 * 1024 * 1024)

	go c.readLoop()
	go c.pingLoop()

	return nil
}

func (c *WSClient) readLoop() {
	for {
		msgType, data, err := c.Conn.Read(c.ctx)
		if err != nil {
			slog.Error("WS read error", "error", err)
			c.cancel()
			return
		}

		if c.OnMessage != nil {
			c.OnMessage(msgType, data)
		}
	}
}

func (c *WSClient) pingLoop() {
	ticker := time.NewTicker(c.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			if c.Conn != nil {
				c.Conn.Ping(c.ctx)
			}
			c.mu.Unlock()

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *WSClient) SendJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return c.Conn.Write(c.ctx, websocket.MessageText, data)
}

func (c *WSClient) Close() {
	c.cancel()

	c.mu.Lock()
	if c.Conn != nil {
		c.Conn.Close(websocket.StatusNormalClosure, "shutdown")
	}
	c.mu.Unlock()
}

func (c *WSClient) BlockUntilClosed() error {
	<-c.ctx.Done()
	return errors.New("connection closed")
}
