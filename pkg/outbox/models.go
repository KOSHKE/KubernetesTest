package outbox

import (
	"time"
)

// OutboxRecord represents an outbox record in the database
// This model is shared across all services for consistency
type OutboxRecord struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	AggregateID string     `gorm:"type:varchar(255);not null;index:idx_aggregate_id"`
	Type        string     `gorm:"type:varchar(100);not null;index:idx_type"`
	Payload     string     `gorm:"type:json;not null"`
	RetryCount  int        `gorm:"type:int;not null;default:0"`
	Processed   bool       `gorm:"type:boolean;not null;default:false;index:idx_processed"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;index:idx_created_at"`
	ProcessedAt *time.Time `gorm:"index:idx_processed_at"`
	FailedAt    *time.Time `gorm:"index:idx_failed_at"`
	Error       string     `gorm:"type:text"`
}

// TableName specifies the table name for OutboxRecord
func (OutboxRecord) TableName() string {
	return "outbox_events"
}
