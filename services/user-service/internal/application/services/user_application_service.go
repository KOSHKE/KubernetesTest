package services

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/validation"
	"ecommerce-platform/services/user-service/internal/application/dto"
	"ecommerce-platform/services/user-service/internal/application/usecases"
	authPorts "ecommerce-platform/services/user-service/internal/domain/ports/auth"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/metrics"
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
	authService authPorts.AuthService,
	logger logger.Logger,
	metrics metrics.UserMetrics,
) *UserApplicationService {
	v := validation.New()

	return &UserApplicationService{
		registerUserUseCase: usecases.NewRegisterUserUseCase(userRepo, logger, metrics),
		loginUserUseCase:    usecases.NewLoginUserUseCase(userRepo, authService, logger, metrics),
		getUserUseCase:      usecases.NewGetUserUseCase(userRepo, logger),
		refreshTokenUseCase: usecases.NewRefreshTokenUseCase(authService, logger),
		logoutUseCase:       usecases.NewLogoutUseCase(authService, logger),
		validator:           v,
		logger:              logger,
	}
}

// RegisterUser registers a new user
func (s *UserApplicationService) RegisterUser(ctx context.Context, req *dto.RegisterUserRequest) (*dto.RegisterUserResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrInvalidUserData
	}

	// Convert DTO to domain parameters
	user, err := s.registerUserUseCase.Execute(ctx, req.Email, req.Password, req.FirstName, req.LastName, req.Phone)
	if err != nil {
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
		return nil, errors.ErrInvalidUserData
	}

	// Convert DTO to domain parameters
	user, tokenPair, err := s.loginUserUseCase.Execute(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	// Convert domain objects to DTO response
	return &dto.LoginResponse{
		UserID:       user.ID,
		Email:        user.Email.Value(),
		FirstName:    user.FirstName.Value(),
		LastName:     user.LastName.Value(),
		Phone:        user.Phone.Value(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

// GetUser retrieves user information
func (s *UserApplicationService) GetUser(ctx context.Context, req *dto.GetUserRequest) (*dto.GetUserResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, errors.ErrInvalidUserData
	}

	// Convert DTO to domain parameters
	user, err := s.getUserUseCase.Execute(ctx, req.UserID)
	if err != nil {
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
		return nil, errors.ErrInvalidUserData
	}

	// Convert DTO to domain parameters
	tokenPair, err := s.refreshTokenUseCase.Execute(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Convert domain object to DTO response
	return &dto.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

// Logout logs out a user by revoking their refresh token
func (s *UserApplicationService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return errors.ErrInvalidUserData
	}

	// Convert DTO to domain parameters
	err := s.logoutUseCase.Execute(ctx, req.RefreshToken)
	if err != nil {
		return err
	}

	return nil
}
