package consumer

import (
	"context"

	"ecommerce-platform/proto-go/events"
)

// StockReservedHandler defines the interface for handling StockReserved events
type StockReservedHandler interface {
	// Handle handles StockReserved event
	Handle(ctx context.Context, event *events.StockReserved) error
}
