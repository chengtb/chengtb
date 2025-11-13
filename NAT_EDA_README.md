# Golang NAT with Event-Driven Architecture (EDA)

A simple and elegant implementation of Network Address Translation (NAT) using Event-Driven Architecture in Go.

## Features

- **NAT Table Management**: Add, remove, and query NAT mappings
- **Event-Driven Architecture**: All NAT operations emit events
- **Thread-Safe**: Concurrent-safe operations using mutexes
- **Simple API**: Easy to use and integrate

## Architecture

The project consists of three main components:

1. **Event Bus**: A publish-subscribe event system
2. **NAT Table**: Manages address mappings with event notifications
3. **Event Types**: Different events for various NAT operations

### Event Types

- `EventMappingAdd`: Triggered when a new mapping is added
- `EventMappingDel`: Triggered when a mapping is deleted
- `EventTranslate`: Triggered when an address is translated
- `EventError`: Triggered when an error occurs

## Installation

```bash
go get github.com/chengtb/chengtb/nat-eda
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    nateda "github.com/chengtb/chengtb/nat-eda"
)

func main() {
    // Create event bus
    eventBus := nateda.NewEventBus()
    
    // Subscribe to mapping events
    eventBus.Subscribe(nateda.EventMappingAdd, func(e nateda.Event) {
        mapping := e.Payload.(nateda.Mapping)
        fmt.Printf("Mapping added: %s -> %s\n", 
            mapping.Internal, mapping.External)
    })
    
    // Create NAT table
    natTable := nateda.NewNATTable(eventBus)
    
    // Add a mapping
    natTable.AddMapping(
        nateda.Address{IP: "192.168.1.10", Port: 8080},
        nateda.Address{IP: "203.0.113.5", Port: 80},
    )
    
    // Translate address
    external, err := natTable.Translate(
        nateda.Address{IP: "192.168.1.10", Port: 8080},
    )
    if err == nil {
        fmt.Printf("Translated to: %s\n", external)
    }
}
```

### Running the Example

```bash
cd example
go run main.go
```

## Testing

Run all tests:

```bash
go test -v
```

Run tests with coverage:

```bash
go test -v -cover
```

## API Reference

### Event Bus

- `NewEventBus()`: Create a new event bus
- `Subscribe(eventType EventType, handler EventHandler)`: Subscribe to events
- `Publish(event Event)`: Publish an event

### NAT Table

- `NewNATTable(eventBus *EventBus)`: Create a new NAT table
- `AddMapping(internal, external Address)`: Add a new mapping
- `RemoveMapping(internal Address)`: Remove a mapping
- `Translate(internal Address)`: Translate an address
- `GetMappings()`: Get all current mappings

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Author

@chengtb
