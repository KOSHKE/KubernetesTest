package usecases

import (
	"context"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/idgenerator"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

// RegisterUserUseCase handles user registration business logic
type RegisterUserUseCase struct {
	userRepo repository.UserRepository
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase
func NewRegisterUserUseCase(
	userRepo repository.UserRepository,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo: userRepo,
	}
}

// Execute performs user registration
func (uc *RegisterUserUseCase) Execute(ctx context.Context, email, password, firstName, lastName, phone string) (*entities.User, error) {
	// Create value objects with validation
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, err
	}

	firstNameVO := valueobjects.NewName(firstName)
	lastNameVO := valueobjects.NewName(lastName)

	phoneVO, err := valueobjects.NewPhone(phone)
	if err != nil {
		return nil, err
	}

	passwordVO, err := valueobjects.NewPassword(password)
	if err != nil {
		return nil, err
	}

	// Check if user already exists
	exists, err := uc.userRepo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.ErrEmailAlreadyExists
	}

	// Create user entity
	userID := idgenerator.GenerateID("user")
	now := time.Now()
	user := entities.NewUser(userID, emailVO, passwordVO, firstNameVO, lastNameVO, phoneVO, now, now)

	// Save user to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
