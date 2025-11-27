package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	"github.com/ChethiyaNishanath/market-data-hub/internal/events"
	marketstate "github.com/ChethiyaNishanath/market-data-hub/internal/market-state"
	ws "github.com/ChethiyaNishanath/market-data-hub/internal/ws"

	"github.com/coder/websocket"
)

type SymbolState struct {
	OrderBook     *OrderBook
	updateCh      chan DepthUpdateEvent
	buffer        []DepthUpdateEvent
	bufferMu      sync.Mutex
	snapshotReady chan struct{}
}

type Streamer struct {
	ctx     context.Context
	bus     *events.Bus
	Symbols map[string]*SymbolState
	mu      sync.RWMutex
	config  config.BinanceConfig
}

type SubscribeAck struct {
	ID     int    `json:"id"`
	Result string `json:"result"`
}

func NewStreamer(ctx context.Context, bus *events.Bus, cfg config.BinanceConfig) *Streamer {
	return &Streamer{
		ctx:    ctx,
		bus:    bus,
		config: cfg,
	}
}

func (s *Streamer) Start(ctx context.Context) {

	symbols := strings.Split(s.config.Subscriptions, ",")

	s.Symbols = make(map[string]*SymbolState)

	for _, sym := range symbols {

		symbol := strings.TrimSpace(sym)
		symbol = strings.ToUpper(symbol)

		if symbol == "" {
			continue
		}

		st := &SymbolState{
			OrderBook:     nil,
			updateCh:      make(chan DepthUpdateEvent, 50000),
			buffer:        make([]DepthUpdateEvent, 0, 1000),
			snapshotReady: make(chan struct{}),
		}

		s.mu.Lock()
		s.Symbols[symbol] = st
		s.mu.Unlock()

		go s.streamDepthUpdates(ctx, symbol, st.updateCh, &st.buffer, &st.bufferMu, st.snapshotReady)

		go s.initializeSymbol(ctx, symbol, st)
	}
}

func (s *Streamer) initializeSymbol(ctx context.Context, symbol string, st *SymbolState) {

	snapshot, err := FetchSnapshot(symbol, s.config)
	if err != nil {
		slog.Error("Snapshot load failed", "symbol", symbol, "error", err)
		return
	}

	s.mu.Lock()
	st.OrderBook = snapshot
	st.OrderBook.Initialized = false
	s.mu.Unlock()

	slog.Info("Snapshot loaded", "symbol", symbol, "lastUpdateId", snapshot.LastUpdateID)

	close(st.snapshotReady)

	time.Sleep(100 * time.Millisecond)

	st.bufferMu.Lock()
	bufferedCopy := make([]DepthUpdateEvent, len(st.buffer))
	copy(bufferedCopy, st.buffer)
	st.bufferMu.Unlock()

	firstApplied := false
	for _, update := range bufferedCopy {
		s.mu.Lock()
		if !firstApplied {
			if update.FirstUpdateEventID <= st.OrderBook.LastUpdateID+1 &&
				update.FinalUpdateEventID >= st.OrderBook.LastUpdateID {
				s.applyDeltaUnsafe(symbol, update, st)
				st.OrderBook.LastUpdateID = update.FinalUpdateEventID
				st.OrderBook.Initialized = true
				marketstate.NewDataStore()
				firstApplied = true
				slog.Info("Order book synchronized live stream in sync", "symbol", symbol,
					"lastUpdateId", st.OrderBook.LastUpdateID)
			}
			continue
		}

		if update.FirstUpdateEventID == st.OrderBook.LastUpdateID+1 {
			s.applyDeltaUnsafe(symbol, update, st)
			st.OrderBook.LastUpdateID = update.FinalUpdateEventID
		} else {
			slog.Warn("Gap detected in buffered updates",
				"symbol", symbol,
				"expected", st.OrderBook.LastUpdateID+1,
				"got", update.FirstUpdateEventID)
		}
		s.mu.Unlock()
	}

	go s.applyDepthEvents(ctx, symbol, st)
}

