package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/user-service/internal/domain/ports/auth"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"

	"ecommerce-platform/pkg/jwt"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/redisclient"
)

// JWTAuthService implements auth.AuthService interface
type JWTAuthService struct {
	jwtManager   *jwt.Manager
	tokenStorage auth.TokenStorage
	userRepo     repository.UserRepository
	config       *Config
	logger       logger.Logger
}

// Config holds JWT authentication configuration
type Config struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	StorageURL         string        // URL for token storage (currently Redis)
	StorageTimeout     time.Duration // timeout for storage operations
}

// NewJWTAuthService creates new JWT authentication service
func NewJWTAuthService(config *Config, userRepo repository.UserRepository, logger logger.Logger) (*JWTAuthService, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	// Parse storage connection details
	storageAddr, storagePassword, storageDB, err := parseStorageURL(config.StorageURL)
	if err != nil {
		return nil, err
	}

	// Create storage client and test connection
	storageClient := redisclient.New(storageAddr, storagePassword, storageDB, logger)
	if err := storageClient.Ping(context.Background()); err != nil {
		logger.Error("failed to connect to storage", "error", err, "storage_url", config.StorageURL)
		return nil, errors.ErrStorageConnectionFailed
	}

	// Create token storage
	tokenStorage := NewRedisTokenStorage(storageClient, logger)
	if tokenStorage == nil {
		logger.Error("failed to create token storage", "storage_addr", storageAddr)
		return nil, errors.ErrTokenStorageFailed
	}

	// Create JWT manager
	jwtConfig := jwt.Config{
		AccessTokenSecret:  config.AccessTokenSecret,
		RefreshTokenSecret: config.RefreshTokenSecret,
		AccessTokenTTL:     config.AccessTokenTTL,
		RefreshTokenTTL:    config.RefreshTokenTTL,
		Issuer:             "user-service",
		Audience:           "ecommerce-platform",
	}
	jwtManager := jwt.NewManager(jwtConfig, logger)
	if jwtManager == nil {
		logger.Error("failed to create JWT manager", "issuer", jwtConfig.Issuer, "audience", jwtConfig.Audience)
		return nil, errors.ErrJWTManagerCreationFailed
	}

	return &JWTAuthService{
		jwtManager:   jwtManager,
		tokenStorage: tokenStorage,
		userRepo:     userRepo,
		config:       config,
		logger:       logger,
	}, nil
}

// validateConfig validates JWT authentication configuration
func validateConfig(config *Config) error {
	if config.AccessTokenSecret == "" {
		return errors.ErrAccessTokenSecretRequired
	}
	if config.RefreshTokenSecret == "" {
		return errors.ErrRefreshTokenSecretRequired
	}
	if config.AccessTokenTTL <= 0 {
		return errors.ErrAccessTokenTTLInvalid
	}
	if config.RefreshTokenTTL <= 0 {
		return errors.ErrRefreshTokenTTLInvalid
	}
	if config.StorageURL == "" {
		return errors.ErrStorageURLRequired
	}
	if config.StorageTimeout <= 0 {
		return errors.ErrStorageTimeoutInvalid
	}
	return nil
}

// parseStorageURL parses storage connection string and returns connection details
func parseStorageURL(storageURL string) (addr, password string, db int, err error) {
	// Handle simple host:port format
	if !strings.Contains(storageURL, "@") && !strings.Contains(storageURL, "/") {
		return storageURL, "", 0, nil
	}

	// Handle redis://user:password@host:port/db format
	storageURL = strings.TrimPrefix(storageURL, "redis://")

	// Split by @ to separate auth from host
	parts := strings.Split(storageURL, "@")
	if len(parts) != 2 {
		return "", "", 0, errors.ErrInvalidStorageURL
	}

	authPart := parts[0]
	hostPart := parts[1]

	// Extract password from auth part
	if strings.Contains(authPart, ":") {
		authParts := strings.Split(authPart, ":")
		if len(authParts) == 2 {
			password = authParts[1]
		}
	}

	// Extract database number from host part
	if strings.Contains(hostPart, "/") {
		hostParts := strings.Split(hostPart, "/")
		if len(hostParts) == 2 {
			hostPart = hostParts[0]
			if dbStr := hostParts[1]; dbStr != "" {
				if dbNum, parseErr := strconv.Atoi(dbStr); parseErr == nil {
					db = dbNum
				}
			}
		}
	}

	return hostPart, password, db, nil
}

