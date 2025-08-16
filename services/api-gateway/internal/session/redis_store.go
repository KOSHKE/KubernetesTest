package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisSessionStore struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisSessionStore(rdb *redis.Client, prefix string) *RedisSessionStore {
	if prefix == "" {
		prefix = "sess:"
	}
	return &RedisSessionStore{rdb: rdb, prefix: prefix}
}

func (s *RedisSessionStore) key(token string) string { return s.prefix + token }

func (s *RedisSessionStore) Create(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	token, err := generateOpaque(32)
	if err != nil {
		return "", err
	}
	if err := s.rdb.Set(ctx, s.key(token), userID, ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *RedisSessionStore) Rotate(ctx context.Context, oldToken string, ttl time.Duration) (string, string, error) {
	userID, err := s.rdb.Get(ctx, s.key(oldToken)).Result()
	if err != nil {
		return "", "", err
	}
	_ = s.rdb.Del(ctx, s.key(oldToken)).Err()
	newToken, err := generateOpaque(32)
	if err != nil {
		return "", "", err
	}
	if err := s.rdb.Set(ctx, s.key(newToken), userID, ttl).Err(); err != nil {
		return "", "", err
	}
	return newToken, userID, nil
}

func (s *RedisSessionStore) Get(ctx context.Context, token string) (string, error) {
	return s.rdb.Get(ctx, s.key(token)).Result()
}

func (s *RedisSessionStore) Delete(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, s.key(token)).Err()
}

func generateOpaque(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
