// FILE: internal/adapters/analyser/adapter.go
//
// DRAFT for the agent-chassis repo (module github.com/gqls/agentchassis).
// Does not compile in the contextkit container — built/deployed in your env.
//
// Analyser adapter dispatcher. Mirrors the thunder adapter
// (internal/adapters/thunder/adapter.go) for Kafka wiring, and REUSES
// platform/orchestration/types for the message contract — including
// FromRequestHeaders/ToResponseHeaders, so the response envelope is built by
// the platform's own header logic rather than hand-populated (the strongest
// way to keep "output messages pass validation"). It does NOT mirror the
// message types locally the way the git adapter did — import-reuse instead.
//
// RECONCILE against your tree before building (two spots only):
//   1. The kafka package import path + the exact signatures of NewConsumer /
//      NewProducer / FetchMessage / Produce / CommitMessages. Copy them from
//      thunder/adapter.go — this file assumes the same shapes thunder uses.
//   2. Which ExecutionContext fields to set before ToResponseHeaders. This sets
//      Sender, Status, IsComplete, IsError; confirm AgentIdentity's fields and
//      whether FromRequestHeaders populates the ToAgent routing fields.

package analyser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	// NOTE: align this import path with the kafka package thunder/adapter.go uses.
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
			// Fetch timeout / transient — loop. (Match thunder's handling.)
			continue
		}
		a.handleMessage(ctx, msg)
	}
}

// startHealthServer serves /health and /ready for the k8s probes (mirrors the
// thunder adapter, which also runs a small HTTP server alongside the consumer).
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

// incoming reuses the canonical RequestHeaders and captures the body as raw
// JSON so it can be unmarshalled into the action's typed request.
type incoming struct {
	Headers types.RequestHeaders `json:"headers"`
	Body    json.RawMessage      `json:"body"`
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
	switch in.Headers.Action {
	case "analyse":
		var req AnalyseRequest
		if execErr = json.Unmarshal(in.Body, &req); execErr == nil {
			result, execErr = a.analyse.Execute(ctx, req)
		}
	default:
		execErr = fmt.Errorf("analyser adapter: action %q not implemented", in.Headers.Action)
	}

	resp := a.buildResponse(ec, result, execErr)

	if in.Headers.ResponsesTopic == "" {
		a.logger.Error("no responses_topic on request — cannot reply",
			zap.String("correlation_id", in.Headers.CorrelationID),
			zap.String("request_id", in.Headers.RequestID))
		_ = a.consumer.CommitMessages(ctx, msg)
		return
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		a.logger.Error("marshal response failed", zap.Error(err))
		_ = a.consumer.CommitMessages(ctx, msg)
		return
	}

	// Produce to the caller's responses topic. Align the exact Produce
	// signature (and whether Kafka message headers are set separately) with
	// thunder/adapter.go.
	if err := a.producer.Produce(ctx, in.Headers.ResponsesTopic, []byte(in.Headers.CorrelationID), payload); err != nil {
		a.logger.Error("produce response failed",
			zap.Error(err), zap.String("topic", in.Headers.ResponsesTopic))
		return // don't commit — let it redeliver (at-least-once)
	}

	_ = a.consumer.CommitMessages(ctx, msg)
	a.logger.Info("analyse response sent",
		zap.String("correlation_id", in.Headers.CorrelationID),
		zap.String("in_response_to_request_id", in.Headers.RequestID),
		zap.String("status", resp.Headers.Status),
		zap.String("topic", in.Headers.ResponsesTopic))
}

// buildResponse constructs a ResponseMessage via the platform's
// ToResponseHeaders, so the envelope matches what the orchestration engine
// expects. Status uses the engine's vocabulary (complete / error_unrecoverable).
func (a *Adapter) buildResponse(ec *types.ExecutionContext, result interface{}, execErr error) types.ResponseMessage {
	if execErr != nil {
		ec.IsError = true
		ec.IsComplete = false
		ec.Status = "error_unrecoverable"
	} else {
		ec.IsError = false
		ec.IsComplete = true
		ec.Status = "complete"
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
