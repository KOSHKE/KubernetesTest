package dto

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	OrderID        string `json:"order_id" validate:"required"`
	AmountAmount   int64  `json:"amount_amount"`
	AmountCurrency string `json:"amount_currency" validate:"required"`
	Method         string `json:"method" validate:"required"`
	UserID         string `json:"user_id" validate:"required"`
}

// ProcessPaymentRequest represents a request to process a payment
type ProcessPaymentRequest struct {
	PaymentID      string          `json:"payment_id" validate:"required"`
	OrderID        string          `json:"order_id" validate:"required"`
	UserID         string          `json:"user_id" validate:"required"`
	AmountAmount   int64           `json:"amount_amount"`
	AmountCurrency string          `json:"amount_currency" validate:"required"`
	Method         string          `json:"method" validate:"required"`
	Details        *PaymentDetails `json:"details,omitempty"`
}

// PaymentDetails represents payment card details
type PaymentDetails struct {
	CardNumber  string `json:"card_number" validate:"required"`
	CardHolder  string `json:"card_holder" validate:"required"`
	ExpiryMonth string `json:"expiry_month" validate:"required"`
	ExpiryYear  string `json:"expiry_year" validate:"required"`
	CVV         string `json:"cvv" validate:"required"`
}
