package repository

import (
	"context"

	"ecommerce-platform/pkg/outbox"
)

// OutboxRepository defines the interface for outbox operations
type OutboxRepository interface {
	SaveEvent(ctx context.Context, event outbox.Event) error
	GetUnprocessedEvents(ctx context.Context, limit int) ([]outbox.Event, error)
	MarkAsProcessed(ctx context.Context, id uint) error
	MarkAsFailed(ctx context.Context, id uint, err string) error
}
