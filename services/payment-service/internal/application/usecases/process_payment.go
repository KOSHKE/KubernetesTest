package usecases

import (
	"context"
	"math/rand"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/payment-service/internal/application/dto"
	"ecommerce-platform/services/payment-service/internal/domain/entities"
	"ecommerce-platform/services/payment-service/internal/domain/ports/repository"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"

	"ecommerce-platform/pkg/idgenerator"
	"ecommerce-platform/pkg/outbox"
)

// ProcessPaymentUseCase handles payment processing business logic
type ProcessPaymentUseCase struct {
	outboxRepo repository.OutboxRepository
}

// NewProcessPaymentUseCase creates a new instance of ProcessPaymentUseCase
func NewProcessPaymentUseCase(
	outboxRepo repository.OutboxRepository,
) *ProcessPaymentUseCase {
	return &ProcessPaymentUseCase{
		outboxRepo: outboxRepo,
	}
}

// Execute processes a payment
func (uc *ProcessPaymentUseCase) Execute(ctx context.Context, orderID, userID string, amount valueobjects.Money, method paymentvalueobjects.PaymentMethod) (*entities.Payment, error) {
	// Create payment entity
	payment := entities.NewPayment(orderID, userID, amount, method)

	// Simulate payment processing
	success := uc.simulatePaymentProcessing()

	if success {
		payment.Status = paymentvalueobjects.PaymentStatusCompleted
		payment.TransactionID = idgenerator.GenerateID("txn")
	} else {
		payment.Status = paymentvalueobjects.PaymentStatusFailed
	}

	// Save event to outbox (to be published later)
	message := "Payment processed successfully"
	if !success {
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
		Success:       success,
		Message:       message,
	}

	event := outbox.Event{
		AggregateID: payment.OrderID,
		Type:        "PaymentProcessed",
		Payload:     eventData,
	}

	if err := uc.outboxRepo.SaveEvent(ctx, event); err != nil {
		return nil, errors.ErrPaymentProcessingFailed
	}

	return payment, nil
}

// simulatePaymentProcessing simulates payment processing logic
func (uc *ProcessPaymentUseCase) simulatePaymentProcessing() bool {
	return rand.Float32() < 0.5
}
