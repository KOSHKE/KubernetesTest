package grpc

import (
	"context"
	"time"

	"ecommerce-platform/pkg/common/grpcutils"
	"ecommerce-platform/pkg/metrics"
	userpb "ecommerce-platform/proto-go/user"
	"ecommerce-platform/services/user-service/internal/application/dto"
	appsvc "ecommerce-platform/services/user-service/internal/application/services"
	userMetrics "ecommerce-platform/services/user-service/internal/metrics"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// PBUserServer implements the gRPC user service
type PBUserServer struct {
	userpb.UnimplementedUserServiceServer
	appService *appsvc.UserApplicationService
	metrics    userMetrics.UserMetrics
}

// NewPBUserServer creates a new user gRPC server
func NewPBUserServer(appService *appsvc.UserApplicationService, metrics userMetrics.UserMetrics) *PBUserServer {
	return &PBUserServer{
		appService: appService,
		metrics:    metrics,
	}
}

// Register creates a new user account
func (s *PBUserServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	start := time.Now()
	method := "Register"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Convert gRPC request to DTO
	appReq := &dto.RegisterUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	// Register user
	response, err := s.appService.RegisterUser(ctx, appReq)
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	grpcResponse := &userpb.RegisterResponse{
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
	}

	// Record business metrics
	s.metrics.EntityEvent(metrics.EntityTypeUser, metrics.ActionCreated, "")

	return grpcResponse, nil
}

// Login authenticates a user and returns access tokens
func (s *PBUserServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	start := time.Now()
	method := "Login"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Convert gRPC request to DTO
	appReq := &dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Login user
	response, err := s.appService.LoginUser(ctx, appReq)
	if err != nil {
		status = "error"
		s.metrics.EntityEvent(metrics.EntityTypeUser, metrics.ActionLoginFailed, "authentication_failed")
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Get user details for response
	userReq := &dto.GetUserRequest{UserID: response.UserID}
	userResponse, err := s.appService.GetUser(ctx, userReq)
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	grpcResponse := &userpb.LoginResponse{
		User:         mapUserResponseToPB(userResponse),
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    int64(time.Until(response.ExpiresAt).Seconds()),
	}

	// Record business metrics
	s.metrics.EntityEvent(metrics.EntityTypeUser, metrics.ActionLoginSuccess, "")

	return grpcResponse, nil
}

// GetUser retrieves user information by ID
func (s *PBUserServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	start := time.Now()
	method := "GetUser"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Convert gRPC request to DTO
	appReq := &dto.GetUserRequest{
		UserID: req.Id,
	}

	// Get user
	response, err := s.appService.GetUser(ctx, appReq)
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	grpcResponse := &userpb.GetUserResponse{
		User: mapUserResponseToPB(response),
	}

	return grpcResponse, nil
}

// RefreshToken generates new access and refresh token pair
func (s *PBUserServer) RefreshToken(ctx context.Context, req *userpb.RefreshTokenRequest) (*userpb.RefreshTokenResponse, error) {
	start := time.Now()
	method := "RefreshToken"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Convert gRPC request to DTO
	appReq := &dto.RefreshTokenRequest{
		SessionID: req.RefreshToken, // Using refresh token as session ID for now
	}

	// Refresh token
	response, err := s.appService.RefreshToken(ctx, appReq)
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	grpcResponse := &userpb.RefreshTokenResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    int64(time.Until(response.ExpiresAt).Seconds()),
	}

	return grpcResponse, nil
}

// Logout revokes refresh token
func (s *PBUserServer) Logout(ctx context.Context, req *userpb.LogoutRequest) (*userpb.LogoutResponse, error) {
	start := time.Now()
	method := "Logout"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Convert gRPC request to DTO
	appReq := &dto.LogoutRequest{
		SessionID: req.RefreshToken, // Using refresh token as session ID for now
	}

	// Logout user
	err := s.appService.Logout(ctx, appReq)
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	grpcResponse := &userpb.LogoutResponse{
		Message: "Logout successful",
	}

	return grpcResponse, nil
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
