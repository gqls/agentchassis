// FILE: cmd/core-manager/main.go
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/ai-persona-system/internal/api"
	// --- Use platform packages ---
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/gqls/ai-persona-system/platform/database"
	"github.com/gqls/ai-persona-system/platform/logger"
	"go.uber.org/zap"
)

func main() {
	// --- Step 1: Load Configuration using the Platform Library ---
	configPath := flag.String("config", "configs/core-manager.yaml", "Path to config file")
	flag.Parse()

	// Use the standardized platform loader.
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to load configuration via platform loader: %v", err)
	}

	// --- Step 2: Initialize Logger using the Platform Library ---
	// Use the standardized platform logger.
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Core Manager Service starting",
		zap.String("service_name", cfg.ServiceInfo.Name),
		zap.String("version", cfg.ServiceInfo.Version),
		zap.String("environment", cfg.ServiceInfo.Environment),
		zap.String("log_level", cfg.Logging.Level),
	)

	// Create a main context that can be cancelled for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Step 3: Initialize Database Connections using the Platform Library ---
	// No more local helper function; we use the reusable platform function.

	// 3a. Create connection pool for the Templates Database.
	templatesPool, err := database.NewPostgresConnection(ctx, cfg.Infrastructure.TemplatesDatabase, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize templates database connection", zap.Error(err))
	}
	defer templatesPool.Close()

	// 3b. Create connection pool for the Clients Database.
	clientsPool, err := database.NewPostgresConnection(ctx, cfg.Infrastructure.ClientsDatabase, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize clients database connection", zap.Error(err))
	}
	defer clientsPool.Close()

	// --- Step 4: Initialize and Start the API Server ---
	// The server now receives the already-initialized dependencies.
	apiServer, err := api.NewServer(ctx, cfg, appLogger, templatesPool, clientsPool)
	if err != nil {
		appLogger.Fatal("Failed to initialize API server", zap.Error(err))
	}

	// Run the server in a goroutine so it doesn't block.
	go func() {
		appLogger.Info("Starting HTTP server...", zap.String("address", apiServer.Address()))
		if err := apiServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("API server ListenAndServe failed unexpectedly", zap.Error(err))
			cancel() // Trigger shutdown on server error
		}
	}()

	// --- Step 5: Handle Graceful Shutdown ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	receivedSignal := <-sigCh
	appLogger.Info("Shutdown signal received.", zap.String("signal", receivedSignal.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	appLogger.Info("Calling API Server Shutdown...")
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Error during API server graceful shutdown", zap.Error(err))
	} else {
		appLogger.Info("API server shutdown complete.")
	}

	appLogger.Info("Core Manager Service fully stopped.")
}
