package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
)

// LogoutUseCase handles user logout business logic
type LogoutUseCase struct {
	sessionRepo repository.SessionRepository
}

// NewLogoutUseCase creates a new LogoutUseCase
func NewLogoutUseCase(sessionRepo repository.SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
	}
}

// Execute performs user logout
func (uc *LogoutUseCase) Execute(ctx context.Context, sessionID string) error {
	// Check if session exists
	_, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return errors.ErrSessionNotFound
	}

	// Delete session
	if err := uc.sessionRepo.Delete(ctx, sessionID); err != nil {
		return errors.ErrSessionDeletionFailed
	}

	return nil
}
