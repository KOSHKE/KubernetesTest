package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	"ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// UpdateOrderStatusUseCase handles order status update business logic
type UpdateOrderStatusUseCase struct{}

// NewUpdateOrderStatusUseCase creates a new update order status use case
func NewUpdateOrderStatusUseCase() *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{}
}

// Execute updates the status of an order
func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, orderID string, status valueobjects.OrderStatus, repo repository.OrderRepositoryFacade) (*aggregates.Order, error) {
	// Get order from repository
	order, err := repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Use domain method to update order status
	if err := order.SetStatus(status); err != nil {
		return nil, errors.ErrOrderStatusUpdateFailed
	}

	// Save updated order
	if err := repo.Update(ctx, order); err != nil {
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
