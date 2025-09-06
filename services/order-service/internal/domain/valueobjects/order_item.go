package valueobjects

import (
	"ecommerce-platform/pkg/common/valueobjects"
)

// OrderItem represents an item in an order as a value object
type OrderItem struct {
	ProductID   string             `gorm:"type:varchar(255);not null"`
	ProductName string             `gorm:"type:varchar(500);not null"`
	Quantity    int32              `gorm:"type:int;not null"`
	Price       valueobjects.Money `gorm:"embedded;embeddedPrefix:price_"`
}

// NewOrderItem creates a new OrderItem value object
func NewOrderItem(productID, productName string, quantity int32, price valueobjects.Money) (*OrderItem, error) {
	return &OrderItem{
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		Price:       price,
	}, nil
}

// Equals checks if two OrderItems are equal
func (oi *OrderItem) Equals(other *OrderItem) bool {
	if other == nil {
		return false
	}
	return oi.ProductID == other.ProductID &&
		oi.ProductName == other.ProductName &&
		oi.Quantity == other.Quantity &&
		oi.Price == other.Price
}
