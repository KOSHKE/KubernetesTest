package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"

	"gorm.io/gorm"
)

// GormOutboxRepository implements OutboxRepository using GORM
type GormOutboxRepository struct {
	db *gorm.DB
}

// NewOutboxRepository creates a new outbox repository
func NewOutboxRepository(db *gorm.DB) repository.OutboxRepository {
	return &GormOutboxRepository{db: db}
}

// SaveEvent saves an event to the outbox table
func (r *GormOutboxRepository) SaveEvent(ctx context.Context, event outbox.Event) error {
	// Convert payload to JSON string
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	record := migration.OutboxRecord{
		AggregateID: event.AggregateID,
		Type:        event.Type,
		Payload:     string(payloadJSON),
		RetryCount:  event.RetryCount,
		Processed:   false,
		ProcessedAt: event.ProcessedAt,
		FailedAt:    event.FailedAt,
		Error:       event.Error,
	}

	return r.db.WithContext(ctx).Create(&record).Error
}

// GetUnprocessedEvents returns unprocessed events
func (r *GormOutboxRepository) GetUnprocessedEvents(ctx context.Context, limit int) ([]outbox.Event, error) {
	var records []migration.OutboxRecord
	err := r.db.WithContext(ctx).
		Where("processed = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	// Convert to outbox.Event
	result := make([]outbox.Event, len(records))
	for i, record := range records {
		// Parse payload back to interface{}
		var payload interface{}
		if err := json.Unmarshal([]byte(record.Payload), &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload for record %d: %w", record.ID, err)
		}

		result[i] = outbox.Event{
			ID:          record.ID,
			AggregateID: record.AggregateID,
			Type:        record.Type,
			Payload:     payload,
			RetryCount:  record.RetryCount,
			CreatedAt:   record.CreatedAt,
			ProcessedAt: record.ProcessedAt,
			FailedAt:    record.FailedAt,
			Error:       record.Error,
		}
	}

	return result, nil
}

// MarkAsProcessed marks an event as processed
func (r *GormOutboxRepository) MarkAsProcessed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&migration.OutboxRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": &now,
			"failed_at":    nil,
			"error":        "",
		}).Error
}

// MarkAsFailed marks an event as failed
func (r *GormOutboxRepository) MarkAsFailed(ctx context.Context, id uint, err string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&migration.OutboxRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"processed":    false,
			"failed_at":    &now,
			"processed_at": nil,
			"error":        err,
			"retry_count":  gorm.Expr("retry_count + 1"),
		}).Error
}
