package usecases

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// GetUserOrdersUseCase handles user orders retrieval business logic
type GetUserOrdersUseCase struct {
	logger    logger.Logger
	orderRepo repository.OrderRepository
}

// NewGetUserOrdersUseCase creates a new get user orders use case
func NewGetUserOrdersUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
) *GetUserOrdersUseCase {
	return &GetUserOrdersUseCase{
		logger:    logger,
		orderRepo: orderRepo,
	}
}

// Execute retrieves orders for a specific user with pagination
func (uc *GetUserOrdersUseCase) Execute(ctx context.Context, userID string, page, limit int) ([]*aggregates.Order, int64, error) {
	// Get orders from repository
	orders, total, err := uc.orderRepo.GetByUserID(ctx, userID, page, limit)
	if err != nil {
		uc.logger.Error("Failed to get user orders", "error", err)
		return nil, 0, err
	}

	return orders, total, nil
}
