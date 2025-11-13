package nateda

import (
	"sync"
)

// EventType represents the type of NAT event
type EventType string

const (
	EventTranslate   EventType = "translate"
	EventMappingAdd  EventType = "mapping_add"
	EventMappingDel  EventType = "mapping_del"
	EventError       EventType = "error"
)

// Event represents a NAT event in the system
type Event struct {
	Type    EventType
	Payload interface{}
}

// EventHandler is a function that handles events
type EventHandler func(Event)

// EventBus is a simple event bus for event-driven architecture
type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish publishes an event to all registered handlers
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}
