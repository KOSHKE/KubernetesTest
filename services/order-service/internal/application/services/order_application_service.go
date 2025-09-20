package services

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/pkg/validation"
	"ecommerce-platform/proto-go/common"
	"ecommerce-platform/services/order-service/internal/application/dto"
	"ecommerce-platform/services/order-service/internal/application/usecases"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"fmt"
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
	orderRepo                  repository.OrderRepositoryFacade
	validator                  *validation.Validate
	logger                     logger.Logger
}

// NewOrderApplicationService creates a new OrderApplicationService instance
func NewOrderApplicationService(
	orderRepo repository.OrderRepositoryFacade,
	l logger.Logger,
) *OrderApplicationService {
	v := validation.New()

	return &OrderApplicationService{
		createOrderUseCase:         usecases.NewCreateOrderUseCase(),
		getOrderUseCase:            usecases.NewGetOrderUseCase(),
		getUserOrdersUseCase:       usecases.NewGetUserOrdersUseCase(),
		updateOrderStatusUseCase:   usecases.NewUpdateOrderStatusUseCase(),
		cancelOrderUseCase:         usecases.NewCancelOrderUseCase(),
		addItemToOrderUseCase:      usecases.NewAddItemToOrderUseCase(),
		removeItemFromOrderUseCase: usecases.NewRemoveItemFromOrderUseCase(),
		orderRepo:                  orderRepo,
		validator:                  v,
		logger:                     l,
	}
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

	var order *aggregates.Order

	// Execute all operations within a transaction
	err := s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepositoryFacade) error {
		// Execute use case with transaction repository
		var err error
		order, err = s.createOrderUseCase.Execute(ctx, req.UserID, req.ShippingAddress, req.Currency, items, txRepo)
		if err != nil {
			return err
		}

		// Save event to outbox table (to be published later)
		eventItems := make([]*common.OrderItem, len(order.Items))
		for i, item := range order.Items {
			eventItems[i] = &common.OrderItem{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price: &common.Money{
					Amount:   item.UnitPrice.Amount,
					Currency: item.UnitPrice.Currency.Code,
				},
				Total: &common.Money{
					Amount:   item.TotalPrice().Amount,
					Currency: item.TotalPrice().Currency.Code,
				},
			}
		}

		eventData := dto.OrderEventDTO{
			UserID:      order.UserID,
			Items:       eventItems,
			TotalAmount: order.TotalAmount.Amount,
			Currency:    order.Currency.Code,
		}
		event := outbox.Event{
			AggregateID: order.ID,
			Type:        "OrderCreated",
			Payload:     eventData,
		}

		// Create outbox service using transaction repository
		outboxService := outbox.NewService(txRepo, s.logger)
		if err := outboxService.SaveEvent(ctx, event); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to create order", "userID", req.UserID, "currency", req.Currency, "error", err)
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}

// GetOrder retrieves an order by ID
func (s *OrderApplicationService) GetOrder(ctx context.Context, req *dto.GetOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	order, err := s.getOrderUseCase.Execute(ctx, req.OrderID, req.UserID, s.orderRepo)
	if err != nil {
		s.logger.Error("failed to get order", "orderID", req.OrderID, "userID", req.UserID, "error", err)
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}

// GetUserOrders retrieves paginated list of user orders
func (s *OrderApplicationService) GetUserOrders(ctx context.Context, req *dto.GetUserOrdersRequest) (*dto.OrdersListResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	// Convert DTO to domain parameters
	orders, total, err := s.getUserOrdersUseCase.Execute(ctx, req.UserID, req.Page, req.Limit, s.orderRepo)
	if err != nil {
		s.logger.Error("failed to get user orders", "userID", req.UserID, "page", req.Page, "limit", req.Limit, "error", err)
		return nil, err
	}

	// Convert aggregates to DTO responses
	orderResponses := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = dto.NewOrderResponse(order)
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

	var order *aggregates.Order

	// Execute all operations within a transaction
	err := s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepositoryFacade) error {
		// Execute use case with transaction repository
		var err error
		order, err = s.updateOrderStatusUseCase.Execute(ctx, req.OrderID, req.Status, txRepo)
		return err
	})

	if err != nil {
		s.logger.Error("failed to update order status", "orderID", req.OrderID, "status", req.Status, "error", err)
		return nil, err
	}

	// Convert Order aggregate to OrderResponse
	return dto.NewOrderResponse(order), nil
}

// CancelOrder cancels an order
func (s *OrderApplicationService) CancelOrder(ctx context.Context, req *dto.CancelOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	var order *aggregates.Order

	// Execute all operations within a transaction
	err := s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepositoryFacade) error {
		// Execute use case with transaction repository
		var err error
		order, err = s.cancelOrderUseCase.Execute(ctx, req.OrderID, req.UserID, txRepo)
		if err != nil {
			return err
		}

		// Publish OrderCancelled event via outbox
		eventItems := make([]*common.OrderItem, len(order.Items))
		for i, item := range order.Items {
			eventItems[i] = &common.OrderItem{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price: &common.Money{
					Amount:   item.UnitPrice.Amount,
					Currency: item.UnitPrice.Currency.Code,
				},
			}
		}

		orderEvent := dto.OrderEventDTO{
			UserID:      order.UserID,
			Items:       eventItems,
			TotalAmount: order.TotalAmount.Amount,
			Currency:    order.Currency.Code,
			Reason:      req.Reason,
		}

		outboxEvent := outbox.Event{
			AggregateID: order.ID,
			Type:        "OrderCancelled",
			Payload:     orderEvent,
		}

		if err := txRepo.SaveEvent(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to save OrderCancelled event: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to cancel order", "orderID", req.OrderID, "userID", req.UserID, "error", err)
		return nil, err
	}

	// Convert Order aggregate to OrderResponse
	return dto.NewOrderResponse(order), nil
}

// AddItemToOrder adds an item to an existing order
func (s *OrderApplicationService) AddItemToOrder(ctx context.Context, req *dto.AddItemToOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	var order *aggregates.Order

	// Execute all operations within a transaction
	err := s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepositoryFacade) error {
		// Execute use case with transaction repository
		var err error
		order, err = s.addItemToOrderUseCase.Execute(ctx, req.OrderID, req.ProductID, req.ProductName, req.Quantity, req.Price, txRepo)
		return err
	})

	if err != nil {
		s.logger.Error("failed to add item to order", "orderID", req.OrderID, "productID", req.ProductID, "error", err)
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}

// RemoveItemFromOrder removes an item from an existing order
func (s *OrderApplicationService) RemoveItemFromOrder(ctx context.Context, req *dto.RemoveItemFromOrderRequest) (*dto.OrderResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrOrderValidationFailed
	}

	var order *aggregates.Order

	// Execute all operations within a transaction
	err := s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepositoryFacade) error {
		// Execute use case with transaction repository
		var err error
		order, err = s.removeItemFromOrderUseCase.Execute(ctx, req.OrderID, req.UserID, req.ProductID, txRepo)
		return err
	})

	if err != nil {
		s.logger.Error("failed to remove item from order", "orderID", req.OrderID, "userID", req.UserID, "productID", req.ProductID, "error", err)
		return nil, err
	}

	return dto.NewOrderResponse(order), nil
}
