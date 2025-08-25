package services

import (
	"context"
	"sync"
	"time"

	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/entities"
	derrors "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/valueobjects"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/metrics"
	procport "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/ports/processor"
)

type PaymentService struct {
	processor procport.PaymentProcessor
	metrics   metrics.PaymentMetrics
	mu        sync.RWMutex
	payments  map[string]*entities.Payment
}

func NewPaymentService(processor procport.PaymentProcessor, metrics metrics.PaymentMetrics) *PaymentService {
	return &PaymentService{
		processor: processor,
		metrics:   metrics,
		payments:  make(map[string]*entities.Payment),
	}
}

// (Clock/IDGenerator injection removed as unused)

type ProcessPaymentRequest struct {
	OrderID string
	UserID  string
	Amount  valueobjects.Money
	Method  entities.PaymentMethod
	// Card details (mock)
	CardNumber string
}

type ProcessPaymentResponse struct {
	Payment *entities.Payment
	Success bool
	Message string
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
	start := time.Now()

	res, err := s.processor.Process(ctx, procport.ProcessRequest{CardNumber: req.CardNumber})
	if err != nil {
		// Record failed payment
		s.metrics.PaymentFailed("processor_error")
		s.metrics.PaymentProcessingDuration(time.Since(start), string(req.Method))
		return nil, err
	}

	now := s.now()
	message := "Payment failed"
	status := entities.PaymentFailed
	if res.Success {
		status = entities.PaymentCompleted
		message = "Payment processed successfully"
	} else if res.FailureReason != "" {
		message = "Payment failed - " + res.FailureReason
	}

	payment := &entities.Payment{
		ID:            s.newID("pay-"),
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Status:        status,
		Method:        req.Method,
		TransactionID: s.newID("txn-"),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// persist in-memory for subsequent reads
	s.mu.Lock()
	s.payments[payment.ID] = payment
	s.mu.Unlock()

	// Record metrics based on result
	if res.Success {
		s.metrics.PaymentSucceeded(string(req.Method))
	} else {
		s.metrics.PaymentFailed(res.FailureReason)
	}
	s.metrics.PaymentProcessingDuration(time.Since(start), string(req.Method))

	if !res.Success {
		return &ProcessPaymentResponse{Payment: payment, Success: res.Success, Message: message}, derrors.ErrPaymentDeclined
	}

	return &ProcessPaymentResponse{Payment: payment, Success: res.Success, Message: message}, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (*entities.Payment, error) {
	s.mu.RLock()
	p := s.payments[id]
	s.mu.RUnlock()
	if p == nil {
		return nil, derrors.ErrPaymentNotFound
	}
	return p, nil
}

func (s *PaymentService) now() time.Time { return time.Now() }

func (s *PaymentService) newID(prefix string) string {
	return prefix + time.Now().Format("20060102150405")
}
