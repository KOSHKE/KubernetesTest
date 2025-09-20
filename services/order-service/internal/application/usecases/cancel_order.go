package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
)

// CancelOrderUseCase handles order cancellation business logic
type CancelOrderUseCase struct{}

// NewCancelOrderUseCase creates a new cancel order use case
func NewCancelOrderUseCase() *CancelOrderUseCase {
	return &CancelOrderUseCase{}
}

// Execute cancels an order
func (uc *CancelOrderUseCase) Execute(ctx context.Context, orderID, userID string, repo repository.OrderRepositoryFacade) (*aggregates.Order, error) {
	// Get order from repository
	order, err := repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.ErrOrderRetrievalFailed
	}

	// Use domain method to cancel order (includes access control and business rules)
	if err := order.CancelOrder(userID); err != nil {
		return nil, err
	}

	// Save updated order
	if err := repo.Update(ctx, order); err != nil {
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
