package usecases

import (
	"context"

	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// GetUserOrdersUseCase handles user orders retrieval business logic
type GetUserOrdersUseCase struct{}

// NewGetUserOrdersUseCase creates a new get user orders use case
func NewGetUserOrdersUseCase() *GetUserOrdersUseCase {
	return &GetUserOrdersUseCase{}
}

// Execute retrieves orders for a specific user with pagination
func (uc *GetUserOrdersUseCase) Execute(ctx context.Context, userID string, page, limit int, repo repository.OrderRepositoryFacade) ([]*aggregates.Order, int64, error) {
	// Get orders from repository
	orders, total, err := repo.GetByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
