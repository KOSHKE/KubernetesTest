package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
)

type Email struct {
	value string
}

// emailRegex is a simple email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates a new Email with validation
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	emailVO := Email{value: email}
	if err := emailVO.Validate(); err != nil {
		return Email{}, err
	}

	return emailVO, nil
}

// Validate validates the email format
func (e Email) Validate() error {
	if e.value == "" {
		return ErrInvalidEmail
	}

	if !emailRegex.MatchString(e.value) {
		return ErrInvalidEmail
	}

	return nil
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
