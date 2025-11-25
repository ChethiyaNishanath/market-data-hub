package events

import (
	"sync"
)

type Event struct {
	Action string
	Topic  string
	Data   any
}

type Subscriber func(event Event)

type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]Subscriber
}

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]Subscriber),
	}
}

func (b *Bus) Publish(action string, topic string, data any) {
	b.mu.RLock()
	subs, ok := b.subscribers[topic]
	b.mu.RUnlock()

	if !ok {
		return
	}

	event := Event{Action: action, Topic: topic, Data: data}
	for _, sub := range subs {
		go sub(event)
	}
}

func (b *Bus) Subscribe(topic string, fn Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], fn)
}
