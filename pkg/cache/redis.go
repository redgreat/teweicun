/**
 * 功能：redis.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package cache documentation
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redgreat/teweicun/internal/config"
	"github.com/redgreat/teweicun/pkg/logger"
	"go.uber.org/zap"
)

var Client *redis.Client

// InitRedis initializes the redis client
func InitRedis(cfg config.RedisConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("unable to connect to redis: %w", err)
	}

	Client = client
	logger.Log.Info("Redis client initialized", zap.Int("db", cfg.DB))

	return nil
}

// CloseRedis closes the redis client
func CloseRedis() {
	if Client != nil {
		if err := Client.Close(); err != nil {
			logger.Log.Error("Failed to close redis client", zap.Error(err))
		} else {
			logger.Log.Info("Redis client closed")
		}
	}
}
