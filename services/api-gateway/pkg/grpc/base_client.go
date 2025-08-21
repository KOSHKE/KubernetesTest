package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// BaseClient provides common functionality for all gRPC clients
type BaseClient struct {
	conn *grpc.ClientConn
}

// NewBaseClient creates a new base client with connection
func NewBaseClient(address string) (*BaseClient, error) {
	conn, err := Dial(address)
	if err != nil {
		return nil, err
	}
	return &BaseClient{conn: conn}, nil
}

// Close closes the underlying gRPC connection
func (c *BaseClient) Close() error {
	return c.conn.Close()
}

// GetConn returns the underlying gRPC connection
func (c *BaseClient) GetConn() *grpc.ClientConn {
	return c.conn
}

// WithTimeoutResult is a generic helper for gRPC calls with timeout
// Can be used with any function that returns (T, error)
func WithTimeoutResult[T any](ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	ctx, cancel := WithTimeout(ctx)
	defer cancel()
	return fn(ctx)
}

// WithTimeoutNoResult is a helper for gRPC calls that don't return values
// Can be used with any function that returns error
func WithTimeoutNoResult(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := WithTimeout(ctx)
	defer cancel()
	return fn(ctx)
}
