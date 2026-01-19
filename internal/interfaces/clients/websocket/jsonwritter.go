package websocket

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/coder/websocket"
)

func WriteJSON(ctx context.Context, conn *websocket.Conn, v interface{}) error {

	data, err := json.Marshal(v)

	if err != nil {
		slog.Error("Failed to marshal websocket response", "error", err)
		return err
	}

	if err = conn.Write(ctx, websocket.MessageText, data); err != nil {
		slog.Error("Failed to write websocket message", "error", err)
		return err
	}

	slog.Debug("WebSocket message sent successfully", "data", string(data))
	return nil
}
