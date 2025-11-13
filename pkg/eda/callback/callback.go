package callback

import (
	"context"
	"fmt"
	"sync"

	"github.com/chengtb/chengtb/pkg/eda/event"
)

// CallbackFunc is a function type for event callbacks
type CallbackFunc func(ctx context.Context, evt event.Event) error

// CallbackHandler manages callback execution
type CallbackHandler struct {
	callbacks map[string][]CallbackFunc
	mu        sync.RWMutex
}

// NewCallbackHandler creates a new callback handler
func NewCallbackHandler() *CallbackHandler {
	return &CallbackHandler{
		callbacks: make(map[string][]CallbackFunc),
	}
}

// Register registers a callback for a specific event type
func (h *CallbackHandler) Register(eventType string, callback CallbackFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callbacks[eventType] = append(h.callbacks[eventType], callback)
}

// Execute executes all callbacks for a given event
func (h *CallbackHandler) Execute(ctx context.Context, evt event.Event) error {
	h.mu.RLock()
	callbacks := h.callbacks[evt.GetType()]
	h.mu.RUnlock()

	if len(callbacks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(callbacks))

	for _, cb := range callbacks {
		wg.Add(1)
		go func(callback CallbackFunc) {
			defer wg.Done()
			if err := callback(ctx, evt); err != nil {
				errChan <- fmt.Errorf("callback error for event %s: %w", evt.GetID(), err)
			}
		}(cb)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("callback execution errors: %v", errors)
	}

	return nil
}

// ExecuteSync executes all callbacks synchronously
func (h *CallbackHandler) ExecuteSync(ctx context.Context, evt event.Event) error {
	h.mu.RLock()
	callbacks := h.callbacks[evt.GetType()]
	h.mu.RUnlock()

	for _, cb := range callbacks {
		if err := cb(ctx, evt); err != nil {
			return fmt.Errorf("callback error for event %s: %w", evt.GetID(), err)
		}
	}

	return nil
}

// GetCallbackCount returns the number of registered callbacks for an event type
func (h *CallbackHandler) GetCallbackCount(eventType string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.callbacks[eventType])
}
