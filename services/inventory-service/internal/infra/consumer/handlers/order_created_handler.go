package handlers

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/services"
)

// OrderCreatedHandler handles OrderCreated events at infrastructure level
type OrderCreatedHandler struct {
	applicationService *services.InventoryApplicationService
	logger             logger.Logger
}

// NewOrderCreatedHandler creates a new OrderCreatedHandler instance
func NewOrderCreatedHandler(applicationService *services.InventoryApplicationService, logger logger.Logger) *OrderCreatedHandler {
	return &OrderCreatedHandler{
		applicationService: applicationService,
		logger:             logger,
	}
}

// Handle processes OrderCreated events with error handling
func (h *OrderCreatedHandler) Handle(ctx context.Context, evt *events.OrderCreated) error {
	// Convert order items to stock reservation items
	items := make([]valueobjects.Item, len(evt.Items))
	for i, item := range evt.Items {
		stockItem, err := valueobjects.NewItem(item.ProductId, item.Quantity)
		if err != nil {
			return err
		}
		items[i] = *stockItem
	}

	// Reserve stock
	req := &dto.ReserveStockRequest{
		OrderID: evt.OrderId,
		UserID:  evt.UserId,
		Items:   convertToStockReservationItems(items),
	}

	_, err := h.applicationService.ReserveStock(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
