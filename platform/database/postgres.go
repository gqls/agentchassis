// FILE: platform/database/postgres.go
package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gqls/agentchassis/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PoolConfig holds connection pool settings
type PoolConfig struct {
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultPoolConfig returns sensible defaults for agent workloads
// Agents are typically short-lived, so we use conservative settings
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConns:        3,                // Reduced from 10 - most agents only need 1-2 concurrent queries
		MinConns:        0,                // Changed from 2 - don't hold connections when idle
		MaxConnLifetime: 30 * time.Minute, // Reduced from 1 hour
		MaxConnIdleTime: 5 * time.Minute,  // Reduced from 30 minutes
	}
}

// PoolConfigFromEnv reads pool configuration from environment variables
// This allows tuning without code changes
func PoolConfigFromEnv() PoolConfig {
	cfg := DefaultPoolConfig()

	if v := os.Getenv("DB_POOL_MAX_CONNS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			cfg.MaxConns = int32(n)
		}
	}

	if v := os.Getenv("DB_POOL_MIN_CONNS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n >= 0 {
			cfg.MinConns = int32(n)
		}
	}

	if v := os.Getenv("DB_POOL_MAX_CONN_LIFETIME_MINUTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.MaxConnLifetime = time.Duration(n) * time.Minute
		}
	}

	if v := os.Getenv("DB_POOL_MAX_CONN_IDLE_MINUTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.MaxConnIdleTime = time.Duration(n) * time.Minute
		}
	}

	return cfg
}

// NewPostgresConnection creates a new PostgreSQL connection pool with retry logic
func NewPostgresConnection(ctx context.Context, dbCfg config.DatabaseConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	return NewPostgresConnectionWithPool(ctx, dbCfg, logger, PoolConfigFromEnv())
}

// NewPostgresConnectionWithPool creates a connection pool with explicit pool settings
func NewPostgresConnectionWithPool(ctx context.Context, dbCfg config.DatabaseConfig, logger *zap.Logger, poolCfg PoolConfig) (*pgxpool.Pool, error) {
	logger.Info("Database configuration",
		zap.String("host", dbCfg.Host),
		zap.Int("port", dbCfg.Port),
		zap.String("user", dbCfg.User),
		zap.String("dbname", dbCfg.DBName),
		zap.String("password_env_var", dbCfg.PasswordEnvVar),
		zap.String("sslmode", dbCfg.SSLMode),
		zap.Int32("pool_max_conns", poolCfg.MaxConns),
		zap.Int32("pool_min_conns", poolCfg.MinConns),
		zap.Duration("pool_max_lifetime", poolCfg.MaxConnLifetime),
		zap.Duration("pool_max_idle_time", poolCfg.MaxConnIdleTime))

	password := os.Getenv(dbCfg.PasswordEnvVar)
	if password == "" {
		return nil, fmt.Errorf("database password environment variable %s is not set", dbCfg.PasswordEnvVar)
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		dbCfg.User, password, dbCfg.Host, dbCfg.Port, dbCfg.DBName, dbCfg.SSLMode)

	pgxPoolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres connection string: %w", err)
	}

	// Apply pool configuration
	pgxPoolConfig.MaxConns = poolCfg.MaxConns
	pgxPoolConfig.MinConns = poolCfg.MinConns
	pgxPoolConfig.MaxConnLifetime = poolCfg.MaxConnLifetime
	pgxPoolConfig.MaxConnIdleTime = poolCfg.MaxConnIdleTime

	// Health check configuration - reduces stale connection issues
	pgxPoolConfig.HealthCheckPeriod = 30 * time.Second

	var pool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, pgxPoolConfig)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			pingErr := pool.Ping(pingCtx)
			cancel() // Don't defer in loop
			if pingErr == nil {
				logger.Info("Successfully connected to PostgreSQL database.",
					zap.String("database", dbCfg.DBName),
					zap.Int32("pool_max_conns", poolCfg.MaxConns))
				return pool, nil
			}
			err = pingErr
			pool.Close() // Close failed pool before retry
		}
		logger.Warn("Failed to connect to PostgreSQL, retrying...",
			zap.Int("attempt", i+1),
			zap.String("database", dbCfg.DBName),
			zap.Error(err),
		)
		time.Sleep(time.Duration(i+1) * 2 * time.Second) // Exponential backoff
	}

	return nil, fmt.Errorf("failed to connect to postgres after multiple attempts: %w", err)
}
