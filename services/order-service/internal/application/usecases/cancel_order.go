package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// CancelOrderUseCase handles order cancellation business logic
type CancelOrderUseCase struct {
	logger    logger.Logger
	orderRepo repository.OrderRepository
}

// NewCancelOrderUseCase creates a new cancel order use case
func NewCancelOrderUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
) *CancelOrderUseCase {
	return &CancelOrderUseCase{
		logger:    logger,
		orderRepo: orderRepo,
	}
}

// Execute cancels an order
func (uc *CancelOrderUseCase) Execute(ctx context.Context, orderID, userID string) (*aggregates.Order, error) {
	// Get order from repository
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to retrieve order for cancellation", "order_id", orderID, "error", err)
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Use domain method to cancel order (includes access control and business rules)
	if err := order.CancelOrder(userID); err != nil {
		uc.logger.Error("Failed to cancel order", "order_id", orderID, "user_id", userID, "error", err)
		return nil, err
	}

	// Save updated order
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		uc.logger.Error("Failed to update cancelled order", "order_id", orderID, "error", err)
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
