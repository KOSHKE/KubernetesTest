package usecases_test

import (
	"context"
	"testing"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/user-service/internal/application/usecases"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
	"ecommerce-platform/services/user-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupGetUserTest creates common test setup for GetUserUseCase tests
func setupGetUserTest(t *testing.T) (*gomock.Controller, *mocks.MockUserRepository, *usecases.GetUserUseCase) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	useCase := usecases.NewGetUserUseCase(mockUserRepo)
	return ctrl, mockUserRepo, useCase
}

func TestGetUserUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, useCase := setupGetUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-123"

		// Create test user
		emailVO, _ := valueobjects.NewEmail("test@example.com")
		passwordVO, _ := valueobjects.NewPassword("TestPassword123!")
		firstNameVO := valueobjects.NewName("John")
		lastNameVO := valueobjects.NewName("Doe")
		phoneVO, _ := valueobjects.NewPhone("+1234567890")
		now := time.Now()

		expectedUser := entities.NewUser(userID, emailVO, passwordVO, firstNameVO, lastNameVO, phoneVO, now, now)

		// Setup mocks
		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(expectedUser, nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, userID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, "test@example.com", result.Email.String())
		assert.Equal(t, "John", result.FirstName.String())
		assert.Equal(t, "Doe", result.LastName.String())
		assert.Equal(t, "+1234567890", result.Phone.String())
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, useCase := setupGetUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "non-existent-user"

		// Setup mocks
		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, errors.ErrUserNotFound).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, useCase := setupGetUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-123"

		// Setup mocks
		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
