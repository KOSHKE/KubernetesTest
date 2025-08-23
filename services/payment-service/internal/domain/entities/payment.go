package entities

import (
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/valueobjects"
	"time"
)

type PaymentStatus string

const (
	PaymentCompleted PaymentStatus = "COMPLETED"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
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
