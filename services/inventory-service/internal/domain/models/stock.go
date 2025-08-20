package models

type Stock struct {
	ProductID         string `gorm:"primaryKey;type:varchar(255)"`
	AvailableQuantity int32  `gorm:"not null;default:0"`
	ReservedQuantity  int32  `gorm:"not null;default:0"`
}
