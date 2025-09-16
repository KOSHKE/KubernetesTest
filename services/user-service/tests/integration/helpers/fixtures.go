package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ecommerce-platform/pkg/jwt"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/redisclient"
	"ecommerce-platform/services/user-service/internal/application/dto"
	"ecommerce-platform/services/user-service/internal/application/services"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
	"ecommerce-platform/services/user-service/internal/infra/repository"
	infraServices "ecommerce-platform/services/user-service/internal/infra/services"
)

// CreateTestUserApplicationService creates a test application service with real dependencies
func CreateTestUserApplicationService(t *testing.T, ctx context.Context, db *gorm.DB, redisClient *redisclient.Client) *services.UserApplicationService {
	// Create repositories
	userRepo := repository.NewGormUserRepository(db)
	sessionRepo := repository.NewRedisSessionRepository(redisClient, 24*time.Hour)

	// Create JWT manager for testing
	jwtConfig := jwt.Config{
		AccessTokenSecret:  "test-access-secret-key-very-long-for-testing",
		RefreshTokenSecret: "test-refresh-secret-key-very-long-for-testing",
		AccessTokenTTL:     1 * time.Hour,
		RefreshTokenTTL:    24 * time.Hour,
		Issuer:             "user-service-test",
		Audience:           "ecommerce-platform-test",
	}
	jwtManager := jwt.NewManager(jwtConfig)
	tokenGenerator := infraServices.NewJWTTokenGenerator(jwtManager)

	// Create logger
	log := logger.NewZapLogger(zap.NewNop().Sugar())

	// Create application service
	appService := services.NewUserApplicationService(userRepo, sessionRepo, tokenGenerator, log)

	return appService
}

// CreateTestUser creates a test user via application service
func CreateTestUser(t *testing.T, ctx context.Context, appService *services.UserApplicationService) *dto.RegisterUserResponse {
	req := &dto.RegisterUserRequest{
		Email:     "test@example.com",
		Password:  "TestPassword123!",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
	}

	user, err := appService.RegisterUser(ctx, req)
	require.NoError(t, err)
	return user
}

// LoginTestUser logs in a test user and returns session data
func LoginTestUser(t *testing.T, ctx context.Context, appService *services.UserApplicationService, email, password string) *dto.LoginResponse {
	req := &dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	loginResp, err := appService.LoginUser(ctx, req)
	require.NoError(t, err)
	return loginResp
}

// CreateTestUserWithLogin creates a user and logs them in
func CreateTestUserWithLogin(t *testing.T, ctx context.Context, appService *services.UserApplicationService) (*dto.RegisterUserResponse, *dto.LoginResponse) {
	user := CreateTestUser(t, ctx, appService)
	loginResp := LoginTestUser(t, ctx, appService, user.Email, "TestPassword123!")
	return user, loginResp
}

// CreateTestUserEntity creates a test user entity directly (for database testing)
func CreateTestUserEntity(t *testing.T, ctx context.Context) *entities.User {
	email, err := valueobjects.NewEmail("test@example.com")
	require.NoError(t, err)

	password, err := valueobjects.NewPassword("TestPassword123!")
	require.NoError(t, err)

	firstName := valueobjects.NewName("John")
	lastName := valueobjects.NewName("Doe")

	phone, err := valueobjects.NewPhone("+1234567890")
	require.NoError(t, err)

	now := time.Now()
	user := entities.NewUser("test-user-123", email, password, firstName, lastName, phone, now, now)

	return user
}

// CreateTestSession creates a test session entity directly (for Redis testing)
func CreateTestSession(t *testing.T, ctx context.Context, userID string) *entities.Session {
	accessToken := valueobjects.NewToken("test-access-token", time.Now().Add(time.Hour))
	refreshToken := valueobjects.NewToken("test-refresh-token", time.Now().Add(24*time.Hour))
	expiresAt := time.Now().Add(24 * time.Hour)

	session := entities.NewSession("test-session-123", userID, accessToken, refreshToken, expiresAt)

	return session
}
