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
	CorrelationID string              `json:"correlation_id"`
	StartTime     time.Time           `json:"start_time"`
	Messages      []TracedMessage     `json:"messages"`
	Tree          map[string][]string `json:"tree"` // parent -> children mapping
}

type TracedMessage struct {
	Timestamp        time.Time               `json:"timestamp"`
	ExecutionContext *types.ExecutionContext `json:"context"`
	Direction        string                  `json:"direction"` // "sent" or "received"
	Topic            string                  `json:"topic"`
	PayloadSize      int                     `json:"payload_size"`
}

func NewTraceLogger(logger *zap.Logger) *TraceLogger {
	return &TraceLogger{
		logger:    logger,
		traces:    make(map[string]*MessageTrace),
		enabled:   os.Getenv("ENABLE_MESSAGE_TRACING") == "true",
		outputDir: os.Getenv("TRACE_OUTPUT_DIR"),
	}
}

func (t *TraceLogger) TraceMessage(execCtx *types.ExecutionContext, direction, topic string, payloadSize int) {
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

	// Add message to trace
	trace.Messages = append(trace.Messages, TracedMessage{
		Timestamp:        time.Now(),
		ExecutionContext: execCtx,
		Direction:        direction,
		Topic:            topic,
		PayloadSize:      payloadSize,
	})

	// Update tree structure
	if execCtx.ParentOrchestrationID != "" {
		children := trace.Tree[execCtx.ParentOrchestrationID]
		found := false
		for _, child := range children {
			if child == execCtx.OrchestrationID {
				found = true
				break
			}
		}
		if !found {
			trace.Tree[execCtx.ParentOrchestrationID] = append(children, execCtx.OrchestrationID)
		}
	}

	// Log the trace entry
	t.logger.Debug("Message traced",
		zap.String("correlation_id", execCtx.CorrelationID),
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("direction", direction),
		zap.Int("total_messages", len(trace.Messages)))
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
