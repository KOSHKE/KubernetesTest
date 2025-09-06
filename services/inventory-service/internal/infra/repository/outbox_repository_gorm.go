package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"

	"gorm.io/gorm"
)

// GormOutboxRepository implements OutboxRepository using GORM
type GormOutboxRepository struct {
	db *gorm.DB
}

// OutboxRecordGorm represents an outbox record in the database with GORM tags
type OutboxRecordGorm struct {
	ID        uint      `gorm:"primaryKey"`
	Type      string    `gorm:"not null"`
	Payload   string    `gorm:"type:json;not null"`
	Processed bool      `gorm:"default:false;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
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

	record := OutboxRecordGorm{
		Type:      event.Type,
		Payload:   string(payloadJSON),
		Processed: false,
	}

	return r.db.WithContext(ctx).Create(&record).Error
}

// GetUnprocessedEvents returns unprocessed events
func (r *GormOutboxRepository) GetUnprocessedEvents(ctx context.Context, limit int) ([]outbox.OutboxRecord, error) {
	var records []OutboxRecordGorm
	err := r.db.WithContext(ctx).
		Where("processed = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	// Convert to outbox.OutboxRecord
	result := make([]outbox.OutboxRecord, len(records))
	for i, record := range records {
		result[i] = outbox.OutboxRecord{
			ID:        record.ID,
			Type:      record.Type,
			Payload:   record.Payload,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
			Processed: record.Processed,
		}
	}

	return result, nil
}

// MarkAsProcessed marks events as processed
func (r *GormOutboxRepository) MarkAsProcessed(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).
		Model(&OutboxRecordGorm{}).
		Where("id IN ?", ids).
		Update("processed", true).Error
}
