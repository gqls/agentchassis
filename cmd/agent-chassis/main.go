// FILE: cmd/agent-chassis/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gqls/agentchassis/platform/agentbase"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/logger"
	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	/*	log.Printf("=== Configuration Debug ===")
		log.Printf("Clients DB Config:")
		log.Printf("  Host: %s", cfg.Infrastructure.ClientsDatabase.Host)
		log.Printf("  Port: %d", cfg.Infrastructure.ClientsDatabase.Port)
		log.Printf("  User: %s", cfg.Infrastructure.ClientsDatabase.User)
		log.Printf("  DBName: %s", cfg.Infrastructure.ClientsDatabase.DBName)
		log.Printf("  PasswordEnvVar: '%s'", cfg.Infrastructure.ClientsDatabase.PasswordEnvVar)
		log.Printf("  SSLMode: %s", cfg.Infrastructure.ClientsDatabase.SSLMode)*/

	// ###### DEBUG ###### //
	// Initialize logger first
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// Get environment variables
	agentType := os.Getenv("AGENT_TYPE")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaTopics := os.Getenv("KAFKA_TOPICS")
	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	parentsResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")
	thisAgentsResponsesTopic := os.Getenv("RESPONSES_TOPIC")
	thisAgentsRequestsTopic := os.Getenv("REQUESTS_TOPIC")

	// Log what we got from environment
	appLogger.Info("Environment variables",
		zap.String("AGENT_TYPE", agentType),
		zap.String("KAFKA_TOPIC", kafkaTopic),
		zap.String("KAFKA_TOPICS", kafkaTopics),
		zap.String("PARENT_RESPONSES_TOPIC", parentsResponsesTopic),
		zap.String("RESPONSES_TOPIC", thisAgentsResponsesTopic),
		zap.String("REQUESTS_TOPIC", thisAgentsRequestsTopic),
		zap.String("KAFKA_CONSUMER_GROUP", consumerGroup))

	// Log ALL database-related environment variables
	appLogger.Info("Database environment variables check",
		zap.String("CLIENTS_DB_PASSWORD_exists", fmt.Sprintf("%v", os.Getenv("CLIENTS_DB_PASSWORD") != "")),
		zap.String("TEMPLATES_DB_PASSWORD_exists", fmt.Sprintf("%v", os.Getenv("TEMPLATES_DB_PASSWORD") != "")),
		zap.String("AUTH_DB_PASSWORD_exists", fmt.Sprintf("%v", os.Getenv("AUTH_DB_PASSWORD") != "")),
		zap.String("CLIENTS_DATABASE_URL", os.Getenv("CLIENTS_DATABASE_URL")),
		zap.String("TEMPLATES_DATABASE_URL", os.Getenv("TEMPLATES_DATABASE_URL")),
		zap.Int("CLIENTS_DB_PASSWORD_len", len(os.Getenv("CLIENTS_DB_PASSWORD"))),
		zap.Int("TEMPLATES_DB_PASSWORD_len", len(os.Getenv("TEMPLATES_DB_PASSWORD"))))

	// Log the loaded configuration
	appLogger.Info("Loaded configuration",
		zap.Any("infrastructure", cfg.Infrastructure),
		zap.Any("custom", cfg.Custom))

	// Check if the environment variable exists
	if cfg.Infrastructure.ClientsDatabase.PasswordEnvVar != "" {
		envValue := os.Getenv(cfg.Infrastructure.ClientsDatabase.PasswordEnvVar)
		appLogger.Info("Environment variable check",
			zap.String("env_var", cfg.Infrastructure.ClientsDatabase.PasswordEnvVar),
			zap.Bool("exists", envValue != ""),
			zap.Int("length", len(envValue)),
		)
	} else {
		appLogger.Warn("PasswordEnvVar is empty!",
			zap.String("env_var", cfg.Infrastructure.ClientsDatabase.PasswordEnvVar),
		)
	}
	appLogger.Info("===================")

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
		cfg.Custom["topic"] = fmt.Sprintf("system.agent.%s.requests", agentType)
	} else {
		cfg.Custom["topic"] = "system.agent.generic.requests"
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

	// Start health check server
	healthPort := os.Getenv("HEALTH_PORT")
	if healthPort == "" {
		healthPort = "8080"
	}

	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			// Could add actual readiness checks here
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("READY"))
		})

		appLogger.Info("Starting health check server", zap.String("port", healthPort))
		if err := http.ListenAndServe(":"+healthPort, mux); err != nil {
			appLogger.Fatal("Health server failed", zap.Error(err))
		}
	}()

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
