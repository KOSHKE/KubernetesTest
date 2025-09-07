package publisher

import (
	"context"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/events"
)

// EventPublisher defines the interface for publishing events
// This interface composes individual publisher interfaces for better separation of concerns
type EventPublisher interface {
	// PublishStockReserved publishes when stock is successfully reserved
	PublishStockReserved(ctx context.Context, event *events.StockReserved) error

	// PublishStockReleased publishes when reserved stock is released
	PublishStockReleased(ctx context.Context, event *events.StockReleased) error

	// PublishStockCommitted publishes when reserved stock is committed
	PublishStockCommitted(ctx context.Context, event *events.StockCommitted) error

	// PublishFromOutbox publishes events directly from outbox
	PublishFromOutbox(ctx context.Context, event outbox.Event) error

	// Close closes the publisher
	Close() error
}

// StockEventsPublisher is an alias for EventPublisher for backward compatibility
// Deprecated: Use EventPublisher instead
type StockEventsPublisher = EventPublisher
