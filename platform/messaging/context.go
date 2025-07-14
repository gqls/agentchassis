// FILE: platform/messaging/context.go
package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

// MessageContext holds the context for processing a single message
type MessageContext struct {
	Message   kafka.Message
	Headers   map[string]string
	Action    string
	StartTime time.Time
	Logger    *zap.Logger
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
