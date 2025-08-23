package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kubernetestest/ecommerce-platform/proto-go/user"
	"github.com/kubernetestest/ecommerce-platform/services/user-service/internal/domain/entities"
	"github.com/kubernetestest/ecommerce-platform/services/user-service/internal/domain/valueobjects"
	"github.com/kubernetestest/ecommerce-platform/services/user-service/internal/ports/auth"
	"github.com/kubernetestest/ecommerce-platform/services/user-service/internal/ports/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	userRepo    repository.UserRepository
	authService auth.AuthService
}

type RegisterUserRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

type LoginRequest struct {
	Email    string
	Password string
}

func NewUserService(userRepo repository.UserRepository, authService auth.AuthService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req *RegisterUserRequest) (*user.User, error) {
	// Check if user exists
	emailVO, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}
	exists, err := s.userRepo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Create user
	passwordVO, err := valueobjects.NewPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}
	// Generate a simple ID (can be replaced with UUID)
	id := "user-" + time.Now().Format("20060102150405")
	userEntity := entities.NewUser(id, emailVO, passwordVO, req.FirstName, req.LastName, req.Phone)

	// Save to database
	if err := s.userRepo.Create(ctx, userEntity); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert domain user to protobuf user
	pbUser := &user.User{
		Id:        userEntity.ID(),
		Email:     userEntity.Email().Value(),
		FirstName: userEntity.FirstName(),
		LastName:  userEntity.LastName(),
		Phone:     userEntity.Phone(),
		CreatedAt: timestamppb.New(userEntity.CreatedAt()),
		UpdatedAt: timestamppb.New(userEntity.UpdatedAt()),
	}

	return pbUser, nil
}

func (s *UserService) LoginUser(ctx context.Context, req *LoginRequest) (*user.LoginResponse, error) {
	// Find user by email
	emailVO, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}
	userEntity, err := s.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if !userEntity.ValidatePassword(req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT tokens
	tokenPair, err := s.authService.GenerateTokenPair(userEntity.ID(), userEntity.Email().Value())
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token in Redis
	if err := s.authService.StoreRefreshToken(ctx, tokenPair.RefreshToken, userEntity.ID()); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Convert domain user to protobuf user
	pbUser := &user.User{
		Id:        userEntity.ID(),
		Email:     userEntity.Email().Value(),
		FirstName: userEntity.FirstName(),
		LastName:  userEntity.LastName(),
		Phone:     userEntity.Phone(),
		CreatedAt: timestamppb.New(userEntity.CreatedAt()),
		UpdatedAt: timestamppb.New(userEntity.UpdatedAt()),
	}

	return &user.LoginResponse{
		User:         pbUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*entities.User, error) {
	userEntity, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return userEntity, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, firstName, lastName, phone string) (*entities.User, error) {
	userEntity, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Use Domain method
	userEntity.UpdateProfile(firstName, lastName, phone)

	if err := s.userRepo.Update(ctx, userEntity); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return userEntity, nil
}

// RefreshToken generates new access token using refresh token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return s.authService.RefreshAccessToken(refreshToken)
}

// Logout revokes refresh token
func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	return s.authService.RevokeRefreshToken(ctx, refreshToken)
}
