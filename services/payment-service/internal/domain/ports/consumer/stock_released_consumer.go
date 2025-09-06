package consumer

import (
	"context"
)

// StockReleasedConsumer defines the interface for consuming stock released events
type StockReleasedConsumer interface {
	Run(ctx context.Context, topics []string) error
	Close() error
}
