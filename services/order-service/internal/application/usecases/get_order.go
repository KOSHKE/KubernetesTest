package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// GetOrderUseCase handles order retrieval business logic
type GetOrderUseCase struct {
	logger    logger.Logger
	orderRepo repository.OrderRepository
}

// NewGetOrderUseCase creates a new get order use case
func NewGetOrderUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
) *GetOrderUseCase {
	return &GetOrderUseCase{
		logger:    logger,
		orderRepo: orderRepo,
	}
}

// Execute retrieves an order by ID
func (uc *GetOrderUseCase) Execute(ctx context.Context, orderID, userID string) (*aggregates.Order, error) {
	// Get order from repository
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Check access control - user can only get their own orders
	if order.UserID != userID {
		uc.logger.Error("Access denied to order", "error", errors.ErrOrderAccessDenied)
		return nil, errors.ErrOrderAccessDenied
	}

	return order, nil
}
