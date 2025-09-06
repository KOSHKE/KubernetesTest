package valueobjects

import (
	"strings"

	"ecommerce-platform/pkg/common/errors"
)

// Currency represents an ISO 4217 currency code
type Currency struct {
	Code string `json:"code" gorm:"type:varchar(3);not null;"`
}

// NewCurrency creates a new Currency instance
func NewCurrency(code string) (Currency, error) {
	currency := Currency{Code: strings.ToUpper(code)}

	if err := currency.Validate(); err != nil {
		return Currency{}, err
	}

	return currency, nil
}

// Validate validates currency data
func (c Currency) Validate() error {
	if c.Code == "" {
		return errors.ErrEmptyCurrency
	}
	if len(c.Code) != 3 {
		return errors.ErrInvalidCurrencyLength
	}
	if !isAlpha(c.Code) {
		return errors.ErrInvalidCurrencyFormat
	}
	return nil
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
