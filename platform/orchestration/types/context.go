package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ExecutionContext represents the complete context for message execution
type ExecutionContext struct {
	// Business Transaction
	CorrelationID string `json:"correlation_id"`
	ClientID      string `json:"client_id"`

	// Execution Instance - UNIQUE per agent
	OrchestrationID       string `json:"orchestration_id"`
	ParentOrchestrationID string `json:"parent_orchestration_id,omitempty"`

	// Group Management
	GroupID        string `json:"group_id,omitempty"`
	FunctionalRole string `json:"functional_role,omitempty"`

	// Message Identity
	MessageID    string `json:"message_id"`
	RequestID    string `json:"request_id"`
	MessageType  string `json:"message_type"` // "request" | "response" | "error"
	InResponseTo string `json:"in_response_to,omitempty"`

	// Routing
	OwnerAgentID  string `json:"owner_agent_id"`
	FromAgentID   string `json:"from_agent_id"`
	FromAgentType string `json:"from_agent_type"`
	ToAgentID     string `json:"to_agent_id"`
	ToAgentType   string `json:"to_agent_type"`
	ReplyToTopic  string `json:"reply_to_topic"`

	// Resource Management
	FuelBudget int `json:"fuel_budget"`

	// Metadata
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// NewExecutionContext creates a new execution context for an agent
func NewExecutionContext(correlationID, clientID, agentID, agentType string) *ExecutionContext {
	return &ExecutionContext{
		CorrelationID:   correlationID,
		ClientID:        clientID,
		OrchestrationID: uuid.New().String(),
		MessageID:       uuid.New().String(),
		RequestID:       uuid.New().String(),
		MessageType:     "request",
		OwnerAgentID:    agentID,
		FromAgentID:     agentID,
		FromAgentType:   agentType,
		FuelBudget:      1000,
		Timestamp:       time.Now().UTC(),
		Version:         "2.0",
	}
}

// CreateChildContext creates a new context for a child orchestration
func (ec *ExecutionContext) CreateChildContext(toAgentID, toAgentType string) *ExecutionContext {
	return &ExecutionContext{
		// Keep business context
		CorrelationID: ec.CorrelationID,
		ClientID:      ec.ClientID,
		GroupID:       ec.GroupID,

		// NEW orchestration for the child
		OrchestrationID:       uuid.New().String(),
		ParentOrchestrationID: ec.OrchestrationID,

		// New message identity
		MessageID:   uuid.New().String(),
		RequestID:   uuid.New().String(),
		MessageType: "request",

		// Routing
		OwnerAgentID:  toAgentID,
		FromAgentID:   ec.OwnerAgentID,
		FromAgentType: ec.FromAgentType,
		ToAgentID:     toAgentID,
		ToAgentType:   toAgentType,
		ReplyToTopic:  fmt.Sprintf("system.agent.%s.responses", ec.FromAgentType),

		// Inherit fuel budget
		FuelBudget: ec.FuelBudget,
		Timestamp:  time.Now().UTC(),
		Version:    "2.0",
	}
}

// CreateResponseContext creates a context for responding to a request
func (ec *ExecutionContext) CreateResponseContext() *ExecutionContext {
	return &ExecutionContext{
		// Keep business context
		CorrelationID: ec.CorrelationID,
		ClientID:      ec.ClientID,
		GroupID:       ec.GroupID,

		// Use PARENT's orchestration if responding to parent
		OrchestrationID:       ec.ParentOrchestrationID,
		ParentOrchestrationID: "", // Clear this for response

		// Response message identity
		MessageID:    uuid.New().String(),
		RequestID:    ec.RequestID, // Keep original request ID
		MessageType:  "response",
		InResponseTo: ec.RequestID,

		// Reverse routing
		FromAgentID:   ec.ToAgentID,
		FromAgentType: ec.ToAgentType,
		ToAgentID:     ec.FromAgentID,
		ToAgentType:   ec.FromAgentType,
		ReplyToTopic:  ec.ReplyToTopic,

		// Remaining fuel
		FuelBudget: ec.FuelBudget,
		Timestamp:  time.Now().UTC(),
		Version:    "2.0",
	}
}

// ToHeaders converts context to Kafka headers
func (ec *ExecutionContext) ToHeaders() map[string]string {
	headers := map[string]string{
		"correlation_id":   ec.CorrelationID,
		"orchestration_id": ec.OrchestrationID,
		"client_id":        ec.ClientID,
		"message_id":       ec.MessageID,
		"request_id":       ec.RequestID,
		"message_type":     ec.MessageType,
		"owner_agent_id":   ec.OwnerAgentID,
		"from_agent_id":    ec.FromAgentID,
		"from_agent_type":  ec.FromAgentType,
		"to_agent_id":      ec.ToAgentID,
		"to_agent_type":    ec.ToAgentType,
		"reply_to_topic":   ec.ReplyToTopic,
		"fuel_budget":      fmt.Sprintf("%d", ec.FuelBudget),
		"timestamp":        ec.Timestamp.Format(time.RFC3339),
		"version":          ec.Version,
	}

	// Add optional fields only if present
	if ec.ParentOrchestrationID != "" {
		headers["parent_orchestration_id"] = ec.ParentOrchestrationID
	}
	if ec.InResponseTo != "" {
		headers["in_response_to"] = ec.InResponseTo
	}
	if ec.GroupID != "" {
		headers["group_id"] = ec.GroupID
	}
	if ec.FunctionalRole != "" {
		headers["functional_role"] = ec.FunctionalRole
	}

	// For backward compatibility
	headers["agent_instance_id"] = ec.OwnerAgentID
	headers["agent_type"] = ec.FromAgentType
	headers["causation_id"] = ec.InResponseTo // Legacy support

	return headers
}

// FromHeaders creates an ExecutionContext from Kafka headers
func FromHeaders(headers map[string]string) (*ExecutionContext, error) {
	// Required fields validation
	if headers["correlation_id"] == "" {
		return nil, fmt.Errorf("missing required header: correlation_id")
	}
	if headers["orchestration_id"] == "" {
		return nil, fmt.Errorf("missing required header: orchestration_id")
	}

	ec := &ExecutionContext{
		CorrelationID:         headers["correlation_id"],
		OrchestrationID:       headers["orchestration_id"],
		ParentOrchestrationID: headers["parent_orchestration_id"],
		ClientID:              headers["client_id"],
		GroupID:               headers["group_id"],
		FunctionalRole:        headers["functional_role"],
		MessageID:             headers["message_id"],
		RequestID:             headers["request_id"],
		MessageType:           headers["message_type"],
		InResponseTo:          headers["in_response_to"],
		OwnerAgentID:          headers["owner_agent_id"],
		FromAgentID:           headers["from_agent_id"],
		FromAgentType:         headers["from_agent_type"],
		ToAgentID:             headers["to_agent_id"],
		ToAgentType:           headers["to_agent_type"],
		ReplyToTopic:          headers["reply_to_topic"],
		Version:               headers["version"],
	}

	// Parse fuel budget
	if fuel := headers["fuel_budget"]; fuel != "" {
		fmt.Sscanf(fuel, "%d", &ec.FuelBudget)
	} else {
		ec.FuelBudget = 1000 // Default
	}

	// Parse timestamp
	if ts := headers["timestamp"]; ts != "" {
		ec.Timestamp, _ = time.Parse(time.RFC3339, ts)
	} else {
		ec.Timestamp = time.Now().UTC()
	}

	// Set defaults for missing fields
	if ec.MessageID == "" {
		ec.MessageID = uuid.New().String()
	}
	if ec.RequestID == "" {
		ec.RequestID = uuid.New().String()
	}
	if ec.MessageType == "" {
		ec.MessageType = "request"
	}
	if ec.Version == "" {
		ec.Version = "2.0"
	}

	// Handle legacy headers
	if ec.InResponseTo == "" && headers["causation_id"] != "" {
		ec.InResponseTo = headers["causation_id"]
	}

	return ec, nil
}

// IsChildOrchestration returns true if this is a child orchestration
func (ec *ExecutionContext) IsChildOrchestration() bool {
	return ec.ParentOrchestrationID != ""
}

// IsResponse returns true if this is a response message
func (ec *ExecutionContext) IsResponse() bool {
	return ec.MessageType == "response"
}

// Validate ensures all required fields are present
func (ec *ExecutionContext) Validate() error {
	if ec.CorrelationID == "" {
		return fmt.Errorf("correlation_id is required")
	}
	if ec.OrchestrationID == "" {
		return fmt.Errorf("orchestration_id is required")
	}
	if ec.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if ec.MessageType == "" {
		return fmt.Errorf("message_type is required")
	}
	if ec.MessageType == "response" && ec.InResponseTo == "" {
		return fmt.Errorf("in_response_to is required for response messages")
	}
	return nil
}

// ToJSON serializes the context to JSON
func (ec *ExecutionContext) ToJSON() ([]byte, error) {
	return json.Marshal(ec)
}

// FromJSON deserializes the context from JSON
func FromJSON(data []byte) (*ExecutionContext, error) {
	var ec ExecutionContext
	if err := json.Unmarshal(data, &ec); err != nil {
		return nil, err
	}
	return &ec, nil
}
