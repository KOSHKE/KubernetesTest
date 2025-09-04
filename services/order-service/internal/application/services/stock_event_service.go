package services

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	domainservices "ecommerce-platform/services/order-service/internal/domain/services"
)

// StockEventService orchestrates stock reservation events processing
type StockEventService struct {
	orderDomainService *domainservices.OrderDomainService
	logger             logger.Logger
}

// NewStockEventService creates a new StockEventService instance
func NewStockEventService(orderDomainService *domainservices.OrderDomainService, logger logger.Logger) *StockEventService {
	return &StockEventService{
		orderDomainService: orderDomainService,
		logger:             logger,
	}
}

// ProcessStockReserved processes StockReserved events
func (s *StockEventService) ProcessStockReserved(ctx context.Context, evt *events.StockReserved) error {
	// Basic nil check only
	if evt == nil {
		return fmt.Errorf("event is nil")
	}

	if err := s.orderDomainService.ConfirmStockReservation(ctx, evt.OrderId); err != nil {
		return err
	}

	return nil
}
