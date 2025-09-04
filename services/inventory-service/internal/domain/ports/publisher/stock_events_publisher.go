package publisher

import (
	"context"
	"ecommerce-platform/proto-go/events"
)

// StockEventsPublisher defines the interface for publishing stock-related events
type StockEventsPublisher interface {
	// PublishStockReserved publishes when stock is successfully reserved
	PublishStockReserved(ctx context.Context, event *events.StockReserved) error

	// PublishStockReservationFailed publishes when stock reservation fails
	PublishStockReservationFailed(ctx context.Context, event *events.StockReservationFailed) error

	// PublishStockReleased publishes when reserved stock is released
	PublishStockReleased(ctx context.Context, event *events.StockReleased) error

	// PublishStockCommitted publishes when reserved stock is committed
	PublishStockCommitted(ctx context.Context, event *events.StockCommitted) error

	// Close closes the publisher
	Close() error
}
