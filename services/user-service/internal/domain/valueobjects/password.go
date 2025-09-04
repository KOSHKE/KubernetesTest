package valueobjects

import (
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	hashedValue string
}

func NewPassword(hashedPassword string) Password {
	return Password{hashedValue: hashedPassword}
}

// NewPasswordFromPlain creates a new Password from plain text with hashing
func NewPasswordFromPlain(plainPassword string) Password {
	// Hash the password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		// In production, this should never happen, but we handle it gracefully
		panic("failed to hash password")
	}

	return Password{hashedValue: string(hashedBytes)}
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

// String returns the hashed password as string
func (p Password) String() string {
	return p.hashedValue
}
