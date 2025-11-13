package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chengtb/chengtb/pkg/eda/event"
	"github.com/chengtb/chengtb/pkg/eda/eventbus"
	"github.com/chengtb/chengtb/pkg/eda/publisher"
	"github.com/chengtb/chengtb/pkg/eda/subscriber"
)

// UserCreatedPayload represents the payload for user created event
type UserCreatedPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// OrderPlacedPayload represents the payload for order placed event
type OrderPlacedPayload struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
}

func main() {
	// Note: This example requires a running NATS server
	// You can start one with: docker run -p 4222:4222 nats:latest
	// Or install nats-server: https://docs.nats.io/running-a-nats-service/introduction/installation

	// Create event bus with NATS connection
	config := eventbus.Config{
		NATSUrl: "nats://localhost:4222",
	}

	bus, err := eventbus.NewEventBus(config)
	if err != nil {
		log.Fatalf("Failed to create event bus: %v", err)
	}
	defer bus.Close()

	fmt.Println("✅ Connected to NATS server")

	// Create publisher and subscriber
	pub := publisher.NewPublisher(bus)
	sub := subscriber.NewSubscriber(bus)

	// Define callbacks for user created event
	userCreatedCallback1 := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("📧 [Callback 1] Sending welcome email for event: %s\n", evt.GetID())
		time.Sleep(100 * time.Millisecond) // Simulate email sending
		fmt.Printf("   ✅ Email sent successfully\n")
		return nil
	}

	userCreatedCallback2 := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("📊 [Callback 2] Updating analytics for event: %s\n", evt.GetID())
		time.Sleep(50 * time.Millisecond) // Simulate analytics update
		fmt.Printf("   ✅ Analytics updated successfully\n")
		return nil
	}

	userCreatedCallback3 := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("🔔 [Callback 3] Sending notification for event: %s\n", evt.GetID())
		time.Sleep(75 * time.Millisecond) // Simulate notification
		fmt.Printf("   ✅ Notification sent successfully\n")
		return nil
	}

	// Subscribe to UserCreated event with multiple callbacks
	fmt.Println("\n📝 Registering callbacks for 'user.created' event...")
	if err := sub.SubscribeWithCallback("user.created", userCreatedCallback1); err != nil {
		log.Fatalf("Failed to subscribe callback 1: %v", err)
	}
	if err := sub.SubscribeWithCallback("user.created", userCreatedCallback2); err != nil {
		log.Fatalf("Failed to subscribe callback 2: %v", err)
	}
	if err := sub.SubscribeWithCallback("user.created", userCreatedCallback3); err != nil {
		log.Fatalf("Failed to subscribe callback 3: %v", err)
	}

	// Define callback for order placed event
	orderPlacedCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("💳 [Order Callback] Processing payment for event: %s\n", evt.GetID())
		time.Sleep(100 * time.Millisecond) // Simulate payment processing
		fmt.Printf("   ✅ Payment processed successfully\n")
		return nil
	}

	// Subscribe to OrderPlaced event
	fmt.Println("📝 Registering callback for 'order.placed' event...")
	if err := sub.SubscribeWithCallback("order.placed", orderPlacedCallback); err != nil {
		log.Fatalf("Failed to subscribe order callback: %v", err)
	}

	// Wait a moment for subscriptions to be ready
	time.Sleep(100 * time.Millisecond)

	// Publish UserCreated event
	fmt.Println("\n🚀 Publishing 'user.created' event...")
	userEvent := event.NewBaseEvent(
		"evt-001",
		"user.created",
		UserCreatedPayload{
			UserID:   "user-123",
			Username: "john_doe",
			Email:    "john@example.com",
		},
	)

	if err := pub.PublishWithValidation(userEvent); err != nil {
		log.Fatalf("Failed to publish user event: %v", err)
	}

	// Wait for callbacks to execute
	time.Sleep(500 * time.Millisecond)

	// Publish OrderPlaced event
	fmt.Println("\n🚀 Publishing 'order.placed' event...")
	orderEvent := event.NewBaseEvent(
		"evt-002",
		"order.placed",
		OrderPlacedPayload{
			OrderID: "order-456",
			UserID:  "user-123",
			Amount:  99.99,
		},
	)

	if err := pub.PublishWithValidation(orderEvent); err != nil {
		log.Fatalf("Failed to publish order event: %v", err)
	}

	// Wait for callbacks to execute
	time.Sleep(500 * time.Millisecond)

	// Publish multiple events
	fmt.Println("\n🚀 Publishing multiple 'user.created' events...")
	for i := 1; i <= 3; i++ {
		evt := event.NewBaseEvent(
			fmt.Sprintf("evt-user-%d", i),
			"user.created",
			UserCreatedPayload{
				UserID:   fmt.Sprintf("user-%d", i),
				Username: fmt.Sprintf("user_%d", i),
				Email:    fmt.Sprintf("user%d@example.com", i),
			},
		)
		if err := pub.Publish(evt); err != nil {
			log.Printf("Failed to publish event %d: %v", i, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Wait for all callbacks to complete
	time.Sleep(1 * time.Second)

	fmt.Println("\n✅ All events published and callbacks executed successfully!")
	fmt.Println("🎉 Event-Driven Architecture demo completed!")
}
