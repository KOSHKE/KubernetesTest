package valueobjects

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	hashedValue string
}

func NewPassword(plainPassword string) (Password, error) {
	if len(plainPassword) < 6 {
		return Password{}, fmt.Errorf("password must be at least 6 characters long")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, fmt.Errorf("failed to hash password: %w", err)
	}

	return Password{hashedValue: string(hashedBytes)}, nil
}

func NewPasswordFromHash(hashedPassword string) Password {
	return Password{hashedValue: hashedPassword}
}

func (p Password) HashedValue() string {
	return p.hashedValue
}

func (p Password) Verify(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hashedValue), []byte(plainPassword))
	return err == nil
}
