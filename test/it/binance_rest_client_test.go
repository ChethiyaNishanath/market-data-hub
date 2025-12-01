package integration

import (
	"log/slog"
	"testing"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	binance "github.com/ChethiyaNishanath/market-data-hub/internal/integration/binance"
	"github.com/stretchr/testify/assert"
)

var cfg = config.BinanceConfig{
	RestApiUrlV3: "https://api.binance.com/api/v3",
}

func TestBinanceRestClientOrderbookSnapshot_Get(t *testing.T) {

	snapshot, err := binance.FetchSnapshot("BTCUSDT", cfg)
	if err != nil {
		slog.Error(err.Error())
	}
	assert.NotNil(t, snapshot)
}

func TestBinanceRestClientExchnageInfo_Get(t *testing.T) {

	exchnageInfo, err := binance.FetchExchangeInfo(cfg)
	if err != nil {
		slog.Error(err.Error())
	}
	assert.NotNil(t, exchnageInfo)
}
