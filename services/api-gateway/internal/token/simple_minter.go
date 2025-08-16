package token

import (
	"context"
	"time"

	"api-gateway/internal/session"
)

type SimpleMinter struct {
	store session.SessionStore
	ttl   time.Duration
}

func NewSimpleMinter(store session.SessionStore, ttl time.Duration) *SimpleMinter {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &SimpleMinter{store: store, ttl: ttl}
}

func (m *SimpleMinter) MintAccessToken(ctx context.Context, userID string) (string, error) {
	return m.store.Create(ctx, userID, m.ttl)
}
