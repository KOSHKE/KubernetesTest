package services

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/validation"
	"ecommerce-platform/services/order-service/internal/application/dto"
	"ecommerce-platform/services/order-service/internal/application/usecases"
	"ecommerce-platform/services/order-service/internal/domain/entities"
	"ecommerce-platform/services/order-service/internal/domain/ports/publisher"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"ecommerce-platform/services/order-service/internal/metrics"
)

// OrderApplicationService orchestrates order operations and provides a unified interface
type OrderApplicationService struct {
	createOrderUseCase         *usecases.CreateOrderUseCase
	getOrderUseCase            *usecases.GetOrderUseCase
	getUserOrdersUseCase       *usecases.GetUserOrdersUseCase
	updateOrderStatusUseCase   *usecases.UpdateOrderStatusUseCase
	cancelOrderUseCase         *usecases.CancelOrderUseCase
	addItemToOrderUseCase      *usecases.AddItemToOrderUseCase
	removeItemFromOrderUseCase *usecases.RemoveItemFromOrderUseCase
	validator                  *validation.Validate
	logger                     logger.Logger
}

// NewOrderApplicationService creates a new OrderApplicationService instance
func NewOrderApplicationService(
	orderRepo repository.OrderRepository,
	orderPublisher publisher.OrderCreatedPublisher,
	l logger.Logger,
	metrics metrics.OrderMetrics,
) *OrderApplicationService {
	v := validation.New()

	return &OrderApplicationService{
		createOrderUseCase:         usecases.NewCreateOrderUseCase(l, orderRepo, orderPublisher, metrics),
		getOrderUseCase:            usecases.NewGetOrderUseCase(l, orderRepo),
		getUserOrdersUseCase:       usecases.NewGetUserOrdersUseCase(l, orderRepo),
		updateOrderStatusUseCase:   usecases.NewUpdateOrderStatusUseCase(l, orderRepo),
		cancelOrderUseCase:         usecases.NewCancelOrderUseCase(l, orderRepo),
		addItemToOrderUseCase:      usecases.NewAddItemToOrderUseCase(l, orderRepo),
		removeItemFromOrderUseCase: usecases.NewRemoveItemFromOrderUseCase(l, orderRepo),
		validator:                  v,
		logger:                     l,
	}
}

// mapOrderItemsToDTO converts domain order items to DTO responses
func mapOrderItemsToDTO(items []*entities.OrderItem) []*dto.OrderItemResponse {
	dtoItems := make([]*dto.OrderItemResponse, len(items))
	for i, item := range items {
		dtoItems[i] = &dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice(),
		}
	}
	return dtoItems
}

// CreateOrder creates a new order
func (s *OrderApplicationService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	items := make([]*orderValueObjects.OrderItem, len(req.Items))
	for i, item := range req.Items {
		orderItem, err := orderValueObjects.NewOrderItem(item.ProductID, item.ProductName, item.Quantity, item.Price)
		if err != nil {
			s.logger.Error("Failed to create order item", "error", err)
			return nil, errors.ErrOrderValidationFailed
		}
		items[i] = orderItem
	}

	order, err := s.createOrderUseCase.Execute(ctx, req.UserID, req.ShippingAddress, req.Currency, items)
	if err != nil {
		return nil, err
	}

	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           mapOrderItemsToDTO(order.Items),
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code(),
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}

// GetOrder retrieves an order by ID
func (s *OrderApplicationService) GetOrder(ctx context.Context, req *dto.GetOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	order, err := s.getOrderUseCase.Execute(ctx, req.OrderID, req.UserID)
	if err != nil {
		return nil, err
	}

	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           mapOrderItemsToDTO(order.Items),
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code(),
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}

// GetUserOrders retrieves paginated list of user orders
func (s *OrderApplicationService) GetUserOrders(ctx context.Context, req *dto.GetUserOrdersRequest) (*dto.OrdersListResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	orders, total, err := s.getUserOrdersUseCase.Execute(ctx, req.UserID, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	// Convert aggregates to DTO responses
	orderResponses := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = &dto.OrderResponse{
			ID:              order.ID,
			UserID:          order.UserID,
			Status:          string(order.Status),
			Items:           mapOrderItemsToDTO(order.Items),
			ShippingAddress: order.ShippingAddress.Value,
			Currency:        order.Currency.Code(),
			TotalAmount:     order.TotalAmount,
			CreatedAt:       order.CreatedAt,
			UpdatedAt:       order.UpdatedAt,
		}
	}

	return &dto.OrdersListResponse{
		Orders: orderResponses,
		Total:  total,
		Page:   int(req.Page),
		Limit:  int(req.Limit),
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderApplicationService) UpdateOrderStatus(ctx context.Context, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	order, err := s.updateOrderStatusUseCase.Execute(ctx, req.OrderID, req.Status)
	if err != nil {
		return nil, err
	}

	// Convert Order aggregate to OrderResponse
	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           mapOrderItemsToDTO(order.Items),
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code(),
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}

// CancelOrder cancels an order
func (s *OrderApplicationService) CancelOrder(ctx context.Context, req *dto.CancelOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	order, err := s.cancelOrderUseCase.Execute(ctx, req.OrderID, req.UserID)
	if err != nil {
		return nil, err
	}

	// Convert Order aggregate to OrderResponse
	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           mapOrderItemsToDTO(order.Items),
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code(),
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}

// AddItemToOrder adds an item to an existing order
func (s *OrderApplicationService) AddItemToOrder(ctx context.Context, req *dto.AddItemToOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	order, err := s.addItemToOrderUseCase.Execute(ctx, req.OrderID, req.ProductID, req.ProductName, req.Quantity, req.Price)
	if err != nil {
		return nil, err
	}

	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           mapOrderItemsToDTO(order.Items),
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code(),
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}

// RemoveItemFromOrder removes an item from an existing order
func (s *OrderApplicationService) RemoveItemFromOrder(ctx context.Context, req *dto.RemoveItemFromOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	order, err := s.removeItemFromOrderUseCase.Execute(ctx, req.OrderID, req.UserID, req.ProductID)
	if err != nil {
		return nil, err
	}

	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           mapOrderItemsToDTO(order.Items),
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code(),
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}
