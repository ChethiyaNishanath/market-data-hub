package memory

import (
	"sync"

	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/orderbook"
)
// FEEDBACK: store/memory naming is completely generic no? i told you to not to use generic names for packages ,but use business terms
type OrderBookStore struct {
	mu    sync.RWMutex
	Books map[string]*orderbook.OrderBook
}

var (
	orderBookStore *OrderBookStore
	once           sync.Once
)
// FEEDBACK: the package name is store/memory, struct is OrderBookStore and create method is NewDataStore ---- the mistakes are repeating
func NewDataStore() *OrderBookStore {
	return &OrderBookStore{
		Books: make(map[string]*orderbook.OrderBook),
	}
}

func (d *OrderBookStore) SetItem(symbol string, book *orderbook.OrderBook) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Books[symbol] = book
}

func (d *OrderBookStore) GetItem(symbol string) (*orderbook.OrderBook, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	value, exists := d.Books[symbol]

	if !exists {
		return nil, false
	}

	cpy := *value
	return &cpy, exists
}

func GetOrderBookStore() *OrderBookStore { // FEEDBACK: Do not use global state via a singleton. Consider injecting dependencies instead.
	once.Do(func() {
		orderBookStore = &OrderBookStore{
			Books: make(map[string]*orderbook.OrderBook),
		}
	})
	return orderBookStore
}

func (d *OrderBookStore) GetAll() map[string]*orderbook.OrderBook {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]*orderbook.OrderBook, len(d.Books))
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
	d.Books = make(map[string]*orderbook.OrderBook)
}
