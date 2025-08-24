package grpc

import (
	"context"

	userpb "github.com/kubernetestest/ecommerce-platform/proto-go/user"
	"github.com/kubernetestest/ecommerce-platform/services/user-service/internal/app/services"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBUserServer struct {
	userpb.UnimplementedUserServiceServer
	svc *services.UserService
}

func NewPBUserServer(svc *services.UserService) *PBUserServer {
	return &PBUserServer{svc: svc}
}

func (s *PBUserServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	resp, err := s.svc.RegisterUser(ctx, &services.RegisterUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.RegisterResponse{
		User: &userpb.User{
			Id:        resp.Id,
			Email:     resp.Email,
			FirstName: resp.FirstName,
			LastName:  resp.LastName,
			Phone:     resp.Phone,
			CreatedAt: resp.CreatedAt,
			UpdatedAt: resp.UpdatedAt,
		},
		Message: "User registered successfully",
	}, nil
}

func (s *PBUserServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	resp, err := s.svc.LoginUser(ctx, &services.LoginRequest{Email: req.Email, Password: req.Password})
	if err != nil {
		// Record failed login metrics
		s.svc.RecordFailedLogin("invalid_credentials")
		return nil, err
	}

	// Debug logging

	return &userpb.LoginResponse{
		User: &userpb.User{
			Id:        resp.User.Id,
			Email:     resp.User.Email,
			FirstName: resp.User.FirstName,
			LastName:  resp.User.LastName,
			Phone:     resp.User.Phone,
			CreatedAt: resp.User.CreatedAt,
			UpdatedAt: resp.User.UpdatedAt,
		},
		Message:      "Login successful",
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

func (s *PBUserServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	u, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserResponse{
		User: &userpb.User{
			Id:        u.ID(),
			Email:     u.Email().Value(),
			FirstName: u.FirstName(),
			LastName:  u.LastName(),
			Phone:     u.Phone(),
			CreatedAt: timestamppb.New(u.CreatedAt()),
			UpdatedAt: timestamppb.New(u.UpdatedAt()),
		},
	}, nil
}

func (s *PBUserServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	u, err := s.svc.UpdateUser(ctx, req.Id, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateUserResponse{
		User: &userpb.User{
			Id:        u.ID(),
			Email:     u.Email().Value(),
			FirstName: u.FirstName(),
			LastName:  u.LastName(),
			Phone:     u.Phone(),
			CreatedAt: timestamppb.New(u.CreatedAt()),
			UpdatedAt: timestamppb.New(u.UpdatedAt()),
		},
		Message: "Profile updated",
	}, nil
}

func (s *PBUserServer) RefreshToken(ctx context.Context, req *userpb.RefreshTokenRequest) (*userpb.RefreshTokenResponse, error) {
	tokenPair, err := s.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &userpb.RefreshTokenResponse{
		AccessToken: tokenPair.AccessToken,
		ExpiresIn:   tokenPair.ExpiresIn,
	}, nil
}

func (s *PBUserServer) Logout(ctx context.Context, req *userpb.LogoutRequest) (*userpb.LogoutResponse, error) {
	err := s.svc.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &userpb.LogoutResponse{
		Message: "Logout successful",
	}, nil
}
