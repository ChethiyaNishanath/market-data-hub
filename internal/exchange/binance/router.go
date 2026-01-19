package binance

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	wsInterface "github.com/ChethiyaNishanath/market-data-hub/internal/interfaces/clients/websocket"
	"github.com/coder/websocket"
)

type HandlerFunc func(ctx context.Context, conn *websocket.Conn, payload json.RawMessage)

type Router struct {
	Routes map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		Routes: make(map[string]HandlerFunc),
	}
}

func (r *Router) Handle(action string, handler HandlerFunc) {
	r.Routes[action] = handler
}

func (r *Router) Dispatch(ctx context.Context, conn *websocket.Conn, msg WSRequest) {
	handler, ok := r.Routes[strings.ToLower(msg.Method)]
	if !ok {
		slog.Warn("Unknown WebSocket action", "action", msg.Method)

		wsmsg := WSMessage{
			Method:  msg.Method,
			Success: false,
			Error:   "unknown action",
		}

		err := wsInterface.WriteJSON(ctx, conn, wsmsg)
		if err != nil {
			return
		}
		return
	}
	handler(ctx, conn, msg.Params)
}
