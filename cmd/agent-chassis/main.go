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
	"time"

	"github.com/gqls/agentchassis/pkg/buildinfo"
	"github.com/gqls/agentchassis/platform/agentbase"
	"github.com/gqls/agentchassis/platform/buildcapability"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/database"
	"github.com/gqls/agentchassis/platform/health"
	kafkaplatform "github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/logger"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	appLogger.Info("build provenance", zap.String("git_commit", buildinfo.GitCommit))

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

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health check server. The endpoints report real state (Kafka
	// reachability, agent started) — they were hardcoded 200s until
	// 2026-07-20, which made the spawner's probes decorative (bugs_open/003).
	healthPort := os.Getenv("HEALTH_PORT")
	if healthPort == "" {
		healthPort = "8080"
	}

	// Probe the SAME broker list the agent's Kafka clients use. They resolve it
	// from the environment via kafka.GetBrokers(); cfg.Infrastructure is only
	// the fallback. Probing cfg when the two disagree reports on a Kafka this
	// process is not talking to — proven in test 2026-07-20 (bugs_open/003).
	healthBrokers := kafkaplatform.GetBrokers()
	if len(healthBrokers) == 0 {
		healthBrokers = cfg.Infrastructure.KafkaBrokers
	}
	kafkaHealth := health.NewKafkaReachability(healthBrokers, appLogger)
	go kafkaHealth.Run(ctx)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", kafkaHealth.HandleHealth)
		mux.HandleFunc("/ready", kafkaHealth.HandleReady)
		mux.Handle("/metrics", promhttp.Handler())

		appLogger.Info("Starting health check server",
			zap.String("port", healthPort),
			zap.Duration("kafka_unhealthy_after", kafkaHealth.UnhealthyAfter))
		if err := http.ListenAndServe(":"+healthPort, mux); err != nil {
			appLogger.Fatal("Health server failed", zap.Error(err))
		}
	}()

	// bugs_open/040-kafka-dial: nothing in this binary ever served /metrics, yet
	// every spawned agent pod is annotated prometheus.io/port "9090" +
	// prometheus.io/path "/metrics" and declares a containerPort named "metrics"
	// on 9090 (spawn_actions.go). Prometheus has therefore been scraping a closed
	// port for the life of the fleet: verified 2026-07-26, the server held ZERO
	// ai_persona_* series, so every counter in platform/observability has been
	// dead on arrival — including the fetch_message error counter that was
	// supposed to make broker trouble visible.
	//
	// Served on its own listener rather than only on the health port because the
	// annotation names 9090 and the health port defaults to 8080; a metric on the
	// wrong port is indistinguishable from no metric at all.
	// Served via observability.MetricsServer rather than a second hand-rolled
	// mux here: that type existed for exactly this job and had no callers, and
	// standing up a rival beside it is how a platform ends up with two ways to
	// do one thing (council objection, round 1, corr 7abe1a57). It was fixed to
	// use its own mux and to stop serving a hardcoded-200 /health before being
	// called — see its doc comment.
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}
	if metricsPort != healthPort {
		go func() {
			appLogger.Info("Starting metrics server", zap.String("port", metricsPort))
			if err := observability.NewMetricsServer(metricsPort).Start(); err != nil {
				// Deliberately not Fatal: losing metrics must not take the agent
				// down. It is logged loudly so the cause is findable.
				appLogger.Error("Metrics server failed", zap.Error(err))
			}
		}()
	}

	// Initialize agent with the modified config
	agent, err := agentbase.New(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize agent", zap.Error(err))
	}

	// RFC_040 stage 2 — publish what THIS binary can do, so "is the new code
	// live yet?" stops being a question only `kubectl exec` can answer.
	//
	// This is the one place in the tree that can see both registries and has
	// config in hand. agentbase was the obvious alternative and is the wrong
	// one: it imports neither registry, so the lists would have to be plumbed
	// through it anyway.
	//
	// DELIBERATELY BEST-EFFORT AND NEVER FATAL. A capability registry that can
	// stop the chassis starting is a worse bargain than the problem it solves,
	// so every failure below is logged and stepped over — including the
	// connection itself. It is a short-lived connection of its own rather than
	// the agent's, because the agent's handle is private to it and widening
	// that for a write nothing depends on would be the larger change.
	recordBuildCapabilities(ctx, cfg, appLogger)

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Run agent in goroutine
	/*errCh := make(chan error, 1)
	go func() {
		if err := agent.Run(); err != nil {
			errCh <- err
		}
	}()*/

	// Run agent in goroutine
	kafkaHealth.SetStarted()
	errCh := make(chan error, 1)
	go func() {
		errCh <- agent.Run() // sends nil on clean exit, error on failure
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigCh:
		appLogger.Info("Shutdown signal received")
		cancel()
		if err := agent.Shutdown(); err != nil {
			appLogger.Error("Shutdown error", zap.Error(err))
		}
		/*case err := <-errCh:
		appLogger.Error("Agent failed", zap.Error(err))
		cancel()
		agent.Shutdown()*/
	case err := <-errCh:
		if err != nil {
			appLogger.Error("Agent failed", zap.Error(err))
		} else {
			appLogger.Info("Agent exited cleanly")
		}
		cancel()
		agent.Shutdown()
	}

	appLogger.Info("Agent stopped")
}

