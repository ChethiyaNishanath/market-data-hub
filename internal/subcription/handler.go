package subcription

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/subscription"
	"github.com/ChethiyaNishanath/market-data-hub/internal/exchange/binance"
	wsInterface "github.com/ChethiyaNishanath/market-data-hub/internal/interfaces/clients/websocket"
	wsserver "github.com/ChethiyaNishanath/market-data-hub/internal/interfaces/servers/websocket"
	orderBookStore "github.com/ChethiyaNishanath/market-data-hub/internal/store/memory"
	"github.com/coder/websocket"
)

type Handler struct {
	router  *binance.Router
	connMgr subscription.ClientConnectionManager
}

func NewHandler(router *binance.Router, connMgr subscription.ClientConnectionManager) *Handler {
	return &Handler{router: router, connMgr: connMgr}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		slog.Error("Failed to accept websocket:", "error", err)
		return
	}

	clientSubscription := wsserver.NewWsClient(conn)

	slog.Info("WebSocket clientSubscription connected")
	h.connMgr.Register(clientSubscription)
	defer h.connMgr.Unregister(clientSubscription)

	background := context.Background()
	if err := conn.Write(background, websocket.MessageText, []byte(`{"client_id":"`+clientSubscription.ID()+`"}`)); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(background)
	defer cancel()

	go clientSubscription.WritePump(ctx)
	go clientSubscription.ReadPump(ctx, h.connMgr)

	for {
		_, data, err := conn.Read(background)
		if err != nil {
			slog.Error("Read error:", "error", err)
			break
		}

		var msg binance.WSRequest
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Error("Invalid Payload:", "error", err)
			continue
		}

		slog.Debug("Received:", "data", string(data))

		go h.router.Dispatch(ctx, conn, msg)
	}

	if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		return
	}
}

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
		msg := binance.WSMessage{
			Method:  "unsubscribe",
			Success: false,
			Error:   "invalid payload or missing topic",
		}

		if writerErr := wsInterface.WriteJSON(ctx, conn, msg); writerErr != nil {
			return
		}

		return
	}

	client := h.connMgr.GetClient(conn)
	if client == nil {
		slog.Warn("Unsubscribe request from unknown client")
		return
	}

	h.connMgr.Unsubscribe(client, data.Topic)

	msg := binance.WSMessage{
		Method:  "unsubscribe",
		Success: true,
		Topic:   data.Topic,
	}

	err := wsInterface.WriteJSON(ctx, conn, msg)
	if err != nil {
		return
	}
}

func (h *Handler) handleDepthSubscription(ctx context.Context, conn *websocket.Conn, symbol string) {
	store := orderBookStore.GetOrderBookStore()
	snapshot, ok := store.GetItem(symbol)

	if !ok {
		h.writeError(conn, "subscribe", "Unknown symbol: "+symbol)
		return
	}

	msg := binance.WSMessage{
		Method:  "subscribe",
		Topic:   symbol + "@depth",
		Success: true,
		Data:    snapshot,
	}
	err := wsInterface.WriteJSON(ctx, conn, msg)
	if err != nil {
		return
	}
}

func (h *Handler) writeError(conn *websocket.Conn, method, errMsg string) {
	msg := binance.WSMessage{
		Method:  method,
		Success: false,
		Error:   errMsg,
	}
	err := wsInterface.WriteJSON(context.Background(), conn, msg)
	if err != nil {
		return
	}
}
