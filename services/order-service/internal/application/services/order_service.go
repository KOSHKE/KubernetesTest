package services

import (
	"context"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/dto"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/usecases"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/ports/productinfo"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/ports/publisher"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/ports/repository"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/metrics"
	"go.uber.org/zap"
)

// OrderService provides the main interface for order operations
type OrderService struct {
	useCases *usecases.OrderUseCases
	metrics  metrics.OrderMetrics
	logger   *zap.SugaredLogger
}

// NewOrderService creates a new OrderService instance
func NewOrderService(
	orderRepo repository.OrderRepository,
	p publisher.EventPublisher,
	prod productinfo.Provider,
	l *zap.Logger,
	m metrics.OrderMetrics,
) *OrderService {
	useCases := usecases.NewOrderUseCases(orderRepo, p, prod, l, m)

	return &OrderService{
		useCases: useCases,
		metrics:  m,
		logger:   l.Sugar(),
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.useCases.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.OrderCreated(order.Currency)
	}

	return dto.NewOrderResponse(order), nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, req *dto.GetOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.useCases.GetOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}

// GetUserOrders retrieves paginated list of user orders
func (s *OrderService) GetUserOrders(ctx context.Context, req *dto.GetUserOrdersRequest) (*dto.OrdersListResponse, error) {
	orders, total, err := s.useCases.GetUserOrders(ctx, req)
	if err != nil {
		return nil, err
	}

	return dto.NewOrdersListResponse(orders, total, int64(req.Page), int64(req.Limit)), nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
	order, err := s.useCases.UpdateOrderStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(ctx context.Context, req *dto.CancelOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.useCases.CancelOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}

// AddItemToOrder adds an item to an existing order
func (s *OrderService) AddItemToOrder(ctx context.Context, req *dto.AddItemToOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.useCases.AddItemToOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}

// RemoveItemFromOrder removes an item from an existing order
func (s *OrderService) RemoveItemFromOrder(ctx context.Context, req *dto.RemoveItemFromOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.useCases.RemoveItemFromOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}
