package dto

import "time"

// PaymentResponse represents a payment response
type PaymentResponse struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	AmountAmount   int64     `json:"amount_amount"`
	AmountCurrency string    `json:"amount_currency"`
	Status         string    `json:"status"`
	Method         string    `json:"method"`
	UserID         string    `json:"user_id"`
	TransactionID  string    `json:"transaction_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PaymentStatusResponse represents a payment status response
type PaymentStatusResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}
