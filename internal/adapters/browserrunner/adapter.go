// FILE: internal/adapters/browserrunner/adapter.go
//
// Browser-runner adapter dispatcher — Tier 4 of the tool verification ladder
// (RUNBOOK_travelling_docs §0 Stage 6; contract pinned in
// PLAN_tool_acceptance_runner rev 2). Mirrors the analyser adapter, which is
// itself grounded in 035_adapter_guide §1 (normative envelope) and the
// git/thunder reference adapters:
//
//   - action from body.action ("run_checks"), payload at body.data (§1.2);
//   - reply topic from headers.responses_topic → headers.reply_to_topic →
//     body.reply_to_topic (§1.2 — accept all three);
//   - response: canonical types.ResponseHeaders in the BODY (real bools —
//     the §1.5 bool trap), request_id REUSED + in_response_to_request_id =
//     incoming request_id + fresh message_id in the KAFKA headers (§1.4);
//   - sent via ProduceWithValidation, never plain Produce (§1.6);
//   - sequential handling, no `go handleMessage` (§2.6).
//
// P0 scope (deliberately small): desktop profile only (1366×900), exactly
// three check types — page_status_ok, selector_exists, no_console_errors.
// Everything else in the criteria is reported as skipped, never faked.

package browserrunner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

const (
	defaultRequestsTopic = "system.adapter.browser-runner.requests"
	defaultConsumerGroup = "browser-runner.adapter.group"
	adapterAgentType     = "browser-runner-adapter"
)

// Adapter consumes run_checks requests from Kafka, drives the deployed page
// in headless Chromium, and replies with per-check results on the caller's
// responses topic. Lifecycle mirrors the analyser/git/thunder adapters.
type Adapter struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.ServiceConfig
	logger *zap.Logger

	consumer      *kafka.Consumer
	producer      kafka.Producer
	requestsTopic string

	runChecks   *RunChecksAction
	renderAudit *RenderAuditAction

	adapterID  uuid.UUID
	senderType string
	podName    string

	healthServer *http.Server
	shutdownOnce sync.Once
	shutdownWg   sync.WaitGroup
}

// NewAdapter wires the Kafka consumer/producer and the check runner. Cleanup
// convention (035 §2.5): every failure path closes everything successfully
// opened before it, then cancels. The browser itself is launched per request
// (P0): a crashed Chromium then poisons one run, not the pod.
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	adapterCtx, cancel := context.WithCancel(ctx)

	// 1. Topics + consumer group (env override or default — 035 §2.10).
	requestsTopic := os.Getenv("REQUESTS_TOPIC")
	if requestsTopic == "" {
		requestsTopic = defaultRequestsTopic
	}
	consumerGroup := os.Getenv("CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = defaultConsumerGroup
	}

	adapterID, _ := uuid.NewUUID()

	// 2. Kafka consumer.
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestsTopic, consumerGroup, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}

	// 3. Kafka producer.
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	senderType := cfg.ServiceInfo.Name
	if senderType == "" {
		senderType = adapterAgentType
	}
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = os.Getenv("HOSTNAME")
	}

	// P3: failure screenshots ride object storage when configured; nil store
	// means the runner behaves exactly as at P0–P2 (evidence, not dependency).
	store := newScreenshotStore(adapterCtx, cfg, logger)

	a := &Adapter{
		ctx:           adapterCtx,
		cancel:        cancel,
		cfg:           cfg,
		logger:        logger.Named("browser_runner_adapter"),
		consumer:      consumer,
		producer:      producer,
		requestsTopic: requestsTopic,
		runChecks:     NewRunChecksAction(logger, store),
		renderAudit:   NewRenderAuditAction(logger),
		adapterID:     adapterID,
		senderType:    senderType,
		podName:       podName,
	}

	logger.Info("Browser-runner adapter initialized",
		zap.Strings("kafka_brokers", cfg.Infrastructure.KafkaBrokers),
		zap.String("requests_topic", requestsTopic),
		zap.String("consumer_group", consumerGroup),
		zap.String("adapter_id", adapterID.String()),
		zap.String("sender_agent_type", senderType),
		zap.Bool("failure_screenshots", store != nil),
	)
	return a, nil
}

