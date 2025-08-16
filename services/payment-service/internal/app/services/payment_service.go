package services

import (
	"context"
	"time"

	"payment-service/internal/domain/entities"
	"payment-service/internal/domain/valueobjects"
	"payment-service/internal/ports/clock"
	"payment-service/internal/ports/idgen"
	procport "payment-service/internal/ports/processor"
)

type PaymentService struct {
	processor procport.PaymentProcessor
	clock     clock.Clock
	ids       idgen.IDGenerator
}

func NewPaymentService(processor procport.PaymentProcessor) *PaymentService {
	return &PaymentService{processor: processor}
}

// WithClock allows injecting a custom clock
func (s *PaymentService) WithClock(c clock.Clock) *PaymentService { s.clock = c; return s }

// WithIDGenerator allows injecting a custom ID generator
func (s *PaymentService) WithIDGenerator(g idgen.IDGenerator) *PaymentService { s.ids = g; return s }

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
	res, err := s.processor.Process(ctx, procport.ProcessRequest{CardNumber: req.CardNumber})
	if err != nil {
		return nil, err
	}

	now := s.now()
	payment := &entities.Payment{
		ID:            s.newID("pay-"),
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Status:        entities.PaymentFailed,
		Method:        req.Method,
		TransactionID: s.newID("txn-"),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	message := "Payment processed successfully"
	if res.Success {
		payment.Status = entities.PaymentCompleted
	} else {
		payment.Status = entities.PaymentFailed
		if res.FailureReason != "" {
			message = "Payment failed - " + res.FailureReason
		}
	}

	return &ProcessPaymentResponse{Payment: payment, Success: res.Success, Message: message}, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (*entities.Payment, error) {
	now := s.now()
	amt, _ := valueobjects.NewMoney(9999, "USD")
	return &entities.Payment{
		ID:            id,
		OrderID:       "order-" + id,
		UserID:        "user-123",
		Amount:        amt,
		Status:        entities.PaymentCompleted,
		Method:        entities.MethodCreditCard,
		TransactionID: "txn-" + id,
		CreatedAt:     now.Add(-time.Hour),
		UpdatedAt:     now.Add(-time.Hour),
	}, nil
}

func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, amount valueobjects.Money, reason string) (*entities.Payment, string, bool, error) {
	// Mock refund success
	now := s.now()
	return &entities.Payment{
		ID:            paymentID,
		OrderID:       "order-" + paymentID,
		UserID:        "user-123",
		Amount:        amount,
		Status:        entities.PaymentRefunded,
		Method:        entities.MethodCreditCard,
		TransactionID: "refund-" + now.Format("20060102150405"),
		CreatedAt:     now.Add(-24 * time.Hour),
		UpdatedAt:     now,
	}, "Refund processed successfully - Reason: " + reason, true, nil
}

func (s *PaymentService) now() time.Time {
	if s.clock != nil {
		return s.clock.Now()
	}
	return time.Now()
}

func (s *PaymentService) newID(prefix string) string {
	if s.ids != nil {
		return s.ids.NewID(prefix)
	}
	return prefix + time.Now().Format("20060102150405")
}
