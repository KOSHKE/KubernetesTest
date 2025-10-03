package clients

import (
	"context"
	"fmt"
	"time"

	dto "ecommerce-platform/pkg/common/dto/user-service"
	"ecommerce-platform/proto-go/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserClient defines the interface for user operations
type UserClient interface {
	Close() error
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.RegisterUserResponse, error)
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	GetUser(ctx context.Context, userID string) (*dto.GetUserResponse, error)
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	Logout(ctx context.Context, sessionID string) error
}

type userClient struct {
	conn   *grpc.ClientConn
	client user.UserServiceClient
}

// NewUserClient creates a new gRPC client for user service
func NewUserClient(address string) (UserClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	return &userClient{
		conn:   conn,
		client: user.NewUserServiceClient(conn),
	}, nil
}

func (c *userClient) Close() error {
	return c.conn.Close()
}

func (c *userClient) Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.RegisterUserResponse, error) {
	pbReq := &user.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	resp, err := c.client.Register(ctx, pbReq)
	if err != nil {
		// return raw gRPC error so HTTP mapper can extract status/message
		return nil, err
	}

	u := resp.GetUser()
	return &dto.RegisterUserResponse{
		UserID:    u.GetId(),
		Email:     u.GetEmail(),
		FirstName: u.GetFirstName(),
		LastName:  u.GetLastName(),
		Phone:     u.GetPhone(),
		CreatedAt: u.GetCreatedAt().AsTime(),
	}, nil
}

func (c *userClient) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	pbReq := &user.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := c.client.Login(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	u := resp.GetUser()
	expiresAt := time.Now().Add(time.Duration(resp.GetExpiresIn()) * time.Second)

	return &dto.LoginResponse{
		UserID:       u.GetId(),
		Email:        u.GetEmail(),
		FirstName:    u.GetFirstName(),
		LastName:     u.GetLastName(),
		Phone:        u.GetPhone(),
		SessionID:    resp.GetSessionId(),
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
		ExpiresAt:    expiresAt,
	}, nil
}

func (c *userClient) GetUser(ctx context.Context, userID string) (*dto.GetUserResponse, error) {
	resp, err := c.client.GetUser(ctx, &user.GetUserRequest{Id: userID})
	if err != nil {
		return nil, err
	}

	u := resp.GetUser()
	return &dto.GetUserResponse{
		UserID:    u.GetId(),
		Email:     u.GetEmail(),
		FirstName: u.GetFirstName(),
		LastName:  u.GetLastName(),
		Phone:     u.GetPhone(),
		CreatedAt: u.GetCreatedAt().AsTime(),
		UpdatedAt: u.GetUpdatedAt().AsTime(),
	}, nil
}

func (c *userClient) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	resp, err := c.client.RefreshToken(ctx, &user.RefreshTokenRequest{SessionId: req.SessionID})
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(resp.GetExpiresIn()) * time.Second)
	return &dto.RefreshTokenResponse{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
		ExpiresAt:    expiresAt,
	}, nil
}

func (c *userClient) Logout(ctx context.Context, sessionID string) error {
	_, err := c.client.Logout(ctx, &user.LogoutRequest{SessionId: sessionID})
	if err != nil {
		return err
	}
	return nil
}
