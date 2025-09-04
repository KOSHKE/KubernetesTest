package consumer

import (
	"context"
	"ecommerce-platform/proto-go/events"
)

// OrderEventsConsumer defines the interface for consuming order-related events
type OrderEventsConsumer interface {
	// HandleOrderCreated handles when a new order is created
	HandleOrderCreated(ctx context.Context, event *events.OrderCreated) error

	// HandlePaymentProcessed handles when payment is processed
	HandlePaymentProcessed(ctx context.Context, event *events.PaymentProcessed) error

	// Close closes the consumer
	Close() error
}
