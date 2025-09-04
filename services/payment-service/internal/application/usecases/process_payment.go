package usecases

import (
	"context"
	"math/rand"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/payment-service/internal/domain/entities"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"

	"github.com/google/uuid"
)

// ProcessPaymentUseCase handles payment processing business logic
type ProcessPaymentUseCase struct {
	logger              logger.Logger
	paymentProcessedPub publisher.PaymentProcessedPublisher
}

// NewProcessPaymentUseCase creates a new instance of ProcessPaymentUseCase
func NewProcessPaymentUseCase(
	paymentProcessedPub publisher.PaymentProcessedPublisher,
	logger logger.Logger,
) *ProcessPaymentUseCase {
	return &ProcessPaymentUseCase{
		logger:              logger,
		paymentProcessedPub: paymentProcessedPub,
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
		payment.TransactionID = uuid.New().String()
	} else {
		payment.Status = paymentvalueobjects.PaymentStatusFailed
	}

	// Publish event
	message := "Payment processed successfully"
	if !success {
		message = "Payment processing failed"
	}

	if err := uc.paymentProcessedPub.PublishPaymentProcessed(ctx, payment, success, message); err != nil {
		uc.logger.Error("Failed to publish PaymentProcessed event", "error", err)
		return nil, errors.ErrPaymentProcessingFailed
	}

	return payment, nil
}

// simulatePaymentProcessing simulates payment processing logic
func (uc *ProcessPaymentUseCase) simulatePaymentProcessing() bool {
	return rand.Float32() < 0.5
}
