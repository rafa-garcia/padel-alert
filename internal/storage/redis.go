package storage

import (
	"context"
	"fmt"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the redis client
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis", "address", opts.Addr)
	return &RedisClient{Client: client}, nil
}

// Close closes the Redis client
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// CheckHealth checks Redis connection
func (r *RedisClient) CheckHealth(ctx context.Context) error {
	if _, err := r.Client.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}
