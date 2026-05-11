// FILE: internal/adapters/thunder/adapter.go
//
// Thunder Compute Adapter — Phase 2 skeleton.
//
// Status: skeleton. Connects to Kafka and Postgres, serves /health and /ready,
// receives messages on system.adapter.thunder.requests and returns
// error_unrecoverable responses with reason "not_implemented" until Phase 3
// lands the real action handlers.
//
// See FOCUS_adapter_design.md for the canonical pattern this follows.
// See 013/033_thunder_adapter_design.md for the design of the full adapter.

package thunder

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
)

// Adapter is the long-running thunder-adapter service.
type Adapter struct {
	ctx       context.Context
	cancel    context.CancelFunc
	cfg       *config.ServiceConfig
	logger    *zap.Logger
	consumer  *kafka.Consumer
	producer  kafka.Producer
	db        *sql.DB
	adapterID uuid.UUID

	requestsTopic string

	healthServer *http.Server
	shutdownOnce sync.Once
	shutdownWg   sync.WaitGroup
}

// NewAdapter wires up dependencies and returns an Adapter ready to Run.
// Returns an error if any dependency can't be initialised.
// Cleanup convention: every error path closes everything successfully
// opened before it (no defer magic; explicit).
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	adapterCtx, cancel := context.WithCancel(ctx)

	// ── 1. Topics + consumer group (env override or default) ──
	requestsTopic := os.Getenv("REQUESTS_TOPIC")
	if requestsTopic == "" {
		requestsTopic = "system.adapter.thunder.requests"
	}
	consumerGroup := os.Getenv("CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = "thunder.adapter.group"
	}

	adapterID, _ := uuid.NewUUID()

	// ── 2. Kafka consumer ──
	consumer, err := kafka.NewConsumer(
		cfg.Infrastructure.KafkaBrokers, requestsTopic, consumerGroup, logger,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	// ── 3. Kafka producer ──
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// ── 4. Postgres connection (owns thunder_instances, thunder_config) ──
	dbDSN, err := buildClientsDBDSN(cfg)
	if err != nil {
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to build db DSN: %w", err)
	}
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	if err := db.PingContext(adapterCtx); err != nil {
		db.Close()
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Sanity check — confirm thunder_config singleton exists (migration 025)
	var capCheck float64
	err = db.QueryRowContext(adapterCtx,
		"SELECT daily_cap_usd FROM thunder_config WHERE singleton = 'X'",
	).Scan(&capCheck)
	if err != nil {
		db.Close()
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("thunder_config row not found — apply migration 025 first: %w", err)
	}
	logger.Info("Thunder adapter db connected", zap.Float64("daily_cap_usd", capCheck))

	a := &Adapter{
		ctx:           adapterCtx,
		cancel:        cancel,
		cfg:           cfg,
		logger:        logger.With(zap.String("component", "thunder-adapter")),
		consumer:      consumer,
		producer:      producer,
		db:            db,
		adapterID:     adapterID,
		requestsTopic: requestsTopic,
	}

	logger.Info("Thunder adapter initialized",
		zap.Strings("kafka_brokers", cfg.Infrastructure.KafkaBrokers),
		zap.String("requests_topic", requestsTopic),
		zap.String("consumer_group", consumerGroup),
		zap.String("adapter_id", adapterID.String()),
	)

	return a, nil
}

// Run is the main message processing loop. Returns nil on graceful shutdown.
// Sequential by design — no `go a.handleMessage(msg)`.
func (a *Adapter) Run() error {
	a.logger.Info("Thunder adapter starting message processing",
		zap.String("topic", a.requestsTopic))

	a.shutdownWg.Add(1)
	defer a.shutdownWg.Done()

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Shutdown signal received, stopping message processing")
			return nil
		default:
			consumeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			msg, err := a.consumer.FetchMessage(consumeCtx)
			cancel()

			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					continue
				}
				select {
				case <-a.ctx.Done():
					return nil
				default:
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			a.handleMessage(msg)
		}
	}
}

// handleMessage parses one inbound request and routes it.
// Phase 2: every action returns error_unrecoverable / not_implemented.
func (a *Adapter) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("request_id", headers["request_id"]),
	)

	// Parse the standard envelope
	var envelope map[string]interface{}
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		l.Error("Failed to unmarshal message", zap.Error(err))
		a.commit(msg)
		return
	}

	body, _ := envelope["body"].(map[string]interface{})
	if body == nil {
		body = make(map[string]interface{})
	}
	action, _ := body["action"].(string)
	replyToTopic, _ := body["reply_to_topic"].(string)

	l.Info("Received request",
		zap.String("action", action),
		zap.String("reply_to_topic", replyToTopic),
	)

	// Phase 2: every action returns error_unrecoverable. Phase 3+ will
	// replace this with a switch dispatching to per-action handlers.
	a.sendErrorResponse(headers, replyToTopic, action,
		"not_implemented",
		"thunder-adapter is a skeleton; action handlers land in Phase 3",
		"error_unrecoverable", l,
	)
	a.commit(msg)
}

