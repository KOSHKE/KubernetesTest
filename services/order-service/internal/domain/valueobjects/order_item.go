package valueobjects

import (
	"ecommerce-platform/pkg/common/valueobjects"
)

// OrderItem represents an item in an order as a value object
type OrderItem struct {
	valueobjects.Item `gorm:"embedded"`
	ProductName       string             `gorm:"type:varchar(500);not null"`
	Price             valueobjects.Money `gorm:"embedded;embeddedPrefix:price_"`
}

// NewOrderItem creates a new OrderItem value object
func NewOrderItem(productID, productName string, quantity int32, price valueobjects.Money) (*OrderItem, error) {
	item, err := valueobjects.NewItem(productID, quantity)
	if err != nil {
		return nil, err
	}

	return &OrderItem{
		Item:        *item,
		ProductName: productName,
		Price:       price,
	}, nil
}

// Equals checks if two OrderItems are equal
func (oi *OrderItem) Equals(other *OrderItem) bool {
	if other == nil {
		return false
	}
	return oi.Item.Equals(&other.Item) &&
		oi.ProductName == other.ProductName &&
		oi.Price == other.Price
}
