package migration

import (
	"time"

	"ecommerce-platform/pkg/outbox"
)

// MigrationRecord tracks applied migrations
type MigrationRecord struct {
	ID          uint      `gorm:"primaryKey"`
	Version     int64     `gorm:"uniqueIndex;not null"`
	Description string    `gorm:"not null"`
	AppliedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// Use shared OutboxRecord from pkg/outbox
// This ensures consistency across all services
type OutboxRecord = outbox.OutboxRecord
