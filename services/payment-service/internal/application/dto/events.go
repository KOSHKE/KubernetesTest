package dto

import (
	"ecommerce-platform/pkg/common/valueobjects"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"
)

// PaymentEventDTO represents payment event data for outbox
type PaymentEventDTO struct {
	OrderID       string                            `json:"orderId"`
	PaymentID     string                            `json:"paymentId"`
	UserID        string                            `json:"userId"`
	Amount        valueobjects.Money                `json:"amount"`
	Status        paymentvalueobjects.PaymentStatus `json:"status"`
	Method        paymentvalueobjects.PaymentMethod `json:"method"`
	TransactionID string                            `json:"transactionId"`
	Success       bool                              `json:"success"`
	Message       string                            `json:"message"`
}
