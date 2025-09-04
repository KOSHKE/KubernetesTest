package services

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	domainservices "ecommerce-platform/services/order-service/internal/domain/services"
)

// PaymentEventService orchestrates payment event processing
type PaymentEventService struct {
	orderDomainService *domainservices.OrderDomainService
	logger             logger.Logger
}

// NewPaymentEventService creates a new PaymentEventService instance
func NewPaymentEventService(orderDomainService *domainservices.OrderDomainService, logger logger.Logger) *PaymentEventService {
	return &PaymentEventService{
		orderDomainService: orderDomainService,
		logger:             logger,
	}
}

// ProcessPaymentEvent processes PaymentProcessed events with error handling and logging
func (s *PaymentEventService) ProcessPaymentEvent(ctx context.Context, evt *events.PaymentProcessed) error {
	// Basic nil check only
	if evt == nil {
		return fmt.Errorf("event is nil")
	}

	// Process event using domain service
	if evt.Success {
		if err := s.orderDomainService.ConfirmPayment(ctx, evt.OrderId); err != nil {
			return err
		}
	} else {
		if err := s.orderDomainService.MarkPaymentFailed(ctx, evt.OrderId, evt.Message); err != nil {
			return err
		}
	}

	return nil
}
