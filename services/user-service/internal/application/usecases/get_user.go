package usecases

import (
	"context"

	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
)

// GetUserUseCase handles retrieving user information
type GetUserUseCase struct {
	userRepo repository.UserRepository
}

// NewGetUserUseCase creates a new GetUserUseCase
func NewGetUserUseCase(userRepo repository.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves user information by ID
func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*entities.User, error) {
	// Get user from repository
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
