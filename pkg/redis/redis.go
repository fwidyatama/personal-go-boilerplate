package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate moq -pkg mock -out mock/redis_client_moq.go . RedisClient
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl int) error
}

// IsKeyNotFoundError checks if the error is a Redis "key not found" error
func IsKeyNotFoundError(err error) bool {
	return err == redis.Nil
}

// IsConnectionError checks if the error is a Redis connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common Redis connection errors
	return err == redis.ErrClosed ||
		err == redis.ErrPoolTimeout
}

type RedisAdapter struct {
	client *redis.Client
}

func NewRedis(addr, password string, db int) RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisAdapter{client: rdb}
}

func (r *RedisAdapter) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	return r.client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
}

func (r *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisAdapter) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

func (r *RedisAdapter) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisAdapter) Expire(ctx context.Context, key string, ttl int) error {
	return r.client.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
}
