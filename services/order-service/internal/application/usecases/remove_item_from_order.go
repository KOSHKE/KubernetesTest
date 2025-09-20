package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// RemoveItemFromOrderUseCase handles removing items from an order
type RemoveItemFromOrderUseCase struct{}

// NewRemoveItemFromOrderUseCase creates a new remove item from order use case
func NewRemoveItemFromOrderUseCase() *RemoveItemFromOrderUseCase {
	return &RemoveItemFromOrderUseCase{}
}

// Execute removes an item from an existing order
func (uc *RemoveItemFromOrderUseCase) Execute(ctx context.Context, orderID, userID, productID string, repo repository.OrderRepositoryFacade) (*aggregates.Order, error) {
	// Get order from repository to verify it exists
	order, err := repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Check access control - user can only modify their own orders
	if order.UserID != userID {
		return nil, errors.ErrOrderAccessDenied
	}

	// Remove item from order aggregate
	if err := order.RemoveItem(productID); err != nil {
		return nil, errors.ErrOrderItemRemovalFailed
	}

	// Save updated order with items
	if err := repo.Update(ctx, order); err != nil {
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
