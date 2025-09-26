package dto

// PaymentEventDTO represents payment event data for outbox
type PaymentEventDTO struct {
	OrderID        string `json:"orderId"`
	PaymentID      string `json:"paymentId"`
	UserID         string `json:"userId"`
	AmountAmount   int64  `json:"amountAmount"`
	AmountCurrency string `json:"amountCurrency"`
	Status         string `json:"status"`
	Method         string `json:"method"`
	TransactionID  string `json:"transactionId"`
	Success        bool   `json:"success"`
	Message        string `json:"message"`
}
