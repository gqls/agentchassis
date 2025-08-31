package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TraceLogger logs complete message traces for debugging
type TraceLogger struct {
	logger    *zap.Logger
	traces    map[string]*MessageTrace
	mu        sync.RWMutex
	enabled   bool
	outputDir string
}

type MessageTrace struct {
	OrchestrationID string              `json:"orchestrationId"`
	CorrelationID   string              `json:"correlation_id"`
	StartTime       time.Time           `json:"start_time"`
	Messages        []TracedMessage     `json:"messages"`
	Tree            map[string][]string `json:"tree"` // parent -> children mapping
}

type TracedMessage struct {
	Timestamp        time.Time               `json:"timestamp"`
	ExecutionContext *types.ExecutionContext `json:"context"`
	Direction        string                  `json:"direction"`
	Topic            string                  `json:"topic"`
	PayloadSize      int                     `json:"payload_size"`
	PayloadPreview   string                  `json:"payload_preview"`
	Action           string                  `json:"action,omitempty"`
	AwaitedSteps     []string                `json:"awaited_steps,omitempty"`
	Error            string                  `json:"error,omitempty"`
}

func NewTraceLogger(logger *zap.Logger) *TraceLogger {
	return &TraceLogger{
		logger:    logger,
		traces:    make(map[string]*MessageTrace),
		enabled:   os.Getenv("ENABLE_MESSAGE_TRACING") == "true",
		outputDir: os.Getenv("TRACE_OUTPUT_DIR"),
	}
}

// Update the TraceMessage method to accept payload
func (t *TraceLogger) TraceMessage(execCtx interface{}, direction, topic string, payload interface{}) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Type assert to ExecutionContext
	ctx, ok := execCtx.(*types.ExecutionContext)
	if !ok {
		t.logger.Warn("Invalid execution context type in TraceMessage")
		return
	}

	// Create payload preview
	payloadPreview := t.createPayloadPreview(payload)
	payloadSize := len(fmt.Sprintf("%v", payload))

	trace, exists := t.traces[ctx.CorrelationID]
	if !exists {
		trace = &MessageTrace{
			OrchestrationID: ctx.OrchestrationID,
			CorrelationID:   ctx.CorrelationID,
			StartTime:       time.Now(),
			Messages:        []TracedMessage{},
			Tree:            make(map[string][]string),
		}
		t.traces[ctx.CorrelationID] = trace
	}

	// Add message to trace
	tracedMsg := TracedMessage{
		Timestamp:        time.Now(),
		ExecutionContext: ctx,
		Direction:        direction,
		Topic:            topic,
		PayloadSize:      payloadSize,
		PayloadPreview:   payloadPreview,
	}

	// Extract action from payload if possible
	if action := t.extractAction(payload); action != "" {
		tracedMsg.Action = action
	}

	trace.Messages = append(trace.Messages, tracedMsg)

	// Update tree structure
	if ctx.ParentOrchestrationID != "" {
		children := trace.Tree[ctx.ParentOrchestrationID]
		found := false
		for _, child := range children {
			if child == ctx.OrchestrationID {
				found = true
				break
			}
		}
		if !found {
			trace.Tree[ctx.ParentOrchestrationID] = append(children, ctx.OrchestrationID)
		}
	}

	// Enhanced logging with payload preview
	t.logger.Info("MESSAGE_TRACE",
		zap.String("correlation_id", ctx.CorrelationID),
		zap.String("orchestration_id", ctx.OrchestrationID),
		zap.String("request_id", ctx.RequestID),
		zap.String("in_response_to", ctx.InResponseTo),
		zap.String("direction", direction),
		zap.String("topic", topic),
		zap.String("from", fmt.Sprintf("%s/%s", ctx.FromAgentID, ctx.FromAgentType)),
		zap.String("to", fmt.Sprintf("%s/%s", ctx.ToAgentID, ctx.ToAgentType)),
		zap.String("payload_preview", payloadPreview),
		zap.Int("message_count", len(trace.Messages)))
}

