package eventbus

import (
	"context"
	"fmt"
	"sync"

	"github.com/chengtb/chengtb/pkg/eda/callback"
	"github.com/chengtb/chengtb/pkg/eda/event"
	"github.com/nats-io/nats.go"
)

// EventBus represents the NATS-based event bus
type EventBus struct {
	conn            *nats.Conn
	callbackHandler *callback.CallbackHandler
	subscriptions   map[string]*nats.Subscription
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// Config holds the configuration for the event bus
type Config struct {
	NATSUrl string
}

// NewEventBus creates a new NATS event bus
func NewEventBus(config Config) (*EventBus, error) {
	conn, err := nats.Connect(config.NATSUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &EventBus{
		conn:            conn,
		callbackHandler: callback.NewCallbackHandler(),
		subscriptions:   make(map[string]*nats.Subscription),
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Publish publishes an event to the event bus
func (eb *EventBus) Publish(evt event.Event) error {
	data, err := evt.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	subject := fmt.Sprintf("events.%s", evt.GetType())
	if err := eb.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Subscribe subscribes to events of a specific type
func (eb *EventBus) Subscribe(eventType string, handler func(event.Event) error) error {
	subject := fmt.Sprintf("events.%s", eventType)

	sub, err := eb.conn.Subscribe(subject, func(msg *nats.Msg) {
		evt, err := event.Unmarshal(msg.Data)
		if err != nil {
			fmt.Printf("Error unmarshaling event: %v\n", err)
			return
		}

		if err := handler(evt); err != nil {
			fmt.Printf("Error handling event: %v\n", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	eb.mu.Lock()
	eb.subscriptions[eventType] = sub
	eb.mu.Unlock()

	return nil
}

// SubscribeWithCallback subscribes to events with callback support
func (eb *EventBus) SubscribeWithCallback(eventType string, cb callback.CallbackFunc) error {
	// Register the callback
	eb.callbackHandler.Register(eventType, cb)

	// Check if already subscribed
	eb.mu.RLock()
	_, exists := eb.subscriptions[eventType]
	eb.mu.RUnlock()

	if exists {
		// Already subscribed, just added callback
		return nil
	}

	// Subscribe to the event type
	subject := fmt.Sprintf("events.%s", eventType)
	sub, err := eb.conn.Subscribe(subject, func(msg *nats.Msg) {
		evt, err := event.Unmarshal(msg.Data)
		if err != nil {
			fmt.Printf("Error unmarshaling event: %v\n", err)
			return
		}

		// Execute callbacks asynchronously
		if err := eb.callbackHandler.Execute(eb.ctx, evt); err != nil {
			fmt.Printf("Error executing callbacks: %v\n", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	eb.mu.Lock()
	eb.subscriptions[eventType] = sub
	eb.mu.Unlock()

	return nil
}

// Unsubscribe unsubscribes from an event type
func (eb *EventBus) Unsubscribe(eventType string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	sub, exists := eb.subscriptions[eventType]
	if !exists {
		return fmt.Errorf("no subscription found for event type: %s", eventType)
	}

	if err := sub.Unsubscribe(); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	delete(eb.subscriptions, eventType)
	return nil
}

// Close closes the event bus connection
func (eb *EventBus) Close() {
	eb.cancel()

	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Unsubscribe all
	for _, sub := range eb.subscriptions {
		sub.Unsubscribe()
	}

	if eb.conn != nil {
		eb.conn.Close()
	}
}

// IsConnected returns whether the event bus is connected to NATS
func (eb *EventBus) IsConnected() bool {
	return eb.conn != nil && eb.conn.IsConnected()
}
