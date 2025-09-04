package consumer

import (
	"context"

	"ecommerce-platform/proto-go/events"
)

// StockReservedConsumer defines the interface for consuming StockReserved events
type StockReservedConsumer interface {
	// HandleStockReserved handles StockReserved event
	HandleStockReserved(ctx context.Context, evt *events.StockReserved) error
}
