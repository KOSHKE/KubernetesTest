package repository

import (
	"context"
)

// OrderRepositoryFacade defines the unified interface for order data access
// This facade combines OrderRepository and OutboxRepository functionality
type OrderRepositoryFacade interface {
	// Order operations
	OrderRepository

	// Outbox operations
	OutboxRepository

	// Transaction support - all repositories participate in the same transaction
	WithTransaction(ctx context.Context, fn func(OrderRepositoryFacade) error) error
}
