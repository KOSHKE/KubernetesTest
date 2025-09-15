package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidPhone = errors.New("invalid phone format")
)

type Phone struct {
	value string
}

// phoneRegex validates international phone numbers (E.164 format)
var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// NewPhone creates a new Phone with validation
func NewPhone(phone string) (Phone, error) {
	phone = strings.TrimSpace(phone)

	phoneVO := Phone{value: phone}
	if err := phoneVO.Validate(); err != nil {
		return Phone{}, err
	}

	return phoneVO, nil
}

// Validate validates the phone format
func (p Phone) Validate() error {
	// Allow empty phone (optional field)
	if p.value == "" {
		return nil
	}

	// Check if it's a valid international format
	if !phoneRegex.MatchString(p.value) {
		return ErrInvalidPhone
	}

	return nil
}

func (p Phone) Value() string {
	return p.value
}

func (p Phone) String() string {
	return p.value
}

func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}
