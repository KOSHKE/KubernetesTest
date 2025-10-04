package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GormRepository implements Repository using GORM
// This is a shared implementation that can be used across all services
type GormRepository struct {
	db        *gorm.DB
	tableName string
}

// NewGormRepository creates a new GORM outbox repository
func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// NewGormRepositoryWithTable creates a new GORM outbox repository targeting a specific table
func NewGormRepositoryWithTable(db *gorm.DB, tableName string) Repository {
	return &GormRepository{db: db, tableName: tableName}
}

func (r *GormRepository) tbl(tx *gorm.DB) *gorm.DB {
	if r.tableName != "" {
		return tx.Table(r.tableName)
	}
	return tx.Model(&OutboxRecord{})
}

// SaveEvent saves an event to the outbox table
func (r *GormRepository) SaveEvent(ctx context.Context, event Event) error {
	// Convert payload to JSON string
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	record := OutboxRecord{
		AggregateID: event.AggregateID,
		Type:        event.Type,
		Payload:     string(payloadJSON),
		RetryCount:  event.RetryCount,
		Processed:   false,
		ProcessedAt: event.ProcessedAt,
		FailedAt:    event.FailedAt,
		Error:       event.Error,
	}

	return r.tbl(r.db.WithContext(ctx)).Create(&record).Error
}

// GetUnprocessedEvents returns unprocessed events
func (r *GormRepository) GetUnprocessedEvents(ctx context.Context, limit int) ([]Event, error) {
	var records []OutboxRecord
	err := r.tbl(r.db.WithContext(ctx)).
		Where("processed = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	// Convert to outbox.Event
	result := make([]Event, len(records))
	for i, record := range records {
		// Parse payload back to interface{}
		var payload interface{}
		if err := json.Unmarshal([]byte(record.Payload), &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload for record %d: %w", record.ID, err)
		}

		result[i] = Event{
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
func (r *GormRepository) MarkAsProcessed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.tbl(r.db.WithContext(ctx)).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": &now,
			"failed_at":    nil,
			"error":        "",
		}).Error
}

// MarkAsFailed marks an event as failed
func (r *GormRepository) MarkAsFailed(ctx context.Context, id uint, err string) error {
	now := time.Now()
	return r.tbl(r.db.WithContext(ctx)).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"processed":    false,
			"failed_at":    &now,
			"processed_at": nil,
			"error":        err,
			"retry_count":  gorm.Expr("retry_count + 1"),
		}).Error
}
