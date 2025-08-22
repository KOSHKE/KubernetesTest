package grpc

import (
	"context"
	userpb "proto-go/user"
	"user-service/internal/app/services"

	"log"

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
		return nil, err
	}

	// Debug logging
	log.Printf("DEBUG: gRPC server: Service response - User: %+v", resp.User)
	log.Printf("DEBUG: gRPC server: Service response - AccessToken: %s", resp.AccessToken)
	log.Printf("DEBUG: gRPC server: Service response - RefreshToken: %s", resp.RefreshToken)
	log.Printf("DEBUG: gRPC server: Service response - ExpiresIn: %d", resp.ExpiresIn)

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
	accessToken, err := s.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &userpb.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   900, // 15 minutes in seconds
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
