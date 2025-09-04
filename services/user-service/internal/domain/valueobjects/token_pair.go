package valueobjects

import "time"

// TokenPair represents a pair of authentication tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewTokenPair creates a new TokenPair
func NewTokenPair(accessToken, refreshToken string, expiresIn int64) *TokenPair {
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
}

// IsExpired checks if the access token is expired
func (tp *TokenPair) IsExpired() bool {
	return time.Now().After(tp.ExpiresAt)
}

// GetExpiresInSeconds returns the number of seconds until expiration
func (tp *TokenPair) GetExpiresInSeconds() int64 {
	expiresIn := time.Until(tp.ExpiresAt)
	if expiresIn < 0 {
		return 0
	}
	return int64(expiresIn.Seconds())
}
