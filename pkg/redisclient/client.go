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

// Set stores a value with TTL (any struct will be JSON-marshaled)
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value and unmarshals it into dest
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
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
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SAdd(ctx, key, members...).Err()
}

// SRem removes one or more members from a Redis set
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SRem(ctx, key, members...).Err()
}

// SMembers returns all members of a Redis set
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

// SAddJSON adds one or more complex objects to a Redis set as JSON
func (c *Client) SAddJSON(ctx context.Context, key string, values ...interface{}) error {
	strValues := make([]interface{}, len(values))
	for i, v := range values {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		strValues[i] = data
	}
	return c.rdb.SAdd(ctx, key, strValues...).Err()
}
