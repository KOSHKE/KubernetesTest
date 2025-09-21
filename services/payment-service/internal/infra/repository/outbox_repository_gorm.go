package repository

import (
	"context"

	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/payment-service/internal/domain/ports/repository"

	"gorm.io/gorm"
)

// OutboxRepositoryGorm implements the OutboxRepository interface using GORM
type OutboxRepositoryGorm struct {
	db *gorm.DB
}

// NewOutboxRepository creates a new outbox repository
func NewOutboxRepository(db *gorm.DB) repository.OutboxRepository {
	return &OutboxRepositoryGorm{db: db}
}

// SaveEvent saves an event to the outbox table
func (r *OutboxRepositoryGorm) SaveEvent(ctx context.Context, event outbox.Event) error {
	return outbox.NewGormRepository(r.db).SaveEvent(ctx, event)
}

// GetUnprocessedEvents retrieves unprocessed events from the outbox table
func (r *OutboxRepositoryGorm) GetUnprocessedEvents(ctx context.Context, limit int) ([]outbox.Event, error) {
	return outbox.NewGormRepository(r.db).GetUnprocessedEvents(ctx, limit)
}

// MarkAsProcessed marks an event as processed
func (r *OutboxRepositoryGorm) MarkAsProcessed(ctx context.Context, id uint) error {
	return outbox.NewGormRepository(r.db).MarkAsProcessed(ctx, id)
}

// MarkAsFailed marks an event as failed with an error message
func (r *OutboxRepositoryGorm) MarkAsFailed(ctx context.Context, id uint, err string) error {
	return outbox.NewGormRepository(r.db).MarkAsFailed(ctx, id, err)
}
