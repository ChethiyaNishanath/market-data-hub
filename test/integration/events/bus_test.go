package events_test

import (
	"sync"
	"testing"
	"time"

	events "github.com/ChethiyaNishanath/market-data-hub/internal/bus"
)

func wait(wg *sync.WaitGroup) bool {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()

	select {
	case <-ch:
		return true
	case <-time.After(2 * time.Second):
		return false
	}
}

func TestPublishSingleSubscriber(t *testing.T) {
	bus := events.New()

	var wg sync.WaitGroup
	wg.Add(1)

	var received events.Event

	bus.Subscribe("btcusd@depth", func(e events.Event) {
		received = e
		wg.Done()
	})

	bus.Publish("depthUpdate", "btcusd@depth", "ABC123")

	waitDone := wait(&wg)
	if !waitDone {
		t.Fatalf("subscriber did not receive event")
	}

	if received.Action != "depthUpdate" {
		t.Errorf("expected action 'depthUpdate', got %s", received.Action)
	}
	if received.Topic != "btcusd@depth" {
		t.Errorf("expected topic 'btcusd@depth', got %s", received.Topic)
	}
	if received.Data != "ABC123" {
		t.Errorf("expected data 'ABC123', got %v", received.Data)
	}
}

func TestPublishMultipleSubscribers(t *testing.T) {
	bus := events.New()

	var wg sync.WaitGroup
	wg.Add(2)

	count := 0
	mu := sync.Mutex{}

	sub := func(e events.Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	}

	bus.Subscribe("btcusd@depth", sub)
	bus.Subscribe("bnbbtc@depth", sub)

	go func() {
		time.Sleep(time.Second * 1)
		bus.Publish("depthUpdate", "btcusd@depth", 99)
	}()
	go func() { bus.Publish("depthUpdate", "bnbbtc@depth", 100) }()

	waitDone := wait(&wg)

	if !waitDone {
		t.Fatalf("subscriber did not receive event")
	}

	if count != 2 {
		t.Errorf("expected 2 subscribers, got %d", count)
	}
}

func TestSubscribeThreadSafe(t *testing.T) {
	bus := events.New()

	var wg sync.WaitGroup
	wg.Add(10)

	for range 10 {
		go func() {
			bus.Subscribe("depthUpdate", func(e events.Event) {})
			wg.Done()
		}()
	}

	if !wait(&wg) {
		t.Fatalf("concurrent Subscribe caused a race or deadlock")
	}
}

func TestPublishThreadSafe(t *testing.T) {
	bus := events.New()

	var received int
	var mu sync.Mutex

	bus.Subscribe("btcusd@depth", func(e events.Event) {
		mu.Lock()
		received++
		mu.Unlock()
	})

	var wg sync.WaitGroup
	wg.Add(50)

	for range 50 {
		go func() {
			bus.Publish("depthUpdate", "btcusd@depth", nil)
			wg.Done()
		}()
	}

	if !wait(&wg) {
		t.Fatalf("publish did not complete")
	}

	time.Sleep(50 * time.Millisecond)

	if received != 50 {
		t.Errorf("expected 50 bus received, got %d", received)
	}
}
