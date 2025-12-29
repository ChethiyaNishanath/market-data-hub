package app

import (
	"context"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	events "github.com/ChethiyaNishanath/market-data-hub/internal/events"
	"github.com/ChethiyaNishanath/market-data-hub/internal/integration/binance"
	ws "github.com/ChethiyaNishanath/market-data-hub/internal/ws"
	"github.com/go-chi/chi/v5"
)

type App struct {
	cfg              *config.Config
	WebSocketHandler *ws.Handler
}

func NewApp(ctx *context.Context, cfg *config.Config) *App {

	bus := events.NewBus()
	connMgr := ws.NewConnectionManager(*ctx)

	websocketModule := ws.NewModule(ctx, connMgr, bus)

	streamer := binance.NewStreamer(*ctx, bus, cfg.Integrations.Binance)
	binance.NewModule(ctx, bus, connMgr, cfg.Integrations.Binance)

	go streamer.Start(*ctx)

	return &App{
		cfg:              cfg,
		WebSocketHandler: websocketModule.Handler,
	}
}

func (a *App) RegisterRoutes(r chi.Router) {
	r.Get("/ws", a.WebSocketHandler.HandleWebSocket)
}
