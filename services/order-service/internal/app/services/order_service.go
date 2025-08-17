package services

import (
	"context"
	"fmt"
	"time"

	"order-service/internal/domain/models"
	"order-service/internal/ports/clock"
	"order-service/internal/ports/idgen"
	"order-service/internal/ports/publisher"
	"order-service/internal/ports/repository"

	events "proto-go/events"
)

type OrderService struct {
	orderRepo repository.OrderRepository
	clock     clock.Clock
	ids       idgen.IDGenerator
	pub       publisher.EventPublisher
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

func NewOrderService(orderRepo repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

// WithClock allows injecting a custom clock
func (s *OrderService) WithClock(c clock.Clock) *OrderService { s.clock = c; return s }

// WithIDGenerator allows injecting a custom ID generator
func (s *OrderService) WithIDGenerator(g idgen.IDGenerator) *OrderService { s.ids = g; return s }

// WithPublisher sets event publisher
func (s *OrderService) WithPublisher(p publisher.EventPublisher) *OrderService { s.pub = p; return s }

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
	if req.Currency == "" {
		return nil, fmt.Errorf("currency is required")
	}

	order := &models.Order{UserID: req.UserID, Status: models.OrderStatusPending, ShippingAddress: req.ShippingAddress, Items: make([]models.OrderItem, 0, len(req.Items)), Currency: req.Currency}
	order.ID = s.newID("order-")
	now := s.now()
	order.CreatedAt, order.UpdatedAt = now, now

	for _, item := range req.Items {
		if err := order.AddItem(item.ProductID, item.ProductName, item.Quantity, item.Price, req.Currency); err != nil {
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
			OccurredAt:  time.Now().Format(time.RFC3339),
		}
		for _, it := range order.Items {
			evt.Items = append(evt.Items, &events.OrderItem{ProductId: it.ProductID, Quantity: it.Quantity})
		}
		_ = s.pub.PublishOrderCreated(ctx, evt)
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID, userID string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("access denied")
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
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if err := order.UpdateStatus(req.Status); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}
	order.UpdatedAt = s.now()
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID, userID string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}
	if err := order.Cancel(); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}
	order.UpdatedAt = s.now()
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	return order, nil
}

func (s *OrderService) AddItemToOrder(ctx context.Context, orderID, userID, productID, productName string, quantity int32, price int64, currency string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}
	if err := order.AddItem(productID, productName, quantity, price, currency); err != nil {
		return nil, fmt.Errorf("failed to add item: %w", err)
	}
	order.UpdatedAt = s.now()
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	return order, nil
}

func (s *OrderService) RemoveItemFromOrder(ctx context.Context, orderID, userID, productID string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}
	if err := order.RemoveItem(productID); err != nil {
		return nil, fmt.Errorf("failed to remove item: %w", err)
	}
	order.UpdatedAt = s.now()
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	return order, nil
}

// helpers
func (s *OrderService) now() time.Time {
	if s.clock != nil {
		return s.clock.Now()
	}
	return time.Now()
}
func (s *OrderService) newID(prefix string) string {
	if s.ids != nil {
		return s.ids.NewID(prefix)
	}
	return prefix + time.Now().Format("20060102150405")
}
