package redisclient

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is a minimal Redis wrapper for basic operations
type Client struct {
	rdb *redis.Client
}

// New creates a new Redis client
func New(addr, password string, db int) *Client {
	return &Client{
		rdb: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping checks Redis connectivity
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Set stores a value with TTL (primitives stored as-is, structs JSON-marshaled)
func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	switch v := value.(type) {
	case string:
		return c.rdb.Set(ctx, key, v, ttl).Err()
	case int, int32, int64, float32, float64, bool:
		return c.rdb.Set(ctx, key, v, ttl).Err()
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return c.rdb.Set(ctx, key, data, ttl).Err()
	}
}

// Get retrieves a value and unmarshals it into dest
func (c *Client) Get(ctx context.Context, key string, dest any) error {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	// For primitives, try direct unmarshaling first
	switch dest.(type) {
	case *string, *int, *int32, *int64, *float32, *float64, *bool:
		// Try direct unmarshaling for primitives
		if err := json.Unmarshal(data, dest); err == nil {
			return nil
		}
		// If JSON unmarshal fails, try direct string conversion for strings
		if strDest, ok := dest.(*string); ok {
			*strDest = string(data)
			return nil
		}
		return err
	default:
		// For complex types, always use JSON unmarshaling
		return json.Unmarshal(data, dest)
	}
}

// Del deletes one or more keys
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists checks if key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	res, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

// FlushDB removes all keys from the current database
func (c *Client) FlushDB(ctx context.Context) error {
	return c.rdb.FlushDB(ctx).Err()
}

// Expire sets TTL for an existing key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// Pipeline returns a Redis pipeline for batch operations
func (c *Client) Pipeline() redis.Pipeliner {
	return c.rdb.Pipeline()
}

// WithTransaction executes a function inside a Redis transaction (MULTI/EXEC)
func (c *Client) WithTransaction(ctx context.Context, fn func(tx redis.Pipeliner) error) error {
	_, err := c.rdb.TxPipelined(ctx, fn)
	return err
}

// SAdd adds one or more members to a Redis set
func (c *Client) SAdd(ctx context.Context, key string, members ...any) error {
	return c.rdb.SAdd(ctx, key, members...).Err()
}

// SRem removes one or more members from a Redis set
func (c *Client) SRem(ctx context.Context, key string, members ...any) error {
	return c.rdb.SRem(ctx, key, members...).Err()
}

// SMembers returns all members of a Redis set
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

// SAddJSON adds one or more complex objects to a Redis set as JSON
func (c *Client) SAddJSON(ctx context.Context, key string, values ...any) error {
	strValues := make([]any, len(values))
	for i, v := range values {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		strValues[i] = data
	}
	return c.rdb.SAdd(ctx, key, strValues...).Err()
}

// MGet retrieves multiple values and unmarshals them into dest slice
func (c *Client) MGet(ctx context.Context, keys []string, dest any) error {
	if len(keys) == 0 {
		return nil
	}

	results, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}

	var raw []json.RawMessage
	for _, r := range results {
		if r == nil {
			continue
		}
		if str, ok := r.(string); ok && str != "" {
			raw = append(raw, json.RawMessage(str))
		}
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}
