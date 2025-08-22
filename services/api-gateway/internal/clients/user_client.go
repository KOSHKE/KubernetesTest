package clients

import (
	"context"
	"fmt"

	grpcutil "api-gateway/pkg/grpc"
	userpb "proto-go/user"
)

// UserClient defines the interface for user-related operations.
type UserClient interface {
	Close() error
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, userID string, req *RegisterRequest) (*User, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
}

type userClient struct {
	*grpcutil.BaseClient
	client userpb.UserServiceClient
}

// User represents a user entity.
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// RegisterRequest contains information for registering a user.
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// LoginRequest contains information for logging in a user.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse contains authentication result.
type AuthResponse struct {
	User         *User  `json:"user"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

// NewUserClient creates a new gRPC client for user service.
func NewUserClient(address string) (UserClient, error) {
	baseClient, err := grpcutil.NewBaseClient(address)
	if err != nil {
		return nil, err
	}
	return &userClient{
		BaseClient: baseClient,
		client:     userpb.NewUserServiceClient(baseClient.GetConn()),
	}, nil
}

// callGRPC wraps a gRPC call with timeout and validates the response containing a user.
func (c *userClient) callGRPC(ctx context.Context, fn func(ctx context.Context) (any, error)) (*User, error) {
	resp, err := grpcutil.WithTimeoutResult(ctx, fn)
	if err != nil {
		return nil, err
	}

	var user *userpb.User
	switch v := resp.(type) {
	case *userpb.RegisterResponse:
		user = v.User
	case *userpb.LoginResponse:
		user = v.User
	case *userpb.GetUserResponse:
		user = v.User
	case *userpb.UpdateUserResponse:
		user = v.User
	default:
		return nil, fmt.Errorf("unsupported response type")
	}

	if user == nil {
		return nil, fmt.Errorf("empty user in response")
	}
	return mapUserFromPB(user), nil
}

// Register registers a new user.
func (c *userClient) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	user, err := c.callGRPC(ctx, func(ctx context.Context) (any, error) {
		return c.client.Register(ctx, &userpb.RegisterRequest{
			Email:     req.Email,
			Password:  req.Password,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Phone:     req.Phone,
		})
	})
	if err != nil {
		return nil, err
	}

	return &AuthResponse{User: user, Message: "registered successfully"}, nil
}

// Login authenticates a user.
func (c *userClient) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Call gRPC directly to get full response with tokens
	resp, err := grpcutil.WithTimeoutResult(ctx, func(ctx context.Context) (any, error) {
		return c.client.Login(ctx, &userpb.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
	})
	if err != nil {
		return nil, err
	}

	// Extract login response
	loginResp, ok := resp.(*userpb.LoginResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	// Map user from protobuf
	user := mapUserFromPB(loginResp.User)

	return &AuthResponse{
		User:         user,
		Message:      "login successful",
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    loginResp.ExpiresIn,
	}, nil
}

// GetUser fetches a user by ID.
func (c *userClient) GetUser(ctx context.Context, userID string) (*User, error) {
	return c.callGRPC(ctx, func(ctx context.Context) (any, error) {
		return c.client.GetUser(ctx, &userpb.GetUserRequest{Id: userID})
	})
}

// UpdateUser updates user details.
func (c *userClient) UpdateUser(ctx context.Context, userID string, req *RegisterRequest) (*User, error) {
	return c.callGRPC(ctx, func(ctx context.Context) (any, error) {
		return c.client.UpdateUser(ctx, &userpb.UpdateUserRequest{
			Id:        userID,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Phone:     req.Phone,
		})
	})
}

// mapUserFromPB converts protobuf user to local User struct.
func mapUserFromPB(u *userpb.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:        u.Id,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		CreatedAt: grpcutil.FormatTimestamp(u.CreatedAt),
		UpdatedAt: grpcutil.FormatTimestamp(u.UpdatedAt),
	}
}

// RefreshToken refreshes access token using refresh token
func (c *userClient) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	resp, err := grpcutil.WithTimeoutResult(ctx, func(ctx context.Context) (any, error) {
		return c.client.RefreshToken(ctx, &userpb.RefreshTokenRequest{
			RefreshToken: refreshToken,
		})
	})
	if err != nil {
		return "", err
	}

	refreshResp, ok := resp.(*userpb.RefreshTokenResponse)
	if !ok {
		return "", fmt.Errorf("unsupported response type")
	}

	return refreshResp.AccessToken, nil
}

// Logout revokes refresh token
func (c *userClient) Logout(ctx context.Context, refreshToken string) error {
	_, err := grpcutil.WithTimeoutResult(ctx, func(ctx context.Context) (any, error) {
		return c.client.Logout(ctx, &userpb.LogoutRequest{
			RefreshToken: refreshToken,
		})
	})
	return err
}
