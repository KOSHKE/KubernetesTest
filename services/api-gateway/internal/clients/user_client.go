package clients

import (
	"context"
	"fmt"
	"time"

	userpb "proto-go/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Interface for DI
type UserClient interface {
	Close() error
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, userID string, req *RegisterRequest) (*User, error)
}

type userClient struct {
	conn   *grpc.ClientConn
	client userpb.UserServiceClient
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User    *User  `json:"user"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

func NewUserClient(address string) (UserClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &userClient{conn: conn, client: userpb.NewUserServiceClient(conn)}, nil
}

func (c *userClient) Close() error { return c.conn.Close() }

func (c *userClient) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}
	resp, err := c.client.Register(ctx, &userpb.RegisterRequest{Email: req.Email, Password: req.Password, FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone})
	if err != nil {
		return nil, err
	}
	return &AuthResponse{User: &User{ID: resp.User.Id, Email: resp.User.Email, FirstName: resp.User.FirstName, LastName: resp.User.LastName, Phone: resp.User.Phone, CreatedAt: resp.User.CreatedAt, UpdatedAt: resp.User.UpdatedAt}, Token: resp.Token, Message: resp.Message}, nil
}

func (c *userClient) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}
	resp, err := c.client.Login(ctx, &userpb.LoginRequest{Email: req.Email, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return &AuthResponse{User: &User{ID: resp.User.Id, Email: resp.User.Email, FirstName: resp.User.FirstName, LastName: resp.User.LastName, Phone: resp.User.Phone, CreatedAt: resp.User.CreatedAt, UpdatedAt: resp.User.UpdatedAt}, Token: resp.Token, Message: resp.Message}, nil
}

func (c *userClient) GetUser(ctx context.Context, userID string) (*User, error) {
	resp, err := c.client.GetUser(ctx, &userpb.GetUserRequest{Id: userID})
	if err != nil {
		return nil, err
	}
	return &User{ID: resp.User.Id, Email: resp.User.Email, FirstName: resp.User.FirstName, LastName: resp.User.LastName, Phone: resp.User.Phone, CreatedAt: resp.User.CreatedAt, UpdatedAt: resp.User.UpdatedAt}, nil
}

func (c *userClient) UpdateUser(ctx context.Context, userID string, req *RegisterRequest) (*User, error) {
	resp, err := c.client.UpdateUser(ctx, &userpb.UpdateUserRequest{Id: userID, FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone})
	if err != nil {
		return nil, err
	}
	return &User{ID: resp.User.Id, Email: resp.User.Email, FirstName: resp.User.FirstName, LastName: resp.User.LastName, Phone: resp.User.Phone, CreatedAt: resp.User.CreatedAt, UpdatedAt: resp.User.UpdatedAt}, nil
}

// Helper function to generate mock JWT
func generateMockJWT(userID, email string) string {
	// Simple mock JWT format: header.payload.signature
	return fmt.Sprintf("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.%s.mock-signature",
		encodeBase64(fmt.Sprintf(`{"user_id":"%s","email":"%s","exp":%d}`,
			userID, email, time.Now().Add(24*time.Hour).Unix())))
}

func encodeBase64(s string) string {
	// Simple base64-like encoding for mock
	return fmt.Sprintf("%x", s)
}
