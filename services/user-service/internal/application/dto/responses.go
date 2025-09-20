package dto

import (
	"time"

	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// RegisterUserResponse represents the response after user registration
type RegisterUserResponse struct {
	UserID    string             `json:"user_id"`
	Email     valueobjects.Email `json:"email"`
	FirstName valueobjects.Name  `json:"first_name"`
	LastName  valueobjects.Name  `json:"last_name"`
	Phone     valueobjects.Phone `json:"phone"`
	CreatedAt time.Time          `json:"created_at"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	UserID       string             `json:"user_id"`
	Email        valueobjects.Email `json:"email"`
	FirstName    valueobjects.Name  `json:"first_name"`
	LastName     valueobjects.Name  `json:"last_name"`
	Phone        valueobjects.Phone `json:"phone"`
	SessionID    string             `json:"session_id"`
	AccessToken  valueobjects.Token `json:"access_token"`
	RefreshToken valueobjects.Token `json:"refresh_token"`
	ExpiresAt    time.Time          `json:"expires_at"`
}

// GetUserResponse represents the user information response
type GetUserResponse struct {
	UserID    string             `json:"user_id"`
	Email     valueobjects.Email `json:"email"`
	FirstName valueobjects.Name  `json:"first_name"`
	LastName  valueobjects.Name  `json:"last_name"`
	Phone     valueobjects.Phone `json:"phone"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// RefreshTokenResponse represents the response after token refresh
type RefreshTokenResponse struct {
	AccessToken  valueobjects.Token `json:"access_token"`
	RefreshToken valueobjects.Token `json:"refresh_token"`
	ExpiresAt    time.Time          `json:"expires_at"`
}
