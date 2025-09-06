package consumer

import (
	"context"
)

// StockReservationFailedConsumer defines the interface for consuming stock reservation failed events
type StockReservationFailedConsumer interface {
	Run(ctx context.Context, topics []string) error
	Close() error
}
