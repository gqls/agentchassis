package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// MessageContext holds the context for processing a single message
type MessageContext struct {
	// Original message
	Message kafka.Message

	// Execution context - primary source of truth
	ExecutionContext *types.ExecutionContext

	// Legacy headers - for backward compatibility only
	Headers map[string]string

	// Processing state
	Action        string
	CollectedData map[string]interface{}
	StartTime     time.Time

	// Logger with context
	Logger *zap.Logger
}

// NewMessageContext creates a new message context with ExecutionContext
func NewMessageContext(msg kafka.Message, headers map[string]string) (*MessageContext, error) {
	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		// Try to create minimal context for error handling
		execCtx = &types.ExecutionContext{
			CorrelationID: headers["correlation_id"],
			ClientID:      headers["client_id"],
			MessageType:   "request",
			Timestamp:     time.Now().UTC(),
			Version:       "2.0",
		}
	}

	return &MessageContext{
		Message:          msg,
		ExecutionContext: execCtx,
		Headers:          headers, // Keep for backward compatibility
		StartTime:        time.Now(),
		CollectedData:    make(map[string]interface{}),
	}, nil
}

// ExtractAction extracts the action from the message payload
func (m *MessageContext) ExtractAction() error {
	var payload struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(m.Message.Value, &payload); err != nil {
		return fmt.Errorf("failed to extract action: %w", err)
	}
	m.Action = payload.Action
	return nil
}

// ValidateContext ensures the ExecutionContext is valid
func (m *MessageContext) ValidateContext() error {
	return m.ExecutionContext.Validate()
}

// ValidateHeaders provides backward compatibility
func (m *MessageContext) ValidateHeaders() error {
	return m.ValidateContext()
}

// IsChildOrchestration returns true if this is a child orchestration
func (m *MessageContext) IsChildOrchestration() bool {
	return m.ExecutionContext.IsChildOrchestration()
}

// GetParentContext returns parent orchestration info if this is a child
func (m *MessageContext) GetParentContext() map[string]interface{} {
	if !m.IsChildOrchestration() {
		return nil
	}
	return map[string]interface{}{
		"orchestration_id": m.ExecutionContext.ParentOrchestrationID,
		"request_id":       m.ExecutionContext.RequestID,
		"reply_to_topic":   m.ExecutionContext.ReplyToTopic,
	}
}

// CreateChildContext creates a new ExecutionContext for calling another agent
func (m *MessageContext) CreateChildContext(toAgentID, toAgentType string) *types.ExecutionContext {
	return m.ExecutionContext.CreateChildContext(toAgentID, toAgentType)
}

// CreateResponseContext creates a context for responding
func (m *MessageContext) CreateResponseContext() *types.ExecutionContext {
	return m.ExecutionContext.CreateResponseContext()
}

// CreateResponseHeaders provides backward compatibility
func (m *MessageContext) CreateResponseHeaders(agentType string) map[string]string {
	responseCtx := m.CreateResponseContext()
	return responseCtx.ToHeaders()
}

// SyncHeadersFromContext updates legacy headers from ExecutionContext
func (m *MessageContext) SyncHeadersFromContext() {
	m.Headers = m.ExecutionContext.ToHeaders()
}

// SyncContextFromHeaders updates ExecutionContext from legacy headers
func (m *MessageContext) SyncContextFromHeaders() error {
	newCtx, err := types.FromHeaders(m.Headers)
	if err != nil {
		return err
	}
	m.ExecutionContext = newCtx
	return nil
}
