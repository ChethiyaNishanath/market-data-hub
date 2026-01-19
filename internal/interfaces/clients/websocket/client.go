package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/exchange"
	"github.com/coder/websocket"
)

type Client struct {
	model  exchange.Client
	conn   *websocket.Conn
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	OnMessage func(msgType websocket.MessageType, data []byte)
}

func New(cont context.Context, model exchange.Client) *Client {
	ctx, cancel := context.WithCancel(cont)
	return &Client{
		model:  model,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			return err
		}
	}

	var err error
	c.conn, _, err = websocket.Dial(c.ctx, c.model.URL, nil)
	if err != nil {
		return err
	}

	c.conn.SetReadLimit(5 * 1024 * 1024)

	go c.readLoop()
	go c.pingLoop()

	return nil
}

func (c *Client) readLoop() {
	for {
		msgType, data, err := c.conn.Read(c.ctx)
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

func (c *Client) pingLoop() {
	ticker := time.NewTicker(c.model.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			if c.conn != nil {
				err := c.conn.Ping(c.ctx)
				if err != nil {
					return
				}
			}
			c.mu.Unlock()

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Client) SendJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return c.conn.Write(c.ctx, websocket.MessageText, data)
}

func (c *Client) Close() {
	c.cancel()

	c.mu.Lock()
	if c.conn != nil {
		if err := c.conn.Close(websocket.StatusNormalClosure, "shutdown"); err != nil {
			return
		}
	}
	c.mu.Unlock()
}

func (c *Client) BlockUntilClosed() error {
	<-c.ctx.Done()
	return errors.New("connection closed")
}
