package valueobjects

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword    = errors.New("invalid password format")
	ErrPasswordTooShort   = errors.New("password too short")
	ErrPasswordTooLong    = errors.New("password too long")
	ErrPasswordHashFailed = errors.New("failed to hash password")
)

type Password struct {
	hashedValue string
}

// NewPassword creates a new Password from plain text with hashing
func NewPassword(plainPassword string) (Password, error) {
	// Validate password before hashing
	if err := validatePassword(plainPassword); err != nil {
		return Password{}, err
	}

	// Hash the password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, ErrPasswordHashFailed
	}

	return Password{hashedValue: string(hashedBytes)}, nil
}

// NewPasswordFromHashed creates a new Password from already hashed value
func NewPasswordFromHashed(hashedPassword string) Password {
	return Password{hashedValue: hashedPassword}
}

// validatePassword validates password strength
func validatePassword(password string) error {
	password = strings.TrimSpace(password)

	// Check minimum length
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	// Check maximum length
	if len(password) > 128 {
		return ErrPasswordTooLong
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrInvalidPassword
	}

	return nil
}

// Verify checks if the plain password matches the hashed password
func (p Password) Verify(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hashedValue), []byte(plainPassword))
	return err == nil
}

// HashedValue returns the hashed password value
func (p Password) HashedValue() string {
	return p.hashedValue
}
