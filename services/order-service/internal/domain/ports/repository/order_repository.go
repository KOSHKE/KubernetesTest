package repository

import (
	"context"

	"ecommerce-platform/services/order-service/internal/domain/aggregates"
)

// OrderRepository defines the interface for order-specific operations
type OrderRepository interface {
	Create(ctx context.Context, order *aggregates.Order) error
	GetByID(ctx context.Context, id string) (*aggregates.Order, error)
	GetByUserID(ctx context.Context, userID string, page, limit int) ([]*aggregates.Order, int64, error)
	Update(ctx context.Context, order *aggregates.Order) error
}
