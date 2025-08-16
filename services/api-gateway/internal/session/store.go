package session

import (
	"context"
	"time"
)

type SessionStore interface {
	Create(ctx context.Context, userID string, ttl time.Duration) (string, error)
	Rotate(ctx context.Context, oldToken string, ttl time.Duration) (newToken string, userID string, err error)
	Get(ctx context.Context, token string) (string, error)
	Delete(ctx context.Context, token string) error
}