// Helper method to create a safe payload preview
func (t *TraceLogger) createPayloadPreview(payload interface{}) string {
	const maxPreviewLength = 500

	var preview string

	switch v := payload.(type) {
	case []byte:
		// Try to parse as JSON first
		var jsonData interface{}
		if err := json.Unmarshal(v, &jsonData); err == nil {
			preview = t.formatJSON(jsonData, maxPreviewLength)
		} else {
			// Fall back to string
			preview = string(v)
		}
	case string:
		preview = v
	case nil:
		preview = "<nil>"
	default:
		// Try to marshal as JSON
		if data, err := json.Marshal(v); err == nil {
			preview = string(data)
		} else {
			preview = fmt.Sprintf("%v", v)
		}
	}

	// Truncate if too long
	if len(preview) > maxPreviewLength {
		preview = preview[:maxPreviewLength] + "..."
	}

	return preview
}

// Helper to format JSON with limited depth
func (t *TraceLogger) formatJSON(data interface{}, maxLen int) string {
	// Marshal with indentation for readability
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}

	result := string(formatted)
	if len(result) > maxLen {
		result = result[:maxLen] + "..."
	}
	return result
}

// Helper to extract action from payload
func (t *TraceLogger) extractAction(payload interface{}) string {
	// Try to extract action field from various payload types
	switch v := payload.(type) {
	case []byte:
		var data map[string]interface{}
		if err := json.Unmarshal(v, &data); err == nil {
			if action, ok := data["action"].(string); ok {
				return action
			}
		}
	case map[string]interface{}:
		if action, ok := v["action"].(string); ok {
			return action
		}
	}
	return ""
}

func (t *TraceLogger) DumpTrace(correlationID string) {
	if !t.enabled {
		return
	}

	t.mu.RLock()
	trace, exists := t.traces[correlationID]
	t.mu.RUnlock()

	if !exists {
		return
	}

	// Write to file if output directory is configured
	if t.outputDir != "" {
		filename := fmt.Sprintf("%s/trace_%s_%d.json",
			t.outputDir, correlationID, time.Now().Unix())

		data, _ := json.MarshalIndent(trace, "", "  ")
		os.WriteFile(filename, data, 0644)

		t.logger.Info("Trace dumped to file",
			zap.String("correlation_id", correlationID),
			zap.String("filename", filename),
			zap.Int("message_count", len(trace.Messages)))
	}

	// Also log a summary
	t.logger.Info("Message trace summary",
		zap.String("correlation_id", correlationID),
		zap.Int("total_messages", len(trace.Messages)),
		zap.Duration("duration", time.Since(trace.StartTime)),
		zap.Any("orchestration_tree", trace.Tree))
}

func (t *TraceLogger) TraceAwaitedSteps(execCtx *types.ExecutionContext, awaitedSteps []string, action string) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	trace, exists := t.traces[execCtx.CorrelationID]
	if !exists {
		trace = &MessageTrace{
			CorrelationID: execCtx.CorrelationID,
			StartTime:     time.Now(),
			Messages:      []TracedMessage{},
			Tree:          make(map[string][]string),
		}
		t.traces[execCtx.CorrelationID] = trace
	}

	trace.Messages = append(trace.Messages, TracedMessage{
		Timestamp:        time.Now(),
		ExecutionContext: execCtx,
		Direction:        "awaited_update",
		Action:           action,
		AwaitedSteps:     awaitedSteps,
	})

	t.logger.Info("AWAITED_STEPS_CHANGED",
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.Strings("awaited_steps", awaitedSteps),
		zap.String("for_action", action),
		zap.String("request_id", execCtx.RequestID))
}

// Add method to trace errors
func (t *TraceLogger) TraceError(execCtx *types.ExecutionContext, err error, context string) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	trace, exists := t.traces[execCtx.CorrelationID]
	if !exists {
		return
	}

	trace.Messages = append(trace.Messages, TracedMessage{
		Timestamp:        time.Now(),
		ExecutionContext: execCtx,
		Direction:        "error",
		Error:            fmt.Sprintf("%s: %v", context, err),
	})
}
