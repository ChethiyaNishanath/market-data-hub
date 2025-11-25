package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	marketstate "github.com/ChethiyaNishanath/market-data-hub/internal/market-state"
	"github.com/coder/websocket"
)

func (h *Handler) HandleSubscribe(ctx context.Context, conn *websocket.Conn, payload json.RawMessage) {
	var data struct {
		Topic string `json:"topic"`
	}

	if err := json.Unmarshal(payload, &data); err != nil || data.Topic == "" {
		h.writeError(conn, "subscribe", "Invalid payload or missing topic")
		return
	}

	client := h.connMgr.GetClient(conn)
	if client == nil {
		slog.Warn("Subscribe request from unknown client")
		return
	}

	h.connMgr.Subscribe(client, data.Topic)

	parts := strings.Split(data.Topic, "@")
	if len(parts) != 2 {
		h.writeError(conn, "subscribe", "Invalid topic format. Expect <symbol>@<event>")
		return
	}

	symbol := strings.ToUpper(parts[0])
	event := strings.ToLower(parts[1])

	switch event {
	case "depth":
		h.handleDepthSubscription(ctx, conn, symbol)
	default:
		h.writeError(conn, "subscribe", "Unsupported event type: "+event)
	}
}

func (h *Handler) HandleUnsubscribe(ctx context.Context, conn *websocket.Conn, payload json.RawMessage) {
	var data struct {
		Topic string `json:"topic"`
	}

	if err := json.Unmarshal(payload, &data); err != nil || data.Topic == "" {
		msg := WSMessage{
			Method:  "unsubscribe",
			Success: false,
			Error:   "invalid payload or missing topic",
		}
		WriteJSON(ctx, conn, msg)
		return
	}

	client := h.connMgr.GetClient(conn)
	if client == nil {
		slog.Warn("Unsubscribe request from unknown client")
		return
	}

	h.connMgr.Unsubscribe(client, data.Topic)

	msg := WSMessage{
		Method:  "unsubscribe",
		Success: true,
		Topic:   data.Topic,
	}

	WriteJSON(ctx, conn, msg)
}

func (h *Handler) handleDepthSubscription(ctx context.Context, conn *websocket.Conn, symbol string) {
	store := marketstate.GetOrderBookStore()
	snapshot, ok := store.GetItem(symbol)

	if !ok {
		h.writeError(conn, "subscribe", "Unknown symbol: "+symbol)
		return
	}

	msg := WSMessage{
		Method:  "subscribe",
		Topic:   symbol + "@depth",
		Success: true,
		Data:    snapshot,
	}
	WriteJSON(ctx, conn, msg)
}

func (h *Handler) writeError(conn *websocket.Conn, method, errMsg string) {
	msg := WSMessage{
		Method:  method,
		Success: false,
		Error:   errMsg,
	}
	WriteJSON(context.Background(), conn, msg)
}
