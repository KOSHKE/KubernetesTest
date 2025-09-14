package services

import (
	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/jwt"
	"ecommerce-platform/services/user-service/internal/domain/ports/services"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// JWTTokenGenerator implements TokenGenerator using JWT
type JWTTokenGenerator struct {
	jwtManager *jwt.Manager
}

// NewJWTTokenGenerator creates a new JWT token generator
func NewJWTTokenGenerator(jwtManager *jwt.Manager) services.TokenGenerator {
	return &JWTTokenGenerator{
		jwtManager: jwtManager,
	}
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (g *JWTTokenGenerator) GenerateTokenPair(userID string) (valueobjects.Token, valueobjects.Token, error) {
	// Generate access token
	accessTokenString, accessExpiresAt, err := g.jwtManager.GenerateAccessToken(userID)
	if err != nil {
		return valueobjects.Token{}, valueobjects.Token{}, errors.ErrTokenGenerationFailed
	}

	// Generate refresh token
	refreshTokenString, refreshExpiresAt, err := g.jwtManager.GenerateRefreshToken(userID)
	if err != nil {
		return valueobjects.Token{}, valueobjects.Token{}, errors.ErrTokenGenerationFailed
	}

	// Create Token value objects
	accessToken := valueobjects.NewToken(accessTokenString, accessExpiresAt)
	refreshToken := valueobjects.NewToken(refreshTokenString, refreshExpiresAt)

	return accessToken, refreshToken, nil
}

// ValidateToken validates a token and returns claims
func (g *JWTTokenGenerator) ValidateToken(token string) (*jwt.Claims, error) {
	claims, err := g.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return nil, errors.ErrTokenValidationFailed
	}

	return claims, nil
}
