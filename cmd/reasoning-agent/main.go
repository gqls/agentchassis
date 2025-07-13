// FILE: cmd/reasoning-agent/main.go
// This is the main entrypoint for our new, specialized Reasoning Agent service.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqls/ai-persona-system/internal/agents/reasoning" // The agent's specific logic
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/gqls/ai-persona-system/platform/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. Load the service's static configuration from its dedicated file.
	configPath := flag.String("config", "configs/reasoning-agent.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize the standardized platform logger.
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// 3. Create a cancellable context for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Initialize the Reasoning Agent.
	// This is where we inject its dependencies from the platform.
	agent, err := reasoning.NewAgent(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize reasoning agent", zap.Error(err))
	}

	// 5. Start the agent's main run loop in a goroutine.
	go func() {
		if err := agent.Run(); err != nil {
			appLogger.Error("Reasoning agent failed to run", zap.Error(err))
			cancel() // Trigger shutdown
		}
	}()

	// 6. Wait for a shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutdown signal received, shutting down reasoning agent...")

	cancel()

	time.Sleep(2 * time.Second)
	appLogger.Info("Reasoning agent service stopped.")
}
```go
// FILE: internal/agents/reasoning/agent.go
// This file contains the core custom logic for the Reasoning Agent.
package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"[github.com/google/uuid](https://github.com/google/uuid)"
	"[github.com/gqls/ai-persona-system/platform/aiservice](https://github.com/gqls/ai-persona-system/platform/aiservice)"
	"[github.com/gqls/ai-persona-system/platform/config](https://github.com/gqls/ai-persona-system/platform/config)"
	"[github.com/gqls/ai-persona-system/platform/kafka](https://github.com/gqls/ai-persona-system/platform/kafka)"
	"go.uber.org/zap"
)

const (
	requestTopic  = "system.agent.reasoning.process"
	responseTopic = "system.responses.reasoning"
	consumerGroup = "reasoning-agent-group"
)

// RequestPayload defines the data this agent expects.
type RequestPayload struct {
	Action string `json:"action"` // e.g., "review_for_clarity", "deductive_reasoning"
	Data   struct {
		ContentToReview string                 `json:"content_to_review"`
		ReviewCriteria  []string               `json:"review_criteria"`
		BriefContext    map[string]interface{} `json:"brief_context"`
	} `json:"data"`
}

// ResponsePayload defines the data this agent produces.
type ResponsePayload struct {
	ReviewPassed bool     `json:"review_passed"`
	Score        float64  `json:"score"`
	Suggestions  []string `json:"suggestions"`
	Reasoning    string   `json:"reasoning"`
}

// Agent is the main struct for this specialized service.
type Agent struct {
	ctx           context.Context
	logger        *zap.Logger
	consumer      *kafka.Consumer
	producer      *kafka.Producer
	aiClient      aiservice.AIService // Using the platform's AI abstraction
}

// NewAgent initializes the agent and its dependencies.
func NewAgent(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error) {
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestTopic, consumerGroup, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Initialize the specific AI client this agent will use.
	// The config for this would be in reasoning-agent.yaml.
	// This demonstrates how a specialized agent can choose its own AI model.
	aiClient, err := aiservice.New(ctx, cfg.Custom["ai_service"]) // Assuming a generic factory
	if err != nil {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &Agent{
		ctx:           ctx,
		logger:        logger,
		consumer:      consumer,
		producer:      producer,
		aiClient:      aiClient,
	}, nil
}

// Run starts the consumer loop.
func (a *Agent) Run() error {
	a.logger.Info("Reasoning Agent is running and waiting for tasks...")
	for {
		select {
		case <-a.ctx.Done():
			a.consumer.Close()
			a.producer.Close()
			return nil
		default:
			msg, err := a.consumer.FetchMessage(a.ctx)
			if err != nil {
				if err == context.Canceled { continue }
				a.logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}
			go a.handleMessage(msg)
		}
	}
}

// handleMessage contains the agent's unique business logic.
func (a *Agent) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(zap.String("correlation_id", headers["correlation_id"]))

	var req RequestPayload
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.Error("Failed to unmarshal request payload", zap.Error(err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// --- THIS IS THE CUSTOM LOGIC ---
	// Here, we would implement the specific reasoning algorithm.
	// For now, we'll use the AI client to simulate it.
	prompt := a.buildReasoningPrompt(req)
	reasoningResult, err := a.aiClient.GenerateText(a.ctx, prompt)
	if err != nil {
		l.Error("AI reasoning call failed", zap.Error(err))
		// Produce an error response or retry
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Assume the AI returns a JSON string that we can parse into our response format.
	var responsePayload ResponsePayload
	if err := json.Unmarshal([]byte(reasoningResult), &responsePayload); err != nil {
		l.Error("Failed to parse reasoning result from AI", zap.Error(err))
		// Fallback response
		responsePayload.Reasoning = "Could not parse AI response, but original content was: " + req.Data.ContentToReview
	}
	// --- END OF CUSTOM LOGIC ---

	// Produce the standard response message.
	responseBytes, _ := json.Marshal(responsePayload)
	responseHeaders := map[string]string{
		"correlation_id": headers["correlation_id"],
		"causation_id":   headers["request_id"],
		"request_id":     uuid.NewString(),
	}
	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders, msg.Key, responseBytes); err != nil {
		l.Error("Failed to produce response message", zap.Error(err))
	}

	a.consumer.CommitMessages(context.Background(), msg)
}

// buildReasoningPrompt creates the prompt for the LLM.
func (a *Agent) buildReasoningPrompt(req RequestPayload) string {
	return fmt.Sprintf(
		"You are a logical reasoning engine. Review the following content based on these criteria: %v. "+
		"The original goal was: %v. Content to review: '%s'. "+
		"Respond with a JSON object with keys: 'review_passed' (boolean), 'score' (float 0-10), 'suggestions' (array of strings), and 'reasoning' (string).",
		req.Data.ReviewCriteria, req.Data.BriefContext, req.Data.ContentToReview,
	)
}
```yaml
# FILE: configs/reasoning-agent.yaml
# This is a dedicated configuration file for our new specialized service.
service_info:
name: "reasoning-agent"
version: "1.0.0"
environment: "development"

server:
port: "8082" # Runs on a different port from other services

logging:
level: "info"

infrastructure:
kafka_brokers:
- "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
# This agent might not need direct DB access if it's purely computational,
# but it could be configured here if needed.
clients_database: {}
templates_database: {}
auth_database: {}
object_storage: {}

# Custom section for this agent's specific needs
custom:
ai_service:
provider: "anthropic"
model: "claude-3-opus-20240229" # Use a powerful model for reasoning tasks
temperature: 0.2
max_tokens: 2048
api_key_env_var: "ANTHROPIC_API_KEY"
```dockerfile
# FILE: Dockerfile.reasoning
# A dedicated Dockerfile for building the reasoning agent service.

# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project context
COPY . .

# Build the specific service binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/reasoning-agent ./cmd/reasoning-agent

# Stage 2: Create the final small image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S -G appgroup appuser

WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/reasoning-agent /app/reasoning-agent

# Copy its specific configuration file
COPY configs/reasoning-agent.yaml /app/configs/reasoning-agent.yaml

RUN chown appuser:appgroup /app/reasoning-agent

USER appuser

# The command to run the service, pointing to its own config
CMD ["./reasoning-agent", "-config", "configs/reasoning-agent.yaml"]
