package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// ErrCacheMiss is returned when a key is not found in the cache.
var ErrCacheMiss = errors.New("cache miss")

// Cache wraps a Redis client and provides caching functionality.
type Cache struct {
	client *redis.Client
}

// NewCache creates a new Cache instance.
func NewCache() (*Cache, error) {
	dsn := viper.GetString("redis.dsn")
	if dsn == "" {
		dsn = "127.0.0.1:6379"
	}

	opts, err := redis.ParseURL("redis://" + dsn)
	if err != nil {
		opts = &redis.Options{Addr: dsn}
	}

	client := redis.NewClient(opts)
	
	// Use a context with timeout for the ping to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Cache{client: client}, nil
}

// Get retrieves a value from the cache.
func (c *Cache) Get(ctx context.Context, key string, v interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), v)
}

// Set sets a cache key to the provided value with expiration.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expires time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expires).Err()
}

// Delete removes a key from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// FlushDB flushes the entire Redis database.
func (c *Cache) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}
