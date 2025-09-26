package services

import (
	"context"

	dto "ecommerce-platform/pkg/common/dto/user-service"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/validation"
	"ecommerce-platform/services/user-service/internal/application/usecases"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/ports/services"
)

// UserApplicationService provides the main interface for user operations
type UserApplicationService struct {
	registerUserUseCase *usecases.RegisterUserUseCase
	loginUserUseCase    *usecases.LoginUserUseCase
	getUserUseCase      *usecases.GetUserUseCase
	refreshTokenUseCase *usecases.RefreshTokenUseCase
	logoutUseCase       *usecases.LogoutUseCase
	validator           *validation.Validate
	logger              logger.Logger
}

// NewUserApplicationService creates a new UserApplicationService instance
func NewUserApplicationService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	tokenGenerator services.TokenGenerator,
	logger logger.Logger,
) *UserApplicationService {
	v := validation.New()

	return &UserApplicationService{
		registerUserUseCase: usecases.NewRegisterUserUseCase(userRepo),
		loginUserUseCase:    usecases.NewLoginUserUseCase(userRepo, sessionRepo, tokenGenerator),
		getUserUseCase:      usecases.NewGetUserUseCase(userRepo),
		refreshTokenUseCase: usecases.NewRefreshTokenUseCase(sessionRepo, tokenGenerator),
		logoutUseCase:       usecases.NewLogoutUseCase(sessionRepo),
		validator:           v,
		logger:              logger,
	}
}

// RegisterUser registers a new user
func (s *UserApplicationService) RegisterUser(ctx context.Context, req *dto.RegisterUserRequest) (*dto.RegisterUserResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain parameters
	user, err := s.registerUserUseCase.Execute(ctx, req.Email, req.Password, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		s.logger.Error("failed to register user", "error", err)
		return nil, err
	}

	// Convert domain entity to DTO response
	return &dto.RegisterUserResponse{
		UserID:    user.ID,
		Email:     user.Email.Value(),
		FirstName: user.FirstName.Value(),
		LastName:  user.LastName.Value(),
		Phone:     user.Phone.Value(),
		CreatedAt: user.CreatedAt,
	}, nil
}

// LoginUser authenticates a user and returns tokens
func (s *UserApplicationService) LoginUser(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain parameters
	user, session, err := s.loginUserUseCase.Execute(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.Error("failed to login user", "email", req.Email, "error", err)
		return nil, err
	}

	// Convert domain objects to DTO response
	return &dto.LoginResponse{
		UserID:       user.ID,
		Email:        user.Email.Value(),
		FirstName:    user.FirstName.Value(),
		LastName:     user.LastName.Value(),
		Phone:        user.Phone.Value(),
		SessionID:    session.ID,
		AccessToken:  session.AccessToken.Value,
		RefreshToken: session.RefreshToken.Value,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

// GetUser retrieves user information
func (s *UserApplicationService) GetUser(ctx context.Context, req *dto.GetUserRequest) (*dto.GetUserResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain parameters
	user, err := s.getUserUseCase.Execute(ctx, req.UserID)
	if err != nil {
		s.logger.Error("failed to get user", "userID", req.UserID, "error", err)
		return nil, err
	}

	// Convert domain entity to DTO response
	return &dto.GetUserResponse{
		UserID:    user.ID,
		Email:     user.Email.Value(),
		FirstName: user.FirstName.Value(),
		LastName:  user.LastName.Value(),
		Phone:     user.Phone.Value(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// RefreshToken refreshes user authentication tokens
func (s *UserApplicationService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain parameters
	session, err := s.refreshTokenUseCase.Execute(ctx, req.SessionID)
	if err != nil {
		s.logger.Error("failed to refresh token", "error", err)
		return nil, err
	}

	// Convert domain object to DTO response
	return &dto.RefreshTokenResponse{
		AccessToken:  session.AccessToken.Value,
		RefreshToken: session.RefreshToken.Value,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

// Logout logs out a user by revoking their session
func (s *UserApplicationService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return err
	}

	// Convert DTO to domain parameters
	err := s.logoutUseCase.Execute(ctx, req.SessionID)
	if err != nil {
		s.logger.Error("failed to logout user", "error", err)
		return err
	}

	return nil
}
