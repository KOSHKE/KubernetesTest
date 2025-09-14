package valueobjects

import (
	"strings"
	"time"
)

// Token represents a JWT token value object
type Token struct {
	value     string
	expiresAt time.Time
}

// NewToken creates a new token
func NewToken(value string, expiresAt time.Time) Token {
	return Token{
		value:     strings.TrimSpace(value),
		expiresAt: expiresAt,
	}
}

// Value returns the token value
func (t Token) Value() string {
	return t.value
}

// ExpiresAt returns the token expiration time
func (t Token) ExpiresAt() time.Time {
	return t.expiresAt
}

// IsExpired checks if the token is expired
func (t Token) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}

// IsEmpty checks if the token is empty
func (t Token) IsEmpty() bool {
	return t.value == ""
}

// Equals checks if two tokens are equal
func (t Token) Equals(other Token) bool {
	return t.value == other.value
}

// String returns the string representation
func (t Token) String() string {
	return t.value
}
