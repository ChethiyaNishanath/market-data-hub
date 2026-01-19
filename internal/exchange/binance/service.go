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

	"github.com/ChethiyaNishanath/market-data-hub/internal/bus"
	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/exchange"
	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/orderbook"
	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/subscription"
	wsInterface "github.com/ChethiyaNishanath/market-data-hub/internal/interfaces/clients/websocket"
	"github.com/ChethiyaNishanath/market-data-hub/internal/store/memory"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Service struct {
	ctx           context.Context
	bus           bus.IBus
	Symbols       map[string]*SymbolState
	config        config.BinanceConfig
	clientConnMgr subscription.ClientConnectionManager
}

type SubscribeAck struct {
	ID     string `json:"id"`
	Result string `json:"result"`
}

func NewService(ctx context.Context, bus bus.IBus, connMgr subscription.ClientConnectionManager, cfg config.BinanceConfig) *Service {
	return &Service{
		ctx:           ctx,
		bus:           bus,
		config:        cfg,
		clientConnMgr: connMgr,
	}
}

func (s *Service) Start(ctx context.Context) {

	s.RegisterEventSubscribers(s.config, s.clientConnMgr, s.bus)

	symbols := strings.Split(s.config.Subscriptions, ",")

	s.Symbols = make(map[string]*SymbolState)

	wg := &sync.WaitGroup{}
	wg.Add(len(symbols))
	validSymbols := make([]string, 0)

	for _, sym := range symbols {
		symbol := strings.TrimSpace(sym)
		symbol = strings.ToUpper(symbol)
		if symbol != "" {
			validSymbols = append(validSymbols, symbol)
		}
	}

	for _, symbol := range validSymbols {
		st := NewMarketState()
		s.Symbols[symbol] = st
		go s.streamDepthUpdates(ctx, symbol, st.UpdateCh, st.SnapshotReady, wg)
	}

	slog.Info("Waiting for all WebSocket connections to be ready", "symbols", s.config.Subscriptions)
	wg.Wait()
	slog.Info("All WebSocket connections ready, starting snapshot fetches")

	for _, symbol := range validSymbols {
		st := s.Symbols[symbol]
		go s.initializeSymbol(ctx, symbol, st)
	}
}

func (s *Service) initializeSymbol(ctx context.Context, symbol string, st *SymbolState) {

	snapshot, err := FetchSnapshot(symbol, s.config)
	if err != nil {
		slog.Error("Snapshot load failed", "symbol", symbol, "error", err)
		return
	}

	st.OrderBook = snapshot
	st.OrderBook.Initialized = false

	slog.Info("Snapshot loaded", "symbol", symbol, "lastUpdateId", st.OrderBook.LastUpdateID)

	close(st.SnapshotReady)
	firstApplied := false

	for {
		select {
		case update := <-st.UpdateCh:
			if !firstApplied {
				if update.FirstUpdateEventID <= st.OrderBook.LastUpdateID+1 &&
					update.FinalUpdateEventID >= st.OrderBook.LastUpdateID {
					s.applyDelta(symbol, update, st)
					st.OrderBook.LastUpdateID = update.FinalUpdateEventID
					st.OrderBook.Initialized = true
					firstApplied = true
					slog.Info("Order book synchronized live stream in sync", "symbol", symbol,
						"lastUpdateId", st.OrderBook.LastUpdateID)
				}
				continue
			}

			if update.FirstUpdateEventID == st.OrderBook.LastUpdateID+1 {
				s.applyDelta(symbol, update, st)
				st.OrderBook.LastUpdateID = update.FinalUpdateEventID
			} else {
				slog.Warn("Gap detected in buffered updates",
					"symbol", symbol,
					"expected", st.OrderBook.LastUpdateID+1,
					"got", update.FirstUpdateEventID)
			}
		default:
			go s.applyDepthEvents(ctx, symbol, st)
			return
		}
	}
}

