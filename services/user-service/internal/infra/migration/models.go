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

// TableName returns dedicated table name for user-service migrations
func (MigrationRecord) TableName() string {
	return "user_migration_records"
}

// UserRecord is a GORM model for users (copied from repository)
type UserRecord struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	FirstName    string    `gorm:"not null"`
	LastName     string    `gorm:"not null"`
	Phone        string    `gorm:"default:null"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for UserRecord
func (UserRecord) TableName() string {
	return "users"
}
