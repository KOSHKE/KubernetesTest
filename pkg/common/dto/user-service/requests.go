package dto

// RegisterUserRequest represents the request to register a new user
type RegisterUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string `json:"phone" validate:"omitempty"`
}

// LoginRequest represents the request to login a user
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
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
