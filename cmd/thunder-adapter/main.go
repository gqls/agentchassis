// FILE: cmd/thunder-adapter/main.go
//
// Thunder Compute Adapter entry point. Phase 2 skeleton (no Thunder API
// or B2 calls yet — see internal/adapters/thunder/adapter.go).
//
// Matches the pattern of cmd/git-adapter/main.go and cmd/web-scrape-adapter/main.go.

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/internal/adapters/thunder"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "configs/thunder-adapter.yaml", "Path to config file")
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

	appLogger.Info("Starting thunder adapter",
		zap.String("service", cfg.ServiceInfo.Name),
		zap.String("version", cfg.ServiceInfo.Version),
		zap.String("environment", cfg.ServiceInfo.Environment),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	adapter, err := thunder.NewAdapter(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize thunder adapter", zap.Error(err))
	}

	healthPort := cfg.Server.Port
	if healthPort == "" {
		healthPort = "8080"
	}
	adapter.StartHealthServer(healthPort)

	// Run adapter in goroutine
	runComplete := make(chan error, 1)
	go func() {
		appLogger.Info("Thunder adapter Run() starting")
		err := adapter.Run()
		if err != nil {
			appLogger.Info("Thunder adapter Run() returned", zap.Error(err))
		} else {
			appLogger.Info("Thunder adapter Run() completed normally")
		}
		runComplete <- err
	}()

	// Wait for shutdown signal or unexpected Run() exit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		appLogger.Info("Shutdown signal received",
			zap.String("signal", sig.String()))
	case err := <-runComplete:
		if err != nil {
			appLogger.Error("Adapter Run() failed unexpectedly", zap.Error(err))
		} else {
			appLogger.Warn("Adapter Run() completed unexpectedly without error")
		}
	}

	// Graceful shutdown
	appLogger.Info("Initiating graceful shutdown")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	adapter.Shutdown()

	select {
	case <-runComplete:
		appLogger.Info("Adapter Run() completed during shutdown")
	case <-shutdownCtx.Done():
		appLogger.Warn("Shutdown timeout exceeded waiting for Run() to complete")
	}

	appLogger.Info("Thunder adapter stopped")
}
