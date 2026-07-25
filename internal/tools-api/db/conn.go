package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBPool constructs a pgxpool.Pool from the DATABASE_URL environment variable.
func NewDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	return pgxpool.New(ctx, dsn)
}
