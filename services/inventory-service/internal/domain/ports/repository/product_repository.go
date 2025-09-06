package repository

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	// Basic CRUD operations
	CreateProduct(ctx context.Context, product *entities.Product) error
	GetProductByID(ctx context.Context, id string) (*entities.Product, error)
	UpdateProduct(ctx context.Context, product *entities.Product) error
	DeleteProduct(ctx context.Context, id string) error
	ListProducts(ctx context.Context, page, limit int, search string) ([]*entities.Product, int32, error)
	ProductExistsByID(ctx context.Context, id string) (bool, error)
}
