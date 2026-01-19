package binance

import (
	"context"
	"fmt"
	"time"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	"github.com/ChethiyaNishanath/market-data-hub/internal/interfaces/clients/rest"
)

func FetchSnapshot(symbol string, cfg config.BinanceConfig) (*OrderBookSnapshot, error) {
	restClient := rest.New(cfg.RestApiUrlV3, 1*time.Second)

	ctx := context.Background()

	requestOpts := rest.RequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	var orderBook OrderBookSnapshot

	path := fmt.Sprintf("/depth?symbol=%s&limit=1000", symbol)

	err := restClient.Get(ctx, path, requestOpts, &orderBook)

	if err != nil {
		return nil, err
	}

	return &orderBook, nil

}
