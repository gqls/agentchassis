// FILE: cmd/agent-chassis/main.go
package main

import (
	"context"
	"flag"
	"fmt"
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

	// Initialize logger first
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// Get environment variables
	agentType := os.Getenv("AGENT_TYPE")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")

	// Log what we got from environment
	appLogger.Info("Environment variables",
		zap.String("AGENT_TYPE", agentType),
		zap.String("KAFKA_TOPIC", kafkaTopic),
		zap.String("KAFKA_CONSUMER_GROUP", consumerGroup))

	// Initialize Custom field as a map if it doesn't exist
	if cfg.Custom == nil {
		cfg.Custom = make(map[string]interface{})
	}

	// OVERRIDE WITH ENVIRONMENT VARIABLES
	if agentType != "" {
		cfg.Custom["agent_type"] = agentType
	} else {
		cfg.Custom["agent_type"] = "generic" // fallback
	}

	if kafkaTopic != "" {
		cfg.Custom["topic"] = kafkaTopic
	} else if agentType != "" {
		cfg.Custom["topic"] = fmt.Sprintf("system.agent.%s.process", agentType)
	} else {
		cfg.Custom["topic"] = "system.agent.generic.process"
	}

	if consumerGroup != "" {
		cfg.Custom["kafka_consumer_group"] = consumerGroup
	} else if agentType != "" {
		cfg.Custom["kafka_consumer_group"] = fmt.Sprintf("%s-group", agentType)
	} else {
		cfg.Custom["kafka_consumer_group"] = "generic-group"
	}

	// Log the actual config being used
	appLogger.Info("Starting agent with configuration",
		zap.String("agent_type", cfg.Custom["agent_type"].(string)),
		zap.String("topic", cfg.Custom["topic"].(string)),
		zap.String("consumer_group", cfg.Custom["kafka_consumer_group"].(string)))

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize agent with the modified config
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
