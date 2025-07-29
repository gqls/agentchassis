package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/internal/agents/contentcreator"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "configs/content-creator-agent.yaml", "Path to config file")
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

	// Create agent with database support if configured
	agent, err := contentcreator.NewAgent(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize content creator agent", zap.Error(err))
	}

	// Start health endpoint
	agent.StartHealthServer("9090")

	// Start the agent's main run loop in a goroutine
	go func() {
		if err := agent.Run(); err != nil {
			appLogger.Error("Content creator agent failed to run", zap.Error(err))
			cancel()
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutdown signal received, shutting down content creator agent...")

	cancel()
	time.Sleep(2 * time.Second)
	appLogger.Info("Content creator agent service stopped.")
}
