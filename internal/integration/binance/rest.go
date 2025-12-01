package binance

import (
	"context"
	"fmt"
	"time"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	rest "github.com/ChethiyaNishanath/market-data-hub/internal/rest-client"
)

func FetchSnapshot(symbol string, cfg config.BinanceConfig) (*OrderBook, error) {
	restClient := rest.NewRestClient(cfg.RestApiUrlV3, 1*time.Second)

	ctx := context.Background()

	requestOpts := rest.RequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	var orderbook OrderBook

	path := fmt.Sprintf("/depth?symbol=%s&limit=1000", symbol)

	err := restClient.Get(ctx, path, requestOpts, &orderbook)

	if err != nil {
		return nil, err
	}

	return &orderbook, nil

}

func FetchExchangeInfo(cfg config.BinanceConfig) (*ExchangeInfo, error) {
	restClient := rest.NewRestClient(cfg.RestApiUrlV3, 1*time.Second)

	ctx := context.Background()

	requestOpts := rest.RequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	var exchangeInfo ExchangeInfo

	path := "/exchangeInfo"

	err := restClient.Get(ctx, path, requestOpts, &exchangeInfo)

	if err != nil {
		return nil, err
	}

	return &exchangeInfo, nil

}
