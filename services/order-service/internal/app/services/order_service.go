package services

import (
	"context"
	"fmt"
	"time"

	derrors "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/models"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/ports/clock"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/ports/productinfo"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/ports/publisher"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/ports/repository"

	events "github.com/kubernetestest/ecommerce-platform/proto-go/events"

	"go.uber.org/zap"
)

type OrderService struct {
	orderRepo repository.OrderRepository
	clock     clock.Clock
	pub       publisher.EventPublisher
	products  productinfo.Provider
	logger    *zap.SugaredLogger
}

type CreateOrderRequest struct {
	UserID          string
	Items           []OrderItemRequest
	ShippingAddress string
	Currency        string
}

type OrderItemRequest struct {
	ProductID   string
	ProductName string
	Quantity    int32
	Price       int64
}

type UpdateOrderStatusRequest struct {
	OrderID string
	Status  models.OrderStatus
}

// NewOrderService creates a fully configured service instance
func NewOrderService(orderRepo repository.OrderRepository, c clock.Clock, p publisher.EventPublisher, prod productinfo.Provider, l *zap.Logger) *OrderService {
	s := &OrderService{orderRepo: orderRepo, clock: c, pub: p, products: prod}
	if l != nil {
		s.logger = l.Sugar()
	}
	return s
}

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*models.Order, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("order must contain at least one item")
	}
	if req.ShippingAddress == "" {
		return nil, fmt.Errorf("shipping address is required")
	}

	// Determine next sequential number per user and stable order ID
	num, err := s.orderRepo.NextOrderNumber(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate order number: %w", err)
	}

	order := &models.Order{UserID: req.UserID, Number: num, Status: models.OrderStatusPending, ShippingAddress: req.ShippingAddress, Items: make([]models.OrderItem, 0, len(req.Items)), Currency: req.Currency}
	order.ID = fmt.Sprintf("%s-#%d", req.UserID, num)
	now := s.now()
	order.CreatedAt, order.UpdatedAt = now, now

	for _, item := range req.Items {
		name := item.ProductName
		price := item.Price
		currency := req.Currency
		if s.products != nil {
			if info, err := s.products.GetProduct(ctx, item.ProductID); err == nil {
				name, price, currency = info.Name, info.Price, info.Currency
			} else if s.logger != nil {
				s.logger.Warnw("product lookup failed; using payload values", "productID", item.ProductID, "error", err)
			}
		}
		if err := order.AddItem(item.ProductID, name, item.Quantity, price, currency); err != nil {
			return nil, fmt.Errorf("failed to add item %s: %w", item.ProductID, err)
		}
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Publish OrderCreated (best-effort)
	if s.pub != nil {
		evt := &events.OrderCreated{
			OrderId:     order.ID,
			UserId:      order.UserID,
			TotalAmount: order.TotalAmount,
			Currency:    order.Currency,
			OccurredAt:  s.now().Format(time.RFC3339),
		}
		for _, it := range order.Items {
			evt.Items = append(evt.Items, &events.OrderItem{ProductId: it.ProductID, Quantity: it.Quantity})
		}
		if err := s.pub.PublishOrderCreated(ctx, evt); err != nil && s.logger != nil {
			s.logger.Errorw("failed to publish OrderCreated", "orderID", order.ID, "error", err)
		}
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID, userID string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derrors.ErrOrderNotFound, err)
	}
	if order.UserID != userID {
		return nil, derrors.ErrOrderAccessDenied
	}
	return order, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID string, page, limit int) ([]*models.Order, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	orders, total, err := s.orderRepo.GetByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}
	return orders, total, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *UpdateOrderStatusRequest) (*models.Order, error) {
	var updated *models.Order
	err := s.modifyOrder(ctx, req.OrderID, "", func(order *models.Order) error {
		if err := order.UpdateStatus(req.Status); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		updated = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID, userID string) (*models.Order, error) {
	var updated *models.Order
	err := s.modifyOrder(ctx, orderID, userID, func(order *models.Order) error {
		if err := order.Cancel(); err != nil {
			return fmt.Errorf("failed to cancel order: %w", err)
		}
		updated = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *OrderService) AddItemToOrder(ctx context.Context, orderID, userID, productID, productName string, quantity int32, price int64, currency string) (*models.Order, error) {
	var updated *models.Order
	err := s.modifyOrder(ctx, orderID, userID, func(order *models.Order) error {
		if err := order.AddItem(productID, productName, quantity, price, currency); err != nil {
			return fmt.Errorf("failed to add item: %w", err)
		}
		updated = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *OrderService) RemoveItemFromOrder(ctx context.Context, orderID, userID, productID string) (*models.Order, error) {
	var updated *models.Order
	err := s.modifyOrder(ctx, orderID, userID, func(order *models.Order) error {
		if err := order.RemoveItem(productID); err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}
		updated = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// helpers
func (s *OrderService) now() time.Time {
	if s.clock != nil {
		return s.clock.Now()
	}
	return time.Now()
}

// modifyOrder centralizes retrieval, optional ownership check, timestamp update, and persistence
func (s *OrderService) modifyOrder(ctx context.Context, orderID, userID string, modifyFunc func(*models.Order) error) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("%w: %v", derrors.ErrOrderNotFound, err)
	}
	if userID != "" && order.UserID != userID {
		return derrors.ErrOrderAccessDenied
	}
	if err := modifyFunc(order); err != nil {
		return err
	}
	order.UpdatedAt = s.now()
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	return nil
}
