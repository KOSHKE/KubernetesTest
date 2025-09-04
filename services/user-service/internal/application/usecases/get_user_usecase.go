package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
)

// GetUserUseCase handles retrieving user information
type GetUserUseCase struct {
	userRepo repository.UserRepository
	logger   logger.Logger
}

// NewGetUserUseCase creates a new GetUserUseCase
func NewGetUserUseCase(userRepo repository.UserRepository, logger logger.Logger) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Execute retrieves user information by ID
func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*entities.User, error) {
	// Get user from repository
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to get user", "error", err)
		return nil, errors.ErrUserNotFound
	}

	return user, nil
}
