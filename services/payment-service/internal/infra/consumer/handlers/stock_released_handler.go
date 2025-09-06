package handlers

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	appservices "ecommerce-platform/services/payment-service/internal/application/services"
)

// StockReleasedHandler handles StockReleased events at infrastructure level
type StockReleasedHandler struct {
	stockEventService *appservices.StockEventService
	logger            logger.Logger
}

// NewStockReleasedHandler creates a new StockReleasedHandler instance
func NewStockReleasedHandler(stockEventService *appservices.StockEventService, logger logger.Logger) *StockReleasedHandler {
	return &StockReleasedHandler{
		stockEventService: stockEventService,
		logger:            logger,
	}
}

// Handle processes StockReleased events with error handling
func (h *StockReleasedHandler) Handle(ctx context.Context, evt *events.StockReleased) error {
	h.logger.Info("processing stock released event", "orderID", evt.OrderId)
	if err := h.stockEventService.ProcessStockReleased(ctx, evt); err != nil {
		h.logger.Error("failed to process stock released event", "orderID", evt.OrderId, "error", err)
		return err
	}
	h.logger.Info("stock released event processed successfully", "orderID", evt.OrderId)
	return nil
}
