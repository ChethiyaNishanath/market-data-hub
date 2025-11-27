package binance

import (
	"context"
	"strings"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	"github.com/ChethiyaNishanath/market-data-hub/internal/events"
	ws "github.com/ChethiyaNishanath/market-data-hub/internal/ws"
)

type Module struct {
	Config  config.BinanceConfig
	ConnMgr *ws.ConnectionManager
}

func NewModule(ctx *context.Context, bus *events.Bus, connMger *ws.ConnectionManager, cfg config.BinanceConfig) *Module {

	module := &Module{
		Config:  cfg,
		ConnMgr: connMger,
	}

	module.registerEventSubscribers(bus)
	return module

}

func (m *Module) registerEventSubscribers(bus *events.Bus) {

	symbols := strings.SplitSeq(strings.ToLower(m.Config.Subscriptions), ",")

	for symbol := range symbols {

		cleaned := strings.TrimSpace(symbol)
		cleaned = strings.ToLower(cleaned)
		if cleaned == "" {
			continue
		}

		depthTopic := cleaned + "@depth"
		resetTopic := cleaned + "@depth.reset"

		bus.Subscribe(depthTopic, func(e events.Event) {
			evt := e.Data.(DepthStreamEvent)
			m.ConnMgr.Broadcast(
				context.Background(),
				e.Topic,
				ws.WSMessage{
					Data: evt,
				},
			)
		})

		bus.Subscribe(resetTopic, func(e events.Event) {
			evt := e.Data.(OrderBookResetEvent)
			m.ConnMgr.Broadcast(
				context.Background(),
				e.Topic,
				ws.WSMessage{
					Method: "orderbook_reset",
					Data:   evt,
				},
			)
		})
	}
}
