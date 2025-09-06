package valueobjects

import (
	"fmt"

	"ecommerce-platform/pkg/common/errors"
)

// Money represents currency in minor units to avoid float inaccuracies
type Money struct {
	Amount   int64    `json:"amount" gorm:"type:bigint;not null"`                     // e.g., cents
	Currency Currency `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"` // ISO 4217
}

// NewMoney creates a new Money instance
func NewMoney(amount int64, currency Currency) Money {
	return Money{Amount: amount, Currency: currency}
}

// Validate validates money data
func (m Money) Validate() error {
	if err := m.Currency.Validate(); err != nil {
		return err
	}
	// Money doesn't have specific validation rules beyond currency
	return nil
}

// Add adds another Money amount (must be same currency)
func (m Money) Add(other Money) (Money, error) {
	if !m.Currency.Equals(other.Currency) {
		return Money{}, errors.ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract subtracts another Money amount (must be same currency)
func (m Money) Subtract(other Money) (Money, error) {
	if !m.Currency.Equals(other.Currency) {
		return Money{}, errors.ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// Multiply multiplies amount by a factor
func (m Money) Multiply(factor int64) Money {
	return Money{Amount: m.Amount * factor, Currency: m.Currency}
}

// IsZero checks if the amount is zero
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive checks if the amount is positive
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsNegative checks if the amount is negative
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// String returns a string representation of Money
func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.Amount, m.Currency)
}
