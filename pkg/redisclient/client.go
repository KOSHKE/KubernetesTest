package redisclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kubernetestest/ecommerce-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Client provides unified Redis operations for caching, tokens, and lists
type Client struct {
	rdb    *redis.Client
	logger logger.Logger
}

// New creates optimized Redis client with minimal configuration
func New(addr, password string, db int, logger logger.Logger) *Client {
	return &Client{
		rdb: redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           db,
			PoolSize:     20, // Increased for better performance
			MinIdleConns: 10, // Higher idle connections
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}),
		logger: logger,
	}
}

// Close terminates Redis connection
func (c *Client) Close() error {
	if err := c.rdb.Close(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to close Redis connection", "error", err)
		}
		return err
	}
	return nil
}

// Ping verifies Redis connectivity
func (c *Client) Ping(ctx context.Context) error {
	if err := c.rdb.Ping(ctx).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("Redis ping failed", "error", err)
		}
		return err
	}
	return nil
}

// === CACHE OPERATIONS (Unified, optimized) ===

// Set stores any value with automatic serialization and TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := c.serialize(value)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to serialize value", "error", err, "key", key)
		}
		return fmt.Errorf("serialize error: %w", err)
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to set cache value", "error", err, "key", key)
		}
		return fmt.Errorf("set cache error: %w", err)
	}
	return nil
}

// Get retrieves and deserializes value
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		if c.logger != nil {
			c.logger.Error("failed to get cache value", "error", err, "key", key)
		}
		return fmt.Errorf("get cache error: %w", err)
	}

	if err := c.deserialize(data, dest); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to deserialize value", "error", err, "key", key)
		}
		return fmt.Errorf("deserialize error: %w", err)
	}
	return nil
}

// Del removes one or more keys (batched)
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to delete keys", "error", err, "keys", keys)
		}
		return fmt.Errorf("delete error: %w", err)
	}
	return nil
}

// Exists checks if key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to check key existence", "error", err, "key", key)
		}
		return false, fmt.Errorf("exists check error: %w", err)
	}
	return exists > 0, nil
}

// Expire sets TTL for existing key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to set expiration", "error", err, "key", key)
		}
		return fmt.Errorf("expire error: %w", err)
	}
	return nil
}

// === TOKEN STORAGE (Optimized for JWT operations) ===

// StoreToken stores refresh token with automatic indexing for fast user-based operations
func (c *Client) StoreToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	tokenKey := fmt.Sprintf("token:%s", token)
	indexKey := fmt.Sprintf("user_tokens:%s", userID)

	// Use pipeline for atomic operations
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, tokenKey, userID, ttl)
	pipe.SAdd(ctx, indexKey, token)
	pipe.Expire(ctx, indexKey, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to store token", "error", err, "user_id", userID)
		}
		return fmt.Errorf("store token error: %w", err)
	}
	return nil
}

// GetTokenUser retrieves user ID for token
func (c *Client) GetTokenUser(ctx context.Context, token string) (string, error) {
	tokenKey := fmt.Sprintf("token:%s", token)

	userID, err := c.rdb.Get(ctx, tokenKey).Result()
	if err != nil {
		if err == redis.Nil {
			if c.logger != nil {
				c.logger.Warn("token not found", "token", token[:8]+"...")
			}
			return "", fmt.Errorf("token not found")
		}
		if c.logger != nil {
			c.logger.Error("failed to get token", "error", err, "token", token[:8]+"...")
		}
		return "", fmt.Errorf("get token error: %w", err)
	}
	return userID, nil
}

// RevokeToken removes single token
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	// Get userID first to clean up index
	userID, err := c.GetTokenUser(ctx, token)
	if err != nil {
		return err // Token doesn't exist, consider it revoked
	}

	tokenKey := fmt.Sprintf("token:%s", token)
	indexKey := fmt.Sprintf("user_tokens:%s", userID)

	// Pipeline for atomic cleanup
	pipe := c.rdb.Pipeline()
	pipe.Del(ctx, tokenKey)
	pipe.SRem(ctx, indexKey, token)

	if _, err := pipe.Exec(ctx); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to revoke token", "error", err, "token", token[:8]+"...")
		}
		return fmt.Errorf("revoke token error: %w", err)
	}
	return nil
}

