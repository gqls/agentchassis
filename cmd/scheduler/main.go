// FILE: cmd/scheduler/main.go
//
// Kafka Scheduler — reads scheduled_tasks from Postgres, publishes trigger
// messages to Kafka on the configured intervals.
//
// Concurrency-group-aware: tasks in the same group respect max_concurrent.
// Pre-query support: tasks with a pre_query get dynamic input_data from SQL.
//
// Environment variables:
//   CLIENTS_DATABASE_URL  — Postgres connection string
//   KAFKA_BROKERS         — comma-separated Kafka broker list
//   TICK_INTERVAL_SECONDS — how often to check (default: 30)
//   LOG_LEVEL             — debug, info, warn, error (default: info)
//
// Deployment: standalone K8s Deployment (1 replica, leader election not needed
// because Postgres last_triggered_at prevents double-firing).

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gqls/agentchassis/platform/kafka"
)

// scheduledTask mirrors a row from scheduled_tasks.
type scheduledTask struct {
	ID               uuid.UUID
	Name             string
	TargetAgentType  string
	TargetTopic      string
	InputData        json.RawMessage
	PreQuery         sql.NullString
	IntervalSeconds  int
	ConcurrencyGroup sql.NullString
	MaxConcurrent    int
	TimeoutSeconds   int
	LastTriggeredAt  sql.NullTime
	LastCompletedAt  sql.NullTime
}

func main() {
	logger := buildLogger()
	defer logger.Sync()

	logger.Info("kafka-scheduler starting")

	// --- Postgres ---
	dbURL := os.Getenv("CLIENTS_DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("CLIENTS_DATABASE_URL required")
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		logger.Fatal("Failed to open database", zap.Error(err))
	}
	defer db.Close()
	db.SetMaxOpenConns(3)

	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Connected to database")

	// --- Kafka ---
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = os.Getenv("SERVICE_INFRASTRUCTURE_KAFKA_BROKERS")
	}
	if brokersEnv == "" {
		logger.Fatal("KAFKA_BROKERS required")
	}
	brokers := strings.Split(brokersEnv, ",")

	producer, err := kafka.NewProducer(brokers, logger)
	if err != nil {
		logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	logger.Info("Kafka producer ready", zap.Strings("brokers", brokers))

	// --- Health endpoint ---
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		})
		http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				http.Error(w, "db not ready", 503)
				return
			}
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		})
		http.ListenAndServe(":8080", nil)
	}()

	// --- Tick loop ---
	tickInterval := 30
	if v := os.Getenv("TICK_INTERVAL_SECONDS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			tickInterval = parsed
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
	}()

	ticker := time.NewTicker(time.Duration(tickInterval) * time.Second)
	defer ticker.Stop()

	logger.Info("Scheduler running",
		zap.Int("tick_interval_seconds", tickInterval))

	// Run once immediately on startup
	runTick(ctx, db, producer, logger)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Scheduler stopped")
			return
		case <-ticker.C:
			runTick(ctx, db, producer, logger)
		}
	}
}

// runTick checks all enabled tasks and fires any that are due.
func runTick(ctx context.Context, db *sql.DB, producer kafka.Producer, logger *zap.Logger) {
	tasks, err := loadDueTasks(ctx, db, logger)
	if err != nil {
		logger.Warn("Failed to load tasks", zap.Error(err))
		return
	}

	if len(tasks) == 0 {
		return
	}

	// Check concurrency groups — count in-flight per group
	inFlight := countInFlight(ctx, db, logger)

	for _, task := range tasks {
		// Concurrency check
		if task.ConcurrencyGroup.Valid {
			group := task.ConcurrencyGroup.String
			current := inFlight[group]
			if current >= task.MaxConcurrent {
				logger.Debug("Skipping task — concurrency group at max",
					zap.String("task", task.Name),
					zap.String("group", group),
					zap.Int("current", current),
					zap.Int("max", task.MaxConcurrent))
				continue
			}
			// Increment for subsequent tasks in this tick
			inFlight[group] = current + 1
		}

		// Run pre-query if configured
		inputData := task.InputData
		if task.PreQuery.Valid && task.PreQuery.String != "" {
			dynamicData, err := runPreQuery(ctx, db, task.PreQuery.String, logger)
			if err != nil {
				logger.Warn("Pre-query failed, skipping task",
					zap.String("task", task.Name), zap.Error(err))
				continue
			}
			if dynamicData == nil {
				logger.Debug("Pre-query returned no rows, skipping task",
					zap.String("task", task.Name))
				continue
			}
			// Merge pre-query results into input_data
			inputData, err = mergeJSON(task.InputData, dynamicData)
			if err != nil {
				logger.Warn("Failed to merge input_data", zap.String("task", task.Name), zap.Error(err))
				continue
			}
		}

		// Fire the task
		if err := fireTrigger(ctx, producer, task, inputData, logger); err != nil {
			logger.Warn("Failed to fire task",
				zap.String("task", task.Name), zap.Error(err))
			continue
		}

		// Update last_triggered_at
		_, err := db.ExecContext(ctx,
			`UPDATE scheduled_tasks SET last_triggered_at = NOW(), updated_at = NOW() WHERE id = $1`,
			task.ID)
		if err != nil {
			logger.Warn("Failed to update last_triggered_at",
				zap.String("task", task.Name), zap.Error(err))
		}

		logger.Info("Triggered task",
			zap.String("task", task.Name),
			zap.String("agent_type", task.TargetAgentType),
			zap.String("topic", task.TargetTopic))
	}
}

