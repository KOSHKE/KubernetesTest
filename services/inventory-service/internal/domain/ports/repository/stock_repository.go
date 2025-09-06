package repository

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
)

// StockRepository defines the interface for stock data access
type StockRepository interface {
	// Basic CRUD operations
	CreateStock(ctx context.Context, stock *entities.Stock) error
	GetStockByID(ctx context.Context, id string) (*entities.Stock, error)
	UpdateStock(ctx context.Context, stock *entities.Stock) error
	StockExistsByID(ctx context.Context, id string) (bool, error)

	// Stock-specific operations
	GetStockByProductID(ctx context.Context, productID string) (*entities.Stock, error)
}
