package app

import (
	"context"

	"github.com/ChethiyaNishanath/market-data-hub/internal/bus"
	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	"github.com/ChethiyaNishanath/market-data-hub/internal/exchange/binance"
	"github.com/ChethiyaNishanath/market-data-hub/internal/store/memory"
	"github.com/ChethiyaNishanath/market-data-hub/internal/subcription"
	"github.com/go-chi/chi/v5"
)

type App struct {
	cfg              *config.Config
	WebSocketHandler *subcription.Handler
}

func NewApp(ctx *context.Context, cfg *config.Config) *App {

	memory.NewDataStore()
	eventBus := bus.New()
	connMgr := subcription.NewConnectionManager()

	subscriptionService := subcription.NewService(connMgr)
	binanceService := binance.NewService(*ctx, eventBus, connMgr, cfg.Integrations.Binance)

	go binanceService.Start(*ctx)

	return &App{
		cfg:              cfg,
		WebSocketHandler: subscriptionService.Handler,
	}
}

func (a *App) RegisterRoutes(r chi.Router) {
	r.Get("/ws", a.WebSocketHandler.HandleWebSocket)
}
