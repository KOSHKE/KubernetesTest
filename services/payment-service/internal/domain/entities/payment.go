package entities

import (
	"time"

	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/valueobjects"
)

type PaymentStatus string

const (
	PaymentCompleted PaymentStatus = "COMPLETED"
	PaymentFailed    PaymentStatus = "FAILED"
)

type PaymentMethod string

const (
	MethodCreditCard PaymentMethod = "CREDIT_CARD"
)

type Payment struct {
	ID            string
	OrderID       string
	UserID        string
	Amount        valueobjects.Money
	Status        PaymentStatus
	Method        PaymentMethod
	TransactionID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