// recordBuildCapabilities publishes this binary's enumerable registries to
// service_binary_capabilities (migration 503, RFC_040 stage 2).
//
// Every arm returns quietly on failure: a chassis that cannot record its
// capabilities must still serve. The log lines are Warn rather than Error
// because an absent row is a *diagnostic* gap, not a service fault — but they
// name the cause, because a silently missing row is exactly the ambiguity this
// mechanism exists to remove.
func recordBuildCapabilities(ctx context.Context, cfg *config.ServiceConfig, appLogger *zap.Logger) {
	podName := os.Getenv("HOSTNAME")
	if podName == "" {
		appLogger.Warn("build capabilities not recorded: HOSTNAME is empty, so the row could not be attributed to a pod")
		return
	}
	if cfg.Infrastructure.ClientsDatabase.Host == "" {
		appLogger.Warn("build capabilities not recorded: no clients database configured")
		return
	}

	db, err := database.NewStdlibConnection(ctx, cfg.Infrastructure.ClientsDatabase, appLogger)
	if err != nil {
		appLogger.Warn("build capabilities not recorded: could not open a clients-db connection", zap.Error(err))
		return
	}
	defer db.Close()

	checkNames := checks.Names()
	actionNames := actions.ListActions()

	if err := buildcapability.Record(ctx, db, "agent-chassis", podName,
		buildcapability.Set{Kind: "discovery_check", Names: checkNames},
		buildcapability.Set{Kind: "action", Names: actionNames},
	); err != nil {
		appLogger.Warn("build capabilities not recorded", zap.Error(err))
		return
	}

	appLogger.Info("build capabilities recorded",
		zap.String("service", "agent-chassis"),
		zap.String("pod", podName),
		zap.String("git_commit", buildinfo.GitCommit),
		zap.Int("discovery_checks", len(checkNames)),
		zap.Int("actions", len(actionNames)))

	// Heartbeat. Rows are pruned after buildcapability.RetentionWindow, so a
	// LONG-LIVED pod must keep saying it is alive or its own rows would be swept
	// while it is still serving — staleness in reverse, and just as misleading.
	//
	// Ephemeral per-job pods (agent-page-rerender-*, agent-page-build-handler-*,
	// …) run this same binary and will start this ticker too. That is harmless
	// and correct: they die long inside the window and are SUPPOSED to age out.
	//
	// A fresh short-lived connection per beat, not a held one. Once every 15
	// minutes per pod is nothing, and it keeps the property the startup write
	// already has — this code never holds a connection open against the pooler.
	go func() {
		ticker := time.NewTicker(buildcapability.TouchInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tdb, err := database.NewStdlibConnection(ctx, cfg.Infrastructure.ClientsDatabase, appLogger)
				if err != nil {
					continue // best-effort, as everything on this path is
				}
				if err := buildcapability.Touch(ctx, tdb, "agent-chassis", podName); err != nil {
					appLogger.Warn("build capability heartbeat failed", zap.Error(err))
				}
				_ = tdb.Close()
			}
		}
	}()
}
