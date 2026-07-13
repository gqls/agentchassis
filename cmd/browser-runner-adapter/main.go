// FILE: cmd/browser-runner-adapter/main.go
//
// Entry point for the browser-runner adapter (Tier-4 headless acceptance
// runner, P0). Mirrors cmd/analyser-adapter/main.go exactly — the standard
// signal-handler pattern: Run() in a goroutine feeding runComplete, select on
// signal-or-completion, then cancel + Shutdown (sync.Once-guarded) + bounded
// wait for Run to exit.

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/internal/adapters/browserrunner"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "configs/browser-runner-adapter.yaml", "Path to config file")
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

	appLogger.Info("Starting browser-runner adapter",
		zap.String("service", cfg.ServiceInfo.Name),
		zap.String("version", cfg.ServiceInfo.Version),
		zap.String("environment", cfg.ServiceInfo.Environment),
		zap.String("server port", cfg.Server.Port),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	adapter, err := browserrunner.NewAdapter(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize browser-runner adapter", zap.Error(err))
	}

	healthPort := cfg.Server.Port
	if healthPort == "" {
		healthPort = "8080"
	}
	adapter.StartHealthServer(healthPort)

	runComplete := make(chan error, 1)

	go func() {
		appLogger.Info("Starting adapter message processing")
		err := adapter.Run()
		if err != nil {
			appLogger.Info("Adapter Run() returned", zap.Error(err))
		} else {
			appLogger.Info("Adapter Run() completed normally")
		}
		runComplete <- err
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		appLogger.Info("Shutdown signal received", zap.String("signal", sig.String()))
	case err := <-runComplete:
		if err != nil {
			appLogger.Error("Adapter Run() failed unexpectedly", zap.Error(err))
		} else {
			appLogger.Warn("Adapter Run() completed unexpectedly without error")
		}
	}

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

	appLogger.Info("Browser-runner adapter stopped successfully")
}
