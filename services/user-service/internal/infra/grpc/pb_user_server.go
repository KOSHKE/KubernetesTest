package grpc

import (
	"context"
	"time"

	"ecommerce-platform/pkg/common/grpcutils"
	userpb "ecommerce-platform/proto-go/user"
	"ecommerce-platform/services/user-service/internal/application/dto"
	appsvc "ecommerce-platform/services/user-service/internal/application/services"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBUserServer struct {
	userpb.UnimplementedUserServiceServer
	svc *appsvc.UserApplicationService
}

func NewPBUserServer(svc *appsvc.UserApplicationService) *PBUserServer {
	return &PBUserServer{svc: svc}
}

func (s *PBUserServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	// Map proto -> app request
	appReq := &dto.RegisterUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	response, err := s.svc.RegisterUser(ctx, appReq)
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &userpb.RegisterResponse{
		User: mapUserResponseToPB(&dto.GetUserResponse{
			UserID:    response.UserID,
			Email:     response.Email,
			FirstName: response.FirstName,
			LastName:  response.LastName,
			Phone:     response.Phone,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.CreatedAt, // Use CreatedAt for newly registered user
		}),
		Message: "User registered successfully",
	}, nil
}

func (s *PBUserServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	// Map proto -> app request
	appReq := &dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := s.svc.LoginUser(ctx, appReq)
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &userpb.LoginResponse{
		User: mapUserResponseToPB(&dto.GetUserResponse{
			UserID:    response.UserID,
			Email:     response.Email,
			FirstName: response.FirstName,
			LastName:  response.LastName,
			Phone:     response.Phone,
			CreatedAt: time.Now(), // Use current time for login response
			UpdatedAt: time.Now(),
		}),
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    int64(time.Until(response.ExpiresAt).Seconds()),
	}, nil
}

func (s *PBUserServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	// Map proto -> app request
	appReq := &dto.GetUserRequest{
		UserID: req.Id,
	}

	response, err := s.svc.GetUser(ctx, appReq)
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &userpb.GetUserResponse{
		User: mapUserResponseToPB(response),
	}, nil
}

// RefreshToken generates new access and refresh token pair
func (s *PBUserServer) RefreshToken(ctx context.Context, req *userpb.RefreshTokenRequest) (*userpb.RefreshTokenResponse, error) {
	// Map proto -> app request
	appReq := &dto.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	response, err := s.svc.RefreshToken(ctx, appReq)
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &userpb.RefreshTokenResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    int64(time.Until(response.ExpiresAt).Seconds()),
	}, nil
}

// Logout revokes refresh token
func (s *PBUserServer) Logout(ctx context.Context, req *userpb.LogoutRequest) (*userpb.LogoutResponse, error) {
	// Map proto -> app request
	appReq := &dto.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}

	err := s.svc.Logout(ctx, appReq)
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &userpb.LogoutResponse{
		Message: "Logout successful",
	}, nil
}

// Mapping helpers
func mapUserResponseToPB(u *dto.GetUserResponse) *userpb.User {
	return &userpb.User{
		Id:        u.UserID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
