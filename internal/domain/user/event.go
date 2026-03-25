package user

// Event is the base interface for all User domain events.
type Event interface {
	// AggregateID returns the id of the User that produced this event.
	AggregateID() string
	// EventType returns a stable, human-readable event name.
	EventType() string
}

// CreatedEvent is emitted after a new user is successfully persisted.
type CreatedEvent struct {
	UserID string
	Name   string
	Email  string
}

func (e CreatedEvent) AggregateID() string { return e.UserID }
func (e CreatedEvent) EventType() string   { return "user.created" }

// UpdatedEvent is emitted after user fields are modified.
type UpdatedEvent struct {
	UserID string
	Name   string
	Email  string
}

func (e UpdatedEvent) AggregateID() string { return e.UserID }
func (e UpdatedEvent) EventType() string   { return "user.updated" }

// DeletedEvent is emitted after a user is removed.
type DeletedEvent struct {
	UserID string
}

func (e DeletedEvent) AggregateID() string { return e.UserID }
func (e DeletedEvent) EventType() string   { return "user.deleted" }
