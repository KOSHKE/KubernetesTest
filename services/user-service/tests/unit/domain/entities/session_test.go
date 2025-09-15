package entities_test

import (
	"testing"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"

	"github.com/stretchr/testify/assert"
)

func TestSession_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		t.Parallel()
		// Arrange
		futureTime := time.Now().Add(time.Hour)
		session := createTestSession(futureTime)

		// Act
		result := session.IsExpired()

		// Assert
		assert.False(t, result)
	})

	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		// Arrange
		pastTime := time.Now().Add(-time.Hour)
		session := createTestSession(pastTime)

		// Act
		result := session.IsExpired()

		// Assert
		assert.True(t, result)
	})

	t.Run("just expired", func(t *testing.T) {
		t.Parallel()
		// Arrange
		justExpired := time.Now().Add(-time.Second)
		session := createTestSession(justExpired)

		// Act
		result := session.IsExpired()

		// Assert
		assert.True(t, result)
	})
}

func TestSession_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		session := createTestSession(time.Now().Add(time.Hour))
		newAccessToken := valueobjects.NewToken("new-access-token", time.Now().Add(30*time.Minute))
		newRefreshToken := valueobjects.NewToken("new-refresh-token", time.Now().Add(2*time.Hour))
		newExpiresAt := time.Now().Add(2 * time.Hour)
		originalUpdatedAt := session.UpdatedAt

		// Act
		session.Refresh(newAccessToken, newRefreshToken, newExpiresAt)

		// Assert
		assert.Equal(t, newAccessToken, session.AccessToken)
		assert.Equal(t, newRefreshToken, session.RefreshToken)
		assert.Equal(t, newExpiresAt, session.ExpiresAt)
		assert.True(t, session.UpdatedAt.After(originalUpdatedAt) || session.UpdatedAt.Equal(originalUpdatedAt))
	})
}

func TestSession_Validate(t *testing.T) {
	t.Run("valid session", func(t *testing.T) {
		t.Parallel()
		// Arrange
		session := createTestSession(time.Now().Add(time.Hour))

		// Act
		err := session.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("empty session ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		session := &entities.Session{
			ID:           "",
			UserID:       "user-123",
			AccessToken:  valueobjects.NewToken("access-token", time.Now().Add(time.Hour)),
			RefreshToken: valueobjects.NewToken("refresh-token", time.Now().Add(2*time.Hour)),
			ExpiresAt:    time.Now().Add(time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Act
		err := session.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidSessionID, err)
	})

	t.Run("empty user ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		session := &entities.Session{
			ID:           "session-123",
			UserID:       "",
			AccessToken:  valueobjects.NewToken("access-token", time.Now().Add(time.Hour)),
			RefreshToken: valueobjects.NewToken("refresh-token", time.Now().Add(2*time.Hour)),
			ExpiresAt:    time.Now().Add(time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Act
		err := session.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidUserID, err)
	})

	t.Run("empty access token", func(t *testing.T) {
		t.Parallel()
		// Arrange
		session := &entities.Session{
			ID:           "session-123",
			UserID:       "user-123",
			AccessToken:  valueobjects.NewToken("", time.Now().Add(time.Hour)),
			RefreshToken: valueobjects.NewToken("refresh-token", time.Now().Add(2*time.Hour)),
			ExpiresAt:    time.Now().Add(time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Act
		err := session.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidAccessToken, err)
	})

	t.Run("empty refresh token", func(t *testing.T) {
		t.Parallel()
		// Arrange
		session := &entities.Session{
			ID:           "session-123",
			UserID:       "user-123",
			AccessToken:  valueobjects.NewToken("access-token", time.Now().Add(time.Hour)),
			RefreshToken: valueobjects.NewToken("", time.Now().Add(2*time.Hour)),
			ExpiresAt:    time.Now().Add(time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Act
		err := session.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidRefreshToken, err)
	})
}

func TestSession_NewSession(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		id := "session-123"
		userID := "user-123"
		accessToken := valueobjects.NewToken("access-token", time.Now().Add(time.Hour))
		refreshToken := valueobjects.NewToken("refresh-token", time.Now().Add(2*time.Hour))
		expiresAt := time.Now().Add(time.Hour)

		// Act
		session := entities.NewSession(id, userID, accessToken, refreshToken, expiresAt)

		// Assert
		assert.Equal(t, id, session.ID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, accessToken, session.AccessToken)
		assert.Equal(t, refreshToken, session.RefreshToken)
		assert.Equal(t, expiresAt, session.ExpiresAt)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.UpdatedAt.IsZero())
		assert.True(t, session.CreatedAt.Equal(session.UpdatedAt))
	})
}

// Helper functions for creating test data
func createTestSession(expiresAt time.Time) *entities.Session {
	id := "session-123"
	userID := "user-123"
	accessToken := valueobjects.NewToken("access-token", expiresAt)
	refreshToken := valueobjects.NewToken("refresh-token", time.Now().Add(2*time.Hour))

	return entities.NewSession(id, userID, accessToken, refreshToken, expiresAt)
}
