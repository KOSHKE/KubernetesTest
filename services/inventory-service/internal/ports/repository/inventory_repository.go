package repository

import (
	"context"
	"inventory-service/internal/domain/models"
)

type InventoryRepository interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	ListProducts(ctx context.Context, categoryID string, page, limit int, search string) ([]*models.Product, int32, error)

	GetStock(ctx context.Context, productID string) (*models.Stock, error)
	Reserve(ctx context.Context, productID string, qty int32) error
	Release(ctx context.Context, productID string, qty int32) error
}
