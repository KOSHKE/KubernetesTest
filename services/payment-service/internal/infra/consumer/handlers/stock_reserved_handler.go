package handlers

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	appservices "ecommerce-platform/services/payment-service/internal/application/services"
)

// StockReservedHandler handles StockReserved events at infrastructure level
type StockReservedHandler struct {
	stockEventService *appservices.StockEventService
	logger            logger.Logger
}

// NewStockReservedHandler creates a new StockReservedHandler instance
func NewStockReservedHandler(stockEventService *appservices.StockEventService, logger logger.Logger) *StockReservedHandler {
	return &StockReservedHandler{
		stockEventService: stockEventService,
		logger:            logger,
	}
}

// Handle processes StockReserved events with error handling
func (h *StockReservedHandler) Handle(ctx context.Context, evt *events.StockReserved) error {
	// Process event using application service
	if err := h.stockEventService.ProcessStockReserved(ctx, evt); err != nil {
		h.logger.Error("failed to process stock reserved event", "error", err, "orderID", evt.OrderId)
		return err
	}

	return nil
}
