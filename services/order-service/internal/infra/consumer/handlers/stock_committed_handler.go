package handlers

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	appservices "ecommerce-platform/services/order-service/internal/application/services"
)

// StockCommittedHandler handles StockCommitted events at infrastructure level
type StockCommittedHandler struct {
	stockEventService *appservices.StockEventService
	logger            logger.Logger
}

// NewStockCommittedHandler creates a new StockCommittedHandler instance
func NewStockCommittedHandler(stockEventService *appservices.StockEventService, logger logger.Logger) *StockCommittedHandler {
	return &StockCommittedHandler{
		stockEventService: stockEventService,
		logger:            logger,
	}
}

// Handle processes StockCommitted events with error handling
func (h *StockCommittedHandler) Handle(ctx context.Context, evt *events.StockCommitted) error {
	h.logger.Info("processing stock committed event", "orderID", evt.OrderId)
	if err := h.stockEventService.ProcessStockCommitted(ctx, evt); err != nil {
		h.logger.Error("failed to process stock committed event", "orderID", evt.OrderId, "error", err)
		return err
	}
	h.logger.Info("stock committed event processed successfully", "orderID", evt.OrderId)
	return nil
}
