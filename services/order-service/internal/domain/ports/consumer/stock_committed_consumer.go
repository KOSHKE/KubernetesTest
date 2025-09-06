package consumer

import (
	"context"
)

// StockCommittedConsumer defines the interface for consuming stock committed events
type StockCommittedConsumer interface {
	Run(ctx context.Context, topics []string) error
	Close() error
}
