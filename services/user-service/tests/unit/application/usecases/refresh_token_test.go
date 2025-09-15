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

// setupRefreshTokenTest creates common test setup for RefreshTokenUseCase tests
func setupRefreshTokenTest(t *testing.T) (*gomock.Controller, *mocks.MockSessionRepository, *mocks.MockTokenGenerator, *usecases.RefreshTokenUseCase) {
	ctrl := gomock.NewController(t)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenGenerator := mocks.NewMockTokenGenerator(ctrl)
	useCase := usecases.NewRefreshTokenUseCase(mockSessionRepo, mockTokenGenerator)
	return ctrl, mockSessionRepo, mockTokenGenerator, useCase
}

func TestRefreshTokenUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, mockTokenGenerator, useCase := setupRefreshTokenTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "session-123"
		userID := "user-123"

		// Create test session
		oldAccessToken := valueobjects.NewToken("old-access-token", time.Now().Add(time.Hour))
		oldRefreshToken := valueobjects.NewToken("old-refresh-token", time.Now().Add(24*time.Hour))
		expiresAt := time.Now().Add(24 * time.Hour)

		session := entities.NewSession(sessionID, userID, oldAccessToken, oldRefreshToken, expiresAt)

		// Create new tokens
		newAccessToken := valueobjects.NewToken("new-access-token", time.Now().Add(time.Hour))
		newRefreshToken := valueobjects.NewToken("new-refresh-token", time.Now().Add(24*time.Hour))

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		mockTokenGenerator.EXPECT().
			GenerateTokenPair(userID).
			Return(newAccessToken, newRefreshToken, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Save(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, session *entities.Session) error {
				assert.Equal(t, newAccessToken, session.AccessToken)
				assert.Equal(t, newRefreshToken, session.RefreshToken)
				assert.Equal(t, userID, session.UserID)
				return nil
			}).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newAccessToken, result.AccessToken)
		assert.Equal(t, newRefreshToken, result.RefreshToken)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("session not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, _, useCase := setupRefreshTokenTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "non-existent-session"

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(nil, errors.ErrSessionNotFound).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrSessionNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("session expired", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, _, useCase := setupRefreshTokenTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "session-123"
		userID := "user-123"

		// Create expired session
		oldAccessToken := valueobjects.NewToken("old-access-token", time.Now().Add(time.Hour))
		oldRefreshToken := valueobjects.NewToken("old-refresh-token", time.Now().Add(24*time.Hour))
		expiredAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago

		expiredSession := entities.NewSession(sessionID, userID, oldAccessToken, oldRefreshToken, expiredAt)

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(expiredSession, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrSessionExpired, err)
		assert.Nil(t, result)
	})

	t.Run("token generation error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, mockTokenGenerator, useCase := setupRefreshTokenTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "session-123"
		userID := "user-123"

		// Create test session
		oldAccessToken := valueobjects.NewToken("old-access-token", time.Now().Add(time.Hour))
		oldRefreshToken := valueobjects.NewToken("old-refresh-token", time.Now().Add(24*time.Hour))
		expiresAt := time.Now().Add(24 * time.Hour)

		session := entities.NewSession(sessionID, userID, oldAccessToken, oldRefreshToken, expiresAt)

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		mockTokenGenerator.EXPECT().
			GenerateTokenPair(userID).
			Return(valueobjects.NewToken("", time.Now()), valueobjects.NewToken("", time.Now()), assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("save error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, mockTokenGenerator, useCase := setupRefreshTokenTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "session-123"
		userID := "user-123"

		// Create test session
		oldAccessToken := valueobjects.NewToken("old-access-token", time.Now().Add(time.Hour))
		oldRefreshToken := valueobjects.NewToken("old-refresh-token", time.Now().Add(24*time.Hour))
		expiresAt := time.Now().Add(24 * time.Hour)

		session := entities.NewSession(sessionID, userID, oldAccessToken, oldRefreshToken, expiresAt)

		// Create new tokens
		newAccessToken := valueobjects.NewToken("new-access-token", time.Now().Add(time.Hour))
		newRefreshToken := valueobjects.NewToken("new-refresh-token", time.Now().Add(24*time.Hour))

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		mockTokenGenerator.EXPECT().
			GenerateTokenPair(userID).
			Return(newAccessToken, newRefreshToken, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Save(ctx, gomock.Any()).
			Return(assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
