package errors

import "errors"

var (
	// ErrPaymentDeclined indicates that payment was declined due to business rules (e.g., insufficient funds)
	ErrPaymentDeclined = errors.New("payment declined")
	// ErrPaymentNotFound indicates that payment with given id does not exist
	ErrPaymentNotFound = errors.New("payment not found")
	// ErrInvalidRefund indicates that refund is not allowed for current payment state
	ErrInvalidRefund = errors.New("invalid refund")
)
