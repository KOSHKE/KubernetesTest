package handlers

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	appservices "ecommerce-platform/services/order-service/internal/application/services"
)

// PaymentProcessedHandler handles PaymentProcessed events at infrastructure level
type PaymentProcessedHandler struct {
	paymentEventService *appservices.PaymentEventService
	logger              logger.Logger
}

// NewPaymentProcessedHandler creates a new PaymentProcessedHandler instance
func NewPaymentProcessedHandler(paymentEventService *appservices.PaymentEventService, logger logger.Logger) *PaymentProcessedHandler {
	return &PaymentProcessedHandler{
		paymentEventService: paymentEventService,
		logger:              logger,
	}
}

// Handle processes PaymentProcessed events with error handling
func (h *PaymentProcessedHandler) Handle(ctx context.Context, evt *events.PaymentProcessed) error {
	// Process event using application service
	if err := h.paymentEventService.ProcessPaymentEvent(ctx, evt); err != nil {
		h.logger.Error("failed to process payment event", "error", err, "paymentID", evt.PaymentId, "orderID", evt.OrderId)
		return fmt.Errorf("failed to process payment event: %w", err)
	}

	return nil
}
