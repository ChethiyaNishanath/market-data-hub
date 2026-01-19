package binance_test

import (
	"context"
	"testing"
	"time"

	events "github.com/ChethiyaNishanath/market-data-hub/internal/bus"
	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	"github.com/ChethiyaNishanath/market-data-hub/internal/exchange/binance"
	"github.com/ChethiyaNishanath/market-data-hub/internal/subcription"
)

func TestNewBinanceModule(t *testing.T) {

	cfg := config.BinanceConfig{
		WsStreamUrl:   "ws://localhost:8080/ws",
		RestApiUrlV3:  "http://localhost:8080/rest",
		Subscriptions: "BNBBTC,BTCUSDT",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*1)*time.Second)

	defer cancel()

	bus := events.New()
	connMgr := subcription.NewConnectionManager()

	module := binance.NewService(ctx, bus, connMgr, cfg)

	if module == nil {
		t.Fatalf("Expected not nil")
	}
}
