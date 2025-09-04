package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/user-service/internal/domain/ports/auth"
)

// LogoutUseCase handles user logout business logic
type LogoutUseCase struct {
	authSvc auth.AuthService
	logger  logger.Logger
}

// NewLogoutUseCase creates a new LogoutUseCase
func NewLogoutUseCase(authSvc auth.AuthService, logger logger.Logger) *LogoutUseCase {
	return &LogoutUseCase{
		authSvc: authSvc,
		logger:  logger,
	}
}

// Execute performs user logout
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	// Revoke refresh token
	if err := uc.authSvc.RevokeRefreshToken(ctx, refreshToken); err != nil {
		uc.logger.Error("failed to revoke refresh token", "error", err)
		return errors.ErrTokenRevocationFailed
	}

	return nil
}
