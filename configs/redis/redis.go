package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ares/dp-vc-webApp/configs/env"
)

var (
	client *redis.Client
)

// Init initializes Redis connection
func Init(config *env.Config) (*redis.Client, error) {
	client = redis.NewClient(&redis.Options{
		Addr:     config.RedisURI,
		Username: config.RedisUser,
		Password: config.RedisPassword,
		DB:       0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// AcquireLock obtains a short-lived distributed lock for scheduled jobs.
func AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return client.SetNX(ctx, buildKey("lock:"+key), "1", expiration).Result()
}

// ReleaseLock releases a distributed lock.
func ReleaseLock(ctx context.Context, key string) error {
	if client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return client.Del(ctx, buildKey("lock:"+key)).Err()
}

// Close closes the Redis connection
func Close() {
	if client != nil {
		_ = client.Close()
	}
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return client
}

// GetPrefix returns the configured key prefix
func GetPrefix() string {
	return env.GetConfig().RedisPrefix
}

// buildKey builds the full key with prefix
func buildKey(key string) string {
	prefix := GetPrefix()
	if prefix != "" {
		return prefix + ":" + key
	}
	return key
}

// Get retrieves a value from Redis
func Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Get(ctx, buildKey(key)).Result()
}

// Set stores a value in Redis
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Set(ctx, buildKey(key), value, expiration).Err()
}

// Delete removes a key from Redis
func Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Del(ctx, buildKey(key)).Err()
}

// Exists checks if a key exists
func Exists(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := client.Exists(ctx, buildKey(key)).Result()
	return count > 0, err
}

// Expire sets expiration on a key
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Expire(ctx, buildKey(key), expiration).Err()
}

// HSet sets a hash field
func HSet(ctx context.Context, key string, field string, value interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.HSet(ctx, buildKey(key), field, value).Err()
}

// HGet gets a hash field
func HGet(ctx context.Context, key string, field string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.HGet(ctx, buildKey(key), field).Result()
}

// HGetAll gets all hash fields
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.HGetAll(ctx, buildKey(key)).Result()
}

// HDel deletes a hash field
func HDel(ctx context.Context, key string, field string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.HDel(ctx, buildKey(key), field).Err()
}

// Incr increments a key's value
func Incr(ctx context.Context, key string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Incr(ctx, buildKey(key)).Result()
}

// SetNX sets a value only if the key doesn't exist
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.SetNX(ctx, buildKey(key), value, expiration).Result()
}