// GenerateTokenPair generates new access and refresh token pair
func (s *JWTAuthService) GenerateTokenPair(userID, email string) (*auth.TokenPair, error) {
	tokenPair, err := s.jwtManager.GenerateTokenPair(userID, email)
	if err != nil {
		s.logger.Error("failed to generate token pair", "error", err, "user_id", userID)
		return nil, errors.ErrTokenGenerationFailed
	}

	return &auth.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// StoreRefreshToken stores refresh token in token storage
func (s *JWTAuthService) StoreRefreshToken(ctx context.Context, refreshToken, userID string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.StorageTimeout)
	defer cancel()

	err := s.tokenStorage.StoreToken(timeoutCtx, refreshToken, userID, s.config.RefreshTokenTTL)
	if err != nil {
		s.logger.Error("failed to store refresh token", "error", err, "user_id", userID)
		return errors.ErrTokenStorageFailed
	}
	return nil
}

// ValidateAccessToken validates access token and returns claims
func (s *JWTAuthService) ValidateAccessToken(tokenString string) (map[string]interface{}, error) {
	claims, err := s.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		s.logger.Error("failed to validate access token", "error", err)
		return nil, errors.ErrTokenValidationFailed
	}

	return s.mapClaimsToMap(claims), nil
}

// ValidateRefreshToken validates refresh token and returns claims
func (s *JWTAuthService) ValidateRefreshToken(tokenString string) (map[string]interface{}, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(tokenString)
	if err != nil {
		return nil, errors.ErrTokenValidationFailed
	}

	// Create context with timeout for Redis operation
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.config.StorageTimeout)
	defer cancel()

	// Check if token exists in storage
	if !s.tokenStorage.IsTokenValid(timeoutCtx, tokenString) {
		s.logger.Warn("refresh token has been revoked", "user_id", claims.UserID)
		return nil, errors.ErrTokenRevoked
	}

	return s.mapClaimsToMap(claims), nil
}

// mapClaimsToMap converts JWT claims to map[string]interface{}
func (s *JWTAuthService) mapClaimsToMap(claims *jwt.Claims) map[string]interface{} {
	return map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"exp":     claims.ExpiresAt.Unix(),
		"iat":     claims.IssuedAt.Unix(),
	}
}

// RefreshAccessToken generates new access and refresh token pair using refresh token
func (s *JWTAuthService) RefreshAccessToken(refreshToken string) (*auth.TokenPair, error) {
	// Validate refresh token first
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.ErrTokenInvalid
	}

	// Generate new token pair using the user information from claims
	tokenPair, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, errors.ErrTokenGenerationFailed
	}

	// Create context with timeout for Redis operations
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.config.StorageTimeout)
	defer cancel()

	// Store new refresh token in token storage
	err = s.StoreRefreshToken(timeoutCtx, tokenPair.RefreshToken, userID)
	if err != nil {
		s.logger.Error("failed to store new refresh token", "error", err, "user_id", userID)
		return nil, errors.ErrTokenStorageFailed
	}

	// Revoke old refresh token
	err = s.RevokeRefreshToken(timeoutCtx, refreshToken)
	if err != nil {
		s.logger.Warn("failed to revoke old refresh token", "error", err, "user_id", userID)
		// Don't fail the operation if revocation fails
	}

	return &auth.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RevokeRefreshToken removes refresh token from token storage
func (s *JWTAuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.StorageTimeout)
	defer cancel()

	return s.tokenStorage.RevokeToken(timeoutCtx, refreshToken)
}

// RevokeAllUserTokens removes all refresh tokens for a specific user from token storage
func (s *JWTAuthService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.StorageTimeout)
	defer cancel()

	return s.tokenStorage.RevokeAllUserTokens(timeoutCtx, userID)
}

// Close closes token storage connection
func (s *JWTAuthService) Close() error {
	return s.tokenStorage.Close()
}
