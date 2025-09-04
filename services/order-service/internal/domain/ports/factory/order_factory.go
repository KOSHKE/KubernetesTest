package factory

import (
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// OrderFactory defines the interface for creating Order entities
type OrderFactory interface {
	// CreateOrder creates a new Order with validation
	CreateOrder(userID string, shippingAddress orderValueObjects.ShippingAddress, currency valueobjects.Currency) (*aggregates.Order, error)
}
