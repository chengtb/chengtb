package subscriber

import (
	"fmt"

	"github.com/chengtb/chengtb/pkg/eda/callback"
	"github.com/chengtb/chengtb/pkg/eda/event"
	"github.com/chengtb/chengtb/pkg/eda/eventbus"
)

// Subscriber provides a convenient interface for subscribing to events
type Subscriber struct {
	eventBus *eventbus.EventBus
}

// NewSubscriber creates a new subscriber
func NewSubscriber(eventBus *eventbus.EventBus) *Subscriber {
	return &Subscriber{
		eventBus: eventBus,
	}
}

// Subscribe subscribes to events of a specific type
func (s *Subscriber) Subscribe(eventType string, handler func(event.Event) error) error {
	if err := s.eventBus.Subscribe(eventType, handler); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", eventType, err)
	}
	return nil
}

// SubscribeWithCallback subscribes with a callback function
func (s *Subscriber) SubscribeWithCallback(eventType string, cb callback.CallbackFunc) error {
	if err := s.eventBus.SubscribeWithCallback(eventType, cb); err != nil {
		return fmt.Errorf("failed to subscribe with callback to %s: %w", eventType, err)
	}
	return nil
}

// SubscribeMultiple subscribes to multiple event types with the same handler
func (s *Subscriber) SubscribeMultiple(eventTypes []string, handler func(event.Event) error) error {
	for _, eventType := range eventTypes {
		if err := s.Subscribe(eventType, handler); err != nil {
			return err
		}
	}
	return nil
}

// SubscribeWithMultipleCallbacks subscribes to an event type with multiple callbacks
func (s *Subscriber) SubscribeWithMultipleCallbacks(eventType string, callbacks []callback.CallbackFunc) error {
	for _, cb := range callbacks {
		if err := s.SubscribeWithCallback(eventType, cb); err != nil {
			return err
		}
	}
	return nil
}

// Unsubscribe unsubscribes from an event type
func (s *Subscriber) Unsubscribe(eventType string) error {
	if err := s.eventBus.Unsubscribe(eventType); err != nil {
		return fmt.Errorf("failed to unsubscribe from %s: %w", eventType, err)
	}
	return nil
}
