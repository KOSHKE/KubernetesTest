package services

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/payment-service/internal/application/dto"
	"ecommerce-platform/services/payment-service/internal/application/usecases"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"
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
	paymentProcessedPub publisher.PaymentProcessedPublisher,
	logger logger.Logger,
	metrics metrics.PaymentMetrics,
) *PaymentApplicationService {
	return &PaymentApplicationService{
		processPaymentUseCase: usecases.NewProcessPaymentUseCase(paymentProcessedPub, logger),
		logger:                logger,
		metrics:               metrics,
	}
}

// GetProcessPaymentUseCase returns the process payment use case for internal use
func (s *PaymentApplicationService) GetProcessPaymentUseCase() *usecases.ProcessPaymentUseCase {
	return s.processPaymentUseCase
}

// ProcessPayment handles the complete payment processing flow
func (s *PaymentApplicationService) ProcessPayment(ctx context.Context, req *dto.ProcessPaymentRequest) (*dto.PaymentResponse, error) {
	// Convert DTO to domain parameters
	payment, err := s.processPaymentUseCase.Execute(ctx, req.OrderID, req.UserID, req.Amount, req.Method)
	if err != nil {
		return nil, err
	}

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

// GetProcessPaymentUseCase returns the process payment use case for external use
func (s *PaymentApplicationService) GetProcessPaymentUseCase() *usecases.ProcessPaymentUseCase {
	return s.processPaymentUseCase
}
