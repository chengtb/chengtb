package main

import (
	"fmt"
	"time"

	nateda "github.com/chengtb/chengtb/nat-eda"
)

func main() {
	fmt.Println("=== Golang NAT with Event-Driven Architecture Demo ===\n")

	// Create event bus
	eventBus := nateda.NewEventBus()

	// Subscribe to events
	eventBus.Subscribe(nateda.EventMappingAdd, func(e nateda.Event) {
		mapping := e.Payload.(nateda.Mapping)
		fmt.Printf("✓ Event: Mapping Added - %s -> %s\n", 
			mapping.Internal, mapping.External)
	})

	eventBus.Subscribe(nateda.EventMappingDel, func(e nateda.Event) {
		mapping := e.Payload.(nateda.Mapping)
		fmt.Printf("✓ Event: Mapping Deleted - %s -> %s\n", 
			mapping.Internal, mapping.External)
	})

	eventBus.Subscribe(nateda.EventTranslate, func(e nateda.Event) {
		mapping := e.Payload.(nateda.Mapping)
		fmt.Printf("✓ Event: Translation - %s -> %s\n", 
			mapping.Internal, mapping.External)
	})

	eventBus.Subscribe(nateda.EventError, func(e nateda.Event) {
		err := e.Payload.(error)
		fmt.Printf("✗ Event: Error - %v\n", err)
	})

	// Create NAT table
	natTable := nateda.NewNATTable(eventBus)

	// Add some mappings
	fmt.Println("Adding NAT mappings...")
	natTable.AddMapping(
		nateda.Address{IP: "192.168.1.10", Port: 8080},
		nateda.Address{IP: "203.0.113.5", Port: 80},
	)
	natTable.AddMapping(
		nateda.Address{IP: "192.168.1.20", Port: 3000},
		nateda.Address{IP: "203.0.113.5", Port: 3000},
	)
	time.Sleep(100 * time.Millisecond) // Wait for events to be processed

	fmt.Println("\nTranslating addresses...")
	// Translate addresses
	if external, err := natTable.Translate(nateda.Address{IP: "192.168.1.10", Port: 8080}); err == nil {
		fmt.Printf("Internal 192.168.1.10:8080 -> External %s\n", external)
	}

	// Try to translate non-existent address
	fmt.Println("\nAttempting to translate non-existent mapping...")
	natTable.Translate(nateda.Address{IP: "192.168.1.99", Port: 9999})
	time.Sleep(100 * time.Millisecond)

	// Display all mappings
	fmt.Println("\nCurrent NAT Mappings:")
	for i, mapping := range natTable.GetMappings() {
		fmt.Printf("%d. %s -> %s\n", i+1, mapping.Internal, mapping.External)
	}

	// Remove a mapping
	fmt.Println("\nRemoving mapping for 192.168.1.10:8080...")
	natTable.RemoveMapping(nateda.Address{IP: "192.168.1.10", Port: 8080})
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nFinal NAT Mappings:")
	for i, mapping := range natTable.GetMappings() {
		fmt.Printf("%d. %s -> %s\n", i+1, mapping.Internal, mapping.External)
	}

	fmt.Println("\n=== Demo Complete ===")
}
