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

// OrderRecord is a GORM model for orders
type OrderRecord struct {
	ID              string    `gorm:"primaryKey;type:varchar(255)"`
	UserID          string    `gorm:"type:varchar(255);not null;index:idx_user_id"`
	Status          string    `gorm:"type:varchar(50);not null;default:'PENDING'"`
	ShippingAddress string    `gorm:"type:text;not null"`
	Currency        string    `gorm:"type:varchar(3);not null;default:'USD'"`
	TotalAmount     int64     `gorm:"type:bigint;not null;default:0"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`

	// GORM relationships
	Items []OrderItemRecord `gorm:"foreignKey:OrderID"`
}

// TableName specifies the table name for OrderRecord
func (OrderRecord) TableName() string {
	return "orders"
}

// OrderItemRecord is a GORM model for order items (copied from repository)
type OrderItemRecord struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	OrderID     string `gorm:"type:varchar(255);not null;uniqueIndex:idx_order_product"`
	ProductID   string `gorm:"type:varchar(255);not null;uniqueIndex:idx_order_product"`
	ProductName string `gorm:"type:varchar(500);not null"`
	Quantity    int32  `gorm:"type:int;not null"`
	UnitPrice   int64  `gorm:"type:bigint;not null"`
	Currency    string `gorm:"type:varchar(3);not null"`

	// GORM relationships
	Order OrderRecord `gorm:"foreignKey:OrderID"`
}

// TableName specifies the table name for OrderItemRecord
func (OrderItemRecord) TableName() string {
	return "order_items"
}
