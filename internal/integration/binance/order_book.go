package binance

import (
	"context"
	"fmt"
	"time"

	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	rest "github.com/ChethiyaNishanath/market-data-hub/internal/rest-client"
)

type OrderBook struct {
	LastUpdateID int        `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
	Initialized  bool       `json:"-"`
}

func (o *OrderBook) ApplySnapshot(snapshot *OrderBook) {
	o.LastUpdateID = snapshot.LastUpdateID
	o.Bids = snapshot.Bids
	o.Asks = snapshot.Asks
}

func (ob *OrderBook) updateBid(price, qty string) {
	for i, b := range ob.Bids {
		if b[0] == price {
			ob.Bids[i][1] = qty
			return
		}
	}
	ob.Bids = append(ob.Bids, []string{price, qty})
}

func (ob *OrderBook) removeBid(price string) {
	for i, b := range ob.Bids {
		if b[0] == price {
			ob.Bids = append(ob.Bids[:i], ob.Bids[i+1:]...)
			return
		}
	}
}

func (ob *OrderBook) updateAsk(price, qty string) {
	for i, a := range ob.Asks {
		if a[0] == price {
			ob.Asks[i][1] = qty
			return
		}
	}
	ob.Asks = append(ob.Asks, []string{price, qty})
}

func (ob *OrderBook) removeAsk(price string) {
	for i, a := range ob.Asks {
		if a[0] == price {
			ob.Asks = append(ob.Asks[:i], ob.Asks[i+1:]...)
			return
		}
	}
}

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

func (s *Streamer) GetOrderBook(symbol string) *OrderBook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	symbolBook, ok := s.Symbols[symbol]

	if s.Symbols == nil || !ok {
		return nil
	}

	ob := *symbolBook.OrderBook
	return &ob
}
