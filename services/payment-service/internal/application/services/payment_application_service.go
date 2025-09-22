package services

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/pkg/validation"
	"ecommerce-platform/services/payment-service/internal/application/dto"
	"ecommerce-platform/services/payment-service/internal/application/usecases"
	"ecommerce-platform/services/payment-service/internal/domain/ports/repository"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"
	"ecommerce-platform/services/payment-service/internal/metrics"
)

// PaymentApplicationService orchestrates payment operations and provides a unified interface
type PaymentApplicationService struct {
	processPaymentUseCase *usecases.ProcessPaymentUseCase
	outboxRepo            repository.OutboxRepository
	validator             *validation.Validate
	logger                logger.Logger
	metrics               metrics.PaymentMetrics
}

// NewPaymentApplicationService creates a new PaymentApplicationService instance
func NewPaymentApplicationService(
	outboxRepo repository.OutboxRepository,
	logger logger.Logger,
	metrics metrics.PaymentMetrics,
) *PaymentApplicationService {
	v := validation.New()

	return &PaymentApplicationService{
		processPaymentUseCase: usecases.NewProcessPaymentUseCase(),
		outboxRepo:            outboxRepo,
		validator:             v,
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
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain parameters
	amount := req.Amount
	method := paymentvalueobjects.PaymentMethod(req.Method)

	// Execute use case with business logic only
	payment, err := s.processPaymentUseCase.Execute(ctx, req.OrderID, req.UserID, amount, method)
	if err != nil {
		s.logger.Error("Failed to process payment", "error", err, "orderID", req.OrderID, "userID", req.UserID)
		return nil, err
	}

	// Save event to outbox (to be published later)
	message := "Payment processed successfully"
	if payment.Status != paymentvalueobjects.PaymentStatusCompleted {
		message = "Payment processing failed"
	}

	eventData := dto.PaymentEventDTO{
		OrderID:       payment.OrderID,
		PaymentID:     payment.ID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		Method:        payment.Method,
		TransactionID: payment.TransactionID,
		Success:       payment.Status == paymentvalueobjects.PaymentStatusCompleted,
		Message:       message,
	}

	event := outbox.Event{
		AggregateID: payment.OrderID,
		Type:        "PaymentProcessed",
		Payload:     eventData,
	}

	// Save event to outbox
	outboxService := outbox.NewService(s.outboxRepo, s.logger)
	if err := outboxService.SaveEvent(ctx, event); err != nil {
		s.logger.Error("Failed to save payment event to outbox", "error", err, "paymentID", payment.ID)
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
