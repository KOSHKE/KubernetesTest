package entities

import (
	"fmt"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/idgenerator"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"
)

type Payment struct {
	ID            string
	OrderID       string
	UserID        string
	Amount        valueobjects.Money
	Status        paymentvalueobjects.PaymentStatus
	Method        paymentvalueobjects.PaymentMethod
	TransactionID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewPayment creates a new Payment entity
func NewPayment(orderID, userID string, amount valueobjects.Money, method paymentvalueobjects.PaymentMethod) *Payment {
	now := time.Now()
	return &Payment{
		ID:            idgenerator.GenerateID("payment"),
		OrderID:       orderID,
		UserID:        userID,
		Amount:        amount,
		Status:        paymentvalueobjects.PaymentStatusPending,
		Method:        method,
		TransactionID: "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// CanBeProcessed checks if payment can be processed
func (p *Payment) CanBeProcessed() error {
	if !p.Status.IsPending() {
		return fmt.Errorf("%w: payment status is %s, expected PENDING", errors.ErrPaymentAlreadyProcessed, p.Status.String())
	}
	return nil
}
