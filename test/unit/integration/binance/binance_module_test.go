package binance_test

import (
	"context"
	"testing"
	"time"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	events "github.com/ChethiyaNishanath/market-data-hub/internal/events"
	"github.com/ChethiyaNishanath/market-data-hub/internal/integration/binance"
	"github.com/ChethiyaNishanath/market-data-hub/internal/ws"
)

func TestNewBinanceModule(t *testing.T) {

	cfg := config.BinanceConfig{
		WsStreamUrl:   "ws://localhost:8080/ws",
		RestApiUrlV3:  "http://localhost:8080/rest",
		Subscriptions: "BNBBTC,BTCUSDT",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*1)*time.Second)

	defer cancel()

	bus := events.NewBus()
	connMgr := ws.NewConnectionManager(ctx)

	module := binance.NewModule(&ctx, bus, connMgr, cfg)

	if module == nil {
		t.Fatalf("Expected not nil")
	}
}
