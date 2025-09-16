package valueobjects

import (
	"strings"
	"time"
)

// Token represents a JWT token value object
type Token struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewToken creates a new token
func NewToken(value string, expiresAt time.Time) Token {
	return Token{
		Value:     strings.TrimSpace(value),
		ExpiresAt: expiresAt,
	}
}

// IsExpired checks if the token is expired
func (t Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsEmpty checks if the token is empty
func (t Token) IsEmpty() bool {
	return t.Value == ""
}

// Equals checks if two tokens are equal
func (t Token) Equals(other Token) bool {
	return t.Value == other.Value
}

// String returns the string representation
func (t Token) String() string {
	return t.Value
}
