package redisclient

import (
	"context"

	"github.com/leora/leora-server/internal/config"
	"github.com/redis/go-redis/v9"
)

// New creates a redis client.
func New(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
