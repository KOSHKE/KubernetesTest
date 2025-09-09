package repository

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
)

// StockRepository defines the interface for stock data access
type StockRepository interface {
	// Basic CRUD operations
	GetStockByID(ctx context.Context, id string) (*entities.Stock, error)
	StockExistsByID(ctx context.Context, id string) (bool, error)

	// Stock-specific operations
	GetStockByProductID(ctx context.Context, productID string) (*entities.Stock, error)
	UpsertStock(ctx context.Context, stock *entities.Stock) error

	// Batch operations for performance optimization
	GetStocksByProductIDs(ctx context.Context, productIDs []string, forUpdate bool) (map[string]*entities.Stock, error)
	UpsertStocks(ctx context.Context, stocks []*entities.Stock) error
}
