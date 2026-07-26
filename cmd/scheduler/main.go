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
	FireMessage      bool
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
		// Concurrency check — refuse if the group is already at its
		// max_concurrent, but do NOT claim the slot yet. A task whose
		// pre-query finds no work must not consume its group's capacity;
		// claiming the slot up-front and then bailing on the no-rows path is
		// what starved the whole `maintenance` group for 79 days
		// (bugs_open/048). The slot is claimed below, only once the task
		// commits to firing or doing work.
		if task.ConcurrencyGroup.Valid {
			group := task.ConcurrencyGroup.String
			if inFlight[group] >= task.MaxConcurrent {
				logger.Info("Skipping task — concurrency group at max",
					zap.String("task", task.Name),
					zap.String("group", group),
					zap.Int("current", inFlight[group]),
					zap.Int("max", task.MaxConcurrent))
				continue
			}
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
				// The pre-query ran and found nothing to do. That is a
				// successful no-op, not a skip: stamp the timestamps so the
				// task rotates to the back of loadDueTasks' ASC NULLS FIRST
				// queue like any completed run. Leaving them unchanged pins
				// the task at the head of its group, where it re-wins the only
				// slot every tick and never lets its group-mates run
				// (bugs_open/048). For a CTE-only task the pre-query IS the
				// work, so it has genuinely run to completion.
				if err := stampCompleted(ctx, db, task.ID); err != nil {
					logger.Warn("Failed to update timestamps",
						zap.String("task", task.Name), zap.Error(err))
				}
				logger.Info("Pre-query found no rows — task ran with nothing to do",
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

		// A workflow authored INSIDE input_data cannot be delivered, and used to
		// be discarded in silence (bugs_open/074). fireTrigger builds the
		// message `config` from this row's COLUMNS, so a workflow written at
		// input_data.config.workflow lands at body.input_data.config.workflow —
		// one level below the only place the chassis reads it
		// (processor.go selectWorkflow, Priority 1: body.config.workflow). The
		// target agent's own workflow then runs instead; for `generic` that is a
		// single complete_workflow no-op, so the task fired, stamped BOTH
		// timestamps, created a COMPLETED orchestration and did nothing. One
		// task sat like that for three months.
		//
		// Refuse rather than fire: manufacturing a green orchestration for work
		// that was never dispatched is the actual harm here. Stamp the
		// timestamps anyway so the row rotates to the back of its group's queue
		// instead of re-winning the only slot every tick (bugs_open/048), and
		// claim no slot — nothing is running.
		//
		// A CHECK constraint (migration 217) refuses this shape at the moment it
		// is authored, so in a healthy database this branch is unreachable. It
		// stays because a constraint can be dropped and an old row restored, and
		// because a discard should be visible where it happens.
		if startStep, carriesWorkflow := discardedInlineWorkflow(inputData); carriesWorkflow {
			logger.Warn("Refusing to fire: task carries a workflow the scheduler cannot deliver",
				zap.String("task", task.Name),
				zap.String("target_agent_type", task.TargetAgentType),
				zap.String("ignored_start_step", startStep),
				zap.String("remedy", "move the workflow into an agent_definitions row and point target_agent_type at it; input_data is the payload only"))
			if err := stampCompleted(ctx, db, task.ID); err != nil {
				logger.Warn("Failed to update timestamps",
					zap.String("task", task.Name), zap.Error(err))
			}
			continue
		}

		// The task has real work to do — claim the concurrency slot now, so
		// other tasks in this group later in the same tick respect
		// max_concurrent. Nothing above this point claimed it: a no-op or an
		// errored pre-query leaves the slot free for its group-mates
		// (bugs_open/048).
		if task.ConcurrencyGroup.Valid {
			inFlight[task.ConcurrencyGroup.String]++
		}

		// CTE-only tasks: pre_query did the work, no Kafka message needed
		if !task.FireMessage {
			if err := stampCompleted(ctx, db, task.ID); err != nil {
				logger.Warn("Failed to update timestamps",
					zap.String("task", task.Name), zap.Error(err))
			}
			logger.Info("Pre-query task completed (no message fired)",
				zap.String("task", task.Name))
			continue
		}

		// Fire the task
		if err := fireTrigger(ctx, producer, task, inputData, logger); err != nil {
			logger.Warn("Failed to fire task",
				zap.String("task", task.Name), zap.Error(err))
			continue
		}

		// Update timestamps. For fire-and-forget tasks, mark completed immediately
		// so the concurrency slot opens for the next tick. The message has been
		// published — we don't wait for the orchestration to finish.
		if err := stampCompleted(ctx, db, task.ID); err != nil {
			logger.Warn("Failed to update timestamps",
				zap.String("task", task.Name), zap.Error(err))
		}

		logger.Info("Triggered task",
			zap.String("task", task.Name),
			zap.String("agent_type", task.TargetAgentType),
			zap.String("topic", task.TargetTopic))
	}
}

