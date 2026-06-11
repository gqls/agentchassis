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
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

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
// ResponseMessage on the caller's responses topic.
type Adapter struct {
	consumer      *kafka.Consumer
	producer      kafka.Producer
	requestsTopic string
	analyse       *AnalyseAction
	logger        *zap.Logger
	podName       string
}

// NewAdapter wires the Kafka consumer/producer, the read-only GitHub source,
// and the analyse action.
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	requestsTopic := os.Getenv("REQUESTS_TOPIC")
	if requestsTopic == "" {
		requestsTopic = defaultRequestsTopic
	}
	consumerGroup := os.Getenv("CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = defaultConsumerGroup
	}

	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestsTopic, consumerGroup, logger)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	// Read-only, repo-scoped GitHub token from a k8s Secret (least privilege).
	source, err := NewGitHubSource(os.Getenv("GITHUB_READ_TOKEN"), os.Getenv("GITHUB_API_BASE"), logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("create github source: %w", err)
	}

	return &Adapter{
		consumer:      consumer,
		producer:      producer,
		requestsTopic: requestsTopic,
		analyse:       NewAnalyseAction(source, logger),
		logger:        logger.Named("analyser_adapter"),
		podName:       os.Getenv("HOSTNAME"),
	}, nil
}

// Run starts the health server and consumes the requests topic until ctx is
// cancelled.
func (a *Adapter) Run(ctx context.Context) error {
	a.startHealthServer()
	a.logger.Info("analyser adapter started", zap.String("requests_topic", a.requestsTopic))
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		msg, err := a.consumer.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			continue // fetch timeout / transient — loop (match webscrape/thunder)
		}
		a.handleMessage(ctx, msg)
	}
}

// startHealthServer serves /health and /ready for the k8s probes.
func (a *Adapter) startHealthServer() {
	port := os.Getenv("HEALTH_PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	mux.HandleFunc("/health", ok)
	mux.HandleFunc("/ready", ok)
	go func() {
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			a.logger.Warn("health server stopped", zap.Error(err))
		}
	}()
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

func (a *Adapter) handleMessage(ctx context.Context, msg kafka.Message) {
	var in incoming
	if err := json.Unmarshal(msg.Value, &in); err != nil {
		a.logger.Error("unparseable request — dropping", zap.Error(err))
		_ = a.consumer.CommitMessages(ctx, msg) // poison message: commit so we don't loop
		return
	}

	ec := types.FromRequestHeaders(in.Headers)
	ec.Sender = types.AgentIdentity{AgentType: adapterAgentType, PodName: a.podName}

	var (
		result  interface{}
		execErr error
	)
	switch in.Body.Action {
	case "analyse":
		var req AnalyseRequest
		if execErr = json.Unmarshal(in.Body.Data, &req); execErr == nil {
			result, execErr = a.analyse.Execute(ctx, req)
		}
	default:
		execErr = fmt.Errorf("analyser adapter: action %q not implemented", in.Body.Action)
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
		_ = a.consumer.CommitMessages(ctx, msg)
		return
	}

	resp := a.buildResponseBody(ec, result, execErr, status)
	payload, err := json.Marshal(resp)
	if err != nil {
		a.logger.Error("marshal response failed", zap.Error(err))
		_ = a.consumer.CommitMessages(ctx, msg)
		return
	}

	// Kafka MESSAGE headers — what the chassis router matches the awaited
	// request on. request_id is the incoming one, reused verbatim.
	kHeaders := buildResponseKafkaHeaders(in.Headers, status, uuid.New().String())

	// ProduceWithValidation, never plain Produce (doc 003). It runs the
	// outgoing validator and still sends error responses even if validation
	// fails. Align the exact signature with webscrape/git.
	if err := a.producer.ProduceWithValidation(ctx, replyTopic, kHeaders, []byte(in.Headers.CorrelationID), payload); err != nil {
		a.logger.Error("produce response failed",
			zap.Error(err), zap.String("topic", replyTopic))
		return // don't commit — let it redeliver (at-least-once)
	}

	_ = a.consumer.CommitMessages(ctx, msg)
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
// the Adapter Response Envelope Contract requires. This is the git adapter's
// ToKafkaHeaders() equivalent — supplied here because the canonical
// types.ResponseHeaders has no such method and no request_id/message_id fields.
// String bools are correct in KAFKA headers (they're byte strings); the typed
// bools live in the JSON body.
func buildResponseKafkaHeaders(req types.RequestHeaders, status, messageID string) map[string]string {
	return map[string]string{
		"message_type":              "response",
		"request_id":                req.RequestID, // REUSED incoming id — router matches on this
		"in_response_to_request_id": req.RequestID,
		"in_response_to_step_id":    req.StepID,
		"in_response_to_step_name":  req.StepName,
		"orchestration_id":          req.OrchestrationID,
		"correlation_id":            req.CorrelationID,
		"client_id":                 req.ClientID,
		"message_id":                messageID, // fresh
		"status":                    status,
		"sender_agent_type":         adapterAgentType,
		"in_response_to_action":     req.Action,
	}
}