func (s *Streamer) streamDepthUpdates(
	ctx context.Context,
	symbol string,
	out chan<- DepthUpdateEvent,
	buffered *[]DepthUpdateEvent,
	bufferMu *sync.Mutex,
	snapshotReady <-chan struct{},
) {

	isBuffering := true

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		client := ws.NewWSClient(ctx, s.config.WsStreamUrl)

		client.OnMessage = func(mt websocket.MessageType, data []byte) {
			var update DepthUpdateEvent
			if err := json.Unmarshal(data, &update); err == nil && update.EventType == "depthUpdate" {

				select {
				case <-snapshotReady:
					isBuffering = false
				default:
				}

				if isBuffering {
					bufferMu.Lock()
					*buffered = append(*buffered, update)
					bufferMu.Unlock()
					return
				}

				select {
				case out <- update:
				case <-ctx.Done():
					return
				}

				return
			}

			var ack SubscribeAck
			if json.Unmarshal(data, &ack) == nil && ack.ID == 1 {
				slog.Info("Subscription ACK received", "symbol", symbol)
			}
		}

		if err := client.Connect(); err != nil {
			slog.Error("WS connect failed", "err", err)
			time.Sleep(time.Second)
			continue
		}

		stream := fmt.Sprintf("%s@depth", strings.ToLower(symbol))
		sub := map[string]any{
			"method": "SUBSCRIBE",
			"params": []string{stream},
			"id":     1,
		}
		if err := client.SendJSON(sub); err != nil {
			slog.Error("Binance subscribe failed", "error", err)
			client.Close()
			continue
		}

		slog.Info("Subscribed", "symbol", symbol)

		if err := client.BlockUntilClosed(); err != nil {
			slog.Warn("WS disconnected - reconnecting", "err", err)
			time.Sleep(time.Second)
			continue
		}
	}
}

func (s *Streamer) applyDepthEvents(ctx context.Context, symbol string, st *SymbolState) {
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-st.updateCh:
			if !ok {
				return
			}

			s.mu.Lock()
			U := update.FirstUpdateEventID
			u := update.FinalUpdateEventID
			last := st.OrderBook.LastUpdateID

			if !st.OrderBook.Initialized {
				if U <= last+1 && u >= last {
					s.applyDeltaUnsafe(symbol, update, st)
					st.OrderBook.LastUpdateID = u
					st.OrderBook.Initialized = true
					slog.Info("Order book synchronized live stream in sync", "symbol", symbol,
						"lastUpdateId", st.OrderBook.LastUpdateID)
					s.mu.Unlock()

					s.broadcastDepthUpdate(update)
					continue
				}
				s.mu.Unlock()
				continue
			}

			if U == last+1 {
				s.applyDeltaUnsafe(symbol, update, st)
				st.OrderBook.LastUpdateID = u
				s.mu.Unlock()

				s.broadcastDepthUpdate(update)
				continue
			}

			slog.Warn("Orderbook desync detected: fetching new snapshot", "symbol", symbol,
				"expected", last+1, "got", U)
			s.mu.Unlock()

			snapshot, err := FetchSnapshot(symbol, s.config)
			if err != nil {
				slog.Error("Snapshot reload failed", "error", err)
				continue
			}

			s.mu.Lock()
			st.OrderBook.ApplySnapshot(snapshot)
			st.OrderBook.Initialized = false
			slog.Info("Snapshot resynced", "lastUpdateId", snapshot.LastUpdateID)
			s.mu.Unlock()
		}
	}
}

func (s *Streamer) applyDeltaUnsafe(symbol string, update DepthUpdateEvent, st *SymbolState) {

	for _, bid := range update.BidsToUpdated {
		price := bid[0]
		quantity := bid[1]

		qty, err := strconv.ParseFloat(quantity, 64)
		if err != nil {
			slog.Error("Failed to convert string to int:", "error", err)
		}

		if qty == 0 {
			st.OrderBook.removeBid(price)
		} else {
			st.OrderBook.updateBid(price, quantity)
		}
	}

	for _, ask := range update.AsksToUpdated {
		price := ask[0]
		quantity := ask[1]

		qty, err := strconv.ParseFloat(quantity, 64)
		if err != nil {
			slog.Error("Failed to convert string to int:", "error", err)
		}

		if qty == 0 {
			st.OrderBook.removeAsk(price)
		} else {
			st.OrderBook.updateAsk(price, quantity)
		}
	}

	marketstate.GetOrderBookStore().SetItem(symbol, (*marketstate.OrderBook)(st.OrderBook))

	st.OrderBook.LastUpdateID = update.FinalUpdateEventID
}

func (s *Streamer) broadcastDepthUpdate(update DepthUpdateEvent) {

	event := DepthStreamEvent{
		EventType:          "depthUpdate",
		EventTime:          update.EventTime,
		Symbol:             update.Symbol,
		FirstUpdateEventID: update.FirstUpdateEventID,
		FinalUpdateEventID: update.FinalUpdateEventID,
		BidsToUpdated:      update.BidsToUpdated,
		AsksToUpdated:      update.AsksToUpdated,
	}

	s.bus.Publish("depthUpdate", fmt.Sprintf("%s@depth", strings.ToLower(update.Symbol)), event)
}
