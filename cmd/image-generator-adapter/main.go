// FILE: cmd/image-generator-adapter/main.go
// This is the entrypoint for our specialized adapter service.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/ai-persona-system/internal/adapters/imagegenerator" // The adapter's specific logic
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/gqls/ai-persona-system/platform/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. Load the service's static configuration.
	// This adapter would have its own config file specifying its topics and the external API details.
	configPath := flag.String("config", "configs/image-adapter.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize the platform logger.
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// 3. Create a cancellable context for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Initialize the adapter.
	// This is where we pass dependencies like the logger and config.
	adapter, err := imagegenerator.NewAdapter(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize image generator adapter", zap.Error(err))
	}

	// 5. Start the adapter's main run loop in a goroutine.
	go func() {
		if err := adapter.Run(); err != nil {
			appLogger.Error("Image generator adapter failed to run", zap.Error(err))
			cancel() // Trigger shutdown
		}
	}()

	// 6. Wait for a shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutdown signal received, shutting down image generator adapter...")

	cancel()

	// Allow a moment for graceful shutdown.
	time.Sleep(2 * time.Second)
	appLogger.Info("Image generator adapter stopped.")
}
