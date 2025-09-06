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

// ProcessStockReservationFailed processes StockReservationFailed events
func (s *StockEventService) ProcessStockReservationFailed(ctx context.Context, evt *events.StockReservationFailed) error {
	// Basic nil check only
	if evt == nil {
		return fmt.Errorf("event is nil")
	}

	if err := s.orderDomainService.CancelOrder(ctx, evt.OrderId, evt.Reason); err != nil {
		return err
	}

	return nil
}

// ProcessStockReleased processes StockReleased events
func (s *StockEventService) ProcessStockReleased(ctx context.Context, evt *events.StockReleased) error {
	// Basic nil check only
	if evt == nil {
		return fmt.Errorf("event is nil")
	}

	if err := s.orderDomainService.HandleStockReleased(ctx, evt.OrderId); err != nil {
		return err
	}

	return nil
}

// ProcessStockCommitted processes StockCommitted events
func (s *StockEventService) ProcessStockCommitted(ctx context.Context, evt *events.StockCommitted) error {
	// Basic nil check only
	if evt == nil {
		return fmt.Errorf("event is nil")
	}

	if err := s.orderDomainService.CompleteOrder(ctx, evt.OrderId); err != nil {
		return err
	}

	return nil
}
