package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/user-service/internal/domain/ports/auth"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// RefreshTokenUseCase handles token refresh business logic
type RefreshTokenUseCase struct {
	authService auth.AuthService
	logger      logger.Logger
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase
func NewRefreshTokenUseCase(
	authService auth.AuthService,
	logger logger.Logger,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		authService: authService,
		logger:      logger,
	}
}

// Execute refreshes user authentication tokens
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, refreshToken string) (*valueobjects.TokenPair, error) {
	var tokenPair *auth.TokenPair
	var domainTokenPair *valueobjects.TokenPair

	// Execute all operations within a transaction
	err := uc.authService.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate refresh token
		claims, err := uc.authService.ValidateRefreshToken(refreshToken)
		if err != nil {
			uc.logger.Error("failed to validate refresh token", "error", err)
			return errors.ErrTokenValidationFailed
		}

		// Extract user ID from claims
		userID, ok := claims["user_id"].(string)
		if !ok {
			uc.logger.Error("invalid token claims - missing user_id")
			return errors.ErrInvalidTokenClaims
		}

		// Generate new token pair
		tokenPair, err = uc.authService.GenerateTokenPair(userID, claims["email"].(string))
		if err != nil {
			uc.logger.Error("failed to generate new tokens", "error", err)
			return errors.ErrTokenGenerationFailed
		}

		// Store new refresh token
		if err := uc.authService.StoreRefreshToken(txCtx, tokenPair.RefreshToken, userID); err != nil {
			uc.logger.Error("failed to store new refresh token", "error", err)
			return errors.ErrTokenStorageFailed
		}

		// Revoke old refresh token
		if err := uc.authService.RevokeRefreshToken(txCtx, refreshToken); err != nil {
			uc.logger.Error("failed to revoke old refresh token", "error", err)
			return errors.ErrTokenRevocationFailed
		}

		// Create domain token pair
		domainTokenPair = valueobjects.NewTokenPair(tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return domainTokenPair, nil
}
