// FILE: platform/database/postgres.go
package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// NewPostgresConnection creates a new PostgreSQL connection pool with retry logic
func NewPostgresConnection(ctx context.Context, dbCfg config.DatabaseConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	password := os.Getenv(dbCfg.PasswordEnvVar)
	if password == "" {
		return nil, fmt.Errorf("database password environment variable %s is not set", dbCfg.PasswordEnvVar)
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		dbCfg.User, password, dbCfg.Host, dbCfg.Port, dbCfg.DBName, dbCfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres connection string: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err = pool.Ping(pingCtx); err == nil {
				logger.Info("Successfully connected to PostgreSQL database.", zap.String("database", dbCfg.DBName))
				return pool, nil
			}
		}
		logger.Warn("Failed to connect to PostgreSQL, retrying...",
			zap.Int("attempt", i+1),
			zap.String("database", dbCfg.DBName),
			zap.Error(err),
		)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to postgres after multiple attempts: %w", err)
}
