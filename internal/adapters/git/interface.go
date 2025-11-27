package git

import (
	"encoding/json"
	"fmt"
	"time"
)

// ============================================================================
// REQUEST STRUCTURES
// ============================================================================

// AdapterRequest matches the message sent by the agent
type AdapterRequest struct {
	Headers AdapterHeaders `json:"headers"`
	Body    AdapterBody    `json:"body"`
}

// AdapterHeaders matches the agent's header structure
type AdapterHeaders struct {
	// Core identifiers
	CorrelationID   string `json:"correlation_id"`
	OrchestrationID string `json:"orchestration_id"`
	RequestID       string `json:"request_id"`
	ClientID        string `json:"client_id"`

	// Parent context (critical for orchestration)
	ParentOrchestrationID string `json:"parent_orchestration_id,omitempty"`
	ParentRequestID       string `json:"parent_request_id,omitempty"`

	// Step context
	StepID   string `json:"step_id,omitempty"`
	StepName string `json:"step_name,omitempty"`

	// Response routing
	ResponsesTopic string `json:"responses_topic"`
}

// AdapterBody is the agent's body structure
type AdapterBody struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"` // The specific payload for the action
}

// ============================================================================
// RESPONSE STRUCTURES (aligned with platform/models)
// ============================================================================

// ResponseMessage is the standard response format
type ResponseMessage struct {
	Headers ResponseHeaders `json:"headers"`
	Body    ResponseBody    `json:"body"`
}

// ResponseHeaders matches the platform's ResponseHeaders struct
type ResponseHeaders struct {
	// Response Tracking
	InResponseToRequestID      string `json:"in_response_to_request_id"`
	ReplyToRequestID           string `json:"reply_to_request_id,omitempty"`
	InResponseToStepID         string `json:"in_response_to_step_id,omitempty"`
	InResponseToStepName       string `json:"in_response_to_step_name,omitempty"`
	InResponseToParentOrchID   string `json:"in_response_to_parent_orchestration_id,omitempty"`
	InResponseToParentOrchName string `json:"in_response_to_parent_orchestration_name,omitempty"`
	InResponseToMessageID      string `json:"in_response_to_message_id,omitempty"`
	InResponseToAction         string `json:"in_response_to_action,omitempty"`
	InResponseTo               string `json:"in_response_to,omitempty"` // step_id shorthand
	RetryCount                 int    `json:"retry_count,omitempty"`

	// The orchestration that should process this response
	OrchestrationID   string `json:"orchestration_id"`
	OrchestrationName string `json:"orchestration_name,omitempty"`

	// My Context - The sender's orchestration (adapters typically don't have these)
	MyOrchestrationID   string `json:"my_orchestration_id,omitempty"`
	MyOrchestrationName string `json:"my_orchestration_name,omitempty"`
	MyRequestsTopic     string `json:"my_requests_topic,omitempty"`
	MyResponsesTopic    string `json:"my_responses_topic,omitempty"`

	// Identity
	CorrelationID   string `json:"correlation_id"`
	CorrelationName string `json:"correlation_name,omitempty"`
	ClientID        string `json:"client_id"`
	MessageType     string `json:"message_type"` // "response"
	MessageID       string `json:"message_id"`
	RequestID       string `json:"request_id"`

	// Parent context
	ParentOrchestrationID string `json:"parent_orchestration_id,omitempty"`
	ParentRequestID       string `json:"parent_request_id,omitempty"`

	// Step context
	StepID   string `json:"step_id,omitempty"`
	StepName string `json:"step_name,omitempty"`

	// Status
	Status  string `json:"status"` // "success" or "error"
	IsError bool   `json:"is_error,omitempty"`

	// Sender identification
	Sender          AgentIdentity `json:"sender,omitempty"`
	SenderAgentID   string        `json:"sender_agent_id"`
	SenderAgentType string        `json:"sender_agent_type"`
	SenderPodName   string        `json:"sender_pod_name,omitempty"`

	// Timing & Resources
	Timestamp           time.Time `json:"timestamp"`
	FuelUsed            int       `json:"fuel_used"`
	RemainingFuelBudget int       `json:"remaining_fuel_budget,omitempty"`
}

// AgentIdentity identifies an agent
type AgentIdentity struct {
	AgentID      string `json:"agent_id"`
	AgentType    string `json:"agent_type"`
	PodName      string `json:"pod_name,omitempty"`
	AgentVersion string `json:"agent_version,omitempty"`
	Role         string `json:"role,omitempty"`
}

// ResponseBody is the standard response body format
type ResponseBody struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code        string `json:"code,omitempty"`
	Type        string `json:"type,omitempty"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// ============================================================================
// GIT-SPECIFIC STRUCTURES
// ============================================================================

// GitCommitData is the expected structure of the 'data' field for a 'commit' action
type GitCommitData struct {
	RepoName      string            `json:"repo_name"`
	Domain        string            `json:"domain"`
	Files         map[string]string `json:"files"`
	CommitMessage string            `json:"commit_message"`
}

// GitHubRepo is a partial struct for GitHub's API response
type GitHubRepo struct {
	Name          string `json:"name"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Owner         struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// TreeEntry is for building the Git tree
type TreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// ToKafkaHeaders converts ResponseHeaders to map[string]string for Kafka
// Note: Only includes fields that should be in Kafka headers (strings only)
func (h *ResponseHeaders) ToKafkaHeaders() map[string]string {
	headers := map[string]string{
		"correlation_id":            h.CorrelationID,
		"orchestration_id":          h.OrchestrationID,
		"client_id":                 h.ClientID,
		"request_id":                h.RequestID,
		"message_type":              h.MessageType,
		"message_id":                h.MessageID,
		"status":                    h.Status,
		"sender_agent_id":           h.SenderAgentID,
		"sender_agent_type":         h.SenderAgentType,
		"in_response_to_request_id": h.InResponseToRequestID,
		"timestamp":                 h.Timestamp.Format(time.RFC3339),
	}

	// Add optional fields if present
	if h.ParentOrchestrationID != "" {
		headers["parent_orchestration_id"] = h.ParentOrchestrationID
	}
	if h.ParentRequestID != "" {
		headers["parent_request_id"] = h.ParentRequestID
	}
	if h.StepID != "" {
		headers["step_id"] = h.StepID
		headers["in_response_to"] = h.StepID
	}
	if h.StepName != "" {
		headers["step_name"] = h.StepName
		headers["in_response_to_step_name"] = h.StepName
	}
	if h.SenderPodName != "" {
		headers["sender_pod_name"] = h.SenderPodName
	}
	// Note: fuel_used is NOT included in Kafka headers (it's in the JSON body)
	// But some receivers expect it, so include as string
	headers["fuel_used"] = fmt.Sprintf("%d", h.FuelUsed)

	return headers
}
