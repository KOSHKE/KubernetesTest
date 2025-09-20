package publisher

import (
	"context"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/events"
)

// EventPublisher defines the interface for publishing events
// This interface composes individual publisher interfaces for better separation of concerns
type EventPublisher interface {
	// PublishOrderCreated publishes when an order is created
	PublishOrderCreated(ctx context.Context, event *events.OrderCreated) error

	// PublishOrderCancelled publishes when an order is cancelled
	PublishOrderCancelled(ctx context.Context, event *events.OrderCancelled) error

	// PublishFromOutbox publishes events directly from outbox
	PublishFromOutbox(ctx context.Context, event outbox.Event) error

	// Close closes the publisher
	Close() error
}

// OrderCreatedPublisher is an alias for EventPublisher for backward compatibility
// Deprecated: Use EventPublisher instead
type OrderCreatedPublisher = EventPublisher

// OrderEventsPublisher is an alias for EventPublisher
type OrderEventsPublisher = EventPublisher
