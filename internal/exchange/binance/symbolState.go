package binance

type SymbolState struct {
	OrderBook     *OrderBookSnapshot
	UpdateCh      chan DepthUpdateMessage
	SnapshotReady chan struct{}
}

func NewMarketState() *SymbolState {
	return &SymbolState{
		OrderBook:     nil,
		UpdateCh:      make(chan DepthUpdateMessage, 100),
		SnapshotReady: make(chan struct{}),
	}
}
