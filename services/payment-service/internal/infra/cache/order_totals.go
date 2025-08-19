package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderTotalsCache interface {
	Set(ctx context.Context, orderID string, amount int64, currency string, ttl time.Duration) error
	Get(ctx context.Context, orderID string) (amount int64, currency string, ok bool, err error)
	Del(ctx context.Context, orderID string) error
	Close() error
}

type RedisOrderTotalsCache struct {
	rdb       *redis.Client
	keyPrefix string
}

func NewRedisOrderTotalsCache(addr string, db int, keyPrefix string) *RedisOrderTotalsCache {
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	return &RedisOrderTotalsCache{rdb: rdb, keyPrefix: keyPrefix}
}

func (c *RedisOrderTotalsCache) key(orderID string) string { return c.keyPrefix + orderID }

func (c *RedisOrderTotalsCache) Set(ctx context.Context, orderID string, amount int64, currency string, ttl time.Duration) error {
	key := c.key(orderID)
	if err := c.rdb.HSet(ctx, key, "amount", strconv.FormatInt(amount, 10), "currency", currency).Err(); err != nil {
		return err
	}
	if ttl > 0 {
		if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisOrderTotalsCache) Get(ctx context.Context, orderID string) (int64, string, bool, error) {
	key := c.key(orderID)
	vals, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, "", false, err
	}
	if len(vals) == 0 {
		return 0, "", false, nil
	}
	amtStr, okAmt := vals["amount"]
	curr, okCur := vals["currency"]
	if !okAmt || !okCur {
		return 0, "", false, nil
	}
	amt, err := strconv.ParseInt(amtStr, 10, 64)
	if err != nil {
		return 0, "", false, err
	}
	return amt, curr, true, nil
}

func (c *RedisOrderTotalsCache) Del(ctx context.Context, orderID string) error {
	return c.rdb.Del(ctx, c.key(orderID)).Err()
}

func (c *RedisOrderTotalsCache) Close() error { return c.rdb.Close() }
