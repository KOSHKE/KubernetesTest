package handlers

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/domain/services"
)

// OrderCreatedHandler handles OrderCreated events at infrastructure level
type OrderCreatedHandler struct {
	domainService *services.InventoryDomainService
	logger        logger.Logger
}

// NewOrderCreatedHandler creates a new OrderCreatedHandler instance
func NewOrderCreatedHandler(domainService *services.InventoryDomainService, logger logger.Logger) *OrderCreatedHandler {
	return &OrderCreatedHandler{
		domainService: domainService,
		logger:        logger,
	}
}

// Handle processes OrderCreated events with error handling
func (h *OrderCreatedHandler) Handle(ctx context.Context, evt *events.OrderCreated) error {
	h.logger.Info("processing order created event", "orderID", evt.OrderId, "userID", evt.UserId)

	// Convert order items to stock reservation items
	items := make([]services.StockReservationItem, len(evt.Items))
	for i, item := range evt.Items {
		items[i] = services.StockReservationItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	// Reserve stock
	failedProducts, err := h.domainService.ReserveStock(ctx, evt.OrderId, evt.UserId, items)
	if err != nil {
		h.logger.Error("failed to reserve stock", "orderID", evt.OrderId, "error", err)
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	if len(failedProducts) > 0 {
		h.logger.Warn("stock reservation failed for some products", "orderID", evt.OrderId, "failedProducts", failedProducts)
	}

	h.logger.Info("order created event processed successfully", "orderID", evt.OrderId)
	return nil
}
