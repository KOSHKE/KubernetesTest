package dto

// RegisterUserRequest represents the request to register a new user
type RegisterUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone"`
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
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenRequest represents the request to refresh authentication tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
