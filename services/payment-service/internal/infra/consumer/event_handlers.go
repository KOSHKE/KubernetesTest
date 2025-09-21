package consumer

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/payment-service/internal/application/dto"
	"ecommerce-platform/services/payment-service/internal/application/services"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"
)

// EventHandlers contains all event handlers for payment service
type EventHandlers struct {
	applicationService *services.PaymentApplicationService
	logger             logger.Logger
}

// NewEventHandlers creates a new EventHandlers instance
func NewEventHandlers(applicationService *services.PaymentApplicationService, logger logger.Logger) *EventHandlers {
	return &EventHandlers{
		applicationService: applicationService,
		logger:             logger,
	}
}

// HandleStockReserved processes StockReserved events
// When stock is reserved, we can proceed with payment processing
func (h *EventHandlers) HandleStockReserved(ctx context.Context, evt *events.StockReserved) error {
	// Convert payment method from string to domain value object
	method := paymentvalueobjects.PaymentMethod(evt.PaymentMethod)

	// Create Money value object from amount and currency
	currency, err := valueobjects.NewCurrency(evt.Currency)
	if err != nil {
		h.logger.Error("invalid currency in stock reserved event", "currency", evt.Currency, "error", err)
		return err
	}
	money := valueobjects.NewMoney(evt.TotalAmount, currency)

	// Create payment processing request
	req := &dto.ProcessPaymentRequest{
		OrderID: evt.OrderId,
		UserID:  evt.UserId,
		Amount:  money,
		Method:  method,
	}

	// Process payment
	_, err = h.applicationService.ProcessPayment(ctx, req)
	if err != nil {
		h.logger.Error("failed to process payment for reserved stock", "orderID", evt.OrderId, "error", err)
		return err
	}

	return nil
}
