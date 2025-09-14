package jwt

import (
	"time"

	"ecommerce-platform/pkg/common/errors"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
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
	config Config // Changed: Config as value, not pointer
}

// NewManager creates new JWT manager
func NewManager(config Config) *Manager { // Changed: Config as value
	return &Manager{
		config: config,
	}
}

// GenerateTokenPair generates new access and refresh token pair
func (m *Manager) GenerateTokenPair(userID string) (*TokenPair, error) {
	// Generate access token
	accessToken, accessExpiresAt, err := m.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, refreshExpiresAt, err := m.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// ValidateAccessToken validates access token and returns claims
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.parseToken(tokenString, m.config.AccessTokenSecret)
}

// ValidateRefreshToken validates refresh token and returns claims
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.parseToken(tokenString, m.config.RefreshTokenSecret)
}

// parseToken parses JWT token and returns Claims
func (m *Manager) parseToken(tokenString, secret string) (*Claims, error) {
	// Build parser options for issuer and audience validation
	var parserOptions []jwt.ParserOption
	if m.config.Issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(m.config.Issuer))
	}
	if m.config.Audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(m.config.Audience))
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrUnexpectedSign
		}
		return []byte(secret), nil
	}, parserOptions...)

	if err != nil {
		return nil, err
	}

	if token.Valid {
		return claims, nil
	}

	return nil, errors.ErrTokenInvalid
}

// RefreshAccessToken generates new access and refresh token pair using refresh token
func (m *Manager) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair (both access and refresh)
	tokenPair, err := m.GenerateTokenPair(claims.UserID)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// GetAccessTokenTTL returns access token TTL from config
func (m *Manager) GetAccessTokenTTL() time.Duration {
	return m.config.AccessTokenTTL
}

// GetRefreshTokenTTL returns refresh token TTL from config
func (m *Manager) GetRefreshTokenTTL() time.Duration {
	return m.config.RefreshTokenTTL
}

// GenerateAccessToken generates access token and returns token string with expiration time
func (m *Manager) GenerateAccessToken(userID string) (string, time.Time, error) {
	// Generate claims first to get exact expiration time
	now := time.Now()

	var audience []string
	if m.config.Audience != "" {
		audience = []string{m.config.Audience}
	}

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
			Audience:  audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.AccessTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, claims.ExpiresAt.Time, nil
}

// GenerateRefreshToken generates refresh token and returns token string with expiration time
func (m *Manager) GenerateRefreshToken(userID string) (string, time.Time, error) {
	// Generate claims first to get exact expiration time
	now := time.Now()

	var audience []string
	if m.config.Audience != "" {
		audience = []string{m.config.Audience}
	}

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
			Audience:  audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.RefreshTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, claims.ExpiresAt.Time, nil
}
