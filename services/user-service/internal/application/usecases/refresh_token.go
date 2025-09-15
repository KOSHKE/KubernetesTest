package usecases

import (
	"context"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/ports/services"
)

// RefreshTokenUseCase handles token refresh business logic
type RefreshTokenUseCase struct {
	sessionRepo    repository.SessionRepository
	tokenGenerator services.TokenGenerator
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase
func NewRefreshTokenUseCase(
	sessionRepo repository.SessionRepository,
	tokenGenerator services.TokenGenerator,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		sessionRepo:    sessionRepo,
		tokenGenerator: tokenGenerator,
	}
}

// Execute refreshes user authentication tokens
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, sessionID string) (*entities.Session, error) {
	// Get existing session
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if session.IsExpired() {
		// Clean up expired session
		_ = uc.sessionRepo.Delete(ctx, sessionID)
		return nil, errors.ErrSessionExpired
	}

	// Generate new token pair
	accessToken, refreshToken, err := uc.tokenGenerator.GenerateTokenPair(session.UserID)
	if err != nil {
		return nil, err
	}

	// Update session with new tokens
	expiresAt := time.Now().Add(24 * time.Hour)
	session.Refresh(accessToken, refreshToken, expiresAt)

	// Save updated session
	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}
