package handlers

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	appservices "ecommerce-platform/services/order-service/internal/application/services"
)

// StockReservedHandler handles StockReserved events at infrastructure level
type StockReservedHandler struct {
	stockEventService *appservices.StockEventService
	logger            logger.Logger
}

// NewStockReservedHandler creates a new StockReservedHandler instance
func NewStockReservedHandler(stockEventService *appservices.StockEventService, logger logger.Logger) *StockReservedHandler {
	return &StockReservedHandler{stockEventService: stockEventService, logger: logger}
}

// Handle processes StockReserved events with error handling
func (h *StockReservedHandler) Handle(ctx context.Context, evt *events.StockReserved) error {
	if err := h.stockEventService.ProcessStockReserved(ctx, evt); err != nil {
		h.logger.Error("failed to process stock reserved event", "error", err, "orderID", evt.OrderId)
		return fmt.Errorf("failed to process stock reserved event: %w", err)
	}

	return nil
}
