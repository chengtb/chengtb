package publisher

import (
	"fmt"

	"github.com/chengtb/chengtb/pkg/eda/event"
	"github.com/chengtb/chengtb/pkg/eda/eventbus"
)

// Publisher provides a convenient interface for publishing events
type Publisher struct {
	eventBus *eventbus.EventBus
}

// NewPublisher creates a new publisher
func NewPublisher(eventBus *eventbus.EventBus) *Publisher {
	return &Publisher{
		eventBus: eventBus,
	}
}

// Publish publishes an event
func (p *Publisher) Publish(evt event.Event) error {
	if err := p.eventBus.Publish(evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

// PublishWithValidation publishes an event with validation
func (p *Publisher) PublishWithValidation(evt event.Event) error {
	// Validate event
	if evt.GetID() == "" {
		return fmt.Errorf("event ID cannot be empty")
	}
	if evt.GetType() == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	return p.Publish(evt)
}
