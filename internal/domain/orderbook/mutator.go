package orderbook

type Mutator interface {
	ApplySnapshot(ob *OrderBook)
	UpdateBid(price, qty string)
	RemoveBid(price string)
	UpdateAsk(price, qty string)
	RemoveAsk(price string)
}
