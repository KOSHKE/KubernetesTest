package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kubernetestest/ecommerce-platform/pkg/logger"
)

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Config represents JWT configuration
type Config struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	Issuer             string // Added: JWT issuer claim
	Audience           string // Added: JWT audience claim
}

// Manager handles JWT operations
type Manager struct {
	config Config        // Changed: Config as value, not pointer
	logger logger.Logger // Added: unified logger
}

// NewManager creates new JWT manager
func NewManager(config Config, logger logger.Logger) *Manager { // Changed: Config as value, added logger
	return &Manager{
		config: config,
		logger: logger,
	}
}

// GenerateTokenPair generates new access and refresh token pair
func (m *Manager) GenerateTokenPair(userID, email string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := m.generateAccessToken(userID, email)
	if err != nil {
		m.logger.Error("failed to generate access token", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := m.generateRefreshToken(userID, email)
	if err != nil {
		m.logger.Error("failed to generate refresh token", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.config.AccessTokenTTL.Seconds()),
	}, nil
}

// ValidateAccessToken validates access token and returns claims
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.parseToken(tokenString, m.config.AccessTokenSecret, "access token")
}

// ValidateRefreshToken validates refresh token and returns claims
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.parseToken(tokenString, m.config.RefreshTokenSecret, "refresh token")
}

// parseToken — общий метод валидации токена
func (m *Manager) parseToken(tokenString, secret, tokenType string) (*Claims, error) {
	// Build parser options for issuer and audience validation
	var parserOptions []jwt.ParserOption
	if m.config.Issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(m.config.Issuer))
	}
	if m.config.Audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(m.config.Audience))
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, parserOptions...)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			m.logger.Error(tokenType+" expired", "error", err)
			return nil, fmt.Errorf("%s expired: %w", tokenType, err)
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			m.logger.Error(tokenType+" not valid yet", "error", err)
			return nil, fmt.Errorf("%s not valid yet: %w", tokenType, err)
		case errors.Is(err, jwt.ErrTokenMalformed):
			m.logger.Error("malformed "+tokenType, "error", err)
			return nil, fmt.Errorf("malformed %s: %w", tokenType, err)
		default:
			m.logger.Error("failed to parse "+tokenType, "error", err)
			return nil, fmt.Errorf("failed to parse %s: %w", tokenType, err)
		}
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	m.logger.Error("invalid " + tokenType + " claims")
	return nil, fmt.Errorf("invalid %s", tokenType)
}

// RefreshAccessToken generates new access and refresh token pair using refresh token
func (m *Manager) RefreshAccessToken(refreshToken string) (*TokenPair, error) { // Changed: returns TokenPair, not just string
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		m.logger.Error("failed to validate refresh token", "error", err)
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Generate new token pair (both access and refresh)
	tokenPair, err := m.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		m.logger.Error("failed to generate new token pair", "user_id", claims.UserID, "error", err)
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	m.logger.Info("successfully refreshed token pair", "user_id", claims.UserID)
	return tokenPair, nil
}

// generateAccessToken generates access token
func (m *Manager) generateAccessToken(userID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,             // Added: issuer claim
			Audience:  []string{m.config.Audience}, // Added: audience claim
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.AccessTokenSecret))
}

// generateRefreshToken generates refresh token
func (m *Manager) generateRefreshToken(userID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,             // Added: issuer claim
			Audience:  []string{m.config.Audience}, // Added: audience claim
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.RefreshTokenSecret))
}

// GenerateRandomString generates random string for token revocation
// bytes parameter specifies the number of random bytes (result string will be longer due to base64 encoding)
func GenerateRandomString(bytes int) (string, error) { // Changed: parameter name from length to bytes
	randomBytes := make([]byte, bytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}
