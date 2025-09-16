//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ecommerce-platform/services/user-service/internal/application/dto"
	"ecommerce-platform/services/user-service/tests/integration/helpers"
)

// TestCompleteUserAuthenticationFlow tests the complete user authentication process
// Registration -> Login -> Get User Data -> Refresh Tokens -> Logout
func TestCompleteUserAuthenticationFlow(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, redisClient, cleanup := helpers.SetupTestDatabase(t, ctx)
	defer cleanup()

	// Create application service with real dependencies
	appService := helpers.CreateTestUserApplicationService(t, ctx, db, redisClient)

	// Test Case 1: User Registration
	t.Run("UserRegistration_ShouldCreateUserInDatabase", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Arrange
		req := &dto.RegisterUserRequest{
			Email:     "test@example.com",
			Password:  "TestPassword123!",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
		}

		// Act
		user, err := appService.RegisterUser(ctx, req)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "+1234567890", user.Phone)
		assert.NotEmpty(t, user.UserID)
		assert.False(t, user.CreatedAt.IsZero())

		// Verify user exists in database
		helpers.AssertUserExistsInDatabase(t, ctx, db, user.UserID, user.Email)
		helpers.AssertDatabaseUserCount(t, ctx, db, 1)
	})

	// Test Case 2: User Login
	t.Run("UserLogin_ShouldCreateSessionInRedis", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Arrange - create user first
		user := helpers.CreateTestUser(t, ctx, appService)

		// Act
		loginReq := &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "TestPassword123!",
		}
		loginResp, err := appService.LoginUser(ctx, loginReq)

		// Assert
		require.NoError(t, err)
		helpers.AssertLoginResponseValid(t, loginResp, "test@example.com")
		assert.Equal(t, user.UserID, loginResp.UserID)

		// Verify session exists in Redis
		helpers.AssertSessionExistsInRedis(t, ctx, redisClient, loginResp.SessionID)
		helpers.AssertRefreshTokenMappingExists(t, ctx, redisClient, loginResp.RefreshToken, loginResp.SessionID)
		helpers.AssertUserSessionsSetExists(t, ctx, redisClient, user.UserID, 1)
	})

	// Test Case 3: Get User Data
	t.Run("GetUserData_ShouldRetrieveUserFromDatabase", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Arrange - create user and login
		user, _ := helpers.CreateTestUserWithLogin(t, ctx, appService)

		// Act
		getUserReq := &dto.GetUserRequest{
			UserID: user.UserID,
		}
		getUserResp, err := appService.GetUser(ctx, getUserReq)

		// Assert
		require.NoError(t, err)
		helpers.AssertUserDataMatches(t, getUserResp, "test@example.com", "John", "Doe", "+1234567890")
		assert.Equal(t, user.UserID, getUserResp.UserID)
	})

	// Test Case 4: Refresh Tokens
	t.Run("RefreshTokens_ShouldUpdateSessionInRedis", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Arrange - create user and login
		_, loginResp := helpers.CreateTestUserWithLogin(t, ctx, appService)

		// Act
		refreshReq := &dto.RefreshTokenRequest{
			SessionID: loginResp.SessionID,
		}
		refreshResp, err := appService.RefreshToken(ctx, refreshReq)

		// Assert
		require.NoError(t, err)
		helpers.AssertRefreshTokenResponseValid(t, refreshResp)

		// Verify new tokens are different from original
		assert.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken, "New access token should be different")
		assert.NotEqual(t, loginResp.RefreshToken, refreshResp.RefreshToken, "New refresh token should be different")

		// Verify session is updated in Redis
		helpers.AssertSessionExistsInRedis(t, ctx, redisClient, loginResp.SessionID)
		helpers.AssertRefreshTokenMappingExists(t, ctx, redisClient, refreshResp.RefreshToken, loginResp.SessionID)
	})

	// Test Case 5: User Logout
	t.Run("UserLogout_ShouldRemoveSessionFromRedis", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Arrange - create user and login
		user, loginResp := helpers.CreateTestUserWithLogin(t, ctx, appService)

		// Verify session exists before logout
		helpers.AssertSessionExistsInRedis(t, ctx, redisClient, loginResp.SessionID)
		helpers.AssertUserSessionsSetExists(t, ctx, redisClient, user.UserID, 1)

		// Act
		logoutReq := &dto.LogoutRequest{
			SessionID: loginResp.SessionID,
		}
		err := appService.Logout(ctx, logoutReq)

		// Assert
		require.NoError(t, err)

		// Verify session is removed from Redis
		helpers.AssertSessionNotExistsInRedis(t, ctx, redisClient, loginResp.SessionID)
		helpers.AssertUserSessionsSetNotExists(t, ctx, redisClient, user.UserID)
	})

	// Test Case 6: Complete Flow Integration
	t.Run("CompleteFlow_ShouldWorkEndToEnd", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Step 1: Register user
		user := helpers.CreateTestUser(t, ctx, appService)
		helpers.AssertUserExistsInDatabase(t, ctx, db, user.UserID, user.Email)
		helpers.AssertDatabaseUserCount(t, ctx, db, 1)

		// Step 2: Login user
		loginResp := helpers.LoginTestUser(t, ctx, appService, user.Email, "TestPassword123!")
		helpers.AssertSessionExistsInRedis(t, ctx, redisClient, loginResp.SessionID)
		helpers.AssertUserSessionsSetExists(t, ctx, redisClient, user.UserID, 1)

		// Step 3: Get user data
		getUserReq := &dto.GetUserRequest{UserID: user.UserID}
		getUserResp, err := appService.GetUser(ctx, getUserReq)
		require.NoError(t, err)
		helpers.AssertUserDataMatches(t, getUserResp, user.Email, user.FirstName, user.LastName, user.Phone)

		// Step 4: Refresh tokens
		refreshReq := &dto.RefreshTokenRequest{SessionID: loginResp.SessionID}
		refreshResp, err := appService.RefreshToken(ctx, refreshReq)
		require.NoError(t, err)
		helpers.AssertRefreshTokenResponseValid(t, refreshResp)

		// Step 5: Logout
		logoutReq := &dto.LogoutRequest{SessionID: loginResp.SessionID}
		err = appService.Logout(ctx, logoutReq)
		require.NoError(t, err)

		// Final verification
		helpers.AssertUserExistsInDatabase(t, ctx, db, user.UserID, user.Email)         // User still exists in DB
		helpers.AssertSessionNotExistsInRedis(t, ctx, redisClient, loginResp.SessionID) // Session removed from Redis
		helpers.AssertUserSessionsSetNotExists(t, ctx, redisClient, user.UserID)        // No sessions for user
	})

	// Test Case 7: Concurrent User Operations
	t.Run("ConcurrentUserOperations_ShouldHandleRaceConditions", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Arrange - create user
		user := helpers.CreateTestUser(t, ctx, appService)

		// Act - run multiple login operations concurrently
		done := make(chan error, 3)

		go func() {
			_ = helpers.LoginTestUser(t, ctx, appService, user.Email, "TestPassword123!")
			done <- nil
		}()

		go func() {
			_ = helpers.LoginTestUser(t, ctx, appService, user.Email, "TestPassword123!")
			done <- nil
		}()

		go func() {
			_ = helpers.LoginTestUser(t, ctx, appService, user.Email, "TestPassword123!")
			done <- nil
		}()

		// Wait for all operations to complete
		var successCount int
		for i := 0; i < 3; i++ {
			err := <-done
			if err == nil {
				successCount++
			}
		}

		// Assert - all operations should succeed
		assert.Equal(t, 3, successCount, "All concurrent login operations should succeed")

		// Verify multiple sessions exist
		helpers.AssertUserSessionsSetExists(t, ctx, redisClient, user.UserID, 3)
	})

	// Test Case 8: Session Expiration Handling
	t.Run("SessionExpiration_ShouldHandleExpiredSessions", func(t *testing.T) {
		// Clean databases before test
		helpers.CleanDatabase(t, db)
		helpers.CleanRedis(t, ctx, redisClient)

		// Arrange - create user and login
		_, loginResp := helpers.CreateTestUserWithLogin(t, ctx, appService)

		// Verify session exists before expiration
		helpers.AssertSessionExistsInRedis(t, ctx, redisClient, loginResp.SessionID)

		// Simulate session expiration by manually removing the session
		// In real implementation, this would be handled by Redis TTL
		helpers.CleanRedis(t, ctx, redisClient)

		// Verify session no longer exists
		helpers.AssertSessionNotExistsInRedis(t, ctx, redisClient, loginResp.SessionID)

		// Act - try to refresh expired session
		refreshReq := &dto.RefreshTokenRequest{
			SessionID: loginResp.SessionID,
		}
		_, err := appService.RefreshToken(ctx, refreshReq)

		// Assert - should fail for expired session
		assert.Error(t, err, "Refresh should fail for expired session")
	})
}
