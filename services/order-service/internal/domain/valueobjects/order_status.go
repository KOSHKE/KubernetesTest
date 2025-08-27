package valueobjects

import (
	"database/sql/driver"
	"fmt"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/errors"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusConfirmed  OrderStatus = "CONFIRMED"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

// AllStatuses contains all valid order statuses
var AllStatuses = []OrderStatus{
	OrderStatusPending,
	OrderStatusConfirmed,
	OrderStatusProcessing,
	OrderStatusShipped,
	OrderStatusDelivered,
	OrderStatusCancelled,
}

// validTransitions defines allowed status transitions
var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
	OrderStatusConfirmed:  {OrderStatusProcessing, OrderStatusCancelled},
	OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
	OrderStatusShipped:    {OrderStatusDelivered},
}

// String returns the string representation of the status
func (os OrderStatus) String() string {
	return string(os)
}

// IsValid checks if the status is valid
func (os OrderStatus) IsValid() bool {
	for _, s := range AllStatuses {
		if os == s {
			return true
		}
	}
	return false
}

// ValidateTransition validates if the status can transition to the new status
func (os OrderStatus) ValidateTransition(to OrderStatus) error {
	if !to.IsValid() {
		return fmt.Errorf("%w: invalid target status %s", errors.ErrInvalidTransition, to)
	}

	// Once cancelled or delivered, no further transitions are allowed
	if os == OrderStatusCancelled {
		return fmt.Errorf("%w: order cannot change status when cancelled", errors.ErrInvalidTransition)
	}
	if os == OrderStatusDelivered {
		return fmt.Errorf("%w: order cannot change status when delivered", errors.ErrInvalidTransition)
	}

	if allowed, exists := validTransitions[os]; exists {
		for _, allowedStatus := range allowed {
			if allowedStatus == to {
				return nil // Valid transition
			}
		}
	}

	return fmt.Errorf("%w: order cannot move from %s to %s", errors.ErrInvalidTransition, os, to)
}

// Validate validates the status
func (os OrderStatus) Validate() error {
	if !os.IsValid() {
		return fmt.Errorf("invalid order status: %s", os)
	}
	return nil
}

// Scan implements the sql.Scanner interface for reading from database
func (os *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		*os = OrderStatusPending
		return nil
	}

	switch v := value.(type) {
	case string:
		*os = OrderStatus(v)
	case []byte:
		*os = OrderStatus(string(v))
	default:
		return fmt.Errorf("cannot scan %T into OrderStatus", value)
	}

	// Validate the scanned value
	if !os.IsValid() {
		return fmt.Errorf("invalid order status scanned from database: %s", *os)
	}

	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (os OrderStatus) Value() (driver.Value, error) {
	if !os.IsValid() {
		return nil, fmt.Errorf("invalid order status: %s", os)
	}
	return string(os), nil
}
