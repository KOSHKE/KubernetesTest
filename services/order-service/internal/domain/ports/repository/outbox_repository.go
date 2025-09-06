package repository

import (
	"context"

	"ecommerce-platform/pkg/outbox"
)

// OutboxRepository defines the interface for outbox operations
type OutboxRepository interface {
	SaveEvent(ctx context.Context, event outbox.Event) error
	GetUnprocessedEvents(ctx context.Context, limit int) ([]outbox.OutboxRecord, error)
	MarkAsProcessed(ctx context.Context, ids []uint) error
}
