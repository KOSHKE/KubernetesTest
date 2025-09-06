package handlers

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/services"
)

// stockOperation represents a stock operation type
type stockOperation string

const (
	OperationCommit  stockOperation = "commit"
	OperationRelease stockOperation = "release"
)

// stockProcessor represents a function that processes stock operations
type stockProcessor func(ctx context.Context, items []dto.StockReservationItem, orderID string) error

// PaymentProcessedHandler handles PaymentProcessed events at infrastructure level
type PaymentProcessedHandler struct {
	applicationService *services.InventoryApplicationService
	logger             logger.Logger
}

// NewPaymentProcessedHandler creates a new PaymentProcessedHandler instance
func NewPaymentProcessedHandler(applicationService *services.InventoryApplicationService, logger logger.Logger) *PaymentProcessedHandler {
	return &PaymentProcessedHandler{
		applicationService: applicationService,
		logger:             logger,
	}
}

// Handle processes PaymentProcessed events with error handling
func (h *PaymentProcessedHandler) Handle(ctx context.Context, evt *events.PaymentProcessed) error {
	// Convert order items to domain value objects
	items := make([]valueobjects.Item, len(evt.Items))
	for i, item := range evt.Items {
		stockItem, err := valueobjects.NewItem(item.ProductId, item.Quantity)
		if err != nil {
			h.logger.Error("invalid item data", "productID", item.ProductId, "quantity", item.Quantity, "error", err)
			return err
		}
		items[i] = *stockItem
	}

	// Process stock based on payment result
	if evt.Success {
		err := h.processStockOperation(ctx, evt.OrderId, items, OperationCommit)
		if err != nil {
			return err
		}
	} else {
		err := h.processStockOperation(ctx, evt.OrderId, items, OperationRelease)
		if err != nil {
			return err
		}
	}

	return nil
}

// processStockOperation processes stock operation using command map pattern
func (h *PaymentProcessedHandler) processStockOperation(ctx context.Context, orderID string, items []valueobjects.Item, op stockOperation) error {
	// Convert to DTO
	stockItems := convertToStockReservationItems(items)

	// Command map pattern - processors for each operation
	processors := map[stockOperation]stockProcessor{
		OperationCommit: func(ctx context.Context, items []dto.StockReservationItem, orderID string) error {
			req := &dto.CommitStockRequest{OrderID: orderID, UserID: "", Items: items}
			_, err := h.applicationService.CommitStock(ctx, req)
			return err
		},
		OperationRelease: func(ctx context.Context, items []dto.StockReservationItem, orderID string) error {
			req := &dto.ReleaseStockRequest{OrderID: orderID, UserID: "", Items: items}
			_, err := h.applicationService.ReleaseStock(ctx, req)
			return err
		},
	}

	// Execute the appropriate processor
	if processor, ok := processors[op]; ok {
		if err := processor(ctx, stockItems, orderID); err != nil {
			h.logger.Error("failed to process stock operation", "orderID", orderID, "operation", op, "error", err)
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown stock operation: %s", op)
}
