package consumer

import (
	"context"

	dto "ecommerce-platform/pkg/common/dto/inventory-service"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/common"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/application/services"
)

// EventHandlers contains all event handlers for inventory service
type EventHandlers struct {
	applicationService *services.InventoryApplicationService
	logger             logger.Logger
}

// NewEventHandlers creates a new EventHandlers instance
func NewEventHandlers(applicationService *services.InventoryApplicationService, logger logger.Logger) *EventHandlers {
	return &EventHandlers{
		applicationService: applicationService,
		logger:             logger,
	}
}

// HandleOrderCreated processes OrderCreated events
func (h *EventHandlers) HandleOrderCreated(ctx context.Context, evt *events.OrderCreated) error {
	// Convert order items to stock reservation items
	items := h.convertToValueObjects(evt.Items)

	// Reserve stock
	req := &dto.ReserveStockRequest{
		OrderID: evt.OrderId,
		Items:   h.convertToStockReservationItems(items),
	}

	_, err := h.applicationService.ReserveStock(ctx, req)
	if err != nil {
		h.logger.Error("failed to reserve stock", "orderID", evt.OrderId, "error", err)
		return err
	}

	return nil
}

// HandlePaymentProcessed processes PaymentProcessed events
func (h *EventHandlers) HandlePaymentProcessed(ctx context.Context, evt *events.PaymentProcessed) error {
	// Convert order items to domain value objects
	items := h.convertToValueObjects(evt.Items)

	// Process stock based on payment result
	if evt.Success {
		return h.commitStock(ctx, evt.OrderId, items)
	}
	return h.releaseStock(ctx, evt.OrderId, items)
}

// commitStock commits reserved stock
func (h *EventHandlers) commitStock(ctx context.Context, orderID string, items []valueobjects.Item) error {
	req := &dto.CommitStockRequest{
		OrderID: orderID,
		Items:   h.convertToStockReservationItems(items),
	}

	_, err := h.applicationService.CommitStock(ctx, req)
	if err != nil {
		h.logger.Error("failed to commit stock", "orderID", orderID, "error", err)
		return err
	}

	return nil
}

// releaseStock releases reserved stock
func (h *EventHandlers) releaseStock(ctx context.Context, orderID string, items []valueobjects.Item) error {
	req := &dto.ReleaseStockRequest{
		OrderID: orderID,
		Items:   h.convertToStockReservationItems(items),
	}

	_, err := h.applicationService.ReleaseStock(ctx, req)
	if err != nil {
		h.logger.Error("failed to release stock", "orderID", orderID, "error", err)
		return err
	}

	return nil
}

// convertToValueObjects converts protobuf items to domain value objects
func (h *EventHandlers) convertToValueObjects(items []*common.OrderItem) []valueobjects.Item {
	result := make([]valueobjects.Item, len(items))
	for i, item := range items {
		stockItem, err := valueobjects.NewItem(item.ProductId, item.Quantity)
		if err != nil {
			h.logger.Error("invalid item data", "productID", item.ProductId, "quantity", item.Quantity, "error", err)
			// Continue with zero value for this item
			result[i] = valueobjects.Item{}
			continue
		}
		result[i] = *stockItem
	}
	return result
}

// convertToStockReservationItems converts valueobjects.Item to dto.StockReservationItem
func (h *EventHandlers) convertToStockReservationItems(items []valueobjects.Item) []dto.StockReservationItem {
	result := make([]dto.StockReservationItem, len(items))
	for i, item := range items {
		result[i] = dto.StockReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return result
}

// HandleOrderCancelled processes OrderCancelled events
func (h *EventHandlers) HandleOrderCancelled(ctx context.Context, evt *events.OrderCancelled) error {
	// Convert order items to domain value objects
	items := h.convertToValueObjects(evt.Items)

	// Release reserved stock for cancelled order
	req := &dto.ReleaseStockRequest{
		OrderID: evt.OrderId,
		Items:   h.convertToStockReservationItems(items),
	}

	_, err := h.applicationService.ReleaseStock(ctx, req)
	if err != nil {
		h.logger.Error("failed to release stock for cancelled order", "orderID", evt.OrderId, "reason", evt.Reason, "error", err)
		return err
	}

	return nil
}
