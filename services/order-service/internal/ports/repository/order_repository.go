package repository

import (
	"context"
	"order-service/internal/domain/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	GetByUserID(ctx context.Context, userID string, page, limit int) ([]*models.Order, int64, error)
	Update(ctx context.Context, order *models.Order) error
	Delete(ctx context.Context, id string) error
	GetByStatus(ctx context.Context, status models.OrderStatus) ([]*models.Order, error)
}
