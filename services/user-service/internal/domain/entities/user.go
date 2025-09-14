package entities

import (
	"time"

	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// User represents the User aggregate root
type User struct {
	ID        string
	Email     valueobjects.Email
	Password  valueobjects.Password
	FirstName valueobjects.Name
	LastName  valueobjects.Name
	Phone     valueobjects.Phone
	CreatedAt time.Time
	UpdatedAt time.Time
}

// VerifyPassword verifies a plain password against the user's hashed password
func (u *User) VerifyPassword(password string) bool {
	return u.Password.Verify(password)
}

// NewUser creates a new User aggregate
func NewUser(id string, email valueobjects.Email, password valueobjects.Password, firstName, lastName valueobjects.Name, phone valueobjects.Phone) *User {
	now := time.Now()
	return &User{
		ID:        id,
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
