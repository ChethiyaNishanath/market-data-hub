package binance

import "github.com/ChethiyaNishanath/market-data-hub/internal/domain/orderbook"

type DepthUpdateEvent struct {
	EventType          string     `json:"e"`
	EventTime          int        `json:"E"`
	Symbol             string     `json:"s"`
	FirstUpdateEventID int        `json:"U"`
	FinalUpdateEventID int        `json:"u"`
	BidsToUpdated      [][]string `json:"b"`
	AsksToUpdated      [][]string `json:"a"`
}

type OrderBookResetEvent struct {
	Symbol    string              `json:"symbol"`
	Snapshot  orderbook.OrderBook `json:"snapshot"`
	Reason    string              `json:"reason"`
	Timestamp int64               `json:"timestamp"`
}