// Run is the main message-processing loop. Returns nil on graceful shutdown.
// Sequential by design — one browser run at a time per pod (035 §2.6; the
// runner PLAN's concurrency answer for P0).
func (a *Adapter) Run() error {
	a.logger.Info("Browser-runner adapter starting message processing",
		zap.String("topic", a.requestsTopic))

	a.shutdownWg.Add(1)
	defer a.shutdownWg.Done()

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Shutdown signal received, stopping message processing")
			return nil
		default:
			fetchCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			msg, err := a.consumer.FetchMessage(fetchCtx)
			cancel()

			if err != nil {
				// errors.Is, not ==: the kafka library may wrap the context error.
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					continue // no messages — re-poll
				}
				select {
				case <-a.ctx.Done():
					return nil // shutdown is the real reason for the error
				default:
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				time.Sleep(time.Second) // backoff before retry
				continue
			}
			a.handleMessage(msg)
		}
	}
}

// StartHealthServer serves /health (liveness) and /ready (readiness). The
// runner's external dependency is Chromium, which is launched per request —
// probing it would launch a browser per probe period for no signal, so /ready
// reports draining state instead (503 once shutdown begins), like the analyser.
func (a *Adapter) StartHealthServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		if a.ctx.Err() != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"draining"}`))
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

// Shutdown stops the adapter gracefully. sync.Once-guarded (035 §2.8).
func (a *Adapter) Shutdown() {
	a.shutdownOnce.Do(func() {
		a.logger.Info("Browser-runner adapter shutting down")
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
		a.logger.Info("Browser-runner adapter shutdown complete")
	})
}

// commit acknowledges a message; logs but doesn't error if the commit fails.
func (a *Adapter) commit(msg kafka.Message) {
	if err := a.consumer.CommitMessages(context.Background(), msg); err != nil {
		a.logger.Warn("Failed to commit message", zap.Error(err))
	}
}

// incoming — canonical RequestHeaders for the envelope; action + payload from
// the BODY (035 §1.2); reply_to_topic captured from the body as one of the
// three accepted sources.
type incoming struct {
	Headers types.RequestHeaders `json:"headers"`
	Body    struct {
		Action       string          `json:"action"`
		Data         json.RawMessage `json:"data"`
		ReplyToTopic string          `json:"reply_to_topic"`
	} `json:"body"`
}

func (a *Adapter) handleMessage(msg kafka.Message) {
	var in incoming
	if err := json.Unmarshal(msg.Value, &in); err != nil {
		a.logger.Error("unparseable request — dropping", zap.Error(err))
		a.commit(msg) // poison message: commit so we don't loop
		return
	}

	ec := types.FromRequestHeaders(in.Headers)
	ec.Sender = types.AgentIdentity{
		AgentType: a.senderType,
		AgentID:   a.adapterID.String(),
		PodName:   a.podName,
	}

	var (
		result  interface{}
		execErr error
	)
	switch in.Body.Action {
	case "run_checks":
		var req RunChecksRequest
		// Execute on a.ctx so a shutdown aborts a long browser run instead of
		// holding the drain window.
		if execErr = json.Unmarshal(in.Body.Data, &req); execErr == nil {
			result, execErr = a.runChecks.Execute(a.ctx, req)
		}
	case "render_audit":
		var req RenderAuditRequest
		// Same ctx for the same reason: an audit of a whole site is a long run.
		if execErr = json.Unmarshal(in.Body.Data, &req); execErr == nil {
			result, execErr = a.renderAudit.Execute(a.ctx, req)
		}
	default:
		execErr = fmt.Errorf("browser-runner adapter: action %q not implemented", in.Body.Action)
	}

	// Shutdown-mid-work: don't reply, don't commit — at-least-once redelivery
	// re-runs it on the next pod (analyser precedent).
	if execErr != nil && a.ctx.Err() != nil {
		a.logger.Warn("run aborted by shutdown — leaving message uncommitted for redelivery",
			zap.String("correlation_id", in.Headers.CorrelationID),
			zap.String("request_id", in.Headers.RequestID))
		return
	}

	status := "complete"
	if execErr != nil {
		// A browser/infra failure may succeed on retry (fresh Chromium, fresh
		// pod); the caller decides. Unknown actions are unrecoverable.
		if in.Body.Action == "run_checks" {
			status = "error_recoverable"
		} else {
			status = "error_unrecoverable"
		}
	}

	replyTopic := in.Headers.ResponsesTopic
	if replyTopic == "" {
		replyTopic = in.Headers.ReplyToTopic
	}
	if replyTopic == "" {
		replyTopic = in.Body.ReplyToTopic
	}
	if replyTopic == "" {
		a.logger.Error("no reply topic on request — cannot reply",
			zap.String("correlation_id", in.Headers.CorrelationID),
			zap.String("request_id", in.Headers.RequestID))
		a.commit(msg)
		return
	}

	resp := a.buildResponseBody(ec, result, execErr, status)
	payload, err := json.Marshal(resp)
	if err != nil {
		a.logger.Error("marshal response failed", zap.Error(err))
		a.commit(msg)
		return
	}

	kHeaders := a.buildResponseKafkaHeaders(in.Headers, status, uuid.New().String())

	prodCtx, prodCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer prodCancel()
	if err := a.producer.ProduceWithValidation(prodCtx, replyTopic, kHeaders, []byte(in.Headers.CorrelationID), payload); err != nil {
		a.logger.Error("produce response failed",
			zap.Error(err), zap.String("topic", replyTopic))
		return // don't commit — let it redeliver (at-least-once)
	}

	a.commit(msg)
	a.logger.Info("run_checks response sent",
		zap.String("correlation_id", in.Headers.CorrelationID),
		zap.String("in_response_to_request_id", in.Headers.RequestID),
		zap.String("status", status),
		zap.String("topic", replyTopic))
}

// buildResponseBody — canonical ToResponseHeaders for the body so
// is_complete/is_error are real JSON bools (035 §1.5, the bool trap).
func (a *Adapter) buildResponseBody(ec *types.ExecutionContext, result interface{}, execErr error, status string) types.ResponseMessage {
	ec.Status = status
	if execErr != nil {
		ec.IsError = true
		ec.IsComplete = false
	} else {
		ec.IsError = false
		ec.IsComplete = true
	}

	body := types.ResponseBody{Success: execErr == nil}
	if execErr != nil {
		body.Error = &types.ErrorInfo{
			Code:        "run_checks_failed",
			Message:     execErr.Error(),
			Recoverable: status == "error_recoverable",
			Timestamp:   time.Now().UTC(),
		}
	} else {
		body.Body = result
	}

	return types.ResponseMessage{
		Headers: ec.ToResponseHeaders(),
		Body:    body,
	}
}

// buildResponseKafkaHeaders — the map[string]string Kafka message headers per
// 035 §1.3–1.4: in_response_to_request_id = incoming request_id (the
// load-bearing claim key), request_id reused, message_id fresh. String bools
// are correct in KAFKA headers; the typed bools live in the JSON body.
func (a *Adapter) buildResponseKafkaHeaders(req types.RequestHeaders, status, messageID string) map[string]string {
	return map[string]string{
		"message_type":              "response",
		"request_id":                req.RequestID,
		"in_response_to_request_id": req.RequestID,
		"in_response_to_step_id":    req.StepID,
		"in_response_to_step_name":  req.StepName,
		"orchestration_id":          req.OrchestrationID,
		"correlation_id":            req.CorrelationID,
		"client_id":                 req.ClientID,
		"message_id":                messageID,
		"status":                    status,
		"sender_agent_type":         a.senderType,
		"sender_agent_id":           a.adapterID.String(),
		"sender_pod_name":           a.podName,
		"time_sent":                 time.Now().UTC().Format(time.RFC3339),
		"in_response_to_action":     req.Action,
	}
}
