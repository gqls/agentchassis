// pkg/models/agent_message.go
package models

import (
	"encoding/json"
	"time"
)

// AgentMessage is the new unified message format for agent-to-agent communication
type AgentMessage struct {
	// Core routing - this is the KEY CHANGE
	MessageID     string `json:"message_id"`
	CorrelationID string `json:"correlation_id"`
	FromAgentID   string `json:"from_agent_id"`  // WHO sent this
	ToAgentID     string `json:"to_agent_id"`    // WHO should process this
	ReplyToTopic  string `json:"reply_to_topic"` // WHERE to send response

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
	headers := make(map[string]string)
	headers["message_id"] = am.MessageID
	headers["correlation_id"] = am.CorrelationID
	headers["from_agent_id"] = am.FromAgentID
	headers["to_agent_id"] = am.ToAgentID
	headers["reply_to_topic"] = am.ReplyToTopic
	headers["message_version"] = am.Version
	headers["message_type"] = am.MessageType
	return headers
}

func AgentMessageFromHeaders(headers map[string]string, body []byte) (*AgentMessage, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &AgentMessage{
		MessageID:     headers["message_id"],
		CorrelationID: headers["correlation_id"],
		FromAgentID:   headers["from_agent_id"],
		ToAgentID:     headers["to_agent_id"],
		ReplyToTopic:  headers["reply_to_topic"],
		MessageType:   headers["message_type"],
		Version:       headers["message_version"],
		Data:          data,
		Timestamp:     time.Now(),
	}, nil
}
