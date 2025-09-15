package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/user-service/internal/application/usecases"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupRegisterUserTest creates common test setup for RegisterUserUseCase tests
func setupRegisterUserTest(t *testing.T) (*gomock.Controller, *mocks.MockUserRepository, *usecases.RegisterUserUseCase) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	useCase := usecases.NewRegisterUserUseCase(mockUserRepo)
	return ctrl, mockUserRepo, useCase
}

func TestRegisterUserUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, useCase := setupRegisterUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "test@example.com"
		password := "TestPassword123!"
		firstName := "John"
		lastName := "Doe"
		phone := "+1234567890"

		// Setup mocks
		mockUserRepo.EXPECT().
			ExistsByEmail(ctx, gomock.Any()).
			Return(false, nil).
			Times(1)

		mockUserRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *entities.User) error {
				assert.Equal(t, email, user.Email.String())
				assert.Equal(t, firstName, user.FirstName.String())
				assert.Equal(t, lastName, user.LastName.String())
				assert.Equal(t, phone, user.Phone.String())
				assert.NotEmpty(t, user.ID)
				return nil
			}).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, email, password, firstName, lastName, phone)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, email, result.Email.String())
		assert.Equal(t, firstName, result.FirstName.String())
		assert.Equal(t, lastName, result.LastName.String())
		assert.Equal(t, phone, result.Phone.String())
		assert.NotEmpty(t, result.ID)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
	})

	t.Run("email already exists", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, useCase := setupRegisterUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "existing@example.com"
		password := "TestPassword123!"
		firstName := "John"
		lastName := "Doe"
		phone := "+1234567890"

		// Setup mocks
		mockUserRepo.EXPECT().
			ExistsByEmail(ctx, gomock.Any()).
			Return(true, nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, email, password, firstName, lastName, phone)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrEmailAlreadyExists, err)
		assert.Nil(t, result)
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, _, useCase := setupRegisterUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		invalidEmail := "invalid-email"
		password := "TestPassword123!"
		firstName := "John"
		lastName := "Doe"
		phone := "+1234567890"

		// Act
		result, err := useCase.Execute(ctx, invalidEmail, password, firstName, lastName, phone)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid password", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, _, useCase := setupRegisterUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "test@example.com"
		invalidPassword := "123" // Too short
		firstName := "John"
		lastName := "Doe"
		phone := "+1234567890"

		// Act
		result, err := useCase.Execute(ctx, email, invalidPassword, firstName, lastName, phone)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid phone", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, _, useCase := setupRegisterUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "test@example.com"
		password := "TestPassword123!"
		firstName := "John"
		lastName := "Doe"
		invalidPhone := "invalid-phone"

		// Act
		result, err := useCase.Execute(ctx, email, password, firstName, lastName, invalidPhone)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
