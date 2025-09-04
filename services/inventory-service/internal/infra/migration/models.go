package migration

import "time"

// ProductRecord represents the product table structure
type ProductRecord struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	Name        string    `gorm:"not null;type:varchar(255)"`
	Description string    `gorm:"type:text"`
	PriceMinor  int64     `gorm:"type:bigint;not null"`
	Currency    string    `gorm:"type:varchar(3);not null;default:'USD'"`
	ImageURL    string    `gorm:"type:text"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for ProductRecord
func (ProductRecord) TableName() string {
	return "products"
}

// StockRecord represents the stock table structure
type StockRecord struct {
	ProductID         string `gorm:"primaryKey;type:varchar(255)"`
	AvailableQuantity int32  `gorm:"not null;default:0"`
	ReservedQuantity  int32  `gorm:"not null;default:0"`
}

// TableName returns the table name for StockRecord
func (StockRecord) TableName() string {
	return "stocks"
}
