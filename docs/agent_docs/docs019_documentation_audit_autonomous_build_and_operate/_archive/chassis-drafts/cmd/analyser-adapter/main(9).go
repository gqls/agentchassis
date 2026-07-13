// FILE: cmd/analyser-adapter/main.go
//
// DRAFT for the agent-chassis repo (module github.com/gqls/agentchassis).
// Does not compile in the contextkit container — built/deployed in your env.
//
// Entry point for the analyser adapter. Standard signal-handler pattern
// (035 §2.16): load config → logger → NewAdapter → StartHealthServer →
// signal handler calling Shutdown → Run → Shutdown (idempotent).
//
// RECONCILE against your tree before building (copy from an existing adapter
// main, e.g. cmd/thunder-adapter/main.go or cmd/git-adapter/main.go):
//   1. The exact config loader call — assumed config.Load(path) here.
//   2. The logger constructor — if the platform has a shared logger helper
//      (level from cfg.Logging.Level), use that instead of the inline zap
//      construction below.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gqls/agentchassis/internal/adapters/analyser"
	"github.com/gqls/agentchassis/platform/config"
)

func main() {
	configPath := flag.String("config", "/app/configs/analyser-adapter.yaml", "path to the service config YAML")
	flag.Parse()

	cfg, err := config.Load(*configPath) // RECONCILE: exact loader name
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config %s: %v\n", *configPath, err)
		os.Exit(1)
	}

	logger, err := newLogger(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	adapter, err := analyser.NewAdapter(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize analyser adapter", zap.Error(err))
	}

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}
	adapter.StartHealthServer(port)

	// Signal handler — Shutdown is sync.Once-guarded, so calling it here and
	// again after Run returns is safe.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		adapter.Shutdown()
	}()

	if err := adapter.Run(); err != nil {
		logger.Error("Adapter run loop exited with error", zap.Error(err))
	}
	adapter.Shutdown()
}

// newLogger builds a production JSON zap logger at the configured level.
// RECONCILE: replace with the platform's shared logger constructor if other
// adapter mains use one.
func newLogger(level string) (*zap.Logger, error) {
	lvl := zapcore.InfoLevel
	if level != "" {
		if err := lvl.UnmarshalText([]byte(level)); err != nil {
			return nil, fmt.Errorf("invalid logging.level %q: %w", level, err)
		}
	}
	zcfg := zap.NewProductionConfig()
	zcfg.Level = zap.NewAtomicLevelAt(lvl)
	return zcfg.Build()
}
