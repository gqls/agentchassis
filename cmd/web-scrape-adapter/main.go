// cmd/webscrape-adapter/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gqls/agentchassis/internal/adapters/webscrape"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logging"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger := logging.NewLogger("webscrape-adapter")
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadServiceConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Create and run adapter
	adapter, err := webscrape.NewAdapter(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create adapter", zap.Error(err))
	}

	logger.Info("Starting webscrape adapter")
	if err := adapter.Run(); err != nil {
		logger.Error("Adapter error", zap.Error(err))
	}
}
