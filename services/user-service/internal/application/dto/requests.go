package dto

import (
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// RegisterUserRequest represents the request to register a new user
type RegisterUserRequest struct {
	Email     valueobjects.Email `json:"email" validate:"required"`
	Password  string             `json:"password" validate:"required"`
	FirstName valueobjects.Name  `json:"first_name" validate:"required"`
	LastName  valueobjects.Name  `json:"last_name" validate:"required"`
	Phone     valueobjects.Phone `json:"phone" validate:"omitempty"`
}

// LoginRequest represents the request to login a user
type LoginRequest struct {
	Email    valueobjects.Email `json:"email" validate:"required"`
	Password string             `json:"password" validate:"required"`
}

// GetUserRequest represents the request to get user information
type GetUserRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// LogoutRequest represents the request to logout a user
type LogoutRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

// RefreshTokenRequest represents the request to refresh authentication tokens
type RefreshTokenRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}
