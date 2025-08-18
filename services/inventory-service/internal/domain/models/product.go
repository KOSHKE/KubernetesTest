package models

type Product struct {
	ID           string `gorm:"primaryKey;type:varchar(255)"`
	Name         string `gorm:"not null;type:varchar(255)"`
	Description  string `gorm:"type:text"`
	PriceMinor   int64  `gorm:"type:bigint;not null"`
	Currency     string `gorm:"type:varchar(3);not null;default:'USD'"`
	CategoryID   string `gorm:"type:varchar(255)"`
	CategoryName string `gorm:"type:varchar(255)"`
	ImageURL     string `gorm:"type:text"`
	IsActive     bool   `gorm:"not null;default:true"`
}
