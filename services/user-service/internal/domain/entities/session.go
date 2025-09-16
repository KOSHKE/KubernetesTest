package entities

import (
	"encoding/json"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// Session represents a user authentication session
type Session struct {
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	AccessToken  valueobjects.Token `json:"access_token"`
	RefreshToken valueobjects.Token `json:"refresh_token"`
	ExpiresAt    time.Time          `json:"expires_at"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// NewSession creates a new session
func NewSession(id, userID string, accessToken, refreshToken valueobjects.Token, expiresAt time.Time) *Session {
	now := time.Now()
	return &Session{
		ID:           id,
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// IsExpired checks if the session is expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Refresh updates the session with new tokens
func (s *Session) Refresh(accessToken, refreshToken valueobjects.Token, expiresAt time.Time) {
	s.AccessToken = accessToken
	s.RefreshToken = refreshToken
	s.ExpiresAt = expiresAt
	s.UpdatedAt = time.Now()
}

// Validate validates session data
func (s *Session) Validate() error {
	if s.ID == "" {
		return errors.ErrInvalidSessionID
	}
	if s.UserID == "" {
		return errors.ErrInvalidUserID
	}
	if s.AccessToken.IsEmpty() {
		return errors.ErrInvalidAccessToken
	}
	if s.RefreshToken.IsEmpty() {
		return errors.ErrInvalidRefreshToken
	}
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler for Redis serialization
func (s *Session) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Redis deserialization
func (s *Session) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}
