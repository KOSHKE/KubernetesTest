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

// setupLogoutTest creates common test setup for LogoutUseCase tests
func setupLogoutTest(t *testing.T) (*gomock.Controller, *mocks.MockSessionRepository, *usecases.LogoutUseCase) {
	ctrl := gomock.NewController(t)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	useCase := usecases.NewLogoutUseCase(mockSessionRepo)
	return ctrl, mockSessionRepo, useCase
}

func TestLogoutUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, useCase := setupLogoutTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "session-123"

		// Create test session
		accessToken := valueobjects.NewToken("access-token", time.Now().Add(time.Hour))
		refreshToken := valueobjects.NewToken("refresh-token", time.Now().Add(24*time.Hour))
		expiresAt := time.Now().Add(24 * time.Hour)

		session := entities.NewSession(sessionID, "user-123", accessToken, refreshToken, expiresAt)

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(nil).
			Times(1)

		// Act
		err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("session not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, useCase := setupLogoutTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "non-existent-session"

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(nil, errors.ErrSessionNotFound).
			Times(1)

		// Act
		err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrSessionNotFound, err)
	})

	t.Run("delete error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockSessionRepo, useCase := setupLogoutTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		sessionID := "session-123"

		// Create test session
		accessToken := valueobjects.NewToken("access-token", time.Now().Add(time.Hour))
		refreshToken := valueobjects.NewToken("refresh-token", time.Now().Add(24*time.Hour))
		expiresAt := time.Now().Add(24 * time.Hour)

		session := entities.NewSession(sessionID, "user-123", accessToken, refreshToken, expiresAt)

		// Setup mocks
		mockSessionRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(assert.AnError).
			Times(1)

		// Act
		err := useCase.Execute(ctx, sessionID)

		// Assert
		assert.Error(t, err)
	})
}
