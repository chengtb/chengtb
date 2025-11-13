package event

import (
	"encoding/json"
	"time"
)

// Event represents the base interface for all events in the EDA system
type Event interface {
	// GetID returns the unique identifier of the event
	GetID() string
	// GetType returns the type/name of the event
	GetType() string
	// GetTimestamp returns when the event was created
	GetTimestamp() time.Time
	// GetPayload returns the event payload data
	GetPayload() interface{}
	// Marshal serializes the event to JSON
	Marshal() ([]byte, error)
}

// BaseEvent provides a default implementation of the Event interface
type BaseEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// NewBaseEvent creates a new base event
func NewBaseEvent(id, eventType string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// GetID returns the event ID
func (e *BaseEvent) GetID() string {
	return e.ID
}

// GetType returns the event type
func (e *BaseEvent) GetType() string {
	return e.Type
}

// GetTimestamp returns the event timestamp
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetPayload returns the event payload
func (e *BaseEvent) GetPayload() interface{} {
	return e.Payload
}

// Marshal serializes the event to JSON
func (e *BaseEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal deserializes JSON data into a BaseEvent
func Unmarshal(data []byte) (*BaseEvent, error) {
	var event BaseEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
