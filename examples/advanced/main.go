package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chengtb/chengtb/pkg/eda/callback"
	"github.com/chengtb/chengtb/pkg/eda/event"
	"github.com/chengtb/chengtb/pkg/eda/eventbus"
	"github.com/chengtb/chengtb/pkg/eda/publisher"
	"github.com/chengtb/chengtb/pkg/eda/subscriber"
)

// PaymentProcessedPayload represents payment event data
type PaymentProcessedPayload struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

func main() {
	// Create event bus
	config := eventbus.Config{
		NATSUrl: "nats://localhost:4222",
	}

	bus, err := eventbus.NewEventBus(config)
	if err != nil {
		log.Fatalf("Failed to create event bus: %v", err)
	}
	defer bus.Close()

	fmt.Println("✅ Connected to NATS server for advanced example")

	pub := publisher.NewPublisher(bus)
	sub := subscriber.NewSubscriber(bus)

	// Callback with error handling
	successCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("✅ [Success Callback] Processing event: %s\n", evt.GetID())
		return nil
	}

	// Callback that simulates an error
	errorCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("❌ [Error Callback] Simulating error for event: %s\n", evt.GetID())
		return errors.New("simulated callback error")
	}

	// Callback with timeout handling
	timeoutCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("⏱️  [Timeout Callback] Starting long operation for event: %s\n", evt.GetID())

		select {
		case <-time.After(2 * time.Second):
			fmt.Printf("   ✅ Long operation completed\n")
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Register callbacks
	fmt.Println("\n📝 Registering callbacks for 'payment.processed' event...")
	sub.SubscribeWithCallback("payment.processed", successCallback)
	sub.SubscribeWithCallback("payment.processed", errorCallback)
	sub.SubscribeWithCallback("payment.processed", timeoutCallback)

	// Wait for subscriptions
	time.Sleep(100 * time.Millisecond)

	// Publish payment event
	fmt.Println("\n🚀 Publishing 'payment.processed' event...")
	paymentEvent := event.NewBaseEvent(
		"evt-payment-001",
		"payment.processed",
		PaymentProcessedPayload{
			PaymentID: "pay-123",
			Amount:    149.99,
			Status:    "completed",
		},
	)

	if err := pub.Publish(paymentEvent); err != nil {
		log.Printf("Failed to publish payment event: %v", err)
	}

	// Wait for callbacks
	time.Sleep(3 * time.Second)

	// Example of chaining callbacks - order processing workflow
	fmt.Println("\n🔗 Demonstrating callback chaining for order workflow...")

	var orderID string

	// Step 1: Validate order
	validateOrderCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("1️⃣  Validating order for event: %s\n", evt.GetID())
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("   ✅ Order validated\n")
		return nil
	}

	// Step 2: Process payment
	processPaymentCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("2️⃣  Processing payment for event: %s\n", evt.GetID())
		time.Sleep(150 * time.Millisecond)
		fmt.Printf("   ✅ Payment processed\n")
		return nil
	}

	// Step 3: Update inventory
	updateInventoryCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("3️⃣  Updating inventory for event: %s\n", evt.GetID())
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("   ✅ Inventory updated\n")
		return nil
	}

	// Step 4: Send confirmation
	sendConfirmationCallback := func(ctx context.Context, evt event.Event) error {
		fmt.Printf("4️⃣  Sending confirmation for event: %s\n", evt.GetID())
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("   ✅ Confirmation sent\n")
		return nil
	}

	// Register all callbacks for order workflow
	callbacks := []callback.CallbackFunc{
		validateOrderCallback,
		processPaymentCallback,
		updateInventoryCallback,
		sendConfirmationCallback,
	}

	if err := sub.SubscribeWithMultipleCallbacks("order.workflow", callbacks); err != nil {
		log.Fatalf("Failed to subscribe callbacks: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Publish order workflow event
	fmt.Println("\n🚀 Publishing 'order.workflow' event...")
	orderEvent := event.NewBaseEvent(
		"evt-order-workflow-001",
		"order.workflow",
		map[string]interface{}{
			"order_id": "order-789",
			"items":    []string{"item1", "item2"},
			"total":    299.99,
		},
	)

	if err := pub.Publish(orderEvent); err != nil {
		log.Printf("Failed to publish order event: %v", err)
	}

	// Wait for workflow to complete
	time.Sleep(1 * time.Second)

	fmt.Println("\n✅ Advanced example completed!")
	fmt.Println("💡 This example demonstrated:")
	fmt.Println("   - Error handling in callbacks")
	fmt.Println("   - Timeout management")
	fmt.Println("   - Callback chaining for workflows")
	fmt.Println("   - Asynchronous callback execution")

	_ = orderID // avoid unused variable warning
}
