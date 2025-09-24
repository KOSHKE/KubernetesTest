package consumer

import (
	"context"

	dto "ecommerce-platform/pkg/common/dto/order-service"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/order-service/internal/application/services"
	"ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// EventHandlers contains all event handlers for order service
type EventHandlers struct {
	applicationService *services.OrderApplicationService
	logger             logger.Logger
}

// NewEventHandlers creates a new EventHandlers instance
func NewEventHandlers(applicationService *services.OrderApplicationService, logger logger.Logger) *EventHandlers {
	return &EventHandlers{
		applicationService: applicationService,
		logger:             logger,
	}
}

// HandleStockReserved processes StockReserved events
func (h *EventHandlers) HandleStockReserved(ctx context.Context, evt *events.StockReserved) error {
	// Update order status to confirmed
	req := &dto.UpdateOrderStatusRequest{
		OrderID: evt.OrderId,
		Status:  valueobjects.OrderStatusConfirmed.String(),
	}

	_, err := h.applicationService.UpdateOrderStatus(ctx, req)
	if err != nil {
		h.logger.Error("failed to update order status to confirmed", "orderID", evt.OrderId, "error", err)
		return err
	}

	return nil
}

// HandleStockReleased processes StockReleased events
func (h *EventHandlers) HandleStockReleased(ctx context.Context, evt *events.StockReleased) error {
	// Update order status to stock released
	req := &dto.UpdateOrderStatusRequest{
		OrderID: evt.OrderId,
		Status:  valueobjects.OrderStatusStockReleased.String(),
	}

	_, err := h.applicationService.UpdateOrderStatus(ctx, req)
	if err != nil {
		h.logger.Error("failed to update order status to stock released", "orderID", evt.OrderId, "error", err)
		return err
	}

	return nil
}

// HandleStockCommitted processes StockCommitted events
func (h *EventHandlers) HandleStockCommitted(ctx context.Context, evt *events.StockCommitted) error {
	// Update order status to completed
	req := &dto.UpdateOrderStatusRequest{
		OrderID: evt.OrderId,
		Status:  valueobjects.OrderStatusCompleted.String(),
	}

	_, err := h.applicationService.UpdateOrderStatus(ctx, req)
	if err != nil {
		h.logger.Error("failed to update order status to completed", "orderID", evt.OrderId, "error", err)
		return err
	}

	return nil
}

// HandlePaymentProcessed processes PaymentProcessed events
func (h *EventHandlers) HandlePaymentProcessed(ctx context.Context, evt *events.PaymentProcessed) error {
	var status valueobjects.OrderStatus
	if evt.Success {
		status = valueobjects.OrderStatusPaid
	} else {
		status = valueobjects.OrderStatusPaymentFailed
	}

	// Update order status based on payment result
	req := &dto.UpdateOrderStatusRequest{
		OrderID: evt.OrderId,
		Status:  status.String(),
	}

	_, err := h.applicationService.UpdateOrderStatus(ctx, req)
	if err != nil {
		h.logger.Error("failed to update order status after payment", "orderID", evt.OrderId, "success", evt.Success, "error", err)
		return err
	}

	return nil
}
