package grpc

import (
	"context"
	"time"
)

// WithTimeout creates a context with timeout for gRPC operations
func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}
