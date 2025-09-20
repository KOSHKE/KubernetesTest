package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// AddItemToOrderUseCase handles adding items to an order
type AddItemToOrderUseCase struct{}

// NewAddItemToOrderUseCase creates a new add item to order use case
func NewAddItemToOrderUseCase() *AddItemToOrderUseCase {
	return &AddItemToOrderUseCase{}
}

// Execute adds an item to an existing order
func (uc *AddItemToOrderUseCase) Execute(ctx context.Context, orderID, productID, productName string, quantity int32, price valueobjects.Money, repo repository.OrderRepositoryFacade) (*aggregates.Order, error) {
	// Get order from repository
	order, err := repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Add item to order
	if err := order.AddItem(productID, productName, quantity, price); err != nil {
		return nil, errors.ErrOrderItemAdditionFailed
	}

	// Save updated order
	if err := repo.Update(ctx, order); err != nil {
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