// discardedInlineWorkflow reports whether a task's input_data carries a workflow
// at input_data.config.workflow, and returns that workflow's start_step for the
// log. Presence is decided structurally, not by shape: a malformed workflow (a
// string, a number, an object with no start_step) is discarded just as silently
// as a well-formed one, so it must be reported just as loudly. start_step is
// best-effort and comes back empty when it cannot be read.
//
// This is deliberately the ONLY key the scheduler inspects inside input_data.
// bugs_closed/054 settled that this column is the payload and that fireTrigger
// must not reach into it — reading a payload field to decide behaviour is how a
// legitimate field named `config`/`action`/`input_data` gets silently eaten.
// Detecting one known-undeliverable shape in order to refuse it is not that:
// nothing here changes what is sent, and the field is never consumed.
func discardedInlineWorkflow(inputData json.RawMessage) (startStep string, found bool) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(inputData, &body); err != nil {
		return "", false
	}
	rawConfig, ok := body["config"]
	if !ok {
		return "", false
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return "", false
	}
	rawWorkflow, ok := config["workflow"]
	if !ok {
		return "", false
	}
	var workflow struct {
		StartStep string `json:"start_step"`
	}
	// Best-effort: a workflow we cannot parse is still a workflow we cannot deliver.
	_ = json.Unmarshal(rawWorkflow, &workflow)
	return workflow.StartStep, true
}

// stampCompleted records that a task ran to completion this tick by advancing
// both last_triggered_at and last_completed_at to NOW(). Advancing both (not
// last_triggered_at alone) keeps the task out of countInFlight and satisfies
// loadDueTasks' in-flight guard, and moves it to the back of its group's
// ASC NULLS FIRST queue so a no-op run cannot pin it at the head forever
// (bugs_open/048).
func stampCompleted(ctx context.Context, db *sql.DB, taskID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_triggered_at = NOW(), last_completed_at = NOW(), updated_at = NOW() WHERE id = $1`,
		taskID)
	return err
}

// loadDueTasks returns enabled tasks whose interval has elapsed AND whose
// previous execution has completed (or timed out). This prevents the same
// task from spawning multiple concurrent pods.
func loadDueTasks(ctx context.Context, db *sql.DB, logger *zap.Logger) ([]scheduledTask, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, target_agent_type, target_topic, input_data,
		       pre_query, interval_seconds, concurrency_group,
		       max_concurrent, timeout_seconds,
		       last_triggered_at, last_completed_at, fire_message
		FROM scheduled_tasks
		WHERE enabled = true
		  AND (
		    last_triggered_at IS NULL
		    OR last_triggered_at + (interval_seconds || ' seconds')::interval <= NOW()
		  )
		  AND (
		    -- Per-task guard: don't re-fire if previous execution is still in-flight.
		    -- CTE-only tasks (fire_message=false) complete instantly so always pass.
		    fire_message = false
		    OR last_triggered_at IS NULL
		    OR last_completed_at >= last_triggered_at
		    OR last_triggered_at + (timeout_seconds || ' seconds')::interval <= NOW()
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
			&t.LastTriggeredAt, &t.LastCompletedAt, &t.FireMessage,
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
		"client_id":          "system",
		"message_type":       "request",
		"action":             "orchestrate",
		"from_agent_type":    "kafka-scheduler",
		"from_agent_id":      "kafka-scheduler-singleton",
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
