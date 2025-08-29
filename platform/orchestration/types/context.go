// platform/orchestration/types/context.go - NEW PACKAGE

package types

import (
	"fmt"

	"github.com/google/uuid"
)

type ExecutionContext struct {
	// Business Transaction
	CorrelationID string
	ClientID      string
	HierarchyLevel	int

	// Execution Instance - UNIQUE per agent
	OrchestrationID       string
	ParentOrchestrationID string

	// Group Management
	GroupID        string
	FunctionalRole string // "content-creator", "domain-analyst", etc.

	// Message Identity
	MessageID    string
	RequestID    string
	MessageType  string // "request" | "response" | "error"
	InResponseTo string

	// Routing
	OwnerAgentID string
	FromAgentID  string
	FromAgentType	string
	ToAgentID    string
	ToAgentType  string
	ReplyTo      string
}

// NewOrchestrationForAgent creates a new orchestration context for a called agent
func (c *ExecutionContext) NewOrchestrationForAgent(toAgentID string) *ExecutionContext {
	return &ExecutionContext{
		// Keep business context
		CorrelationID: c.CorrelationID,
		ClientID:      c.ClientID,
		GroupID:       c.GroupID,

		// NEW orchestration for the called agent
		OrchestrationID:       uuid.New().String(),
		ParentOrchestrationID: c.OrchestrationID,
		OwnerAgentID:          toAgentID,

		// Will be set when sending message
		FromAgentID: c.OwnerAgentID,
		ToAgentID:   toAgentID,
	}
}

func (c *ExecutionContext) ToHeaders() map[string]string {
	return map[string]string{
		"correlation_id":          c.CorrelationID,
		"orchestration_id":        c.OrchestrationID,
		"parent_orchestration_id": c.ParentOrchestrationID,
		"client_id":               c.ClientID,
		"group_id":                c.GroupID,
		"functional_role":         c.FunctionalRole,
		"message_id":              c.MessageID,
		"request_id":              c.RequestID,
		"message_type":            c.MessageType,
		"in_response_to":          c.InResponseTo,
		"owner_agent_id":          c.OwnerAgentID,
		"from_agent_id":           c.FromAgentID,
		"to_agent_id":             c.ToAgentID,
		"reply_to":                c.ReplyTo,
	}
}

func ExecutionContextFromHeaders(headers map[string]string) (*ExecutionContext, error) {
	// Required fields
	if headers["orchestration_id"] == "" {
		return nil, fmt.Errorf("missing orchestration_id")
	}
	if headers["correlation_id"] == "" {
		return nil, fmt.Errorf("missing correlation_id")
	}

	return &ExecutionContext{
		CorrelationID:         headers["correlation_id"],
		HierarchyLevel:		   headers["hierarchy_level"],
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
		ToAgentID:             headers["to_agent_id"],
		FromAgentType:		   headers["from_agent_type"],
		ToAgentType			   headers["to_agent_type"],
		ReplyTo:               headers["reply_to"],
	}, nil
}
