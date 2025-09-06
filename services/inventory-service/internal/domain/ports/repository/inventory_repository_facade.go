package repository

import (
	"context"
)

// InventoryRepositoryFacade defines the unified interface for inventory data access
// This interface composes ProductRepository, StockRepository and OutboxRepository functionality
type InventoryRepositoryFacade interface {
	ProductRepository
	StockRepository
	OutboxRepository

	// Transaction support - all repositories participate in the same transaction
	WithTransaction(ctx context.Context, fn func(repo InventoryRepositoryFacade) error) error
}
