package valueobjects

import "fmt"

// Money represents currency in minor units to avoid float inaccuracies
type Money struct {
	Amount   int64  // e.g., cents
	Currency string // ISO 4217
}

func NewMoney(amount int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, fmt.Errorf("currency is required")
	}
	return Money{Amount: amount, Currency: currency}, nil
}
