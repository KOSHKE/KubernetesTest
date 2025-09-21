package usecases

import (
	"context"
	"math/rand"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/payment-service/internal/domain/entities"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"

	"ecommerce-platform/pkg/idgenerator"
)

// ProcessPaymentUseCase handles payment processing business logic
type ProcessPaymentUseCase struct{}

// NewProcessPaymentUseCase creates a new instance of ProcessPaymentUseCase
func NewProcessPaymentUseCase() *ProcessPaymentUseCase {
	return &ProcessPaymentUseCase{}
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

	return payment, nil
}

// simulatePaymentProcessing simulates payment processing logic
func (uc *ProcessPaymentUseCase) simulatePaymentProcessing() bool {
	return rand.Float32() < 0.5
}
