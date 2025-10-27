// cmd/image-generator-adapter/main.go
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
	// Parse command line flags
	configPath := flag.String("config", "configs/image-adapter.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
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

	// Log startup information
	appLogger.Info("Starting image generator adapter",
		zap.String("service", cfg.ServiceInfo.Name),
		zap.String("version", cfg.ServiceInfo.Version),
		zap.String("environment", cfg.ServiceInfo.Environment),
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the adapter
	adapter, err := imagegenerator.NewDynamicImageAdapter(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize image generator adapter", zap.Error(err))
	}

	// Start health check server
	healthPort := cfg.Server.Port
	if healthPort == "" {
		healthPort = "9090"
	}
	adapter.StartHealthServer(healthPort)
	appLogger.Info("Health server started", zap.String("port", healthPort))

	// Start the adapter's main run loop in a goroutine
	go func() {
		appLogger.Info("Starting adapter message processing")
		if err := adapter.Run(); err != nil {
			appLogger.Error("Image generator adapter error", zap.Error(err))
			cancel()
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	appLogger.Info("Shutdown signal received, shutting down image generator adapter",
		zap.String("signal", sig.String()),
	)

	// Graceful shutdown
	cancel()

	// Give components time to clean up
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown adapter
	adapter.Shutdown()

	// Wait for shutdown or timeout
	select {
	case <-shutdownCtx.Done():
		appLogger.Warn("Shutdown timeout exceeded")
	case <-time.After(2 * time.Second):
		appLogger.Info("Graceful shutdown completed")
	}

	appLogger.Info("Image generator adapter stopped")
}
