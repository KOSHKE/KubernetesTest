package auth

import (
	"context"
)

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// AuthService defines authentication service interface
type AuthService interface {
	// GenerateTokenPair generates new access and refresh token pair
	GenerateTokenPair(userID, email string) (*TokenPair, error)

	// StoreRefreshToken stores refresh token in Redis
	StoreRefreshToken(ctx context.Context, refreshToken, userID string) error

	// ValidateRefreshToken validates refresh token and returns claims
	ValidateRefreshToken(tokenString string) (map[string]interface{}, error)

	// RevokeRefreshToken removes refresh token from Redis
	RevokeRefreshToken(ctx context.Context, refreshToken string) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}
