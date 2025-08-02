// FILE: internal/agents/reasoning/agent.go
package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

const (
	requestTopic  = "system.agent.reasoning.process"
	responseTopic = "system.responses.reasoning"
	consumerGroup = "reasoning-agent-group"
)

// RequestPayload defines the data this agent expects
type RequestPayload struct {
	Action string `json:"action"`
	Data   struct {
		ContentToReview string                 `json:"content_to_review"`
		ReviewCriteria  []string               `json:"review_criteria"`
		BriefContext    map[string]interface{} `json:"brief_context"`
	} `json:"data"`
}

// ResponsePayload defines the response format
type ResponsePayload struct {
	ReviewPassed bool     `json:"review_passed"`
	Score        float64  `json:"score"`
	Suggestions  []string `json:"suggestions"`
	Reasoning    string   `json:"reasoning"`
}

// Agent is the reasoning specialist
type Agent struct {
	ctx      context.Context
	logger   *zap.Logger
	consumer *kafka.Consumer
	producer kafka.Producer
	aiClient aiservice.AIService
}

// NewAgent creates a new reasoning agent
func NewAgent(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error) {
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestTopic, consumerGroup, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Initialize AI client from custom config
	aiConfig := cfg.Custom["ai_service"].(map[string]interface{})
	aiClient, err := aiservice.NewAnthropicClient(ctx, aiConfig)
	if err != nil {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &Agent{
		ctx:      ctx,
		logger:   logger,
		consumer: consumer,
		producer: producer,
		aiClient: aiClient,
	}, nil
}

// Run starts the agent's main loop
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
				if err == context.Canceled {
					continue
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}
			go a.handleMessage(msg)
		}
	}
}

// handleMessage processes a single reasoning request
func (a *Agent) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(zap.String("correlation_id", headers["correlation_id"]))

	var req RequestPayload
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.Error("Failed to unmarshal request", zap.Error(err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Build the reasoning prompt
	prompt := a.buildReasoningPrompt(req)

	// Call the AI service
	result, err := a.aiClient.GenerateText(a.ctx, prompt, nil)
	if err != nil {
		l.Error("AI reasoning call failed", zap.Error(err))
		a.sendErrorResponse(headers, "Failed to perform reasoning")
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Parse the AI response
	// Extract and parse the JSON response
	jsonStr := a.extractJSON(result)
	var responsePayload ResponsePayload

	if err := json.Unmarshal([]byte(jsonStr), &responsePayload); err != nil {
		l.Error("Failed to parse AI response", zap.Error(err), zap.String("extracted_json", jsonStr))
		// Fallback response with the full reasoning text
		responsePayload = ResponsePayload{
			ReviewPassed: false,
			Score:        0,
			Suggestions:  []string{"Could not parse AI response"},
			Reasoning:    result,
		}
	}

	// Send response
	a.sendResponse(headers, responsePayload)

	// Commit message
	a.consumer.CommitMessages(context.Background(), msg)
}

// extractJSON attempts to extract JSON from a response that might contain additional text
func (a *Agent) extractJSON(response string) string {
	// First, try to find JSON object boundaries
	startIdx := strings.Index(response, "{")
	if startIdx == -1 {
		return response // No JSON found, return as is
	}

	// Find the matching closing brace
	braceCount := 0
	endIdx := -1
	inString := false
	escapeNext := false

	for i := startIdx; i < len(response); i++ {
		char := response[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' && !escapeNext {
			inString = !inString
			continue
		}

		if !inString {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					endIdx = i + 1
					break
				}
			}
		}
	}

	if endIdx != -1 {
		return response[startIdx:endIdx]
	}

	// If we couldn't find proper JSON boundaries, return the original
	return response
}

// buildReasoningPrompt creates the prompt for the LLM
func (a *Agent) buildReasoningPrompt(req RequestPayload) string {
	return fmt.Sprintf(`You are a logical reasoning engine. Review the following content based on these criteria: %v.

Context: %v

Content to review: "%s"

IMPORTANT: You must respond with ONLY a valid JSON object, nothing else. Do not include any explanatory text before or after the JSON. Do not wrap the JSON in markdown code blocks.

Return your analysis as a pure JSON object with exactly this structure:
{
    "review_passed": boolean,
    "score": number (0-10),
    "suggestions": ["suggestion1", "suggestion2", ...],
    "reasoning": "detailed explanation of your analysis"
}

Example response:
{
    "review_passed": true,
    "score": 8.5,
    "suggestions": ["Consider A/B testing the budget allocation", "Define specific KPIs for each channel"],
    "reasoning": "The strategy shows strong alignment with the target demographic."
}

Be thorough but concise in your reasoning. Remember: respond with ONLY the JSON object.`,
		req.Data.ReviewCriteria,
		req.Data.BriefContext,
		req.Data.ContentToReview,
	)
}

// sendResponse sends a successful response
func (a *Agent) sendResponse(headers map[string]string, payload ResponsePayload) {
	responseBytes, _ := json.Marshal(payload)
	responseHeaders := map[string]string{
		"correlation_id": headers["correlation_id"],
		"causation_id":   headers["request_id"],
		"request_id":     uuid.NewString(),
	}

	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to produce response", zap.Error(err))
	}
}

// sendErrorResponse sends an error response
func (a *Agent) sendErrorResponse(headers map[string]string, errorMsg string) {
	payload := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}
	responseBytes, _ := json.Marshal(payload)
	responseHeaders := map[string]string{
		"correlation_id": headers["correlation_id"],
		"causation_id":   headers["request_id"],
		"request_id":     uuid.NewString(),
	}

	a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes)
}

// StartHealthServer starts a simple HTTP server for health checks
func (a *Agent) StartHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"agent":  "reasoning-agent",
		})
	})

	go func() {
		a.logger.Info("Starting health server", zap.String("port", port))
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			a.logger.Error("Health server failed", zap.Error(err))
		}
	}()
}
