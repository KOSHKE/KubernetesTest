package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/auth"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
	"ecommerce-platform/services/user-service/internal/metrics"
)

// LoginUserUseCase handles user login business logic
type LoginUserUseCase struct {
	userRepo    repository.UserRepository
	authService auth.AuthService
	logger      logger.Logger
	metrics     metrics.UserMetrics
}

// NewLoginUserUseCase creates a new LoginUserUseCase
func NewLoginUserUseCase(
	userRepo repository.UserRepository,
	authService auth.AuthService,
	logger logger.Logger,
	metrics metrics.UserMetrics,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:    userRepo,
		authService: authService,
		logger:      logger,
		metrics:     metrics,
	}
}

// Execute performs user login
func (uc *LoginUserUseCase) Execute(ctx context.Context, email, password string) (*entities.User, *valueobjects.TokenPair, error) {
	// Create email value object
	emailVO := valueobjects.NewEmail(email)

	// Find user by email
	user, err := uc.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		uc.logger.Error("failed to find user", "error", err)
		return nil, nil, errors.ErrUserNotFound
	}

	// Verify password
	if !user.Password.Verify(password) {
		uc.logger.Warn("invalid password for user", "email", email)
		if uc.metrics != nil {
			uc.metrics.UserLoginFailed("invalid_password")
		}
		return nil, nil, errors.ErrInvalidCredentials
	}

	var domainTokenPair *valueobjects.TokenPair

	// Execute token generation and storage within a transaction
	err = uc.authService.WithTransaction(ctx, func(txCtx context.Context) error {
		// Generate authentication tokens
		tokenPair, err := uc.authService.GenerateTokenPair(user.ID, user.Email.Value())
		if err != nil {
			uc.logger.Error("failed to generate tokens", "error", err)
			return errors.ErrTokenGenerationFailed
		}

		// Store refresh token
		if err := uc.authService.StoreRefreshToken(txCtx, tokenPair.RefreshToken, user.ID); err != nil {
			uc.logger.Error("failed to store refresh token", "error", err)
			return errors.ErrTokenStorageFailed
		}

		// Create domain token pair
		domainTokenPair = valueobjects.NewTokenPair(tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Record metrics for successful login
	if uc.metrics != nil {
		uc.metrics.UserLoginSuccess()
	}

	return user, domainTokenPair, nil
}
