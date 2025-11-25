package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Handler struct {
	router  *Router
	connMgr *ConnectionManager
}

func NewHandler(router *Router, connMgr *ConnectionManager) *Handler {
	return &Handler{router: router, connMgr: connMgr}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		slog.Error("Failed to accept websocket:", "error", err)
		return
	}

	client := &Client{
		ID:     uuid.New(),
		Conn:   conn,
		SendCh: make(chan []byte, 256),
		Topics: make(map[string]bool),
	}

	defer h.connMgr.Unregister(client)

	slog.Info("WebSocket client connected")
	h.connMgr.Register(client)

	conn.Write(context.Background(), websocket.MessageText, []byte(`{"client_id":"`+client.ID.String()+`"}`))

	ctx, cancel := context.WithCancel(h.connMgr.ctx)
	defer cancel()

	go client.WritePump(ctx)
	go client.ReadPump(ctx, h.connMgr)

	for {
		_, data, err := conn.Read(context.Background())
		if err != nil {
			slog.Error("Read error:", "error", err)
			break
		}

		var msg WSRequest
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Error("Invalid Paylaod:", "error", err)
			continue
		}

		slog.Debug("Received:", "data", string(data))

		go h.router.Dispatch(ctx, conn, msg)
	}
	conn.Close(websocket.StatusNormalClosure, "")
}
