package exchange

import (
	"context"

	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/orderbook"
)

type IStreamer interface {
	Start(ctx context.Context)
	GetOrderBook(symbol string) *orderbook.OrderBook
	BroadcastOrderBookReset(symbol string, reason string, snapshot *orderbook.OrderBook)
}
