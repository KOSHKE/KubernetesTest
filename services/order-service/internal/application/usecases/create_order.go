package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// CreateOrderUseCase handles order creation business logic
type CreateOrderUseCase struct{}

// NewCreateOrderUseCase creates a new create order use case
func NewCreateOrderUseCase() *CreateOrderUseCase {
	return &CreateOrderUseCase{}
}

// Execute creates a new order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, userID, shippingAddress, currency string, items []*orderValueObjects.OrderItem, repo repository.OrderRepositoryFacade) (*aggregates.Order, error) {
	// Create value objects
	shippingAddr, err := orderValueObjects.NewShippingAddress(shippingAddress)
	if err != nil {
		return nil, errors.ErrOrderValidationFailed
	}

	currencyObj, err := valueobjects.NewCurrency(currency)
	if err != nil {
		return nil, errors.ErrOrderValidationFailed
	}

	// Create order using domain constructor
	order, err := aggregates.NewOrder(userID, shippingAddr, currencyObj)
	if err != nil {
		return nil, errors.ErrOrderCreationFailed
	}

	// Add items to order with product details
	for _, item := range items {
		if err := order.AddItem(item.ProductID, item.ProductName, item.Quantity, item.Price); err != nil {
			return nil, errors.ErrOrderItemAdditionFailed
		}
	}

	// Save order to repository
	if err := repo.Create(ctx, order); err != nil {
		return nil, errors.ErrOrderPersistenceFailed
	}

	return order, nil
}
