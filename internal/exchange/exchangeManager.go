package exchange

import (
	"context"

	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/orderbook"
)

type Manager interface {
	Start(ctx context.Context)
	GetOrderBook(symbol string) *orderbook.OrderBook
}
