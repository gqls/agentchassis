// FILE: platform/database/mysql.go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gqls/agentchassis/platform/config"
	"go.uber.org/zap"
)

// NewMySQLConnection creates a new MySQL database connection pool with retry logic
func NewMySQLConnection(ctx context.Context, dbCfg config.DatabaseConfig, logger *zap.Logger) (*sql.DB, error) {
	password := os.Getenv(dbCfg.PasswordEnvVar)
	if password == "" {
		return nil, fmt.Errorf("database password environment variable %s is not set", dbCfg.PasswordEnvVar)
	}

	// DSN format for MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		dbCfg.User, password, dbCfg.Host, dbCfg.Port, dbCfg.DBName)

	var db *sql.DB
	var err error

	// Retry loop for initial connection
	for i := 0; i < 5; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err = db.PingContext(pingCtx); err == nil {
				// Set connection pool parameters
				db.SetMaxOpenConns(10)
				db.SetMaxIdleConns(5)
				db.SetConnMaxLifetime(time.Hour)

				logger.Info("Successfully connected to MySQL database.", zap.String("database", dbCfg.DBName))
				return db, nil
			}
		}
		logger.Warn("Failed to connect to MySQL, retrying...",
			zap.Int("attempt", i+1),
			zap.String("database", dbCfg.DBName),
			zap.Error(err),
		)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to mysql after multiple attempts: %w", err)
}
