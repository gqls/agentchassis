// FILE: cmd/image-generator-adapter/main.go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/internal/adapters/imagegenerator"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "configs/image-adapter.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	adapter, err := imagegenerator.NewAdapter(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize image generator adapter", zap.Error(err))
	}

	// Start health endpoint HERE
	adapter.StartHealthServer("9090")

	// Start the adapter's main run loop in a goroutine
	go func() {
		if err := adapter.Run(); err != nil {
			appLogger.Error("Image generator adapter failed to run", zap.Error(err))
			cancel()
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutdown signal received, shutting down image generator adapter...")

	cancel()
	time.Sleep(2 * time.Second)
	appLogger.Info("Image generator adapter stopped.")
}
