// FILE: internal/adapters/analyser/adapter.go
//
// DRAFT for the agent-chassis repo (module github.com/gqls/agentchassis).
// Does not compile in the contextkit container — built/deployed in your env.
//
// Analyser adapter dispatcher. Mirrors the thunder/git/image-generator adapters.
// The response envelope is grounded in the orchestrator (coordinator.go) and the
// three working adapters, not the docs — which disagree (003 vs FOCUS) and were
// resolved empirically:
//
// REQUEST parse:
//   - action comes from the BODY (body.action), never the headers — matching
//     thunder (body["action"]), git (req.Body.Action), image-generator
//     (req.Body.Data). The payload is at body.data.
//   - reply topic is mixed in the codebase: git/websearch read it from the
//     headers, thunder/image-generator from the body. Both are accepted.
//
// RESPONSE build (the part that makes responses route instead of timing out):
//   - The chassis claims the awaited request on in_response_to_request_id first
//     (coordinator.go ProcessResponse), falling back to request_id only if that
//     is empty. So in_response_to_request_id (= the incoming request_id) is the
//     load-bearing field. Following git, request_id is also reused from the
//     incoming request and a fresh message_id is added — this covers both the
//     primary and fallback match paths. See buildResponseKafkaHeaders.
//   - The reply BODY uses the canonical types.ResponseHeaders (via
//     ToResponseHeaders), so is_complete/is_error marshal as real JSON bools.
//     This is the "bool trap": the orchestrator unmarshals into a struct with
//     bool fields, and a map[string]string sending the string "true" fails the
//     unmarshal (documented in thunder-adapter, lines ~634-643). websearch is
//     the lone adapter still on the string-bool map + plain Produce — the
//     deprecated path the FOCUS skeleton happens to describe; do not copy it.
//   - Sent via ProduceWithValidation, never plain Produce (git/thunder/dynamic).
//
// Import-reuse note: the canonical types.ResponseHeaders has NO request_id /
// message_id field and NO ToKafkaHeaders method (verified against
// platform/orchestration/types/context.go) — which is exactly why the git
// adapter carries its own ResponseHeaders mirror. The adapter envelope needs
// request_id/message_id in the Kafka headers, so this file supplies
// buildResponseKafkaHeaders locally (the same fields git's ToKafkaHeaders emits)
// while the body reuses the canonical type for the real-bool guarantee.
//
// RECONCILE against your tree before building:
//   1. The kafka package import path + the exact signatures of NewConsumer /
//      NewProducer / FetchMessage / CommitMessages / ProduceWithValidation —
//      copy from webscrape or git (the contract's reference adapters). This
//      file assumes ProduceWithValidation(ctx, topic, headers, key, value)
//      (verified against webscrape_actions.go).
//   2. The ExecutionContext fields ToResponseHeaders reads (Sender/Status/
//      IsComplete/IsError) and AgentIdentity's field names.

package analyser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/reposource"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	// NOTE: align this import path with the kafka package webscrape/git use.
	"github.com/gqls/agentchassis/platform/kafka"
)

const (
	defaultRequestsTopic = "system.adapter.analyser.requests"
	defaultConsumerGroup = "analyser.adapter.group"
	adapterAgentType     = "analyser-adapter"
)

// Adapter consumes analyse requests from Kafka and replies with a
// ResponseMessage on the caller's responses topic. Lifecycle mirrors the
// thunder/git adapters: NewAdapter → StartHealthServer → Run; Shutdown is
// sync.Once-guarded and safe to call from both the signal handler and a test
// harness.
type Adapter struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.ServiceConfig
	logger *zap.Logger

	consumer      *kafka.Consumer
	producer      kafka.Producer
	requestsTopic string

	analyse *AnalyseAction

	// Identity for response headers: senderType is cfg.ServiceInfo.Name
	// (fallback adapterAgentType); adapterID is this instance's run UUID.
	adapterID  uuid.UUID
	senderType string
	podName    string

	healthServer *http.Server
	shutdownOnce sync.Once
	shutdownWg   sync.WaitGroup
}

