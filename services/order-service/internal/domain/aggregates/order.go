package aggregates

import (
	"fmt"
	"time"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/entities"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// Business Rules:
// 1. Each user can have orders with sequential numbers (1, 2, 3...)
// 2. Each product can appear only once per order
// 3. If the same product is added again, use AddItem to update quantity

// Order represents the main aggregate for order management
type Order struct {
	ID              string
	UserID          string
	Number          int64
	Status          valueobjects.OrderStatus
	Items           []*entities.OrderItem
	ShippingAddress string
	Currency        string
	TotalAmount     *valueobjects.Money
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewOrder creates a new Order aggregate
func NewOrder(id, userID string, number int64, shippingAddress, currency string) (*Order, error) {
	// Business rules validation
	if err := validateOrderBusinessRules(id, userID, shippingAddress, currency); err != nil {
		return nil, fmt.Errorf("failed to validate order business rules: %w", err)
	}

	// Create zero money for initial total
	zeroMoney, err := valueobjects.NewMoney(0, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create zero money: %w", err)
	}

	now := time.Now()
	return &Order{
		ID:              id,
		UserID:          userID,
		Number:          number,
		Status:          valueobjects.OrderStatusPending,
		Items:           make([]*entities.OrderItem, 0),
		ShippingAddress: shippingAddress,
		Currency:        zeroMoney.Currency,
		TotalAmount:     zeroMoney,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// validateOrderBusinessRules validates business rules for order creation
func validateOrderBusinessRules(id, userID, shippingAddress, currency string) error {
	if id == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if shippingAddress == "" {
		return fmt.Errorf("shipping address cannot be empty")
	}
	if currency == "" {
		return fmt.Errorf("currency cannot be empty")
	}
	if len(currency) != 3 {
		return fmt.Errorf("currency must be exactly 3 characters")
	}
	return nil
}

// AddItem adds a new item to the order or updates quantity if product already exists
// This method handles both new items and existing product updates gracefully
func (o *Order) AddItem(productID, productName string, quantity int32, unitPrice *valueobjects.Money) error {
	if unitPrice.Currency != o.Currency {
		return fmt.Errorf("currency mismatch: item currency %s differs from order currency %s", unitPrice.Currency, o.Currency)
	}

	// Check if product already exists and update quantity
	for _, existingItem := range o.Items {
		if existingItem.ProductID == productID {
			// Update existing item quantity
			existingItem.Quantity += quantity
			if err := o.recalculateTotal(); err != nil {
				return fmt.Errorf("failed to recalculate total: %w", err)
			}
			o.UpdatedAt = time.Now()
			return nil
		}
	}

	// Add new item if product doesn't exist
	item := entities.NewOrderItem(productID, productName, quantity, unitPrice)
	o.Items = append(o.Items, item)
	if err := o.recalculateTotal(); err != nil {
		return fmt.Errorf("failed to recalculate total: %w", err)
	}
	o.UpdatedAt = time.Now()

	return nil
}

// RemoveItem removes an item from the order by product ID
func (o *Order) RemoveItem(productID string) error {
	for i, item := range o.Items {
		if item.ProductID == productID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			if err := o.recalculateTotal(); err != nil {
				return fmt.Errorf("failed to recalculate total: %w", err)
			}
			o.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("item with product ID %s not found", productID)
}

// UpdateStatus updates the order status
func (o *Order) UpdateStatus(newStatus valueobjects.OrderStatus) error {
	if err := o.Status.ValidateTransition(newStatus); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	o.Status = newStatus
	o.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if err := o.Status.ValidateTransition(valueobjects.OrderStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	o.Status = valueobjects.OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

// recalculateTotal recalculates the total amount based on items
func (o *Order) recalculateTotal() error {
	total, err := valueobjects.NewMoney(0, o.Currency)
	if err != nil {
		return fmt.Errorf("failed to create zero money: %w", err)
	}

	for _, item := range o.Items {
		newTotal, err := total.Add(item.TotalPrice())
		if err != nil {
			return fmt.Errorf("failed to add item total: %w", err)
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

// CanBeModified checks if the order can be modified
func (o *Order) CanBeModified() bool {
	return o.Status == valueobjects.OrderStatusPending || o.Status == valueobjects.OrderStatusConfirmed
}
