package migration

import (
	"time"
)

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
