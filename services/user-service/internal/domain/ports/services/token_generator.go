package services

import (
	"ecommerce-platform/pkg/jwt"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// TokenGenerator defines the interface for token generation
type TokenGenerator interface {
	// GenerateTokenPair generates both access and refresh tokens for a user
	GenerateTokenPair(userID string) (valueobjects.Token, valueobjects.Token, error)

	// ValidateToken validates a token and returns claims
	ValidateToken(token string) (*jwt.Claims, error)
}
