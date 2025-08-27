package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kubernetestest/ecommerce-platform/proto-go/events"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/dto"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/validators"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/aggregates"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/ports/productinfo"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/ports/publisher"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/ports/repository"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/metrics"
	"go.uber.org/zap"
)

// OrderUseCases handles business logic for order operations
type OrderUseCases struct {
	orderRepo repository.OrderRepository
	pub       publisher.EventPublisher
	products  productinfo.Provider
	logger    *zap.SugaredLogger
	validator *validators.Validator
	metrics   metrics.OrderMetrics
}

// NewOrderUseCases creates a new OrderUseCases instance
func NewOrderUseCases(
	orderRepo repository.OrderRepository,
	p publisher.EventPublisher,
	prod productinfo.Provider,
	l *zap.Logger,
	m metrics.OrderMetrics,
) *OrderUseCases {
	return &OrderUseCases{
		orderRepo: orderRepo,
		pub:       p,
		products:  prod,
		logger:    l.Sugar(),
		validator: validators.NewValidator(),
		metrics:   m,
	}
}

// CreateOrder creates a new order
func (uc *OrderUseCases) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*aggregates.Order, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Determine next sequential number per user
	num, err := uc.orderRepo.NextOrderNumber(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate order number: %w", err)
	}

	// Generate secure order ID
	orderID := "ORD-" + uuid.New().String()

	// Create order aggregate
	order, err := aggregates.NewOrder(orderID, req.UserID, num, req.ShippingAddress, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create order aggregate: %w", err)
	}

	// Add items to order
	for _, item := range req.Items {
		name := item.ProductName
		price := item.Price

		// Try to get product info from inventory service
		if uc.products != nil {
			if info, err := uc.products.GetProduct(ctx, item.ProductID); err == nil {
				name = info.Name
				price = info.Price
			} else if uc.logger != nil {
				uc.logger.Warnw("product lookup failed; using payload values", "productID", item.ProductID, "error", err)
			}
		}

		if err := order.AddItem(item.ProductID, name, item.Quantity, price); err != nil {
			return nil, fmt.Errorf("failed to add item %s: %w", item.ProductID, err)
		}
	}

	// Save order to repository
	if err := uc.orderRepo.Create(ctx, order); err != nil {
		if uc.logger != nil {
			uc.logger.Errorw("failed to create order in database", "orderID", order.ID, "userID", order.UserID, "error", err)
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Publish OrderCreated event (best-effort)
	uc.publishOrderCreated(ctx, order)

	return order, nil
}

// GetOrder retrieves an order by ID with ownership check
func (uc *OrderUseCases) GetOrder(ctx context.Context, req *dto.GetOrderRequest) (*aggregates.Order, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order, err := uc.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrOrderNotFound, err)
	}

	if !order.IsOwnedBy(req.UserID) {
		if uc.logger != nil {
			uc.logger.Warnw("access denied to order", "orderID", req.OrderID, "requestedUserID", req.UserID, "orderUserID", order.UserID)
		}
		return nil, errors.ErrOrderAccessDenied
	}

	return order, nil
}

// GetUserOrders retrieves paginated list of user orders
func (uc *OrderUseCases) GetUserOrders(ctx context.Context, req *dto.GetUserOrdersRequest) ([]*aggregates.Order, int64, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}

	orders, total, err := uc.orderRepo.GetByUserID(ctx, req.UserID, req.Page, req.Limit)
	if err != nil {
		if uc.logger != nil {
			uc.logger.Errorw("failed to retrieve user orders from database", "userID", req.UserID, "error", err)
		}
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}

	return orders, total, nil
}

// UpdateOrderStatus updates the status of an order
func (uc *OrderUseCases) UpdateOrderStatus(ctx context.Context, req *dto.UpdateOrderStatusRequest) (*aggregates.Order, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	order, err := uc.modifyOrder(ctx, req.OrderID, "", func(order *aggregates.Order) error {
		if err := order.UpdateStatus(req.Status); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// CancelOrder cancels an order
func (uc *OrderUseCases) CancelOrder(ctx context.Context, req *dto.CancelOrderRequest) (*aggregates.Order, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	order, err := uc.modifyOrder(ctx, req.OrderID, req.UserID, func(order *aggregates.Order) error {
		if err := order.Cancel(); err != nil {
			return fmt.Errorf("failed to cancel order: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// AddItemToOrder adds an item to an existing order
func (uc *OrderUseCases) AddItemToOrder(ctx context.Context, req *dto.AddItemToOrderRequest) (*aggregates.Order, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("failed to add item to order: %w", err)
	}

	order, err := uc.modifyOrder(ctx, req.OrderID, req.UserID, func(order *aggregates.Order) error {
		if err := order.AddItem(req.ProductID, req.ProductName, req.Quantity, req.Price); err != nil {
			return fmt.Errorf("failed to add item: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// RemoveItemFromOrder removes an item from an existing order
func (uc *OrderUseCases) RemoveItemFromOrder(ctx context.Context, req *dto.RemoveItemFromOrderRequest) (*aggregates.Order, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("failed to remove item from order: %w", err)
	}

	order, err := uc.modifyOrder(ctx, req.OrderID, req.UserID, func(order *aggregates.Order) error {
		if err := order.RemoveItem(req.ProductID); err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// modifyOrder centralizes order modification logic
func (uc *OrderUseCases) modifyOrder(ctx context.Context, orderID, userID string, modifyFunc func(*aggregates.Order) error) (*aggregates.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrOrderNotFound, err)
	}

	if userID != "" && !order.IsOwnedBy(userID) {
		return nil, errors.ErrOrderAccessDenied
	}

	if !order.CanBeModified() {
		return nil, fmt.Errorf("order with status %s cannot be modified", order.Status)
	}

	if err := modifyFunc(order); err != nil {
		return nil, err
	}

	order.UpdatedAt = time.Now()
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		if uc.logger != nil {
			uc.logger.Errorw("failed to save modified order", "orderID", orderID, "error", err)
		}
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return order, nil
}

// publishOrderCreated publishes OrderCreated event asynchronously
func (uc *OrderUseCases) publishOrderCreated(ctx context.Context, order *aggregates.Order) {
	if uc.pub == nil {
		return
	}

	// Publish event asynchronously to avoid blocking the caller
	go func() {
		// Create OrderCreated event
		evt := &events.OrderCreated{
			OrderId:     order.ID,
			UserId:      order.UserID,
			TotalAmount: order.TotalAmount.Amount,
			Currency:    order.Currency,
			OccurredAt:  order.CreatedAt.Format(time.RFC3339),
			Items:       make([]*events.OrderItem, len(order.Items)),
		}

		// Convert order items to event items
		for i, item := range order.Items {
			evt.Items[i] = &events.OrderItem{
				ProductId: item.ProductID,
				Quantity:  item.Quantity,
			}
		}

		// Publish event with background context to avoid cancellation
		bgCtx := context.Background()
		if err := uc.pub.PublishOrderCreated(bgCtx, evt); err != nil {
			if uc.logger != nil {
				uc.logger.Errorw("failed to publish OrderCreated event", "orderID", order.ID, "error", err)
			}
		}
	}()
}