func (s *Service) streamDepthUpdates(
	ctx context.Context,
	symbol string,
	updateCh chan<- DepthUpdateMessage,
	snapshotReady <-chan struct{},
	wg *sync.WaitGroup,
) {

	internalRequestId, _ := uuid.NewUUID()

	go func() {
		select {
		case <-snapshotReady:
			slog.Debug("Buffering stopped", "symbol", symbol)
		case <-ctx.Done():
			return
		}
	}()

	for {

		readyCh := make(chan struct{})
		wsReadyOnce := sync.Once{}

		go func() {
			select {
			case <-readyCh:
				wg.Done()
			case <-ctx.Done():
				wg.Done()
			}
		}()

		newStream := exchange.New(s.config.WsStreamUrl)
		client := wsInterface.New(ctx, newStream)

		client.OnMessage = func(mt websocket.MessageType, data []byte) {

			var ack SubscribeAck
			if json.Unmarshal(data, &ack) == nil && ack.ID == internalRequestId.String() {
				wsReadyOnce.Do(func() {
					close(readyCh)
					slog.Info("WebSocket ready", "symbol", symbol)
				})
			}

			var update DepthUpdateMessage
			if err := json.Unmarshal(data, &update); err == nil && update.EventType == OrderBookUpdate {
				select {
				case updateCh <- update:
				default:
					slog.Warn("Dropping depth update", "symbol", symbol)
				}
				return
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
			"id":     internalRequestId.String(),
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

func (s *Service) applyDepthEvents(ctx context.Context, symbol string, st *SymbolState) {
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-st.UpdateCh:
			if !ok {
				return
			}

			U := update.FirstUpdateEventID
			u := update.FinalUpdateEventID
			last := st.OrderBook.LastUpdateID

			if !st.OrderBook.Initialized {
				if U <= last+1 && u >= last {
					s.applyDelta(symbol, update, st)
					st.OrderBook.LastUpdateID = u
					st.OrderBook.Initialized = true
					slog.Info("Order book synchronized live stream in sync", "symbol", symbol,
						"lastUpdateId", st.OrderBook.LastUpdateID)

					s.broadcastDepthUpdate(update)
					continue
				}
				continue
			}

			if U == last+1 {
				s.applyDelta(symbol, update, st)
				st.OrderBook.LastUpdateID = u

				s.broadcastDepthUpdate(update)
				continue
			}

			slog.Warn("Order book de-sync detected: fetching new snapshot", "symbol", symbol,
				"expected", last+1, "got", U)

			snapshot, err := FetchSnapshot(symbol, s.config)
			if err != nil {
				slog.Error("Snapshot reload failed", "error", err)
				continue
			}

			st.OrderBook.ApplySnapshot(snapshot)
			st.OrderBook.Initialized = false
			slog.Info("Snapshot resynced", "lastUpdateId", st.OrderBook)
			orderBookSnapshot := st.OrderBook.ToOrderBook()
			s.BroadcastOrderBookReset(symbol, "Orderbook desync detected", orderBookSnapshot)
		}
	}
}

func (s *Service) applyDelta(symbol string, update DepthUpdateMessage, st *SymbolState) {

	for _, bid := range update.BidsToUpdated {
		price := bid[0]
		quantity := bid[1]

		qty, err := strconv.ParseFloat(quantity, 64)
		if err != nil {
			slog.Error("Failed to convert string to int:", "error", err)
		}

		if qty == 0 {
			st.OrderBook.RemoveBid(price)
		} else {
			st.OrderBook.UpdateBid(price, quantity)
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
			st.OrderBook.RemoveAsk(price)
		} else {
			st.OrderBook.UpdateAsk(price, quantity)
		}
	}

	orderBookSnapshot := st.OrderBook.ToOrderBook()

	memory.GetOrderBookStore().SetItem(symbol, &orderBookSnapshot)

	st.OrderBook.LastUpdateID = update.FinalUpdateEventID
}

func (s *Service) broadcastDepthUpdate(update DepthUpdateMessage) {

	event := DepthUpdateEvent{
		EventType:          "depthUpdate",
		EventTime:          update.EventTime,
		Symbol:             update.Symbol,
		FirstUpdateEventID: update.FirstUpdateEventID,
		FinalUpdateEventID: update.FinalUpdateEventID,
		BidsToUpdated:      update.BidsToUpdated,
		AsksToUpdated:      update.AsksToUpdated,
	}

	s.bus.Publish(OrderBookUpdate, fmt.Sprintf("%s@depth", strings.ToLower(update.Symbol)), event)
}

func (s *Service) GetOrderBook(symbol string) *orderbook.OrderBook {
	symbolBook, ok := s.Symbols[symbol]

	if s.Symbols == nil || !ok {
		return nil
	}

	ob := *symbolBook.OrderBook
	snapshot := ob.ToOrderBook()
	return &snapshot
}

func (s *Service) BroadcastOrderBookReset(symbol string, reason string, orderBook orderbook.OrderBook) {

	event := OrderBookResetEvent{
		Symbol:    symbol,
		Snapshot:  orderBook,
		Reason:    reason,
		Timestamp: time.Now().Unix(),
	}

	s.bus.Publish(OrderBookReset, fmt.Sprintf("%s@depth.reset", strings.ToLower(symbol)), event)
}

func (s *Service) RegisterEventSubscribers(config config.BinanceConfig, connMgr subscription.ClientConnectionManager, eventBus bus.IBus) {

	symbols := strings.SplitSeq(strings.ToLower(config.Subscriptions), ",")

	for symbol := range symbols {

		cleaned := strings.TrimSpace(symbol)
		cleaned = strings.ToLower(cleaned)
		if cleaned == "" {
			continue
		}

		depthTopic := cleaned + "@depth"
		resetTopic := cleaned + "@depth.reset"

		eventBus.Subscribe(depthTopic, func(e bus.Event) {
			evt := e.Data.(DepthUpdateEvent)
			connMgr.Broadcast(e.Topic, WSMessage{
				Data: evt,
			})
		})

		eventBus.Subscribe(resetTopic, func(e bus.Event) {
			evt := e.Data.(OrderBookResetEvent)
			connMgr.Broadcast(e.Topic, WSMessage{
				Method: "orderbook_reset",
				Data:   evt,
			})
		})
	}
}
