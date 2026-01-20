package bus

import (
	"sync"
)

type IBus interface {
	Publish(action string, topic string, data any)
	Subscribe(topic string, fn Subscriber)
}

type Subscriber func(event Event)

type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]Subscriber
}

func New() *Bus {
	return &Bus{
		subscribers: make(map[string][]Subscriber),
	}
}

func (b *Bus) Publish(action string, topic string, data any) {
	b.mu.RLock()
	subs, ok := b.subscribers[topic] // FEEDBACK: why do we need to lock here for reading subscribers? If subscriptions changes during running still the slice is not protected no
	b.mu.RUnlock()

	if !ok {
		return
	}

	event := Event{Action: action, Topic: topic, Data: data}
	for _, sub := range subs {
		go sub(event) // FEEDBACK: Creating a goroutine per message can lead to unbounded goroutine growth. Also it may create out-of-order processing issues.
	}
}

func (b *Bus) Subscribe(topic string, fn Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], fn) // FEEDBACK: using a functional interface for callback is not idiomatic Go
}
