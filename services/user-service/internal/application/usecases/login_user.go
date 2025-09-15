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
)

// LoginUserUseCase handles user login business logic
type LoginUserUseCase struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	tokenGenerator services.TokenGenerator
}

// NewLoginUserUseCase creates a new LoginUserUseCase
func NewLoginUserUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	tokenGenerator services.TokenGenerator,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		tokenGenerator: tokenGenerator,
	}
}

// Execute performs user login
func (uc *LoginUserUseCase) Execute(ctx context.Context, email, password string) (*entities.User, *entities.Session, error) {
	// Create email value object with validation
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, nil, err
	}

	// Find user by email
	user, err := uc.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		return nil, nil, err
	}

	// Verify password
	if !user.VerifyPassword(password) {
		return nil, nil, errors.ErrInvalidCredentials
	}

	// Create session - all business logic in use case
	session, err := uc.createSession(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, session, nil
}

// createSession creates a new session for the user
func (uc *LoginUserUseCase) createSession(ctx context.Context, user *entities.User) (*entities.Session, error) {
	// Generate token pair (both access and refresh tokens)
	accessToken, refreshToken, err := uc.tokenGenerator.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return session, nil
}
