package valueobjects

// PaymentMethod represents the method of payment
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "CREDIT_CARD"
)

// String returns the string representation of the method
func (pm PaymentMethod) String() string {
	return string(pm)
}
