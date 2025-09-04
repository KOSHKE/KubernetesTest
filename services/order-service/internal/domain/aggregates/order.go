package aggregates

import (
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/entities"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// Business Rules:
// 1. Each product can appear only once per order
// 2. If the same product is added again, use AddItem to update quantity
// 3. All items in order must have the same currency as the order

// Order represents an order aggregate
type Order struct {
	ID              string
	UserID          string
	Status          orderValueObjects.OrderStatus
	Items           []*entities.OrderItem
	ShippingAddress orderValueObjects.ShippingAddress
	Currency        valueobjects.Currency
	TotalAmount     valueobjects.Money
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewOrder creates a new Order aggregate
func NewOrder(userID string, shippingAddress orderValueObjects.ShippingAddress, currency valueobjects.Currency) (*Order, error) {
	// Create zero money for initial total
	zeroMoney := valueobjects.NewMoney(0, currency)

	return &Order{
		ID:              "", // Will be set by GORM when saving
		UserID:          userID,
		Status:          orderValueObjects.OrderStatusPending,
		Items:           make([]*entities.OrderItem, 0),
		ShippingAddress: shippingAddress,
		Currency:        currency,
		TotalAmount:     zeroMoney,
		CreatedAt:       time.Time{}, // Zero value, will be set by GORM
		UpdatedAt:       time.Time{}, // Zero value, will be set by GORM
	}, nil
}

// AddItem adds a new item to the order or updates quantity if product already exists
// This method handles both new items and existing product updates gracefully
func (o *Order) AddItem(productID, productName string, quantity int32, unitPrice valueobjects.Money) error {
	if !unitPrice.Currency.Equals(o.Currency) {
		return errors.ErrCurrencyMismatch
	}

	// Check if product already exists and update quantity
	for _, existingItem := range o.Items {
		if existingItem.ProductID == productID {
			// Update existing item quantity
			existingItem.Quantity += quantity
			if err := o.recalculateTotal(); err != nil {
				return errors.ErrOrderItemProcessingFailed
			}
			return nil
		}
	}

	// Add new item if product doesn't exist
	item := entities.NewOrderItem(productID, productName, quantity, unitPrice)
	o.Items = append(o.Items, item)
	if err := o.recalculateTotal(); err != nil {
		return errors.ErrOrderItemProcessingFailed
	}

	return nil
}

// RemoveItem removes an item from the order by product ID
func (o *Order) RemoveItem(productID string) error {
	for i, item := range o.Items {
		if item.ProductID == productID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			if err := o.recalculateTotal(); err != nil {
				return errors.ErrOrderItemProcessingFailed
			}
			return nil
		}
	}
	return errors.ErrOrderItemNotFound
}

// SetStatus updates the order status
func (o *Order) SetStatus(newStatus orderValueObjects.OrderStatus) error {
	o.Status = newStatus
	return nil
}

// CancelOrder cancels the order with business rules validation
func (o *Order) CancelOrder(userID string) error {
	// Check access control - user can only cancel their own orders
	if !o.IsOwnedBy(userID) {
		return errors.ErrOrderAccessDenied
	}

	// Business rule: only pending orders can be cancelled
	if o.Status != orderValueObjects.OrderStatusPending {
		return errors.ErrOrderCancellationFailed
	}

	// Cancel the order
	o.Status = orderValueObjects.OrderStatusCancelled

	return nil
}

// recalculateTotal recalculates the total amount based on items
func (o *Order) recalculateTotal() error {
	total := valueobjects.NewMoney(0, o.Currency)

	for _, item := range o.Items {
		newTotal, err := total.Add(item.TotalPrice())
		if err != nil {
			return errors.ErrOrderItemProcessingFailed
		}
		total = newTotal
	}

	o.TotalAmount = total
	return nil
}

// IsOwnedBy checks if the order belongs to the specified user
func (o *Order) IsOwnedBy(userID string) bool {
	return o.UserID == userID
}
