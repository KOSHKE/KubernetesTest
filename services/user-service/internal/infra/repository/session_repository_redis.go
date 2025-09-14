package repository

import (
	"context"
	"time"

	"ecommerce-platform/pkg/common/errors"
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
	key := r.sessionKey(session.ID)
	refreshKey := r.refreshTokenKey(session.RefreshToken.Value())
	userSessionsKey := r.userSessionsKey(session.UserID)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, session, r.ttl)
	pipe.Set(ctx, refreshKey, session.ID, r.ttl)
	pipe.SAdd(ctx, userSessionsKey, session.ID)
	pipe.Expire(ctx, userSessionsKey, r.ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.ErrSessionCreationFailed
	}
	return nil
}

// GetByID retrieves a session by ID
func (r *RedisSessionRepository) GetByID(ctx context.Context, id string) (*entities.Session, error) {
	key := r.sessionKey(id)
	var session entities.Session
	if err := r.client.Get(ctx, key, &session); err != nil {
		return nil, errors.ErrSessionNotFound
	}
	return &session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *RedisSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	userSessionsKey := r.userSessionsKey(userID)
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey)
	if err != nil {
		return nil, errors.ErrSessionNotFound
	}

	var sessions []*entities.Session
	for _, sessionID := range sessionIDs {
		session, err := r.GetByID(ctx, sessionID)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *RedisSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error) {
	refreshKey := r.refreshTokenKey(refreshToken)
	var sessionID string
	if err := r.client.Get(ctx, refreshKey, &sessionID); err != nil {
		return nil, errors.ErrSessionNotFound
	}
	return r.GetByID(ctx, sessionID)
}

// Delete removes a session
func (r *RedisSessionRepository) Delete(ctx context.Context, id string) error {
	session, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	key := r.sessionKey(id)
	refreshKey := r.refreshTokenKey(session.RefreshToken.Value())
	userSessionsKey := r.userSessionsKey(session.UserID)

	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, refreshKey)
	pipe.SRem(ctx, userSessionsKey, id)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return errors.ErrSessionDeletionFailed
	}
	return nil
}

// DeleteByUserID removes all sessions for a user
func (r *RedisSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	userSessionsKey := r.userSessionsKey(userID)
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	for _, sessionID := range sessionIDs {
		pipe.Del(ctx, r.sessionKey(sessionID))
	}
	pipe.Del(ctx, userSessionsKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return errors.ErrSessionDeletionFailed
	}
	return nil
}

// Helper methods for Redis key generation
func (r *RedisSessionRepository) sessionKey(sessionID string) string {
	return "session:" + sessionID
}

func (r *RedisSessionRepository) refreshTokenKey(refreshToken string) string {
	return "refresh_token:" + refreshToken
}

func (r *RedisSessionRepository) userSessionsKey(userID string) string {
	return "user_sessions:" + userID
}