// RevokeAllUserTokens efficiently removes all tokens for user using pre-built index
func (c *Client) RevokeAllUserTokens(ctx context.Context, userID string) error {
	indexKey := fmt.Sprintf("user_tokens:%s", userID)

	// Get all tokens for user
	tokens, err := c.rdb.SMembers(ctx, indexKey).Result()
	if err != nil && err != redis.Nil {
		if c.logger != nil {
			c.logger.Error("failed to get user tokens", "error", err, "user_id", userID)
		}
		return fmt.Errorf("get user tokens error: %w", err)
	}

	if len(tokens) == 0 {
		return nil // No tokens to revoke
	}

	// Batch delete all tokens and index
	pipe := c.rdb.Pipeline()
	for _, token := range tokens {
		pipe.Del(ctx, fmt.Sprintf("token:%s", token))
	}
	pipe.Del(ctx, indexKey)

	if _, err := pipe.Exec(ctx); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to revoke user tokens", "error", err, "user_id", userID, "count", len(tokens))
		}
		return fmt.Errorf("revoke user tokens error: %w", err)
	}
	return nil
}

// IsTokenValid checks token existence efficiently
func (c *Client) IsTokenValid(ctx context.Context, token string) bool {
	tokenKey := fmt.Sprintf("token:%s", token)
	exists, err := c.rdb.Exists(ctx, tokenKey).Result()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to check token validity", "error", err, "token", token[:8]+"...")
		}
		return false
	}
	return exists > 0
}

// === LIST OPERATIONS (Optimized for batch operations) ===

// LPush pushes multiple values to left of list
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	if err := c.rdb.LPush(ctx, key, values...).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to lpush", "error", err, "key", key, "count", len(values))
		}
		return fmt.Errorf("lpush error: %w", err)
	}
	return nil
}

// RPush pushes multiple values to right of list
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	if err := c.rdb.RPush(ctx, key, values...).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to rpush", "error", err, "key", key, "count", len(values))
		}
		return fmt.Errorf("rpush error: %w", err)
	}
	return nil
}

// LPop pops value from left
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	result, err := c.rdb.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list empty")
		}
		if c.logger != nil {
			c.logger.Error("failed to lpop", "error", err, "key", key)
		}
		return "", fmt.Errorf("lpop error: %w", err)
	}
	return result, nil
}

// RPop pops value from right
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	result, err := c.rdb.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list empty")
		}
		if c.logger != nil {
			c.logger.Error("failed to rpop", "error", err, "key", key)
		}
		return "", fmt.Errorf("rpop error: %w", err)
	}
	return result, nil
}

// LRange gets range of list values
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	result, err := c.rdb.LRange(ctx, key, start, stop).Result()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to lrange", "error", err, "key", key)
		}
		return nil, fmt.Errorf("lrange error: %w", err)
	}
	return result, nil
}

// LLen gets list length
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	result, err := c.rdb.LLen(ctx, key).Result()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to llen", "error", err, "key", key)
		}
		return 0, fmt.Errorf("llen error: %w", err)
	}
	return result, nil
}

// === BATCH OPERATIONS (High-performance multi-key operations) ===

// MSet sets multiple key-value pairs atomically
func (c *Client) MSet(ctx context.Context, pairs map[string]interface{}) error {
	if len(pairs) == 0 {
		return nil
	}

	// Convert to Redis format
	values := make([]interface{}, 0, len(pairs)*2)
	for key, value := range pairs {
		data, err := c.serialize(value)
		if err != nil {
			if c.logger != nil {
				c.logger.Error("failed to serialize batch value", "error", err, "key", key)
			}
			return fmt.Errorf("serialize batch error: %w", err)
		}
		values = append(values, key, data)
	}

	if err := c.rdb.MSet(ctx, values...).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to mset", "error", err, "count", len(pairs))
		}
		return fmt.Errorf("mset error: %w", err)
	}
	return nil
}

// MGet gets multiple values at once
func (c *Client) MGet(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	results, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to mget", "error", err, "keys", keys)
		}
		return nil, fmt.Errorf("mget error: %w", err)
	}

	result := make(map[string]string, len(keys))
	for i, key := range keys {
		if results[i] != nil {
			if str, ok := results[i].(string); ok {
				result[key] = str
			}
		}
	}
	return result, nil
}

// === UTILITY METHODS ===

// serialize handles different data types efficiently
func (c *Client) serialize(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return []byte(fmt.Sprintf("%d", v)), nil
	case float32, float64:
		return []byte(fmt.Sprintf("%f", v)), nil
	case bool:
		return []byte(fmt.Sprintf("%t", v)), nil
	default:
		return json.Marshal(value)
	}
}

// deserialize handles different destination types efficiently
func (c *Client) deserialize(data string, dest interface{}) error {
	switch d := dest.(type) {
	case *string:
		*d = data
		return nil
	case *[]byte:
		*d = []byte(data)
		return nil
	default:
		return json.Unmarshal([]byte(data), dest)
	}
}
