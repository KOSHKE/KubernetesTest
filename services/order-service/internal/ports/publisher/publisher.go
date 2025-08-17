package publisher

import (
	"context"

	events "proto-go/events"
)

// EventPublisher defines minimal contract for emitting domain events
type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, evt *events.OrderCreated) error
}
