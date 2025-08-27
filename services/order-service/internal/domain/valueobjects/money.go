package valueobjects

import (
	"errors"
	"fmt"
	"strings"
)

// ErrCurrencyMismatch is returned when trying to perform operations with different currencies
var ErrCurrencyMismatch = errors.New("currency mismatch")

// Money represents a monetary value with currency
type Money struct {
	Amount   int64  // Amount in smallest currency unit (cents, kopecks, etc.)
	Currency string // Currency code (USD, EUR, RUB, etc.)
}

// NewMoney creates a new Money value object
func NewMoney(amount int64, currency string) (*Money, error) {
	// Normalize currency to uppercase to prevent mismatches
	currency = strings.ToUpper(currency)

	if len(currency) != 3 {
		return nil, fmt.Errorf("currency code must be exactly 3 characters: %s", currency)
	}

	// Validate that currency contains only letters (ISO 4217 standard)
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return nil, fmt.Errorf("currency code must contain only letters: %s", currency)
		}
	}

	return &Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

// Add adds another Money value (must be same currency)
func (m *Money) Add(other *Money) (*Money, error) {
	if m.Currency != other.Currency {
		return nil, ErrCurrencyMismatch
	}

	newAmount := m.Amount + other.Amount

	return &Money{
		Amount:   newAmount,
		Currency: m.Currency,
	}, nil
}

// Subtract subtracts another Money value (must be same currency)
func (m *Money) Subtract(other *Money) (*Money, error) {
	if m.Currency != other.Currency {
		return nil, ErrCurrencyMismatch
	}

	newAmount := m.Amount - other.Amount

	return &Money{
		Amount:   newAmount,
		Currency: m.Currency,
	}, nil
}

// IsZero checks if the money amount is zero
func (m *Money) IsZero() bool {
	return m.Amount == 0
}

// Format formats the money for display
func (m *Money) Format() string {
	// Convert to decimal representation (e.g., 1000 -> 10.00)
	whole := m.Amount / 100
	fraction := m.Amount % 100

	if fraction == 0 {
		return fmt.Sprintf("%d %s", whole, m.Currency)
	}

	return fmt.Sprintf("%d.%02d %s", whole, fraction, m.Currency)
}

// Equals checks if two Money values are equal
func (m *Money) Equals(other *Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// Multiply multiplies the money amount by a factor
func (m *Money) Multiply(factor int64) *Money {
	newAmount := m.Amount * factor

	return &Money{
		Amount:   newAmount,
		Currency: m.Currency,
	}
}
