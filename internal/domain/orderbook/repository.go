package orderbook

type Repository interface {
	GetAll() map[string]*OrderBook
	GetItem(symbol string) (*OrderBook, bool)
	SetItem(symbol string, book *OrderBook)
	DeleteItem(symbol string)
	Clear()
}
