package binance
// FEEDBACK : Why is this public. its not used outsode the package no?
type SymbolState struct {
	OrderBook     *OrderBookSnapshot
	UpdateCh      chan DepthUpdateMessage
	SnapshotReady chan struct{}
}

func NewMarketState() *SymbolState { // FEEDBACK: Why this is public
	return &SymbolState{
		OrderBook:     nil,
		UpdateCh:      make(chan DepthUpdateMessage, 100),
		SnapshotReady: make(chan struct{}),
	}
}
