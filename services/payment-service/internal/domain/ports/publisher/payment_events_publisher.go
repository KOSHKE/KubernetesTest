package publisher

import (
	"context"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/events"
)

// PaymentEventsPublisher defines the interface for publishing payment events
type PaymentEventsPublisher interface {
	// PublishPaymentProcessed publishes when payment is processed
	PublishPaymentProcessed(ctx context.Context, event *events.PaymentProcessed) error

	// PublishFromOutbox publishes events directly from outbox
	PublishFromOutbox(ctx context.Context, event outbox.Event) error

	// Close closes the publisher
	Close() error
}
