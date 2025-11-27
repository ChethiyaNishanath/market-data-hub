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

	symbols := strings.Split(strings.ToLower(m.Config.Subscriptions), ",")

	binanceDepthTopics := []string{}

	for _, symbol := range symbols {
		cleaned := strings.TrimSpace(symbol)
		cleaned = strings.ToLower(cleaned)

		if cleaned == "" {
			continue
		}
		binanceDepthTopics = append(binanceDepthTopics, cleaned+"@depth")
	}

	for _, sbdt := range binanceDepthTopics {

		bus.Subscribe(sbdt, func(e events.Event) {
			evt := e.Data.(DepthStreamEvent)
			m.ConnMgr.Broadcast(
				context.Background(),
				e.Topic,
				ws.WSMessage{
					Data: evt,
				},
			)
		})
	}
}
