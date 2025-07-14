// FILE: cmd/agent-chassis/main.go (updated)
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gqls/agentchassis/platform/agentbase"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	configPath := flag.String("config", "configs/agent-chassis.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize agent
	agent, err := agentbase.New(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize agent", zap.Error(err))
	}

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Run agent in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := agent.Run(); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigCh:
		appLogger.Info("Shutdown signal received")
		cancel()
		if err := agent.Shutdown(); err != nil {
			appLogger.Error("Shutdown error", zap.Error(err))
		}
	case err := <-errCh:
		appLogger.Error("Agent failed", zap.Error(err))
		cancel()
		agent.Shutdown()
	}

	appLogger.Info("Agent stopped")
}
