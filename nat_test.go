package nateda

import (
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	eb := NewEventBus()
	
	eventReceived := false
	eb.Subscribe(EventMappingAdd, func(e Event) {
		eventReceived = true
	})

	eb.Publish(Event{Type: EventMappingAdd, Payload: nil})
	time.Sleep(10 * time.Millisecond) // Wait for goroutine

	if !eventReceived {
		t.Error("Event was not received by subscriber")
	}
}

func TestNATTableAddMapping(t *testing.T) {
	eb := NewEventBus()
	nt := NewNATTable(eb)

	internal := Address{IP: "192.168.1.10", Port: 8080}
	external := Address{IP: "203.0.113.5", Port: 80}

	err := nt.AddMapping(internal, external)
	if err != nil {
		t.Errorf("Failed to add mapping: %v", err)
	}

	// Try to add duplicate
	err = nt.AddMapping(internal, external)
	if err == nil {
		t.Error("Expected error when adding duplicate mapping")
	}
}

func TestNATTableTranslate(t *testing.T) {
	eb := NewEventBus()
	nt := NewNATTable(eb)

	internal := Address{IP: "192.168.1.10", Port: 8080}
	external := Address{IP: "203.0.113.5", Port: 80}

	nt.AddMapping(internal, external)

	result, err := nt.Translate(internal)
	if err != nil {
		t.Errorf("Translation failed: %v", err)
	}

	if result.IP != external.IP || result.Port != external.Port {
		t.Errorf("Translation incorrect: got %v, want %v", result, external)
	}
}

func TestNATTableTranslateNotFound(t *testing.T) {
	eb := NewEventBus()
	nt := NewNATTable(eb)

	internal := Address{IP: "192.168.1.99", Port: 9999}

	_, err := nt.Translate(internal)
	if err == nil {
		t.Error("Expected error for non-existent mapping")
	}
}

func TestNATTableRemoveMapping(t *testing.T) {
	eb := NewEventBus()
	nt := NewNATTable(eb)

	internal := Address{IP: "192.168.1.10", Port: 8080}
	external := Address{IP: "203.0.113.5", Port: 80}

	nt.AddMapping(internal, external)

	err := nt.RemoveMapping(internal)
	if err != nil {
		t.Errorf("Failed to remove mapping: %v", err)
	}

	// Try to translate after removal
	_, err = nt.Translate(internal)
	if err == nil {
		t.Error("Expected error after mapping removal")
	}
}

func TestNATTableGetMappings(t *testing.T) {
	eb := NewEventBus()
	nt := NewNATTable(eb)

	nt.AddMapping(Address{IP: "192.168.1.10", Port: 8080}, Address{IP: "203.0.113.5", Port: 80})
	nt.AddMapping(Address{IP: "192.168.1.20", Port: 3000}, Address{IP: "203.0.113.5", Port: 3000})

	mappings := nt.GetMappings()
	if len(mappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(mappings))
	}
}

func TestAddressString(t *testing.T) {
	addr := Address{IP: "192.168.1.10", Port: 8080}
	expected := "192.168.1.10:8080"
	
	if addr.String() != expected {
		t.Errorf("Address.String() = %s, want %s", addr.String(), expected)
	}
}
