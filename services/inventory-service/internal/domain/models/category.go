package models

type Category struct {
	ID          string `gorm:"primaryKey;type:varchar(255)"`
	Name        string `gorm:"not null;type:varchar(255)"`
	Description string `gorm:"type:text"`
	IsActive    bool   `gorm:"not null;default:true"`
}
