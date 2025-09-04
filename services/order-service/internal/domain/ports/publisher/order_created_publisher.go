package publisher

import (
	"context"

	"ecommerce-platform/proto-go/events"
)

// OrderCreatedPublisher defines the interface for publishing OrderCreated events
type OrderCreatedPublisher interface {
	// PublishOrderCreated publishes OrderCreated event
	PublishOrderCreated(ctx context.Context, evt *events.OrderCreated) error
}
