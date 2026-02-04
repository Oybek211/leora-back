package redisclient

import (
	"context"

	"github.com/leora/leora-server/internal/config"
	"github.com/redis/go-redis/v9"
)

// New creates a redis client.
// If cfg.URL (REDIS_URL) is set, it is parsed and used; otherwise
// the individual addr/password/db fields are used.
func New(cfg config.RedisConfig) (*redis.Client, error) {
	var opts *redis.Options
	if cfg.URL != "" {
		var err error
		opts, err = redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, err
		}
	} else {
		opts = &redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
