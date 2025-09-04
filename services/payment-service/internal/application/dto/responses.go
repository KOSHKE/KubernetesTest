package dto

import (
	"time"

	"ecommerce-platform/pkg/common/valueobjects"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"
)

// PaymentResponse represents a payment response
type PaymentResponse struct {
	ID            string                            `json:"id"`
	OrderID       string                            `json:"order_id"`
	Amount        valueobjects.Money                `json:"amount"`
	Status        paymentvalueobjects.PaymentStatus `json:"status"`
	Method        paymentvalueobjects.PaymentMethod `json:"method"`
	UserID        string                            `json:"user_id"`
	TransactionID string                            `json:"transaction_id,omitempty"`
	CreatedAt     time.Time                         `json:"created_at"`
	UpdatedAt     time.Time                         `json:"updated_at"`
}

// PaymentStatusResponse represents a payment status response
type PaymentStatusResponse struct {
	PaymentID string                            `json:"payment_id"`
	Status    paymentvalueobjects.PaymentStatus `json:"status"`
	Message   string                            `json:"message,omitempty"`
}
