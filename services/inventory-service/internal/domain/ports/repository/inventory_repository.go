package repository

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/aggregates"
)

// InventoryRepository defines the interface for inventory data access
type InventoryRepository interface {
	// Product operations
	CreateProduct(ctx context.Context, product *aggregates.Product) error
	GetProductByID(ctx context.Context, id string) (*aggregates.Product, error)
	UpdateProduct(ctx context.Context, product *aggregates.Product) error
	DeleteProduct(ctx context.Context, id string) error
	ListProducts(ctx context.Context, page, limit int, search string) ([]*aggregates.Product, int32, error)

	// Stock operations
	UpdateStock(ctx context.Context, productID string, availableQuantity, reservedQuantity int32) error
	GetStockByProductID(ctx context.Context, productID string) (*aggregates.Product, error)
	GetStocksByProductIDs(ctx context.Context, productIDs []string) (map[string]*aggregates.Product, error)

	// Transaction support
	WithTx(ctx context.Context, fn func(repo InventoryRepository) error) error
}