// sendErrorResponse sends a chassis-shaped error response on the reply topic.
// Includes all the correlation headers the chassis needs to match the response
// to its awaited_requests row.
func (a *Adapter) sendErrorResponse(
	reqHeaders map[string]string,
	replyToTopic string,
	action string,
	errCode string, // short code, e.g. "not_implemented", "thunder_api_unreachable"
	errMsg string, // human-readable detail
	status string, // "error_recoverable" or "error_unrecoverable"
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic in request; cannot send response",
			zap.String("action", action))
		return
	}

	respHeaders := buildResponseHeaders(reqHeaders, status, false, a.cfg.ServiceInfo.Name, a.adapterID)

	body := map[string]interface{}{
		"success": false,
		"action":  action,
		"error":   errCode,
		"detail":  errMsg,
	}
	envelope := map[string]interface{}{
		"headers": respHeaders,
		"body":    body,
	}
	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		l.Error("Failed to marshal response envelope", zap.Error(err))
		return
	}

	produceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.producer.Produce(
		produceCtx,
		replyToTopic,
		respHeaders,
		[]byte(reqHeaders["correlation_id"]), // key
		envelopeBytes,                        // value
	); err != nil {
		l.Error("Failed to produce response", zap.Error(err))
		return
	}
	l.Info("Sent error response",
		zap.String("action", action),
		zap.String("status", status),
		zap.String("error_code", errCode),
	)
}

// buildResponseHeaders constructs the response header map per the FOCUS doc
// adapter spec. Includes the in_response_to_* headers the chassis needs to
// match this response back to its awaited request.
func buildResponseHeaders(
	reqHeaders map[string]string,
	status string,
	success bool,
	senderType string,
	senderID uuid.UUID,
) map[string]string {
	return map[string]string{
		"correlation_id":            reqHeaders["correlation_id"],
		"orchestration_id":          reqHeaders["orchestration_id"],
		"in_response_to_request_id": reqHeaders["request_id"],
		"request_id":                uuid.New().String(),
		"client_id":                 reqHeaders["client_id"],
		"message_type":              "response",
		"status":                    status,
		"is_complete":               "true",
		"is_error":                  boolStr(!success),
		"sender_agent_type":         senderType,
		"sender_agent_id":           senderID.String(),
		"in_response_to_step_name":  reqHeaders["step_name"],
		"in_response_to_step_id":    reqHeaders["step_id"],
		"time_sent":                 time.Now().UTC().Format(time.RFC3339),
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// commit acknowledges a message; logs but doesn't error if the commit fails.
func (a *Adapter) commit(msg kafka.Message) {
	if err := a.consumer.CommitMessages(context.Background(), msg); err != nil {
		a.logger.Warn("Failed to commit message", zap.Error(err))
	}
}

// StartHealthServer starts /health (liveness) and /ready (readiness) endpoints.
func (a *Adapter) StartHealthServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"db_unreachable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	a.healthServer = &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := a.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Health server error", zap.Error(err))
		}
	}()
	a.logger.Info("Health server started", zap.String("port", port))
}

// Shutdown stops the adapter gracefully. Safe to call multiple times.
func (a *Adapter) Shutdown() {
	a.shutdownOnce.Do(func() {
		a.logger.Info("Thunder adapter shutting down")
		a.cancel()

		if a.healthServer != nil {
			sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = a.healthServer.Shutdown(sctx)
		}

		done := make(chan struct{})
		go func() {
			a.shutdownWg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			a.logger.Warn("Shutdown wait exceeded; closing connections anyway")
		}

		a.consumer.Close()
		a.producer.Close()
		if a.db != nil {
			_ = a.db.Close()
		}
		a.logger.Info("Thunder adapter shutdown complete")
	})
}

// buildClientsDBDSN constructs a Postgres DSN from the ServiceConfig.
func buildClientsDBDSN(cfg *config.ServiceConfig) (string, error) {
	dbc := cfg.Infrastructure.ClientsDatabase
	password := os.Getenv(dbc.PasswordEnvVar)
	if password == "" {
		return "", fmt.Errorf("clients_db password env var %q not set", dbc.PasswordEnvVar)
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbc.Host, dbc.Port, dbc.User, password, dbc.DBName, dbc.SSLMode,
	)
	return dsn, nil
}
