package services

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// OrderDomainService handles order business logic in domain layer
type OrderDomainService struct {
	orderRepo repository.OrderRepository
	logger    logger.Logger
}

// NewOrderDomainService creates a new OrderDomainService instance
func NewOrderDomainService(orderRepo repository.OrderRepository, logger logger.Logger) *OrderDomainService {
	return &OrderDomainService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

// ConfirmPayment confirms order payment and updates status to paid
func (s *OrderDomainService) ConfirmPayment(ctx context.Context, orderID string) error {
	return s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		order, err := txRepo.GetByID(ctx, orderID)
		if err != nil {
			s.logger.Error("Failed to get order for payment confirmation", "order_id", orderID, "error", err)
			return errors.ErrOrderNotFound
		}

		if err := order.SetStatus(orderValueObjects.OrderStatusPaid); err != nil {
			s.logger.Error("Failed to set order status to paid", "order_id", orderID, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		if err := txRepo.Update(ctx, order); err != nil {
			s.logger.Error("Failed to update order after payment confirmation", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		return nil
	})
}

// MarkPaymentFailed marks order payment as failed
func (s *OrderDomainService) MarkPaymentFailed(ctx context.Context, orderID string, reason string) error {
	return s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		order, err := txRepo.GetByID(ctx, orderID)
		if err != nil {
			s.logger.Error("Failed to get order for payment failure marking", "order_id", orderID, "error", err)
			return errors.ErrOrderNotFound
		}

		if err := order.SetStatus(orderValueObjects.OrderStatusPaymentFailed); err != nil {
			s.logger.Error("Failed to set order status to payment failed", "order_id", orderID, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		if err := txRepo.Update(ctx, order); err != nil {
			s.logger.Error("Failed to update order after payment failure marking", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		return nil
	})
}

// ConfirmStockReservation confirms stock reservation and updates order status to confirmed
func (s *OrderDomainService) ConfirmStockReservation(ctx context.Context, orderID string) error {
	return s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		order, err := txRepo.GetByID(ctx, orderID)
		if err != nil {
			s.logger.Error("Failed to get order for stock reservation confirmation", "order_id", orderID, "error", err)
			return errors.ErrOrderNotFound
		}

		if err := order.SetStatus(orderValueObjects.OrderStatusConfirmed); err != nil {
			s.logger.Error("Failed to set order status to confirmed", "order_id", orderID, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		if err := txRepo.Update(ctx, order); err != nil {
			s.logger.Error("Failed to update order after stock reservation confirmation", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		return nil
	})
}

// CancelOrder cancels an order with a reason
func (s *OrderDomainService) CancelOrder(ctx context.Context, orderID string, reason string) error {
	return s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		order, err := txRepo.GetByID(ctx, orderID)
		if err != nil {
			s.logger.Error("Failed to get order for cancellation", "order_id", orderID, "error", err)
			return errors.ErrOrderNotFound
		}

		if err := order.SetStatus(orderValueObjects.OrderStatusCancelled); err != nil {
			s.logger.Error("Failed to set order status to cancelled", "order_id", orderID, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		if err := txRepo.Update(ctx, order); err != nil {
			s.logger.Error("Failed to update order after cancellation", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		s.logger.Info("Order cancelled successfully", "order_id", orderID, "reason", reason)
		return nil
	})
}

// HandleStockReleased handles stock release event
func (s *OrderDomainService) HandleStockReleased(ctx context.Context, orderID string) error {
	return s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		order, err := txRepo.GetByID(ctx, orderID)
		if err != nil {
			s.logger.Error("Failed to get order for stock release handling", "order_id", orderID, "error", err)
			return errors.ErrOrderNotFound
		}

		// Update order status to reflect stock release
		if err := order.SetStatus(orderValueObjects.OrderStatusStockReleased); err != nil {
			s.logger.Error("Failed to set order status to stock released", "order_id", orderID, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		if err := txRepo.Update(ctx, order); err != nil {
			s.logger.Error("Failed to update order after stock release", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		s.logger.Info("Order stock released successfully", "order_id", orderID)
		return nil
	})
}

// CompleteOrder completes an order after stock commit
func (s *OrderDomainService) CompleteOrder(ctx context.Context, orderID string) error {
	return s.orderRepo.WithTransaction(ctx, func(txRepo repository.OrderRepository) error {
		order, err := txRepo.GetByID(ctx, orderID)
		if err != nil {
			s.logger.Error("Failed to get order for completion", "order_id", orderID, "error", err)
			return errors.ErrOrderNotFound
		}

		if err := order.SetStatus(orderValueObjects.OrderStatusCompleted); err != nil {
			s.logger.Error("Failed to set order status to completed", "order_id", orderID, "error", err)
			return errors.ErrOrderStatusUpdateFailed
		}

		if err := txRepo.Update(ctx, order); err != nil {
			s.logger.Error("Failed to update order after completion", "order_id", orderID, "error", err)
			return errors.ErrOrderPersistenceFailed
		}

		s.logger.Info("Order completed successfully", "order_id", orderID)
		return nil
	})
}
