package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// AddItemToOrderUseCase handles adding items to an order
type AddItemToOrderUseCase struct {
	logger    logger.Logger
	orderRepo repository.OrderRepository
}

// NewAddItemToOrderUseCase creates a new add item to order use case
func NewAddItemToOrderUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
) *AddItemToOrderUseCase {
	return &AddItemToOrderUseCase{
		logger:    logger,
		orderRepo: orderRepo,
	}
}

// Execute adds an item to an existing order
func (uc *AddItemToOrderUseCase) Execute(ctx context.Context, orderID, productID, productName string, quantity int32, price valueobjects.Money) (*aggregates.Order, error) {
	// Get order from repository
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to retrieve order", "order_id", orderID, "error", err)
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Add item to order
	if err := order.AddItem(productID, productName, quantity, price); err != nil {
		uc.logger.Error("Failed to add item to order", "order_id", orderID, "product_id", productID, "error", err)
		return nil, errors.ErrOrderItemAdditionFailed
	}

	// Save updated order
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		uc.logger.Error("Failed to update order", "order_id", orderID, "error", err)
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
