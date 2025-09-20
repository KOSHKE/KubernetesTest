package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// GetOrderUseCase handles order retrieval business logic
type GetOrderUseCase struct{}

// NewGetOrderUseCase creates a new get order use case
func NewGetOrderUseCase() *GetOrderUseCase {
	return &GetOrderUseCase{}
}

// Execute retrieves an order by ID
func (uc *GetOrderUseCase) Execute(ctx context.Context, orderID, userID string, repo repository.OrderRepositoryFacade) (*aggregates.Order, error) {
	// Get order from repository
	order, err := repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Check access control - user can only get their own orders
	if order.UserID != userID {
		return nil, errors.ErrOrderAccessDenied
	}

	return order, nil
}
