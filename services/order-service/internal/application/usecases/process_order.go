package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	"ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// ProcessOrderUseCase handles order processing business logic
type ProcessOrderUseCase struct {
	logger    logger.Logger
	orderRepo repository.OrderRepository
}

// NewProcessOrderUseCase creates a new process order use case
func NewProcessOrderUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
) *ProcessOrderUseCase {
	return &ProcessOrderUseCase{
		logger:    logger,
		orderRepo: orderRepo,
	}
}

// Execute processes an order
func (uc *ProcessOrderUseCase) Execute(ctx context.Context, orderID string) (*aggregates.Order, error) {
	var order *aggregates.Order

	// Execute all operations within a transaction
	err := uc.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		// Get order from repository
		var err error
		order, err = txRepo.GetByID(ctx, orderID)
		if err != nil {
			uc.logger.Error("Failed to retrieve order for processing", "order_id", orderID, "error", err)
			return errors.ErrOrderRetrievalFailed
		}

		// Process order - update status to CONFIRMED using domain method
		if err := order.SetStatus(valueobjects.OrderStatusConfirmed); err != nil {
			uc.logger.Error("Failed to set order status to confirmed", "order_id", orderID, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		// Update order in repository
		if err := txRepo.Update(ctx, order); err != nil {
			uc.logger.Error("Failed to update processed order", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}
