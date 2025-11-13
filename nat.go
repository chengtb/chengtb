package nateda

import (
	"fmt"
	"sync"
)

// Address represents a network address
type Address struct {
	IP   string
	Port int
}

func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

// Mapping represents a NAT mapping between internal and external addresses
type Mapping struct {
	Internal Address
	External Address
}

// NATTable manages NAT address mappings
type NATTable struct {
	mu       sync.RWMutex
	mappings map[string]Mapping
	eventBus *EventBus
}

// NewNATTable creates a new NAT table with event bus
func NewNATTable(eventBus *EventBus) *NATTable {
	return &NATTable{
		mappings: make(map[string]Mapping),
		eventBus: eventBus,
	}
}

// AddMapping adds a new NAT mapping
func (nt *NATTable) AddMapping(internal, external Address) error {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	key := internal.String()
	if _, exists := nt.mappings[key]; exists {
		return fmt.Errorf("mapping already exists for %s", key)
	}

	mapping := Mapping{Internal: internal, External: external}
	nt.mappings[key] = mapping

	// Publish event
	if nt.eventBus != nil {
		nt.eventBus.Publish(Event{
			Type:    EventMappingAdd,
			Payload: mapping,
		})
	}

	return nil
}

// RemoveMapping removes a NAT mapping
func (nt *NATTable) RemoveMapping(internal Address) error {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	key := internal.String()
	mapping, exists := nt.mappings[key]
	if !exists {
		return fmt.Errorf("mapping not found for %s", key)
	}

	delete(nt.mappings, key)

	// Publish event
	if nt.eventBus != nil {
		nt.eventBus.Publish(Event{
			Type:    EventMappingDel,
			Payload: mapping,
		})
	}

	return nil
}

// Translate translates an internal address to external address
func (nt *NATTable) Translate(internal Address) (Address, error) {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	key := internal.String()
	mapping, exists := nt.mappings[key]
	if !exists {
		err := fmt.Errorf("no mapping found for %s", key)
		if nt.eventBus != nil {
			nt.eventBus.Publish(Event{
				Type:    EventError,
				Payload: err,
			})
		}
		return Address{}, err
	}

	// Publish event
	if nt.eventBus != nil {
		nt.eventBus.Publish(Event{
			Type:    EventTranslate,
			Payload: mapping,
		})
	}

	return mapping.External, nil
}

// GetMappings returns all current mappings
func (nt *NATTable) GetMappings() []Mapping {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	mappings := make([]Mapping, 0, len(nt.mappings))
	for _, mapping := range nt.mappings {
		mappings = append(mappings, mapping)
	}
	return mappings
}
