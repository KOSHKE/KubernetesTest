package services

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/payment-service/internal/application/dto"
	"ecommerce-platform/services/payment-service/internal/application/usecases"
	"ecommerce-platform/services/payment-service/internal/domain/ports/repository"
	"ecommerce-platform/services/payment-service/internal/metrics"
)

// PaymentApplicationService orchestrates payment operations and provides a unified interface
type PaymentApplicationService struct {
	processPaymentUseCase *usecases.ProcessPaymentUseCase
	logger                logger.Logger
	metrics               metrics.PaymentMetrics
}

// NewPaymentApplicationService creates a new PaymentApplicationService instance
func NewPaymentApplicationService(
	outboxRepo repository.OutboxRepository,
	logger logger.Logger,
	metrics metrics.PaymentMetrics,
) *PaymentApplicationService {
	return &PaymentApplicationService{
		processPaymentUseCase: usecases.NewProcessPaymentUseCase(outboxRepo),
		logger:                logger,
		metrics:               metrics,
	}
}

// GetProcessPaymentUseCase returns the process payment use case
func (s *PaymentApplicationService) GetProcessPaymentUseCase() *usecases.ProcessPaymentUseCase {
	return s.processPaymentUseCase
}

// ProcessPayment handles the complete payment processing flow
func (s *PaymentApplicationService) ProcessPayment(ctx context.Context, req *dto.ProcessPaymentRequest) (*dto.PaymentResponse, error) {
	// Convert DTO to domain parameters
	payment, err := s.processPaymentUseCase.Execute(ctx, req.OrderID, req.UserID, req.Amount, req.Method)
	if err != nil {
		s.logger.Error("Failed to process payment", "error", err, "orderID", req.OrderID, "userID", req.UserID)
		return nil, err
	}

	s.logger.Info("Payment processed successfully", "paymentID", payment.ID, "orderID", req.OrderID, "userID", req.UserID, "status", payment.Status)

	// Convert entity to DTO response
	return &dto.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		Method:        payment.Method,
		TransactionID: payment.TransactionID,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}
