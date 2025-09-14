package migration

import (
	"time"
)

// MigrationRecord tracks applied migrations
type MigrationRecord struct {
	ID          uint      `gorm:"primaryKey"`
	Version     int64     `gorm:"uniqueIndex;not null"`
	Description string    `gorm:"not null"`
	AppliedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// ProductRecord represents the product table structure
type ProductRecord struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)"`
	Name      string    `gorm:"not null;type:varchar(255)"`
	Price     int64     `gorm:"type:bigint;not null"`
	Currency  string    `gorm:"type:varchar(3);not null;default:'USD'"`
	ImageURL  string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// GORM relationships
	Stock *StockRecord `gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:RESTRICT"`
}

// TableName returns the table name for ProductRecord
func (ProductRecord) TableName() string {
	return "products"
}

// StockRecord represents the stock table structure
type StockRecord struct {
	ProductID         string    `gorm:"primaryKey;type:varchar(255)"`
	AvailableQuantity int32     `gorm:"not null;default:0"`
	ReservedQuantity  int32     `gorm:"not null;default:0"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`

	// GORM relationships
	Product ProductRecord `gorm:"foreignKey:ProductID;references:ID"`
}

// TableName returns the table name for StockRecord
func (StockRecord) TableName() string {
	return "stocks"
}

// OutboxRecord represents the outbox table structure
type OutboxRecord struct {
	ID          uint   `gorm:"primaryKey"`
	AggregateID string `gorm:"not null"`
	Type        string `gorm:"not null"`
	Payload     string `gorm:"type:json;not null"`
	RetryCount  int    `gorm:"default:0"`
	Processed   bool   `gorm:"default:false;not null"`
	ProcessedAt *time.Time
	FailedAt    *time.Time
	Error       string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for OutboxRecord
func (OutboxRecord) TableName() string {
	return "outbox_records"
}
