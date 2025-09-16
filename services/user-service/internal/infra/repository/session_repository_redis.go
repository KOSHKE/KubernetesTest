package repository

import (
	"context"
	"time"

	"ecommerce-platform/pkg/redisclient"
	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
)

// RedisSessionRepository implements SessionRepository using Redis
type RedisSessionRepository struct {
	client *redisclient.Client
	ttl    time.Duration
}

// NewRedisSessionRepository creates a new Redis session repository
func NewRedisSessionRepository(client *redisclient.Client, ttl time.Duration) repository.SessionRepository {
	return &RedisSessionRepository{
		client: client,
		ttl:    ttl,
	}
}

// Save saves or updates a session
func (r *RedisSessionRepository) Save(ctx context.Context, session *entities.Session) error {
	pipe := r.client.Pipeline()
	pipe.Set(ctx, "session:"+session.ID, session, r.ttl)
	pipe.Set(ctx, "refresh_token:"+session.RefreshToken.Value, session.ID, r.ttl)
	pipe.SAdd(ctx, "user_sessions:"+session.UserID, session.ID)
	pipe.Expire(ctx, "user_sessions:"+session.UserID, r.ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a session by ID
func (r *RedisSessionRepository) GetByID(ctx context.Context, id string) (*entities.Session, error) {
	var session entities.Session
	if err := r.client.Get(ctx, "session:"+id, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *RedisSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	sessionIDs, err := r.client.SMembers(ctx, "user_sessions:"+userID)
	if err != nil {
		return nil, err
	}

	if len(sessionIDs) == 0 {
		return []*entities.Session{}, nil
	}

	// Build session keys for MGET
	sessionKeys := make([]string, len(sessionIDs))
	for i, sessionID := range sessionIDs {
		sessionKeys[i] = "session:" + sessionID
	}

	// Use MGET to fetch all sessions in one request
	var sessions []*entities.Session
	if err := r.client.MGet(ctx, sessionKeys, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *RedisSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error) {
	var sessionID string
	if err := r.client.Get(ctx, "refresh_token:"+refreshToken, &sessionID); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, sessionID)
}

// Delete removes a session
func (r *RedisSessionRepository) Delete(ctx context.Context, id string) error {
	session, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, "session:"+id)
	pipe.Del(ctx, "refresh_token:"+session.RefreshToken.Value)
	pipe.SRem(ctx, "user_sessions:"+session.UserID, id)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// DeleteByUserID removes all sessions for a user
func (r *RedisSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	sessionIDs, err := r.client.SMembers(ctx, "user_sessions:"+userID)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	for _, sessionID := range sessionIDs {
		pipe.Del(ctx, "session:"+sessionID)
	}
	pipe.Del(ctx, "user_sessions:"+userID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisSessionRepository) Close() error {
	return r.client.Close()
}
