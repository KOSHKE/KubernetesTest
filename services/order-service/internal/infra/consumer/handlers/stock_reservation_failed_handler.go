package handlers

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	appservices "ecommerce-platform/services/order-service/internal/application/services"
)

// StockReservationFailedHandler handles StockReservationFailed events at infrastructure level
type StockReservationFailedHandler struct {
	stockEventService *appservices.StockEventService
	logger            logger.Logger
}

// NewStockReservationFailedHandler creates a new StockReservationFailedHandler instance
func NewStockReservationFailedHandler(stockEventService *appservices.StockEventService, logger logger.Logger) *StockReservationFailedHandler {
	return &StockReservationFailedHandler{
		stockEventService: stockEventService,
		logger:            logger,
	}
}

// Handle processes StockReservationFailed events with error handling
func (h *StockReservationFailedHandler) Handle(ctx context.Context, evt *events.StockReservationFailed) error {
	h.logger.Info("processing stock reservation failed event", "orderID", evt.OrderId, "reason", evt.Reason)
	if err := h.stockEventService.ProcessStockReservationFailed(ctx, evt); err != nil {
		h.logger.Error("failed to process stock reservation failed event", "orderID", evt.OrderId, "error", err)
		return err
	}
	h.logger.Info("stock reservation failed event processed successfully", "orderID", evt.OrderId)
	return nil
}
