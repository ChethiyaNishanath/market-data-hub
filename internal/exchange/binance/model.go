package binance

import "github.com/ChethiyaNishanath/market-data-hub/internal/domain/orderbook"

type OrderBookSnapshot struct {
	LastUpdateID int        `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
	Initialized  bool       `json:"-"`
}

func (s *OrderBookSnapshot) ApplySnapshot(snapshot *OrderBookSnapshot) {
	s.LastUpdateID = snapshot.LastUpdateID
	s.Bids = snapshot.Bids
	s.Asks = snapshot.Asks
}

func (s *OrderBookSnapshot) UpdateBid(price, qty string) {
	for i, b := range s.Bids {
		if b[0] == price {
			s.Bids[i][1] = qty
			return
		}
	}
	s.Bids = append(s.Bids, []string{price, qty})
}

func (s *OrderBookSnapshot) RemoveBid(price string) {
	for i, b := range s.Bids {
		if b[0] == price {
			s.Bids = append(s.Bids[:i], s.Bids[i+1:]...)
			return
		}
	}
}

func (s *OrderBookSnapshot) UpdateAsk(price, qty string) {
	for i, a := range s.Asks {
		if a[0] == price {
			s.Asks[i][1] = qty
			return
		}
	}
	s.Asks = append(s.Asks, []string{price, qty})
}

func (s *OrderBookSnapshot) RemoveAsk(price string) {
	for i, a := range s.Asks {
		if a[0] == price {
			s.Asks = append(s.Asks[:i], s.Asks[i+1:]...)
			return
		}
	}
}

func (s *OrderBookSnapshot) ToOrderBook() orderbook.OrderBook {
	return orderbook.OrderBook{
		LastUpdateID: s.LastUpdateID,
		Bids:         s.Bids,
		Asks:         s.Asks,
		Initialized:  s.Initialized,
	}
}

func (s *OrderBookSnapshot) ToSnapshot() *OrderBookSnapshot {
	return s
}
