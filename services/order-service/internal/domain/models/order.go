package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Order - Aggregate Root
type Order struct {
	ID              string      `gorm:"primaryKey;type:varchar(255)"`
	UserID          string      `gorm:"not null;type:varchar(255);index:idx_user_number,unique"`
	Number          int64       `gorm:"not null;default:0;index:idx_user_number,unique"`
	Status          OrderStatus `gorm:"type:varchar(20);not null;default:'PENDING'"`
	Items           []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	TotalAmount     int64       `gorm:"type:bigint;not null"`
	Currency        string      `gorm:"type:varchar(3);not null;default:'USD'"`
	ShippingAddress string      `gorm:"type:text;not null"`
	CreatedAt       time.Time   `gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime"`
}

// OrderItem - Entity within Order Aggregate
type OrderItem struct {
	ID          string `gorm:"primaryKey;type:varchar(255)"`
	OrderID     string `gorm:"not null;type:varchar(255);index"`
	ProductID   string `gorm:"not null;type:varchar(255)"`
	ProductName string `gorm:"not null;type:varchar(255)"`
	Quantity    int32  `gorm:"not null"`
	Price       int64  `gorm:"type:bigint;not null"`
	Total       int64  `gorm:"type:bigint;not null"`
	Currency    string `gorm:"type:varchar(3);not null;default:'USD'"`
}

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusConfirmed  OrderStatus = "CONFIRMED"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

// TableName sets the table name
func (Order) TableName() string     { return "orders" }
func (OrderItem) TableName() string { return "order_items" }

// Domain methods for Order Aggregate

// AddItem adds item to order
// AddItem adds item to order
func (o *Order) AddItem(productID, productName string, quantity int32, price int64, currency string) error {
	if o.Status == OrderStatusCancelled {
		return errors.New("cannot add items to cancelled order")
	}
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if price < 0 {
		return errors.New("price cannot be negative")
	}
	if o.Currency == "" {
		o.Currency = currency
	}
	if o.Currency != currency {
		return errors.New("mixed currencies are not supported in a single order")
	}

	item := OrderItem{
		ID:          generateOrderItemID(o.ID, productID),
		OrderID:     o.ID,
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		Price:       price,
		Total:       int64(quantity) * price,
		Currency:    currency,
	}
	o.Items = append(o.Items, item)
	o.recalculateTotal()
	return nil
}

// RemoveItem removes item from order
func (o *Order) RemoveItem(productID string) error {
	if o.Status == OrderStatusCancelled {
		return errors.New("cannot modify cancelled order")
	}
	for i, item := range o.Items {
		if item.ProductID == productID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			o.recalculateTotal()
			return nil
		}
	}
	return errors.New("item not found")
}

// UpdateStatus updates order status
func (o *Order) UpdateStatus(status OrderStatus) error {
	if !o.canTransitionTo(status) {
		return errors.New("invalid status transition")
	}
	o.Status = status
	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.Status == OrderStatusDelivered {
		return errors.New("cannot cancel delivered order")
	}
	if o.Status == OrderStatusCancelled {
		return errors.New("order already cancelled")
	}
	o.Status = OrderStatusCancelled
	return nil
}

// GetItemCount returns total item count
func (o *Order) GetItemCount() int32 {
	var total int32
	for _, item := range o.Items {
		total += item.Quantity
	}
	return total
}

// IsModifiable checks if order can be modified
func (o *Order) IsModifiable() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

// Private methods
func (o *Order) recalculateTotal() {
	var total int64
	for _, item := range o.Items {
		total += item.Total
	}
	o.TotalAmount = total
}

func (o *Order) canTransitionTo(newStatus OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed:  {OrderStatusProcessing, OrderStatusCancelled},
		OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:    {OrderStatusDelivered},
		OrderStatusDelivered:  {},
		OrderStatusCancelled:  {},
	}
	allowed, ok := transitions[o.Status]
	if !ok {
		return false
	}
	for _, st := range allowed {
		if st == newStatus {
			return true
		}
	}
	return false
}

// GORM Hooks
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if len(o.Items) == 0 {
		return errors.New("order must have at least one item")
	}
	if o.ShippingAddress == "" {
		return errors.New("shipping address is required")
	}
	o.recalculateTotal()
	return nil
}

func (o *Order) BeforeUpdate(tx *gorm.DB) error { o.recalculateTotal(); return nil }

// Helper functions
func generateOrderItemID(orderID, productID string) string {
	return orderID + "-" + productID
}
