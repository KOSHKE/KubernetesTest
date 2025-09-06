package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	"ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// UpdateOrderStatusUseCase handles order status update business logic
type UpdateOrderStatusUseCase struct {
	logger    logger.Logger
	orderRepo repository.OrderRepository
}

// NewUpdateOrderStatusUseCase creates a new update order status use case
func NewUpdateOrderStatusUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
) *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{
		logger:    logger,
		orderRepo: orderRepo,
	}
}

// Execute updates the status of an order
func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, orderID string, status valueobjects.OrderStatus) (*aggregates.Order, error) {
	var order *aggregates.Order

	// Execute all operations within a transaction
	err := uc.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		// Get order from repository
		var err error
		order, err = txRepo.GetByID(ctx, orderID)
		if err != nil {
			uc.logger.Error("Failed to retrieve order for status update", "order_id", orderID, "error", err)
			return errors.ErrOrderRetrievalFailed
		}

		// Use domain method to update order status
		if err := order.SetStatus(status); err != nil {
			uc.logger.Error("Failed to set order status", "order_id", orderID, "new_status", status, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		// Save updated order
		if err := txRepo.Update(ctx, order); err != nil {
			uc.logger.Error("Failed to save updated order status", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}
