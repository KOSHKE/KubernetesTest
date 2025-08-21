package types

// Money is a Value Object for monetary amounts in minor units with currency code.
type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}
