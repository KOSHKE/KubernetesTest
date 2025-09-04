package valueobjects

import (
	"strings"
)

type Email struct {
	value string
}

// NewEmail creates a new Email (assumes validation already done)
func NewEmail(email string) Email {
	email = strings.TrimSpace(strings.ToLower(email))
	return Email{value: email}
}

func (e Email) Value() string {
	return e.value
}

func (e Email) String() string {
	return e.value
}

func (e Email) Equals(other Email) bool {
	return e.value == other.value
}
