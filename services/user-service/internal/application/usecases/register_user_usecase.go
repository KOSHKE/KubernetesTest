package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/idgenerator"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
	"ecommerce-platform/services/user-service/internal/metrics"
)

// RegisterUserUseCase handles user registration business logic
type RegisterUserUseCase struct {
	userRepo repository.UserRepository
	metrics  metrics.UserMetrics
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase
func NewRegisterUserUseCase(
	userRepo repository.UserRepository,
	metrics metrics.UserMetrics,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo: userRepo,
		metrics:  metrics,
	}
}

// Execute performs user registration
func (uc *RegisterUserUseCase) Execute(ctx context.Context, email, password, firstName, lastName, phone string) (*entities.User, error) {
	// Create value objects
	emailVO := valueobjects.NewEmail(email)
	firstNameVO := valueobjects.NewName(firstName)
	lastNameVO := valueobjects.NewName(lastName)
	phoneVO := valueobjects.NewPhone(phone)
	passwordVO := valueobjects.NewPasswordFromPlain(password)

	var user *entities.User

	// Execute all operations within a transaction
	err := uc.userRepo.WithTransaction(ctx, func(txRepo repository.UserRepository) error {
		// Check if user already exists
		exists, err := txRepo.ExistsByEmail(ctx, emailVO)
		if err != nil {
			return errors.ErrDatabaseOperationFailed
		}

		if exists {
			return errors.ErrEmailAlreadyExists
		}

		// Create user entity
		userID := idgenerator.GenerateID("user")
		user = entities.NewUser(userID, emailVO, passwordVO, firstNameVO, lastNameVO, phoneVO)

		// Save user to repository
		if err := txRepo.Create(ctx, user); err != nil {
			return errors.ErrUserCreationFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Record metrics for successful registration
	if uc.metrics != nil {
		uc.metrics.UserCreated()
	}

	return user, nil
}
