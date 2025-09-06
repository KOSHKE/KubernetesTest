package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// RemoveItemFromOrderUseCase handles removing items from an order
type RemoveItemFromOrderUseCase struct {
	logger    logger.Logger
	orderRepo repository.OrderRepository
}

// NewRemoveItemFromOrderUseCase creates a new remove item from order use case
func NewRemoveItemFromOrderUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
) *RemoveItemFromOrderUseCase {
	return &RemoveItemFromOrderUseCase{
		logger:    logger,
		orderRepo: orderRepo,
	}
}

// Execute removes an item from an existing order
func (uc *RemoveItemFromOrderUseCase) Execute(ctx context.Context, orderID, userID, productID string) (*aggregates.Order, error) {
	var order *aggregates.Order

	// Execute all operations within a transaction
	err := uc.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		// Get order from repository to verify it exists
		var err error
		order, err = txRepo.GetByID(ctx, orderID)
		if err != nil {
			uc.logger.Error("Failed to retrieve order for item removal", "order_id", orderID, "error", err)
			return errors.ErrOrderRetrievalFailed
		}

		// Check access control - user can only modify their own orders
		if order.UserID != userID {
			uc.logger.Error("Access denied to order", "error", errors.ErrOrderAccessDenied)
			return errors.ErrOrderAccessDenied
		}

		// Remove item from order aggregate
		if err := order.RemoveItem(productID); err != nil {
			uc.logger.Error("Failed to remove item from order aggregate", "order_id", orderID, "product_id", productID, "error", err)
			return errors.ErrOrderItemRemovalFailed
		}

		// Save updated order with items
		if err := txRepo.Update(ctx, order); err != nil {
			uc.logger.Error("Failed to save updated order after item removal", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}
