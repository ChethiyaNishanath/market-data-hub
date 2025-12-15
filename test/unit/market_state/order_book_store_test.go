package marketstate_test

import (
	"testing"

	marketstate "github.com/ChethiyaNishanath/market-data-hub/internal/market-state"
)

func TestNewDataStore(t *testing.T) {
	store := marketstate.NewDataStore()

	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if len(store.Books) != 0 {
		t.Fatalf("expected empty store, got %d", len(store.Books))
	}
}

func TestSetAndGetItem(t *testing.T) {
	store := marketstate.NewDataStore()

	store.SetItem("BNBBTC", &marketstate.OrderBook{LastUpdateID: 12345})

	item, exists := store.GetItem("BNBBTC")
	if !exists {
		t.Fatalf("expected item to exist")
	}
	if item.LastUpdateID != 12345 {
		t.Errorf("expected LastUpdateID=12345, got %d", item.LastUpdateID)
	}
}

func TestGetItemNotExists(t *testing.T) {
	store := marketstate.NewDataStore()

	store.SetItem("BNBBTC", &marketstate.OrderBook{LastUpdateID: 1})

	_, exists := store.GetItem("ETHBTC")

	if exists {
		t.Fatalf("expected item not to exist")
	}
}

func TestGetAll(t *testing.T) {
	store := marketstate.NewDataStore()

	store.SetItem("BNBBTC", &marketstate.OrderBook{LastUpdateID: 1})
	store.SetItem("ETHBTC", &marketstate.OrderBook{LastUpdateID: 1})

	all := store.GetAll()

	if len(all) != 2 {
		t.Fatalf("expected 2 items, got %d", len(all))
	}

	all["BNBBTC"].LastUpdateID = 11111

	got, _ := store.GetItem("BNBBTC")

	if got.LastUpdateID == 11111 {
		t.Errorf("GetAll leaked internal state map")
	}
}

func TestDeleteItem(t *testing.T) {
	store := marketstate.NewDataStore()

	store.SetItem("BNBBTC", &marketstate.OrderBook{LastUpdateID: 1})

	store.DeleteItem("BNBBTC")

	_, exists := store.GetItem("BNBBTC")
	if exists {
		t.Fatal("expected BNBBTC to be deleted")
	}
}

func TestDeleteItemWithMultipleItemsInMap(t *testing.T) {
	store := marketstate.NewDataStore()

	store.SetItem("BNBBTC", &marketstate.OrderBook{LastUpdateID: 1})
	store.SetItem("ETHBTC", &marketstate.OrderBook{LastUpdateID: 1})

	store.DeleteItem("BNBBTC")

	_, exists_bnbbtc := store.GetItem("BNBBTC")
	if exists_bnbbtc {
		t.Fatal("expected BNBBTC to be deleted")
	}

	_, exists_ethbtc := store.GetItem("ETHBTC")
	if !exists_ethbtc {
		t.Fatal("expected ETHBTC to be exist")
	}
}
