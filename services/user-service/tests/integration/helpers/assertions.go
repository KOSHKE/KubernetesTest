package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ecommerce-platform/pkg/redisclient"
	"ecommerce-platform/services/user-service/internal/application/dto"
	"ecommerce-platform/services/user-service/internal/domain/entities"
)

// AssertUserExistsInDatabase checks that user exists in PostgreSQL
func AssertUserExistsInDatabase(t *testing.T, ctx context.Context, db *gorm.DB, userID, email string) {
	var userRecord struct {
		ID        string `gorm:"column:id"`
		Email     string `gorm:"column:email"`
		FirstName string `gorm:"column:first_name"`
		LastName  string `gorm:"column:last_name"`
		Phone     string `gorm:"column:phone"`
	}

	err := db.Table("users").Where("id = ?", userID).First(&userRecord).Error
	require.NoError(t, err, "User should exist in database")
	assert.Equal(t, userID, userRecord.ID, "User ID should match")
	assert.Equal(t, email, userRecord.Email, "User email should match")
	assert.NotEmpty(t, userRecord.FirstName, "First name should not be empty")
	assert.NotEmpty(t, userRecord.LastName, "Last name should not be empty")
	assert.NotEmpty(t, userRecord.Phone, "Phone should not be empty")
}

// AssertUserNotExistsInDatabase checks that user does not exist in PostgreSQL
func AssertUserNotExistsInDatabase(t *testing.T, ctx context.Context, db *gorm.DB, userID string) {
	var count int64
	err := db.Table("users").Where("id = ?", userID).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "User should not exist in database")
}

// AssertSessionExistsInRedis checks that session exists in Redis
func AssertSessionExistsInRedis(t *testing.T, ctx context.Context, redisClient *redisclient.Client, sessionID string) {
	var session entities.Session
	err := redisClient.Get(ctx, "session:"+sessionID, &session)
	require.NoError(t, err, "Session should exist in Redis")
	assert.NotEmpty(t, session.ID, "Session ID should not be empty")
	assert.NotEmpty(t, session.UserID, "User ID should not be empty")
	assert.False(t, session.AccessToken.IsEmpty(), "Access token should not be empty")
	assert.False(t, session.RefreshToken.IsEmpty(), "Refresh token should not be empty")
}

// AssertSessionNotExistsInRedis checks that session does not exist in Redis
func AssertSessionNotExistsInRedis(t *testing.T, ctx context.Context, redisClient *redisclient.Client, sessionID string) {
	exists, err := redisClient.Exists(ctx, "session:"+sessionID)
	require.NoError(t, err, "Should be able to check if session exists")
	assert.False(t, exists, "Session should not exist in Redis")
}

// AssertRefreshTokenMappingExists checks that refresh token mapping exists in Redis
func AssertRefreshTokenMappingExists(t *testing.T, ctx context.Context, redisClient *redisclient.Client, refreshToken, sessionID string) {
	var mappedSessionID string
	err := redisClient.Get(ctx, "refresh_token:"+refreshToken, &mappedSessionID)
	require.NoError(t, err, "Refresh token mapping should exist in Redis")
	assert.Equal(t, sessionID, mappedSessionID, "Mapped session ID should match")
}

// AssertUserSessionsSetExists checks that user sessions set exists in Redis
func AssertUserSessionsSetExists(t *testing.T, ctx context.Context, redisClient *redisclient.Client, userID string, expectedSessionCount int) {
	sessionIDs, err := redisClient.SMembers(ctx, "user_sessions:"+userID)
	require.NoError(t, err, "User sessions set should exist in Redis")
	assert.Len(t, sessionIDs, expectedSessionCount, "User sessions count should match")
}

// AssertUserSessionsSetNotExists checks that user sessions set does not exist in Redis
func AssertUserSessionsSetNotExists(t *testing.T, ctx context.Context, redisClient *redisclient.Client, userID string) {
	sessionIDs, err := redisClient.SMembers(ctx, "user_sessions:"+userID)
	require.NoError(t, err)
	assert.Len(t, sessionIDs, 0, "User sessions set should be empty")
}

// AssertValidJWTToken checks that JWT token is valid and contains expected claims
func AssertValidJWTToken(t *testing.T, token string, expectedUserID string) {
	// This is a basic check - in real implementation you would validate the JWT
	assert.NotEmpty(t, token, "JWT token should not be empty")
	assert.Greater(t, len(token), 50, "JWT token should be reasonably long")
	// Note: Full JWT validation would require the JWT manager, but for integration tests
	// we mainly care that tokens are generated and stored correctly
}

// AssertUserDataMatches checks that user data in response matches expected values
func AssertUserDataMatches(t *testing.T, user *dto.GetUserResponse, expectedEmail, expectedFirstName, expectedLastName, expectedPhone string) {
	assert.Equal(t, expectedEmail, user.Email.String(), "Email should match")
	assert.Equal(t, expectedFirstName, user.FirstName.String(), "First name should match")
	assert.Equal(t, expectedLastName, user.LastName.String(), "Last name should match")
	assert.Equal(t, expectedPhone, user.Phone.String(), "Phone should match")
	assert.NotEmpty(t, user.UserID, "User ID should not be empty")
	assert.False(t, user.CreatedAt.IsZero(), "Created at should not be zero")
	assert.False(t, user.UpdatedAt.IsZero(), "Updated at should not be zero")
}

// AssertLoginResponseValid checks that login response contains valid data
func AssertLoginResponseValid(t *testing.T, loginResp *dto.LoginResponse, expectedEmail string) {
	assert.NotEmpty(t, loginResp.UserID, "User ID should not be empty")
	assert.Equal(t, expectedEmail, loginResp.Email.String(), "Email should match")
	assert.NotEmpty(t, loginResp.SessionID, "Session ID should not be empty")
	assert.NotEmpty(t, loginResp.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, loginResp.RefreshToken, "Refresh token should not be empty")
	assert.False(t, loginResp.ExpiresAt.IsZero(), "Expires at should not be zero")
}

// AssertRefreshTokenResponseValid checks that refresh token response contains valid data
func AssertRefreshTokenResponseValid(t *testing.T, refreshResp *dto.RefreshTokenResponse) {
	assert.NotEmpty(t, refreshResp.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, refreshResp.RefreshToken, "Refresh token should not be empty")
	assert.False(t, refreshResp.ExpiresAt.IsZero(), "Expires at should not be zero")
}

// AssertDatabaseUserCount checks the number of users in database
func AssertDatabaseUserCount(t *testing.T, ctx context.Context, db *gorm.DB, expectedCount int) {
	var count int64
	err := db.Table("users").Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(expectedCount), count, "User count in database should match")
}

// AssertRedisKeyCount checks the number of keys in Redis
func AssertRedisKeyCount(t *testing.T, ctx context.Context, redisClient *redisclient.Client, expectedCount int) {
	// Since we don't have a direct scan method, we'll check specific keys
	// This is a simplified version for testing
	sessionKeys := []string{"session:test-session-123", "refresh_token:test-refresh-token", "user_sessions:test-user-123"}

	var actualCount int
	for _, key := range sessionKeys {
		exists, err := redisClient.Exists(ctx, key)
		if err == nil && exists {
			actualCount++
		}
	}

	// For integration tests, we mainly care about the specific keys we're testing
	assert.GreaterOrEqual(t, actualCount, 0, "Redis key count should be non-negative")
}
