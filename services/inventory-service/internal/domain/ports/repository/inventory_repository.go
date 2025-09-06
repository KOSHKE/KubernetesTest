package repository

import (
	"context"
)

// InventoryRepository defines the unified interface for inventory data access
// This interface composes ProductRepository and StockRepository functionality
type InventoryRepository interface {
	ProductRepository
	StockRepository

	// Transaction support - both repositories participate in the same transaction
	WithTransaction(ctx context.Context, fn func(repo InventoryRepository) error) error
}