// NewAdapter wires the Kafka consumer/producer, the read-only GitHub source,
// and the analyse action. Cleanup convention (035 §2.5): every failure path
// closes everything successfully opened before it, then cancels.
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	adapterCtx, cancel := context.WithCancel(ctx)

	// 1. Topics + consumer group (env override or default — matches thunder).
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

	// 4. Read-only, repo-scoped GitHub token from a k8s Secret (least
	// privilege). NewGitHubSource fails fast on an empty token. The env var
	// names are documented in configs/analyser-adapter.yaml (custom.github);
	// direct env reads match the thunder pattern.
	source, err := reposource.NewGitHubSource(os.Getenv("GITHUB_READ_TOKEN"), os.Getenv("GITHUB_API_BASE"), logger)
	if err != nil {
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("create github source: %w", err)
	}

	// Identity for response headers. sender_agent_type comes from the config
	// (service_info.name), falling back to the const so the adapter still
	// answers sensibly if the config field is blank.
	senderType := cfg.ServiceInfo.Name
	if senderType == "" {
		senderType = adapterAgentType
	}
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = os.Getenv("HOSTNAME")
	}

	a := &Adapter{
		ctx:           adapterCtx,
		cancel:        cancel,
		cfg:           cfg,
		logger:        logger.Named("analyser_adapter"),
		consumer:      consumer,
		producer:      producer,
		requestsTopic: requestsTopic,
		analyse:       NewAnalyseAction(source, logger),
		adapterID:     adapterID,
		senderType:    senderType,
		podName:       podName,
	}

	logger.Info("Analyser adapter initialized",
		zap.Strings("kafka_brokers", cfg.Infrastructure.KafkaBrokers),
		zap.String("requests_topic", requestsTopic),
		zap.String("consumer_group", consumerGroup),
		zap.String("adapter_id", adapterID.String()),
		zap.String("sender_agent_type", senderType),
	)
	return a, nil
}

