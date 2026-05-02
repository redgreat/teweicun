/**
 * 功能：postgres.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package database documentation
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redgreat/teweicun/internal/config"
	"github.com/redgreat/teweicun/pkg/logger"
	"go.uber.org/zap"
)

var Pool *pgxpool.Pool

// InitPostgres initializes the pgx connection pool
func InitPostgres(cfg config.DatabaseConfig) error {
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, sslmode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("unable to parse database config: %w", err)
	}
	timezone := cfg.Timezone
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = map[string]string{}
	}
	poolConfig.ConnConfig.RuntimeParams["timezone"] = timezone

	// Setup connection pool settings
	poolConfig.MaxConns = cfg.PoolMaxConns
	if poolConfig.MaxConns == 0 {
		poolConfig.MaxConns = 20
	}
	poolConfig.MinConns = cfg.PoolMinConns
	if poolConfig.MinConns == 0 {
		poolConfig.MinConns = 5
	}
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Pool = pool
	logger.Log.Info(
		"Database connection pool initialized",
		zap.String("db", cfg.DBName),
		zap.String("timezone", timezone),
	)

	return nil
}

// ClosePostgres closes the connection pool
func ClosePostgres() {
	if Pool != nil {
		Pool.Close()
		logger.Log.Info("Database connection pool closed")
	}
}
