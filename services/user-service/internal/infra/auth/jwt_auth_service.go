package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jwt"
	"redis"
	"user-service/internal/ports/auth"
	"user-service/internal/ports/repository"
)

// JWTAuthService implements auth.AuthService interface
type JWTAuthService struct {
	jwtManager  *jwt.Manager
	redisClient *redis.Client
	userRepo    repository.UserRepository
	config      *Config
}

// Config holds JWT authentication configuration
type Config struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	RedisURL           string
}

// NewJWTAuthService creates new JWT authentication service
func NewJWTAuthService(config *Config, userRepo repository.UserRepository) (*JWTAuthService, error) {

	// Parse Redis URL
	redisAddr := config.RedisURL
	redisPassword := ""
	redisDB := 0

	if strings.Contains(redisAddr, "@") {
		parts := strings.Split(redisAddr, "@")
		if len(parts) == 2 {
			authPart := parts[0]
			hostPart := parts[1]

			if strings.Contains(authPart, ":") {
				authParts := strings.Split(authPart, ":")
				if len(authParts) == 2 {
					redisPassword = authParts[1]
				}
			}
			redisAddr = hostPart
		}
	}

	// Create Redis client
	redisClient := redis.NewClient(redisAddr, redisPassword, redisDB)

	// Test Redis connection
	if err := redisClient.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create JWT manager
	jwtConfig := &jwt.Config{
		AccessTokenSecret:  config.AccessTokenSecret,
		RefreshTokenSecret: config.RefreshTokenSecret,
		AccessTokenTTL:     config.AccessTokenTTL,
		RefreshTokenTTL:    config.RefreshTokenTTL,
	}
	jwtManager := jwt.NewManager(jwtConfig)

	return &JWTAuthService{
		jwtManager:  jwtManager,
		redisClient: redisClient,
		userRepo:    userRepo,
		config:      config,
	}, nil
}

// GenerateTokenPair generates new access and refresh token pair
func (s *JWTAuthService) GenerateTokenPair(userID, email string) (*auth.TokenPair, error) {

	tokenPair, err := s.jwtManager.GenerateTokenPair(userID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	return &auth.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// StoreRefreshToken stores refresh token in Redis
func (s *JWTAuthService) StoreRefreshToken(ctx context.Context, refreshToken, userID string) error {
	return s.redisClient.StoreRefreshToken(ctx, refreshToken, userID, s.config.RefreshTokenTTL)
}

// ValidateAccessToken validates access token and returns claims
func (s *JWTAuthService) ValidateAccessToken(tokenString string) (map[string]interface{}, error) {
	claims, err := s.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to validate access token: %w", err)
	}

	return map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"exp":     claims.ExpiresAt.Unix(),
		"iat":     claims.IssuedAt.Unix(),
	}, nil
}

// ValidateRefreshToken validates refresh token and returns claims
func (s *JWTAuthService) ValidateRefreshToken(tokenString string) (map[string]interface{}, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Check if token exists in Redis
	if !s.redisClient.IsRefreshTokenValid(context.Background(), tokenString) {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	return map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"exp":     claims.ExpiresAt.Unix(),
		"iat":     claims.IssuedAt.Unix(),
	}, nil
}

// RefreshAccessToken generates new access token using refresh token
func (s *JWTAuthService) RefreshAccessToken(refreshToken string) (string, error) {
	// Validate refresh token first
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to validate refresh token: %w", err)
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID in token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", fmt.Errorf("invalid email in token")
	}

	// Generate new access token using the user information from claims
	accessToken, err := s.jwtManager.GenerateTokenPair(userID, email)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessToken.AccessToken, nil
}

// RevokeRefreshToken removes refresh token from Redis
func (s *JWTAuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	return s.redisClient.RevokeRefreshToken(ctx, refreshToken)
}

// RevokeAllUserTokens removes all refresh tokens for a specific user
func (s *JWTAuthService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.redisClient.RevokeAllUserTokens(ctx, userID)
}

// Close closes Redis connection
func (s *JWTAuthService) Close() error {
	return s.redisClient.Close()
}
