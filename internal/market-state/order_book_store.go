package marketstate

import (
	"sync"
)

type OrderBookStore struct {
	mu    sync.RWMutex
	Books map[string]*OrderBook
}

var (
	orderBookStore *OrderBookStore
	once           sync.Once
)

func NewDataStore() *OrderBookStore {
	return &OrderBookStore{
		Books: make(map[string]*OrderBook),
	}
}

func (d *OrderBookStore) SetItem(symbol string, book *OrderBook) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Books[symbol] = book
}

func (d *OrderBookStore) GetItem(symbol string) (*OrderBook, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	value, exists := d.Books[symbol]

	if !exists {
		return nil, false
	}

	cpy := *value
	return &cpy, exists
}

func GetOrderBookStore() *OrderBookStore {
	once.Do(func() {
		orderBookStore = &OrderBookStore{
			Books: make(map[string]*OrderBook),
		}
	})
	return orderBookStore
}

func (d *OrderBookStore) GetAll() map[string]*OrderBook {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]*OrderBook, len(d.Books))
	for k, v := range d.Books {
		cpy := *v
		result[k] = &cpy
	}

	return result
}

func (d *OrderBookStore) DeleteItem(symbol string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.Books, symbol)
}

func (d *OrderBookStore) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Books = make(map[string]*OrderBook)
}
