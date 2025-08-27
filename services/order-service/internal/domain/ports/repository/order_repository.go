package repository

import (
	"context"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/aggregates"
)

type OrderRepository interface {
	Create(ctx context.Context, order *aggregates.Order) error
	GetByID(ctx context.Context, id string) (*aggregates.Order, error)
	GetByUserID(ctx context.Context, userID string, page, limit int) ([]*aggregates.Order, int64, error)
	Update(ctx context.Context, order *aggregates.Order) error
	Delete(ctx context.Context, id string) error
	NextOrderNumber(ctx context.Context, userID string) (int64, error)
}
