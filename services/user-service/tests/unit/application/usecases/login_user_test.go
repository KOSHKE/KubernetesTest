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

// setupLoginUserTest creates common test setup for LoginUserUseCase tests
func setupLoginUserTest(t *testing.T) (*gomock.Controller, *mocks.MockUserRepository, *mocks.MockSessionRepository, *mocks.MockTokenGenerator, *usecases.LoginUserUseCase) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenGenerator := mocks.NewMockTokenGenerator(ctrl)
	useCase := usecases.NewLoginUserUseCase(mockUserRepo, mockSessionRepo, mockTokenGenerator)
	return ctrl, mockUserRepo, mockSessionRepo, mockTokenGenerator, useCase
}

func TestLoginUserUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, mockSessionRepo, mockTokenGenerator, useCase := setupLoginUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "test@example.com"
		password := "TestPassword123!"

		// Create test user
		emailVO, _ := valueobjects.NewEmail(email)
		passwordVO, _ := valueobjects.NewPassword(password)
		firstNameVO := valueobjects.NewName("John")
		lastNameVO := valueobjects.NewName("Doe")
		phoneVO, _ := valueobjects.NewPhone("+1234567890")
		now := time.Now()

		user := entities.NewUser("user-123", emailVO, passwordVO, firstNameVO, lastNameVO, phoneVO, now, now)

		// Create test tokens
		accessToken := valueobjects.NewToken("access-token", time.Now().Add(time.Hour))
		refreshToken := valueobjects.NewToken("refresh-token", time.Now().Add(24*time.Hour))

		// Setup mocks
		mockUserRepo.EXPECT().
			GetByEmail(ctx, gomock.Any()).
			Return(user, nil).
			Times(1)

		mockTokenGenerator.EXPECT().
			GenerateTokenPair("user-123").
			Return(accessToken, refreshToken, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Save(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, session *entities.Session) error {
				assert.Equal(t, "user-123", session.UserID)
				assert.Equal(t, accessToken, session.AccessToken)
				assert.Equal(t, refreshToken, session.RefreshToken)
				assert.NotEmpty(t, session.ID)
				return nil
			}).
			Times(1)

		// Act
		resultUser, resultSession, err := useCase.Execute(ctx, email, password)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resultUser)
		assert.NotNil(t, resultSession)
		assert.Equal(t, user.ID, resultUser.ID)
		assert.Equal(t, user.ID, resultSession.UserID)
		assert.Equal(t, accessToken, resultSession.AccessToken)
		assert.Equal(t, refreshToken, resultSession.RefreshToken)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, _, _, useCase := setupLoginUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "notfound@example.com"
		password := "TestPassword123!"

		// Setup mocks
		mockUserRepo.EXPECT().
			GetByEmail(ctx, gomock.Any()).
			Return(nil, errors.ErrUserNotFound).
			Times(1)

		// Act
		resultUser, resultSession, err := useCase.Execute(ctx, email, password)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		assert.Nil(t, resultUser)
		assert.Nil(t, resultSession)
	})

	t.Run("invalid password", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, _, _, useCase := setupLoginUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "test@example.com"
		wrongPassword := "wrongpassword"

		// Create test user with different password
		emailVO, _ := valueobjects.NewEmail(email)
		passwordVO, _ := valueobjects.NewPassword("correctpassword")
		firstNameVO := valueobjects.NewName("John")
		lastNameVO := valueobjects.NewName("Doe")
		phoneVO, _ := valueobjects.NewPhone("+1234567890")
		now := time.Now()

		user := entities.NewUser("user-123", emailVO, passwordVO, firstNameVO, lastNameVO, phoneVO, now, now)

		// Setup mocks
		mockUserRepo.EXPECT().
			GetByEmail(ctx, gomock.Any()).
			Return(user, nil).
			Times(1)

		// Act
		resultUser, resultSession, err := useCase.Execute(ctx, email, wrongPassword)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidCredentials, err)
		assert.Nil(t, resultUser)
		assert.Nil(t, resultSession)
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, _, _, _, useCase := setupLoginUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		invalidEmail := "invalid-email"
		password := "TestPassword123!"

		// Act
		resultUser, resultSession, err := useCase.Execute(ctx, invalidEmail, password)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resultUser)
		assert.Nil(t, resultSession)
	})

	t.Run("token generation error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockUserRepo, _, mockTokenGenerator, useCase := setupLoginUserTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		email := "test@example.com"
		password := "TestPassword123!"

		// Create test user
		emailVO, _ := valueobjects.NewEmail(email)
		passwordVO, _ := valueobjects.NewPassword(password)
		firstNameVO := valueobjects.NewName("John")
		lastNameVO := valueobjects.NewName("Doe")
		phoneVO, _ := valueobjects.NewPhone("+1234567890")
		now := time.Now()

		user := entities.NewUser("user-123", emailVO, passwordVO, firstNameVO, lastNameVO, phoneVO, now, now)

		// Setup mocks
		mockUserRepo.EXPECT().
			GetByEmail(ctx, gomock.Any()).
			Return(user, nil).
			Times(1)

		mockTokenGenerator.EXPECT().
			GenerateTokenPair("user-123").
			Return(valueobjects.NewToken("", time.Now()), valueobjects.NewToken("", time.Now()), assert.AnError).
			Times(1)

		// Act
		resultUser, resultSession, err := useCase.Execute(ctx, email, password)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resultUser)
		assert.Nil(t, resultSession)
	})
}
