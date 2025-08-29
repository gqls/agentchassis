// FILE: platform/messaging/context.go
package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

// MessageContext holds the context for processing a single message
type MessageContext struct {
	// Original message
	Message      kafka.Message
	AgentMessage *models.AgentMessage // Parsed message

	// Execution context
	OrchestrationID       string
	ParentOrchestrationID string
	RequestID             string
	ParentRequestID       string // What parent is waiting for

	// Headers (complete set)
	Headers map[string]string

	// Processing state
	Action        string
	CollectedData map[string]interface{}
	StartTime     time.Time

	// Logger with context
	Logger *zap.Logger
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

// ValidateHeaders ensures required headers are present
func (m *MessageContext) ValidateHeaders() error {
	required := []string{"correlation_id", "request_id", "client_id", "agent_instance_id"}
	for _, key := range required {
		if m.Headers[key] == "" {
			return fmt.Errorf("missing required header: %s", key)
		}
	}
	return nil
}

// CreateResponseHeaders creates headers for a response message
func (m *MessageContext) CreateResponseHeaders(agentType string) map[string]string {
	return map[string]string{
		"correlation_id": m.Headers["correlation_id"],
		"causation_id":   m.Headers["request_id"],
		"request_id":     uuid.NewString(),
		"client_id":      m.Headers["client_id"],
		"agent_type":     agentType,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
}

func NewMessageContext(msg kafka.Message, headers map[string]string) *MessageContext {
	return &MessageContext{
		Message:               msg,
		Headers:               headers,
		OrchestrationID:       headers["orchestration_id"],
		ParentOrchestrationID: headers["parent_orchestration_id"],
		RequestID:             headers["request_id"],
		ParentRequestID:       headers["parent_request_id"],
		StartTime:             time.Now(),
		CollectedData:         make(map[string]interface{}),
	}
}

func (mc *MessageContext) IsChildOrchestration() bool {
	return mc.ParentOrchestrationID != "" && mc.ParentRequestID != ""
}

func (mc *MessageContext) GetParentContext() map[string]interface{} {
	if !mc.IsChildOrchestration() {
		return nil
	}
	return map[string]interface{}{
		"orchestration_id": mc.ParentOrchestrationID,
		"request_id":       mc.ParentRequestID,
		"reply_to_topic":   mc.Headers["reply_to_topic"],
	}
}
