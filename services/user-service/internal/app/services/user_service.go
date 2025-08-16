package services

import (
	"context"
	"fmt"
	"time"

	"user-service/internal/domain/entities"
	"user-service/internal/domain/valueobjects"
	"user-service/internal/ports/repository"
)

type UserService struct {
	userRepo repository.UserRepository
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

type AuthResponse struct {
	User  *entities.User
	Token string
}

func NewUserService(userRepo repository.UserRepository, _ interface{}) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) RegisterUser(ctx context.Context, req *RegisterUserRequest) (*AuthResponse, error) {
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
	user := entities.NewUser(id, emailVO, passwordVO, req.FirstName, req.LastName, req.Phone)

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &AuthResponse{User: user, Token: ""}, nil
}

func (s *UserService) LoginUser(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Find user by email
	emailVO, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	user, err := s.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if !user.ValidatePassword(req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return &AuthResponse{User: user, Token: ""}, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, firstName, lastName, phone string) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Use Domain method
	user.UpdateProfile(firstName, lastName, phone)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
