package consumer

import (
	"context"

	"ecommerce-platform/proto-go/events"
)

// PaymentProcessedConsumer defines the interface for consuming PaymentProcessed events
type PaymentProcessedConsumer interface {
	// HandlePaymentProcessed handles PaymentProcessed event
	HandlePaymentProcessed(ctx context.Context, evt *events.PaymentProcessed) error
}