// Run is the main message-processing loop. Returns nil on graceful shutdown.
// Sequential by design — no `go a.handleMessage(msg)` (035 §2.6).
func (a *Adapter) Run() error {
	a.logger.Info("Analyser adapter starting message processing",
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
				if err == context.Canceled || err == context.DeadlineExceeded {
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

// StartHealthServer serves /health (liveness) and /ready (readiness) on the
// given port. Called by main before Run. The analyser's only external
// dependency is the GitHub API; pinging it on every probe period would burn
// rate limit for no signal, so /ready instead reports draining state: 503 once
// shutdown has begun (so k8s pulls the pod from Service endpoints), 200
// otherwise.
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

// Shutdown stops the adapter gracefully. sync.Once makes it safe to call more
// than once — the signal handler in main() and a test harness can both call it.
func (a *Adapter) Shutdown() {
	a.shutdownOnce.Do(func() {
		a.logger.Info("Analyser adapter shutting down")
		a.cancel() // signals Run() to exit at next iteration

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
		a.logger.Info("Analyser adapter shutdown complete")
	})
}

// commit acknowledges a message; logs but doesn't error if the commit fails.
// Uses a background context so commits still land during shutdown drain.
func (a *Adapter) commit(msg kafka.Message) {
	if err := a.consumer.CommitMessages(context.Background(), msg); err != nil {
		a.logger.Warn("Failed to commit message", zap.Error(err))
	}
}

// incoming reuses the canonical RequestHeaders for the envelope headers, and
// reads the action + payload from the BODY — matching every working adapter
// (thunder body["action"], git req.Body.Action, image-generator req.Body.Data).
// The action is deliberately NOT taken from the headers. reply_to_topic is in
// the body on some adapters (thunder, image-generator) and in the headers on
// others (git, websearch), so it is captured here and resolved at dispatch.
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
	case "analyse":
		var req AnalyseRequest
		// Execute on a.ctx so a shutdown aborts a long-running analysis
		// (tarball fetch + AST walk) instead of holding the drain window.
		if execErr = json.Unmarshal(in.Body.Data, &req); execErr == nil {
			result, execErr = a.analyse.Execute(a.ctx, req)
		}
	default:
		execErr = fmt.Errorf("analyser adapter: action %q not implemented", in.Body.Action)
	}

	// Shutdown-mid-work: the failure is the drain, not the request. Don't
	// reply (an error_unrecoverable here would wrongly tell the caller not to
	// retry) and don't commit — at-least-once redelivery re-runs it on the
	// next pod.
	if execErr != nil && a.ctx.Err() != nil {
		a.logger.Warn("analysis aborted by shutdown — leaving message uncommitted for redelivery",
			zap.String("correlation_id", in.Headers.CorrelationID),
			zap.String("request_id", in.Headers.RequestID))
		return
	}

	status := "complete"
	if execErr != nil {
		status = "error_unrecoverable"
	}

	// Reply topic source is mixed across the working adapters: git/websearch
	// read it from the headers, thunder/image-generator from the body. Accept
	// all three keys rather than assume one.
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

	// Kafka MESSAGE headers — what the chassis router matches the awaited
	// request on. request_id is the incoming one, reused verbatim.
	kHeaders := a.buildResponseKafkaHeaders(in.Headers, status, uuid.New().String())

	// ProduceWithValidation, never plain Produce (035 §1.6). Background-derived
	// timeout so a completed analysis still gets its reply out during drain.
	prodCtx, prodCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer prodCancel()
	if err := a.producer.ProduceWithValidation(prodCtx, replyTopic, kHeaders, []byte(in.Headers.CorrelationID), payload); err != nil {
		a.logger.Error("produce response failed",
			zap.Error(err), zap.String("topic", replyTopic))
		return // don't commit — let it redeliver (at-least-once)
	}

	a.commit(msg)
	a.logger.Info("analyse response sent",
		zap.String("correlation_id", in.Headers.CorrelationID),
		zap.String("in_response_to_request_id", in.Headers.RequestID),
		zap.String("status", status),
		zap.String("topic", replyTopic))
}

// buildResponseBody constructs the ResponseMessage body via the platform's
// ToResponseHeaders, so is_complete/is_error are real JSON bools (the bool
// trap). The Kafka-header routing fields are set separately (see
// buildResponseKafkaHeaders) because the canonical ResponseHeaders does not
// carry request_id/message_id.
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
			Code:        "analyse_failed",
			Message:     execErr.Error(),
			Recoverable: false,
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

// buildResponseKafkaHeaders builds the map[string]string Kafka message headers
// the envelope contract (035 §1.3-1.4) requires. This is the git adapter's
// ToKafkaHeaders() equivalent — supplied here because the canonical
// types.ResponseHeaders has no such method and no request_id/message_id fields.
// String bools are correct in KAFKA headers (they're byte strings); the typed
// bools live in the JSON body.
func (a *Adapter) buildResponseKafkaHeaders(req types.RequestHeaders, status, messageID string) map[string]string {
	return map[string]string{
		"message_type":              "response",
		"request_id":                req.RequestID, // REUSED incoming id — covers the fallback match path
		"in_response_to_request_id": req.RequestID, // the load-bearing claim key (035 §1.4)
		"in_response_to_step_id":    req.StepID,
		"in_response_to_step_name":  req.StepName,
		"orchestration_id":          req.OrchestrationID,
		"correlation_id":            req.CorrelationID,
		"client_id":                 req.ClientID,
		"message_id":                messageID, // fresh
		"status":                    status,
		"sender_agent_type":         a.senderType,
		"sender_agent_id":           a.adapterID.String(),
		"sender_pod_name":           a.podName,
		"time_sent":                 time.Now().UTC().Format(time.RFC3339),
		"in_response_to_action":     req.Action,
	}
}
