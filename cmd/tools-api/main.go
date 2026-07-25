package main

import (
	"context"
	"log"

	"github.com/gqls/agentchassis/internal/tools-api/api"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := db.NewDBPool(ctx)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer pool.Close()

	r := api.NewRouter(pool, cfg)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
