package usecases

import (
	"context"
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/idgenerator"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/ports/services"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
	"ecommerce-platform/services/user-service/internal/metrics"
)

// LoginUserUseCase handles user login business logic
type LoginUserUseCase struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	tokenGenerator services.TokenGenerator
	metrics        metrics.UserMetrics
}

// NewLoginUserUseCase creates a new LoginUserUseCase
func NewLoginUserUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	tokenGenerator services.TokenGenerator,
	metrics metrics.UserMetrics,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		tokenGenerator: tokenGenerator,
		metrics:        metrics,
	}
}

// Execute performs user login
func (uc *LoginUserUseCase) Execute(ctx context.Context, email, password string) (*entities.User, *entities.Session, error) {
	// Create email value object
	emailVO := valueobjects.NewEmail(email)

	// Find user by email
	user, err := uc.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		return nil, nil, errors.ErrUserNotFound
	}

	// Verify password
	if !user.VerifyPassword(password) {
		if uc.metrics != nil {
			uc.metrics.UserLoginFailed("invalid_password")
		}
		return nil, nil, errors.ErrInvalidCredentials
	}

	// Create session - вся бизнес-логика в use case
	session, err := uc.createSession(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	// Record metrics for successful login
	if uc.metrics != nil {
		uc.metrics.UserLoginSuccess()
	}

	return user, session, nil
}

// createSession creates a new session for the user
func (uc *LoginUserUseCase) createSession(ctx context.Context, user *entities.User) (*entities.Session, error) {
	// Generate token pair (both access and refresh tokens)
	accessToken, refreshToken, err := uc.tokenGenerator.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, errors.ErrTokenGenerationFailed
	}

	// Create session
	sessionID := idgenerator.GenerateID("session")
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours session

	session := entities.NewSession(
		sessionID,
		user.ID,
		accessToken,
		refreshToken,
		expiresAt,
	)

	// Validate session
	if err := session.Validate(); err != nil {
		return nil, err
	}

	// Save session
	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, errors.ErrSessionCreationFailed
	}

	return session, nil
}
