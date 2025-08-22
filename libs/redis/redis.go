package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client represents Redis client
type Client struct {
	client *redis.Client
}

// NewClient creates new Redis client
func NewClient(addr, password string, db int) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &Client{client: client}
}

// Close closes Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Ping checks Redis connection
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// StoreRefreshToken stores refresh token with user ID and expiration
func (c *Client) StoreRefreshToken(ctx context.Context, refreshToken, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	data := map[string]string{
		"user_id": userID,
		"token":   refreshToken,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	err = c.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

// GetRefreshToken retrieves refresh token data
func (c *Client) GetRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("refresh token not found")
		}
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	var tokenData map[string]string
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return "", fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	userID, exists := tokenData["user_id"]
	if !exists {
		return "", fmt.Errorf("invalid token data")
	}

	return userID, nil
}

// RevokeRefreshToken removes refresh token from Redis
func (c *Client) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens removes all refresh tokens for a specific user
func (c *Client) RevokeAllUserTokens(ctx context.Context, userID string) error {
	pattern := "refresh_token:*"

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// Get token data to check if it belongs to the user
		data, err := c.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var tokenData map[string]string
		if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
			continue
		}

		if tokenData["user_id"] == userID {
			if err := c.client.Del(ctx, key).Err(); err != nil {
				return fmt.Errorf("failed to revoke user token: %w", err)
			}
		}
	}

	return nil
}

// IsRefreshTokenValid checks if refresh token exists and is valid
func (c *Client) IsRefreshTokenValid(ctx context.Context, refreshToken string) bool {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)

	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false
	}

	return exists > 0
}
