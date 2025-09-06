package auth

import (
	"context"
	"time"
)

// TokenStorage defines interface for token storage operations
type TokenStorage interface {
	// StoreToken stores refresh token with TTL
	StoreToken(ctx context.Context, token, userID string, ttl time.Duration) error

	// IsTokenValid checks if token exists and is valid
	IsTokenValid(ctx context.Context, token string) bool

	// RevokeToken removes specific token
	RevokeToken(ctx context.Context, token string) error

	// RevokeAllUserTokens removes all tokens for specific user
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(context.Context) error) error

	// Close closes the storage connection
	Close() error
}
