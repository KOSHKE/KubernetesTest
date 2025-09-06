package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/publisher"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"ecommerce-platform/services/order-service/internal/metrics"
)

// CreateOrderUseCase handles order creation business logic
type CreateOrderUseCase struct {
	logger         logger.Logger
	orderRepo      repository.OrderRepository
	orderPublisher publisher.OrderCreatedPublisher
	metrics        metrics.OrderMetrics
}

// NewCreateOrderUseCase creates a new create order use case
func NewCreateOrderUseCase(
	logger logger.Logger,
	orderRepo repository.OrderRepository,
	orderPublisher publisher.OrderCreatedPublisher,
	metrics metrics.OrderMetrics,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		logger:         logger,
		orderRepo:      orderRepo,
		orderPublisher: orderPublisher,
		metrics:        metrics,
	}
}

// Execute creates a new order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, userID, shippingAddress, currency string, items []*orderValueObjects.OrderItem) (*aggregates.Order, error) {
	// Create value objects
	shippingAddr, err := orderValueObjects.NewShippingAddress(shippingAddress)
	if err != nil {
		uc.logger.Error("Failed to create shipping address", "error", err)
		uc.metrics.OrderCreationFailed("shipping_address_invalid")
		return nil, errors.ErrOrderValidationFailed
	}

	currencyObj, err := valueobjects.NewCurrency(currency)
	if err != nil {
		uc.logger.Error("Failed to create currency", "error", err)
		uc.metrics.OrderCreationFailed("currency_invalid")
		return nil, errors.ErrOrderValidationFailed
	}

	// Create order using domain constructor
	order, err := aggregates.NewOrder(userID, shippingAddr, currencyObj)
	if err != nil {
		uc.logger.Error("Failed to create order aggregate", "error", err)
		uc.metrics.OrderCreationFailed("aggregate_creation_failed")
		return nil, errors.ErrOrderCreationFailed
	}

	// Add items to order with product details
	for _, item := range items {
		if err := order.AddItem(item.ProductID, item.ProductName, item.Quantity, item.Price); err != nil {
			uc.logger.Error("Failed to add item to order", "product_id", item.ProductID, "error", err)
			uc.metrics.OrderCreationFailed("item_addition_failed")
			return nil, errors.ErrOrderItemAdditionFailed
		}
	}

	// Execute order creation and event publishing within a transaction
	if err := uc.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		// Save order to repository
		if err := txRepo.Create(ctx, order); err != nil {
			uc.logger.Error("Failed to save order", "error", err)
			uc.metrics.OrderCreationFailed("persistence_failed")
			return errors.ErrOrderPersistenceFailed
		}

		// Publish OrderCreated event - CRITICAL: must succeed
		if uc.orderPublisher != nil {
			event := &events.OrderCreated{
				OrderId:     order.ID,
				UserId:      order.UserID,
				TotalAmount: order.TotalAmount.Amount,
				Currency:    order.Currency.String(),
			}
			if err := uc.orderPublisher.PublishOrderCreated(ctx, event); err != nil {
				uc.logger.Error("Failed to publish OrderCreated event", "error", err, "orderID", order.ID)
				// If event publishing fails, the transaction will be rolled back
				return errors.ErrOrderEventPublishFailed
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Record success metrics
	uc.metrics.OrderCreated(currency)

	return order, nil
}
