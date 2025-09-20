package entities

import (
	"ecommerce-platform/pkg/common/valueobjects"
)

// OrderItem represents an individual item within an order
type OrderItem struct {
	ProductID   string
	ProductName string
	Quantity    int32
	UnitPrice   valueobjects.Money
}

// NewOrderItem creates a new OrderItem
func NewOrderItem(productID, productName string, quantity int32, unitPrice valueobjects.Money) *OrderItem {
	return &OrderItem{
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
	}
}

// ChangeQuantity changes the item quantity
func (oi *OrderItem) ChangeQuantity(newQuantity int32) {
	oi.Quantity = newQuantity
}

// TotalPrice calculates and returns the total price for this item
func (oi *OrderItem) TotalPrice() valueobjects.Money {
	return oi.UnitPrice.Multiply(int64(oi.Quantity))
}

// Currency returns the currency of the item
func (oi *OrderItem) Currency() string {
	return oi.UnitPrice.Currency.Code
}

// UnitPriceAmount returns the unit price amount in minor units
func (oi *OrderItem) UnitPriceAmount() int64 {
	return oi.UnitPrice.Amount
}

// TotalPriceAmount returns the total price amount in minor units
func (oi *OrderItem) TotalPriceAmount() int64 {
	return oi.TotalPrice().Amount
}
