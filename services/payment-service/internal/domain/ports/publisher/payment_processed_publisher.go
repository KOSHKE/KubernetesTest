package publisher

import (
	"context"

	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/payment-service/internal/domain/entities"
)

// PaymentProcessedPublisher defines the interface for publishing PaymentProcessed events
type PaymentProcessedPublisher interface {
	// PublishPaymentProcessed publishes PaymentProcessed event
	PublishPaymentProcessed(ctx context.Context, payment *entities.Payment, success bool, message string) error
	// PublishFromOutbox publishes events directly from outbox
	PublishFromOutbox(ctx context.Context, event outbox.Event) error
}
