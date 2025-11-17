// cmd/git-adapter/main.go - Fixed to handle Run() properly

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/internal/adapters/git"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/git-adapter.yaml", "Path to config file")
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
	appLogger.Info("Starting git adapter",
		zap.String("service", cfg.ServiceInfo.Name),
		zap.String("version", cfg.ServiceInfo.Version),
		zap.String("environment", cfg.ServiceInfo.Environment),
		zap.String("server port", cfg.Server.Port),
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the adapter
	adapter, err := git.NewAdapter(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize git adapter", zap.Error(err))
	}

	// Start health check server
	healthPort := cfg.Server.Port
	if healthPort == "" {
		healthPort = "8080"
	}
	adapter.StartHealthServer(healthPort)
	appLogger.Info("Health server started", zap.String("port", healthPort))

	// Channel to signal when Run() completes
	runComplete := make(chan error, 1)

	// Start the adapter's main run loop in a goroutine
	go func() {
		appLogger.Info("Starting adapter message processing")
		err := adapter.Run()

		// Run() should only return when shutting down
		// Log the return but don't treat it as an error unless it's unexpected
		if err != nil {
			appLogger.Info("Adapter Run() returned", zap.Error(err))
		} else {
			appLogger.Info("Adapter Run() completed normally")
		}

		runComplete <- err
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either shutdown signal or Run() to complete unexpectedly
	select {
	case sig := <-quit:
		// Received shutdown signal
		appLogger.Info("Shutdown signal received",
			zap.String("signal", sig.String()),
		)

	case err := <-runComplete:
		// Run() completed unexpectedly
		if err != nil {
			appLogger.Error("Adapter Run() failed unexpectedly", zap.Error(err))
		} else {
			appLogger.Warn("Adapter Run() completed unexpectedly without error")
		}
	}

	// Initiate graceful shutdown
	appLogger.Info("Initiating graceful shutdown")

	// Cancel the context to signal shutdown to adapter
	cancel()

	// Give adapter time to shut down cleanly
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown adapter
	adapter.Shutdown()

	// Wait for Run() to complete or timeout
	select {
	case <-runComplete:
		appLogger.Info("Adapter Run() completed during shutdown")
	case <-shutdownCtx.Done():
		appLogger.Warn("Shutdown timeout exceeded waiting for Run() to complete")
	}

	appLogger.Info("Git adapter stopped successfully")
}

// Alternative simpler main if you prefer
func mainSimple() {
	configPath := flag.String("config", "configs/git-adapter.yaml", "Path to config file")
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

	adapter, err := git.NewAdapter(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize git adapter", zap.Error(err))
	}

	adapter.StartHealthServer("9090")

	// Run adapter in goroutine - don't care about its return
	go func() {
		adapter.Run() // Ignore return - it runs until shutdown
	}()

	// Just wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down...")
	cancel()
	adapter.Shutdown()
	time.Sleep(2 * time.Second)
	appLogger.Info("Stopped")
}
