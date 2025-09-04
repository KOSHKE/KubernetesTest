package consumer

import (
	"context"

	"ecommerce-platform/proto-go/events"
)

// OrderCreatedHandler defines the interface for handling OrderCreated events
type OrderCreatedHandler interface {
	// Handle handles OrderCreated event
	Handle(ctx context.Context, event *events.OrderCreated) error
}
