package messaging

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// MessageContext holds the context for processing a single message
type MessageContext struct {
	Message          kafka.Message
	Headers          map[string]string
	ExecutionContext *types.ExecutionContext // NEW: Primary context
	Logger           *zap.Logger
	CollectedData    map[string]interface{}
	StartTime        time.Time

	// For stateless operation
	OrchestrationID   string
	OrchestrationName string
	IsStateless       bool
}

// IsChildOrchestration checks if this is a child orchestration
func (mc *MessageContext) IsChildOrchestration() bool {
	if mc.ExecutionContext != nil {
		return mc.ExecutionContext.ParentOrchestrationID != ""
	}
	return mc.Headers["parent_orchestration_id"] != ""
}

// ValidateContext validates the message context
func (mc *MessageContext) ValidateContext() error {
	if mc.ExecutionContext != nil {
		return mc.ExecutionContext.Validate()
	}

	// Basic validation for legacy messages
	if mc.Headers["correlation_id"] == "" {
		return fmt.Errorf("missing correlation_id")
	}
	if mc.Headers["orchestration_id"] == "" {
		return fmt.Errorf("missing orchestration_id")
	}
	return nil
}

// ValidateHeaders provides backward compatibility
func (m *MessageContext) ValidateHeaders() error {
	return m.ValidateContext()
}

// GetParentContext returns parent orchestration info if this is a child
func (m *MessageContext) GetParentContext() map[string]interface{} {
	if !m.IsChildOrchestration() {
		return nil
	}
	return map[string]interface{}{
		"orchestration_id": m.ExecutionContext.ParentOrchestrationID,
		"request_id":       m.ExecutionContext.RequestID,
		"responses_topic":  m.ExecutionContext.ResponsesTopic,
	}
}

// CreateChildContext creates a new ExecutionContext for calling another agent
func (m *MessageContext) CreateChildContext(toAgentID, toAgentType string) *types.ExecutionContext {
	return m.ExecutionContext.CreateChildContext(toAgentType)
}

// CreateResponseContext creates a response context for sending responses
func (mc *MessageContext) CreateResponseContext() *types.ExecutionContext {
	if mc.ExecutionContext != nil {
		return mc.ExecutionContext.CreateResponseContext("complete", 100)
	}

	// Fallback for legacy messages
	return &types.ExecutionContext{
		MessageID:       uuid.New().String(),
		CorrelationID:   mc.Headers["correlation_id"],
		OrchestrationID: mc.Headers["orchestration_id"],
		RequestID:       mc.Headers["request_id"],
		InResponseTo: &types.ResponseContext{
			RequestID:             mc.Headers["request_id"],
			ParentOrchestrationID: mc.Headers["orchestration_id"],
			MessageID:             mc.Headers["message_id"],
		},
		FromAgentID:    mc.Headers["agent_id"],
		FromAgentType:  mc.Headers["agent_type"],
		ToAgentID:      mc.Headers["from_agent_id"],
		ToAgentType:    mc.Headers["from_agent_type"],
		ResponsesTopic: mc.Headers["responses_topic"],
		Timestamp:      time.Now(),
		Version:        "2.0",
	}
}

// CreateResponseHeaders creates response headers for the current context
func (mc *MessageContext) CreateResponseHeaders(agentType string) map[string]string {
	if mc.ExecutionContext != nil {
		responseCtx := mc.ExecutionContext.CreateResponseContext("complete", 100)
		return responseCtx.ToHeaders()
	}

	// Legacy header creation
	return map[string]string{
		"correlation_id":   mc.Headers["correlation_id"],
		"orchestration_id": mc.Headers["orchestration_id"],
		"in_response_to":   mc.Headers["request_id"],
		"message_type":     "response",
		"from_agent_type":  agentType,
		"timestamp":        time.Now().Format(time.RFC3339),
	}
}

// SyncHeadersFromContext syncs ExecutionContext back to headers
func (mc *MessageContext) SyncHeadersFromContext() {
	if mc.ExecutionContext != nil {
		mc.Headers = mc.ExecutionContext.ToHeaders()
	}
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

func NewMessageContext(msg kafka.Message, headers map[string]string, receivingAgentType, receivingAgentRole string, logger *zap.Logger) (*MessageContext, error) {

	logger.Info("IN NewMessageContext",
		zap.Any("incoming headers are:", headers))

	// Parse sender's context from all headers.
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		// Log the error but continue, as the validation step will catch critical missing fields.
		logger.Warn("Could not fully parse headers into ExecutionContext", zap.Error(err))
		if execCtx == nil {
			// If parsing failed completely, create a minimal context to avoid a nil pointer.
			execCtx = &types.ExecutionContext{}
		}
	}

	// The previous logic incorrectly created a new, empty context for regular requests,
	// discarding essential fields like message_type.
	// By using the parsed context directly, we preserve all incoming header information.
	// The process() function in the message processor is responsible for deciding
	// if a new orchestration ID needs to be generated if one isn't present.

	// Set the identity of this agent as the 'Sender' for any subsequent messages it might create.
	execCtx.Sender = types.AgentIdentity{
		AgentType:    receivingAgentType,
		AgentID:      os.Getenv("AGENT_ID"),
		PodName:      os.Getenv("HOSTNAME"),
		AgentVersion: os.Getenv("AGENT_VERSION"),
		Role:         receivingAgentRole,
	}

	return &MessageContext{
		Message:          msg,
		ExecutionContext: execCtx,
		Headers:          headers, // Keep original headers for reference
		StartTime:        time.Now(),
		CollectedData:    make(map[string]interface{}),
	}, nil
}