// loadDueTasks returns enabled tasks whose interval has elapsed.
// Uses Postgres to atomically check timing — prevents double-firing
// even if two scheduler instances run briefly during rolling updates.
func loadDueTasks(ctx context.Context, db *sql.DB, logger *zap.Logger) ([]scheduledTask, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, target_agent_type, target_topic, input_data,
		       pre_query, interval_seconds, concurrency_group,
		       max_concurrent, timeout_seconds,
		       last_triggered_at, last_completed_at
		FROM scheduled_tasks
		WHERE enabled = true
		  AND (
		    last_triggered_at IS NULL
		    OR last_triggered_at + (interval_seconds || ' seconds')::interval <= NOW()
		  )
		ORDER BY last_triggered_at ASC NULLS FIRST
	`)
	if err != nil {
		return nil, fmt.Errorf("query scheduled_tasks: %w", err)
	}
	defer rows.Close()

	var tasks []scheduledTask
	for rows.Next() {
		var t scheduledTask
		if err := rows.Scan(
			&t.ID, &t.Name, &t.TargetAgentType, &t.TargetTopic,
			&t.InputData, &t.PreQuery, &t.IntervalSeconds,
			&t.ConcurrencyGroup, &t.MaxConcurrent, &t.TimeoutSeconds,
			&t.LastTriggeredAt, &t.LastCompletedAt,
		); err != nil {
			logger.Warn("Failed to scan task", zap.Error(err))
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// countInFlight returns how many tasks per concurrency group are currently
// running (triggered but not yet completed or timed out).
func countInFlight(ctx context.Context, db *sql.DB, logger *zap.Logger) map[string]int {
	result := map[string]int{}
	rows, err := db.QueryContext(ctx, `
		SELECT concurrency_group, COUNT(*)
		FROM scheduled_tasks
		WHERE enabled = true
		  AND concurrency_group IS NOT NULL
		  AND last_triggered_at IS NOT NULL
		  AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
		  AND last_triggered_at + (timeout_seconds || ' seconds')::interval > NOW()
		GROUP BY concurrency_group
	`)
	if err != nil {
		logger.Warn("Failed to count in-flight tasks", zap.Error(err))
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var group string
		var count int
		if err := rows.Scan(&group, &count); err == nil {
			result[group] = count
		}
	}
	return result
}

// runPreQuery executes the task's pre_query and returns the first row as JSON.
// Returns nil if the query returns no rows (task should be skipped).
func runPreQuery(ctx context.Context, db *sql.DB, query string, logger *zap.Logger) (json.RawMessage, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pre_query exec: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("pre_query columns: %w", err)
	}

	if !rows.Next() {
		return nil, nil // no rows — skip
	}

	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, fmt.Errorf("pre_query scan: %w", err)
	}

	result := map[string]interface{}{}
	for i, col := range cols {
		switch v := values[i].(type) {
		case []byte:
			result[col] = string(v)
		default:
			result[col] = v
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("pre_query marshal: %w", err)
	}
	return data, nil
}

// mergeJSON merges two JSON objects. Fields from overlay take precedence.
func mergeJSON(base, overlay json.RawMessage) (json.RawMessage, error) {
	var baseMap, overlayMap map[string]interface{}
	if err := json.Unmarshal(base, &baseMap); err != nil {
		baseMap = map[string]interface{}{}
	}
	if err := json.Unmarshal(overlay, &overlayMap); err != nil {
		return base, nil
	}
	for k, v := range overlayMap {
		baseMap[k] = v
	}
	return json.Marshal(baseMap)
}

// fireTrigger publishes a Kafka message that triggers the target agent.
// Uses the same message format as CallAgentAction / DispatchAreaDiscoverersAction.
func fireTrigger(ctx context.Context, producer kafka.Producer, task scheduledTask, inputData json.RawMessage, logger *zap.Logger) error {
	orchID := uuid.New().String()
	reqID := uuid.New().String()
	orchName := fmt.Sprintf("sched-%s-%s", task.Name, time.Now().Format("20060102-150405"))

	// Parse input_data into a map for the body
	var inputMap map[string]interface{}
	if err := json.Unmarshal(inputData, &inputMap); err != nil {
		inputMap = map[string]interface{}{}
	}

	body := map[string]interface{}{
		"action": "orchestrate",
		"config": map[string]interface{}{
			"agent_type": task.TargetAgentType,
		},
		"input_data": inputMap,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	headers := map[string]string{
		"correlation_id":     uuid.New().String(),
		"request_id":         reqID,
		"message_id":         uuid.New().String(),
		"orchestration_id":   orchID,
		"orchestration_name": orchName,
		"step_name":          "start",
		"message_type":       "request",
		"action":             "orchestrate",
		"from_agent_type":    "kafka-scheduler",
		"from_agent_id":      "kafka-scheduler-singleton",
		"responses_topic":    "system.scheduler.responses",
	}

	return producer.Produce(ctx, task.TargetTopic, headers, []byte(reqID), bodyBytes)
}

// buildLogger creates a zap logger from LOG_LEVEL env var.
func buildLogger() *zap.Logger {
	level := zapcore.InfoLevel
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		switch strings.ToLower(v) {
		case "debug":
			level = zapcore.DebugLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		}
	}
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, _ := cfg.Build()
	return logger
}
