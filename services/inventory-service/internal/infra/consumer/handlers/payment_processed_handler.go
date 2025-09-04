package handlers

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/domain/services"
)

// PaymentProcessedHandler handles PaymentProcessed events at infrastructure level
type PaymentProcessedHandler struct {
	domainService *services.InventoryDomainService
	logger        logger.Logger
}

// NewPaymentProcessedHandler creates a new PaymentProcessedHandler instance
func NewPaymentProcessedHandler(domainService *services.InventoryDomainService, logger logger.Logger) *PaymentProcessedHandler {
	return &PaymentProcessedHandler{
		domainService: domainService,
		logger:        logger,
	}
}

// Handle processes PaymentProcessed events with error handling
func (h *PaymentProcessedHandler) Handle(ctx context.Context, evt *events.PaymentProcessed) error {
	h.logger.Info("processing payment processed event", "orderID", evt.OrderId, "success", evt.Success)

	// TODO: Get order items from order service or store them locally
	// For now, we'll need to implement a way to get the order items
	// This could be done by:
	// 1. Storing order items locally when order is created
	// 2. Calling order service to get order details
	// 3. Using a different event that includes order items

	if evt.Success {
		// Payment successful - commit reserved stock
		// TODO: Implement commit stock logic
		h.logger.Info("payment successful, committing reserved stock", "orderID", evt.OrderId)
	} else {
		// Payment failed - release reserved stock
		// TODO: Implement release stock logic
		h.logger.Info("payment failed, releasing reserved stock", "orderID", evt.OrderId)
	}

	return nil
}
