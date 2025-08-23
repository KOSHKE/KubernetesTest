package events

import (
	"context"

	events "github.com/kubernetestest/ecommerce-platform/proto-go/events"
)

// Publisher defines contract to publish inventory domain events
type Publisher interface {
	PublishStockReserved(ctx context.Context, evt *events.StockReserved) error
	PublishStockReservationFailed(ctx context.Context, evt *events.StockReservationFailed) error
}
