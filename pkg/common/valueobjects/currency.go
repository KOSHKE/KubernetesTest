package valueobjects

import (
	"strings"

	"ecommerce-platform/pkg/common/errors"
)

// Currency represents an ISO 4217 currency code
type Currency struct {
	Code string `json:"code" gorm:"type:varchar(3);not null;default:'USD'"`
}

// Common currency codes
const (
	USD = "USD"
	EUR = "EUR"
	RUB = "RUB"
	GBP = "GBP"
	JPY = "JPY"
)

// NewCurrency creates a new Currency instance
func NewCurrency(code string) (Currency, error) {
	if code == "" {
		return Currency{}, errors.ErrEmptyCurrency
	}
	if len(code) != 3 {
		return Currency{}, errors.ErrInvalidCurrencyLength
	}
	if !isAlpha(code) {
		return Currency{}, errors.ErrInvalidCurrencyFormat
	}
	return Currency{Code: strings.ToUpper(code)}, nil
}

// isAlpha checks if string contains only letters
func isAlpha(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

// String returns the currency code as string
func (c Currency) String() string {
	return c.Code
}

// Equals checks if two currencies are equal
func (c Currency) Equals(other Currency) bool {
	return c.Code == other.Code
}

// IsZero checks if currency is empty
func (c Currency) IsZero() bool {
	return c.Code == ""
}
