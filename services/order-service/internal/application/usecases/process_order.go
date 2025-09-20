package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	"ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// ProcessOrderUseCase handles order processing business logic
type ProcessOrderUseCase struct{}

// NewProcessOrderUseCase creates a new process order use case
func NewProcessOrderUseCase() *ProcessOrderUseCase {
	return &ProcessOrderUseCase{}
}

// Execute processes an order
func (uc *ProcessOrderUseCase) Execute(ctx context.Context, orderID string, repo repository.OrderRepositoryFacade) (*aggregates.Order, error) {
	// Get order from repository
	order, err := repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Process order - update status to CONFIRMED using domain method
	if err := order.SetStatus(valueobjects.OrderStatusConfirmed); err != nil {
		return nil, errors.ErrOrderStatusUpdateFailed
	}

	// Update order in repository
	if err := repo.Update(ctx, order); err != nil {
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
