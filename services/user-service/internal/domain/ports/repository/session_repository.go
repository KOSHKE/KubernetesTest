package repository

import (
	"context"

	"ecommerce-platform/services/user-service/internal/domain/entities"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	// Save saves or updates a session
	Save(ctx context.Context, session *entities.Session) error

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id string) (*entities.Session, error)

	// GetByUserID retrieves all sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error)

	// GetByRefreshToken retrieves a session by refresh token
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error)

	// Delete removes a session
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error
}
