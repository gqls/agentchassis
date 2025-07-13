// FILE: cmd/agent-chassis/main.go
// This is the entrypoint for our generic agent service.
package main

import (
	"context"
	"flag"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/platform/agentbase"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
)

func main() {
	// 1. Load the service's static configuration from a file.
	configPath := flag.String("config", "configs/agent-chassis.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize the logger.
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// 3. Create a cancellable context for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Initialize the base agent framework from the platform library.
	// This one call sets up all connections (DB, Kafka) and the main consumer loop logic.
	agent, err := agentbase.New(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize agent base", zap.Error(err))
	}

	// After creating the agent
	agent, err = agentbase.New(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize agent base", zap.Error(err))
	}

	// Start health endpoint
	agent.StartHealthServer("9090")

	// 5. Start the agent's main run loop in a goroutine.
	go func() {
		if err := agent.Run(); err != nil {
			appLogger.Error("Agent chassis failed to run", zap.Error(err))
			cancel() // Trigger shutdown if the run loop exits with an error
		}
	}()

	// 6. Wait for a shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutdown signal received, shutting down agent chassis...")

	// Trigger the context cancellation, which will gracefully stop the agent's Run loop.
	cancel()

	// Allow a moment for graceful shutdown to complete.
	time.Sleep(2 * time.Second)
	appLogger.Info("Agent chassis service stopped.")
}
