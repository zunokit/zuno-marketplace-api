package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
)

// Init initializes the Redis client
func Init(addr string) error {
	opts, err := redis.ParseURL(addr)
	if err != nil {
		return fmt.Errorf("parse redis url: %w", err)
	}

	// Upstash-specific optimizations
	opts.PoolSize = 10
	opts.MinIdleConns = 5
	opts.MaxRetries = 3
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	client = redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return client
}

// Close closes the Redis client
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}
