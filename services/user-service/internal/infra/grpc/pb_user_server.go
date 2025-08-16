package grpc

import (
	"context"
	userpb "proto-go/user"
	"user-service/internal/app/services"
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
			Id:        resp.User.ID(),
			Email:     resp.User.Email().Value(),
			FirstName: resp.User.FirstName(),
			LastName:  resp.User.LastName(),
			Phone:     resp.User.Phone(),
			CreatedAt: resp.User.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: resp.User.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		},
		Token:   resp.Token,
		Message: "User registered successfully",
	}, nil
}

func (s *PBUserServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	resp, err := s.svc.LoginUser(ctx, &services.LoginRequest{Email: req.Email, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return &userpb.LoginResponse{
		User: &userpb.User{
			Id:        resp.User.ID(),
			Email:     resp.User.Email().Value(),
			FirstName: resp.User.FirstName(),
			LastName:  resp.User.LastName(),
			Phone:     resp.User.Phone(),
			CreatedAt: resp.User.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: resp.User.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		},
		Token:   resp.Token,
		Message: "Login successful",
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
			CreatedAt: u.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: u.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
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
			CreatedAt: u.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: u.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		},
		Message: "Profile updated",
	}, nil
}
