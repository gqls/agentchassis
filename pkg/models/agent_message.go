// pkg/models/agent_message.go
package models

import (
	"encoding/json"
	"strings"
	"time"
)

// AgentMessage is the new unified message format for agent-to-agent communication
type AgentMessage struct {
	// Core routing - this is the KEY CHANGE
	MessageID         string `json:"message_id"`
	RequestID         string `json:"request_id"` // replicates MessageID, get rid whenever we can
	CorrelationID     string `json:"correlation_id"`
	OrchestrationID   string `json:"orchestration_id"`
	OrchestrationName string `json:"orchestration_name"`
	FromAgentID       string `json:"from_agent_id"`  // WHO sent this
	ToAgentID         string `json:"to_agent_id"`    // WHO should process this
	ReplyToTopic      string `json:"reply_to_topic"` // WHERE to send response

	AgentInstanceID string `json:"agent_instance_id"`

	// Backward compatibility with existing TaskRequest/TaskResponse
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`

	// Tree tracking (new capability, optional initially)
	TreePath  []string `json:"tree_path,omitempty"`
	SubtreeID string   `json:"subtree_id,omitempty"`
	TreeDepth int      `json:"tree_depth,omitempty"`

	// Metadata
	Timestamp   time.Time `json:"timestamp"`
	MessageType string    `json:"message_type"` // "request" or "response"
	Version     string    `json:"version"`      // "2.0" for new format
}

// Helper functions for conversion
func (am *AgentMessage) ToTaskRequest() TaskRequest {
	return TaskRequest{
		Action: am.Action,
		Data:   am.Data,
	}
}

func (am *AgentMessage) ToHeaders() map[string]string {
	headers := map[string]string{
		"message_id":         am.MessageID,
		"request_id":         am.MessageID,
		"correlation_id":     am.CorrelationID,
		"orchestration_id":   am.OrchestrationID,
		"orchestration_name": am.OrchestrationName,
		"from_agent_id":      am.FromAgentID,
		"to_agent_id":        am.ToAgentID,
		"reply_to_topic":     am.ReplyToTopic,
		"message_type":       am.MessageType,
		"message_version":    am.Version,
		"timestamp":          am.Timestamp.Format(time.RFC3339),
	}

	// Add tree path if present
	if len(am.TreePath) > 0 {
		headers["tree_path"] = strings.Join(am.TreePath, ",")
	}

	if am.SubtreeID != "" {
		headers["subtree_id"] = am.SubtreeID
	}

	return headers
}

func AgentMessageFromHeaders(headers map[string]string, body []byte) (*AgentMessage, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &AgentMessage{
		MessageID:         headers["message_id"],
		CorrelationID:     headers["correlation_id"],
		OrchestrationID:   headers["orchestration_id"],
		OrchestrationName: headers["orchestration_name"],
		FromAgentID:       headers["from_agent_id"],
		ToAgentID:         headers["to_agent_id"],
		ReplyToTopic:      headers["reply_to_topic"],
		MessageType:       headers["message_type"],
		Version:           headers["message_version"],
		Data:              data,
		Timestamp:         time.Now(),
	}, nil
}
