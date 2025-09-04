package auth

import (
	"context"
	"time"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/redisclient"
)

// RedisTokenStorage implements TokenStorage interface using Redis
type RedisTokenStorage struct {
	client *redisclient.Client
	logger logger.Logger
}

// NewRedisTokenStorage creates new Redis token storage
func NewRedisTokenStorage(client *redisclient.Client, logger logger.Logger) *RedisTokenStorage {
	return &RedisTokenStorage{
		client: client,
		logger: logger,
	}
}

// StoreToken stores refresh token in Redis with TTL
func (s *RedisTokenStorage) StoreToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	err := s.client.StoreToken(ctx, token, userID, ttl)
	if err != nil {
		s.logger.Error("failed to store token in Redis", "error", err, "user_id", userID)
		return err
	}
	return nil
}

// IsTokenValid checks if token exists and is valid in Redis
func (s *RedisTokenStorage) IsTokenValid(ctx context.Context, token string) bool {
	return s.client.IsTokenValid(ctx, token)
}

// RevokeToken removes specific token from Redis
func (s *RedisTokenStorage) RevokeToken(ctx context.Context, token string) error {
	err := s.client.RevokeToken(ctx, token)
	if err != nil {
		s.logger.Error("failed to revoke token in Redis", "error", err)
		return err
	}
	return nil
}

// RevokeAllUserTokens removes all tokens for specific user from Redis
func (s *RedisTokenStorage) RevokeAllUserTokens(ctx context.Context, userID string) error {
	err := s.client.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		s.logger.Error("failed to revoke all user tokens in Redis", "error", err, "user_id", userID)
		return err
	}
	return nil
}

// Close closes Redis connection
func (s *RedisTokenStorage) Close() error {
	return s.client.Close()
}
