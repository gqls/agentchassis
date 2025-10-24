package types

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// ExecutionContext represents the complete context for message execution
type ExecutionContext struct {
	// Core identity
	CorrelationID     string `json:"correlation_id"`
	OrchestrationID   string `json:"orchestration_id"`
	ClientID          string `json:"client_id"`
	CorrelationName   string `json:"correlation_name,omitempty"`
	OrchestrationName string `json:"orchestration_name,omitempty"`

	// hierarchy
	ParentOrchestrationID   string `json:"parent_orchestration_id,omitempty"`
	ParentOrchestrationName string `json:"parent_orchestration_name,omitempty"`
	ReplyToRequestID        string `json:"reply_to_request_id,omitempty"`

	// Group Management
	GroupID        string `json:"group_id,omitempty"`
	FunctionalRole string `json:"functional_role,omitempty"`

	// Step Context (for requests)
	StepID   string `json:"step_id,omitempty"`
	StepName string `json:"step_name,omitempty"`
	Action   string `json:"action,omitempty"`

	// Message Identity
	MessageID    string `json:"message_id"`
	RequestID    string `json:"request_id"`
	MessageType  string `json:"message_type"` // "request" | "response" | "error"
	RetryVersion int    `json:"retry_version"`

	// Response context
	InResponseTo        *ResponseContext `json:"in_response_to,omitempty"`
	Status              string           `json:"status,omitempty"` // awaiting|processing|complete|error
	IsComplete          bool             `json:"is_complete,omitempty"`
	IsError             bool             `json:"is_error,omitempty"`
	IsMultipartResponse bool             `json:"is_multipart_response,omitempty"`
	PartCount           int              `json:"part_count,omitempty"`

	// Routing
	Sender         AgentIdentity `json:"sender"` // Who sent this message
	ToAgentType    string        `json:"to_agent_type"`
	RequestsTopic  string        `json:"requests_topic,omitempty"`
	ResponsesTopic string        `json:"responses_topic,omitempty"`
	ReplyToTopic   string        `json:"reply_to_topic,omitempty"`

	// Processing info
	ProcessingNode string `json:"processing_node,omitempty"`
	ToAgentID      string `json:"to_agent_id,omitempty"`
	FromAgentID    string `json:"from_agent_id,omitempty"`
	FromAgentType  string `json:"from_agent_type,omitempty"`

	// Resource Management
	FuelBudget     int           `json:"fuel_budget"`
	FuelUsed       int           `json:"fuel_used,omitempty"`
	TimeoutSeconds int           `json:"timeout_seconds,omitempty"`
	TimeSpent      time.Duration `json:"time_spent,omitempty"`

	// Metadata
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// AgentIdentity now represents the sender, focusing on its type and pod name.
type AgentIdentity struct {
	AgentID      string `json:"agent_id"`
	AgentType    string `json:"agent_type"`
	PodName      string `json:"pod_name"` // The new "ID" for stateless agents
	AgentVersion string `json:"agent_version"`
	Role         string `json:"role,omitempty"`
}

// Request Message Structure
type RequestMessage struct {
	Headers RequestHeaders `json:"headers"`
	Body    interface{}    `json:"body"`
}

type RequestHeaders struct {
	// Identity
	CorrelationID   string `json:"correlation_id"`
	ClientID        string `json:"client_id"`
	CorrelationName string `json:"correlation_name"`

	Sender         AgentIdentity `json:"sender"`
	FunctionalRole string        `json:"functional_role"`

	// Orchestration Context
	OrchestrationID         string `json:"orchestration_id"` // Sender's orchestration
	OrchestrationName       string `json:"orchestration_name"`
	StepID                  string `json:"step_id"` // Sender's step
	StepName                string `json:"step_name"`
	RequestID               string `json:"request_id"`              // Unique per step task
	RetryVersion            int    `json:"retry_version"`           // 0, 1, 2 for retries
	ParentOrchestrationID   string `json:"parent_orchestration_id"` // If spawned
	ParentOrchestrationName string `json:"parent_orchestration_name"`
	ReplyToRequestID        string `json:"reply_to_request_id"` // If spawned
	ParentResponsesTopic    string `json:"parent_responses_topic"`

	// Message Metadata
	MessageID   string    `json:"message_id"`   // Unique per message
	MessageType string    `json:"message_type"` // "request"
	FromAgent   string    `json:"from_agent"`
	ToAgent     string    `json:"to_agent"`
	ToAgentType string    `json:"to_agent_type"`
	Action      string    `json:"action"`
	Timestamp   time.Time `json:"timestamp"`

	// Resource Management
	FuelBudget     int `json:"fuel_budget"`
	TimeoutSeconds int `json:"timeout_seconds"`

	// Routing
	RequestsTopic  string `json:"requests_topic"`
	ResponsesTopic string `json:"responses_topic"`
}

// Response Message Structure - Highly Transparent Format
type ResponseMessage struct {
	Headers ResponseHeaders `json:"headers"`
	Body    ResponseBody    `json:"body"`
}

type ResponseHeaders struct {
	// Response Tracking
	InResponseToRequestID      string `json:"in_response_to_request_id"`
	ReplyToRequestID           string `json:"reply_to_request_id"`
	InResponseToStepID         string `json:"in_response_to_step_id"`
	InResponseToStepName       string `json:"in_response_to_step_name"`
	InResponseToParentOrchID   string `json:"in_response_to_parent_orchestration_id"`
	InResponseToParentOrchName string `json:"in_response_to_parent_orchestration_name"`
	InResponseToMessageID      string `json:"in_response_to_message_id"`
	InResponseToAction         string `json:"in_response_to_action"`
	RetryCount                 int    `json:"retry_count"`

	// The orchestration that should process this response
	OrchestrationID   string `json:"orchestration_id"`
	OrchestrationName string `json:"orchestration_name"`

	// My Context - The sender's orchestration
	MyOrchestrationID   string `json:"my_orchestration_id"`
	MyOrchestrationName string `json:"my_orchestration_name"`
	MyRequestsTopic     string `json:"my_requests_topic"`
	MyResponsesTopic    string `json:"my_responses_topic"`

	// Identity
	CorrelationID   string `json:"correlation_id"`
	CorrelationName string `json:"correlation_name"`
	ClientID        string `json:"client_id"`
	MessageType     string `json:"message_type"`
	FromAgent       string `json:"from_agent"`
	ToAgent         string `json:"to_agent"`
	ToAgentType     string `json:"to_agent_type"`

	// Status Flags
	IsComplete          bool   `json:"is_complete"`
	IsError             bool   `json:"is_error"`
	IsMultipartResponse bool   `json:"is_multipart_response"`
	PartCount           int    `json:"part_count"`
	Status              string `json:"status"` // awaiting|processing|complete|error_recoverable|error_unrecoverable

	Sender AgentIdentity `json:"sender"`

	// Timing & Resources
	TimeSent                   time.Time     `json:"time_sent"`
	TimeSpent                  time.Duration `json:"time_spent"`
	OverallTimeBudgetRemaining int           `json:"overall_time_budget_remaining"`
	TopicSentTo                string        `json:"topic_sent_to"`
	FuelUsed                   int           `json:"fuel_used"`
	RemainingFuelBudget        int           `json:"remaining_fuel_budget"`
}

// In types package
type ResponseBody struct {
	Success bool                   `json:"success"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Body    interface{}            `json:"body"`
	Error   *ErrorInfo             `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Recoverable bool                   `json:"recoverable"`
	Timestamp   time.Time              `json:"timestamp"`
	RetryAfter  int                    `json:"retry_after_seconds,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ResponseContext captures what we're responding to
type ResponseContext struct {
	RequestID               string `json:"request_id"`
	StepID                  string `json:"step_id"`
	StepName                string `json:"step_name"`
	MessageID               string `json:"message_id"`
	Action                  string `json:"action"`
	ParentOrchestrationID   string `json:"parent_orchestration_id"`
	ParentOrchestrationName string `json:"parent_orchestration_name"`
}

// ToRequestHeaders converts ExecutionContext to RequestHeaders for sending requests
func (ec *ExecutionContext) ToRequestHeaders() RequestHeaders {
	return RequestHeaders{
		// Who Am I
		Sender: ec.Sender,

		// Identity
		CorrelationID:   ec.CorrelationID,
		CorrelationName: ec.CorrelationName,
		ClientID:        ec.ClientID,
		FunctionalRole:  ec.FunctionalRole,

		// Orchestration Context
		OrchestrationID:         ec.OrchestrationID,
		OrchestrationName:       ec.OrchestrationName,
		StepID:                  ec.StepID,
		StepName:                ec.StepName,
		RequestID:               ec.RequestID,
		RetryVersion:            ec.RetryVersion,
		ParentOrchestrationID:   ec.ParentOrchestrationID,
		ParentOrchestrationName: ec.ParentOrchestrationName,
		ReplyToRequestID:        ec.ReplyToRequestID,

		// Message Metadata
		MessageID:   ec.MessageID,
		MessageType: "request",
		FromAgent:   ec.Sender.AgentID, // Legacy field
		ToAgentType: ec.ToAgentType,
		Action:      ec.Action,
		Timestamp:   ec.Timestamp,

		// Resource Management
		FuelBudget:     ec.FuelBudget,
		TimeoutSeconds: ec.TimeoutSeconds,

		// Routing
		RequestsTopic:        ec.RequestsTopic,
		ResponsesTopic:       ec.ResponsesTopic,
		ParentResponsesTopic: ec.ReplyToTopic,
	}
}

// ToResponseHeaders converts ExecutionContext to ResponseHeaders for sending responses
func (ec *ExecutionContext) ToResponseHeaders() ResponseHeaders {

	// Explicitly determine where this response should be routed
	targetOrchID, targetOrchName := ec.DetermineResponseOrchestrationTarget()

	headers := ResponseHeaders{
		// Who Am I
		Sender: ec.Sender,

		// Where should this response be routed?
		OrchestrationID:   targetOrchID,
		OrchestrationName: targetOrchName,

		// Response tracking - what we're responding to
		InResponseToRequestID:      ec.RequestID,
		InResponseToStepID:         ec.StepID,
		InResponseToStepName:       ec.StepName,
		InResponseToParentOrchID:   ec.ParentOrchestrationID,
		InResponseToParentOrchName: ec.ParentOrchestrationName,
		InResponseToMessageID:      ec.MessageID,
		InResponseToAction:         ec.Action,
		RetryCount:                 ec.RetryVersion,

		// My Context
		MyOrchestrationID:   ec.OrchestrationID,
		MyOrchestrationName: ec.OrchestrationName,
		MyRequestsTopic:     ec.RequestsTopic,
		MyResponsesTopic:    ec.ResponsesTopic,

		// Identity
		CorrelationID:   ec.CorrelationID,
		CorrelationName: ec.CorrelationName,
		ClientID:        ec.ClientID,
		MessageType:     "response",
		FromAgent:       ec.Sender.AgentID,
		ToAgent:         ec.FromAgentID, // Responding back to sender
		ToAgentType:     ec.FromAgentType,

		// Status Flags
		IsComplete:          ec.IsComplete,
		IsError:             ec.IsError,
		IsMultipartResponse: ec.IsMultipartResponse,
		PartCount:           ec.PartCount,
		Status:              ec.Status,

		// Timing & Resources
		TimeSent:                   ec.Timestamp,
		TimeSpent:                  ec.TimeSpent,
		OverallTimeBudgetRemaining: ec.TimeoutSeconds - int(ec.TimeSpent.Seconds()),
		TopicSentTo:                ec.ResponsesTopic,
		FuelUsed:                   ec.FuelUsed,
		RemainingFuelBudget:        ec.FuelBudget - ec.FuelUsed,
	}

	// If this is a response to something, populate those fields
	if ec.InResponseTo != nil {
		headers.InResponseToRequestID = ec.InResponseTo.RequestID
		headers.InResponseToStepID = ec.InResponseTo.StepID
		headers.InResponseToStepName = ec.InResponseTo.StepName
		headers.InResponseToParentOrchID = ec.InResponseTo.ParentOrchestrationID
		headers.InResponseToParentOrchName = ec.InResponseTo.ParentOrchestrationName
		headers.InResponseToMessageID = ec.InResponseTo.MessageID
		headers.InResponseToAction = ec.InResponseTo.Action
		headers.RetryCount = ec.RetryVersion
	}

	return headers
}

// FromRequestHeaders creates an ExecutionContext from incoming RequestHeaders
func FromRequestHeaders(headers RequestHeaders) *ExecutionContext {
	return &ExecutionContext{
		// Core identity
		CorrelationID:     headers.CorrelationID,
		CorrelationName:   headers.CorrelationName,
		OrchestrationID:   headers.OrchestrationID,
		OrchestrationName: headers.OrchestrationName,
		ClientID:          headers.ClientID,

		// Hierarchy
		ParentOrchestrationID:   headers.ParentOrchestrationID,
		ParentOrchestrationName: headers.ParentOrchestrationName,
		ReplyToRequestID:        headers.ReplyToRequestID,

		// Step Context
		StepID:   headers.StepID,
		StepName: headers.StepName,
		Action:   headers.Action,

		// Message Identity
		MessageID:    headers.MessageID,
		RequestID:    headers.RequestID,
		MessageType:  headers.MessageType,
		RetryVersion: headers.RetryVersion,

		// Routing
		Sender:         headers.Sender,
		ToAgentType:    headers.ToAgentType,
		RequestsTopic:  headers.RequestsTopic,
		ResponsesTopic: headers.ResponsesTopic,

		// Resource Management
		FuelBudget:     headers.FuelBudget,
		TimeoutSeconds: headers.TimeoutSeconds,

		// Metadata
		Timestamp: headers.Timestamp,
	}
}

// FromResponseHeaders creates an ExecutionContext from incoming ResponseHeaders
func FromResponseHeaders(headers ResponseHeaders) *ExecutionContext {
	return &ExecutionContext{
		// Core identity
		CorrelationID:     headers.CorrelationID,
		CorrelationName:   headers.CorrelationName,
		OrchestrationID:   headers.MyOrchestrationID,
		OrchestrationName: headers.MyOrchestrationName,
		ClientID:          headers.ClientID,

		// Response Context
		InResponseTo: &ResponseContext{
			RequestID:               headers.InResponseToRequestID,
			StepID:                  headers.InResponseToStepID,
			StepName:                headers.InResponseToStepName,
			MessageID:               headers.InResponseToMessageID,
			Action:                  headers.InResponseToAction,
			ParentOrchestrationID:   headers.InResponseToParentOrchID,
			ParentOrchestrationName: headers.InResponseToParentOrchName,
		},
		Status:              headers.Status,
		IsComplete:          headers.IsComplete,
		IsError:             headers.IsError,
		IsMultipartResponse: headers.IsMultipartResponse,
		PartCount:           headers.PartCount,
		RetryVersion:        headers.RetryCount,

		// Routing
		Sender:         headers.Sender,
		RequestsTopic:  headers.MyRequestsTopic,
		ResponsesTopic: headers.MyResponsesTopic,

		// Resource Management
		FuelUsed:   headers.FuelUsed,
		FuelBudget: headers.RemainingFuelBudget + headers.FuelUsed,
		TimeSpent:  headers.TimeSpent,

		// Metadata
		MessageType: headers.MessageType,
		Timestamp:   headers.TimeSent,
	}
}

// CreateChildContext creates a new ExecutionContext for a spawned agent
func (ec *ExecutionContext) CreateChildContext(childAgentType string) *ExecutionContext {
	child := &ExecutionContext{
		// Inherit correlation
		CorrelationID:   ec.CorrelationID,
		CorrelationName: ec.CorrelationName,
		ClientID:        ec.ClientID,

		// New orchestration for child
		OrchestrationID:   uuid.New().String(),
		OrchestrationName: GenerateReadableName(childAgentType, "orchestration"),

		// Parent references
		ParentOrchestrationID:   ec.OrchestrationID,
		ParentOrchestrationName: ec.OrchestrationName,
		ReplyToRequestID:        ec.RequestID,

		// Inherit group if applicable
		GroupID:        ec.GroupID,
		FunctionalRole: "", // Child determines its own role

		// New message identity
		MessageID:    uuid.New().String(),
		RequestID:    uuid.New().String(),
		MessageType:  "request",
		RetryVersion: 0,

		// Resource inheritance (deduct some)
		FuelBudget:     ec.FuelBudget - 100,
		TimeoutSeconds: ec.TimeoutSeconds - 5,

		// Metadata
		Timestamp: time.Now(),
		Version:   ec.Version,
	}

	return child
}

// CreateResponseContext prepares context for sending a response
func (ec *ExecutionContext) CreateResponseContext(status string, fuelUsed int) *ExecutionContext {
	startTime := ec.Timestamp
	timeSpent := time.Since(startTime)

	responseCtx := &ExecutionContext{
		// Keep identity
		CorrelationID:     ec.CorrelationID,
		CorrelationName:   ec.CorrelationName,
		OrchestrationID:   ec.OrchestrationID,
		OrchestrationName: ec.OrchestrationName,
		ClientID:          ec.ClientID,

		// Response specifics
		MessageID:   uuid.New().String(),
		MessageType: "response",
		Status:      status,
		IsComplete:  status == "complete",
		IsError:     status == "error_unrecoverable" || status == "error_recoverable",

		// What we're responding to
		InResponseTo: &ResponseContext{
			RequestID:               ec.ReplyToRequestID, // changed from request id
			StepID:                  ec.StepID,
			StepName:                ec.StepName,
			MessageID:               ec.MessageID,
			Action:                  ec.Action,
			ParentOrchestrationID:   ec.ParentOrchestrationID,
			ParentOrchestrationName: ec.ParentOrchestrationName,
		},

		// Resource tracking
		FuelUsed:       fuelUsed,
		FuelBudget:     ec.FuelBudget,
		TimeSpent:      timeSpent,
		TimeoutSeconds: ec.TimeoutSeconds,

		// Metadata
		Timestamp:      time.Now(),
		RetryVersion:   ec.RetryVersion,
		Version:        ec.Version,
		ResponsesTopic: ec.ResponsesTopic,
	}

	return responseCtx
}

func GenerateReadableName(agentType, suffix string) string {
	timestamp := time.Now().Format("0102-1504")
	return fmt.Sprintf("%s-%s-%s", agentType, suffix, timestamp)
}

// NewExecutionContext creates a new execution context for an agent
func NewExecutionContext(correlationID, clientID, agentID, agentType string) *ExecutionContext {
	return &ExecutionContext{
		CorrelationID:     correlationID,
		ClientID:          clientID,
		OrchestrationID:   uuid.New().String(),
		OrchestrationName: GenerateReadableName(agentType, "orchestration"),
		MessageID:         uuid.New().String(),
		RequestID:         uuid.New().String(),
		MessageType:       "request",
		FuelBudget:        1000,
		Timestamp:         time.Now().UTC(),
		Version:           "2.0",
	}
}

// FromHeaders creates an ExecutionContext from Kafka headers
func FromHeaders(headers map[string]string) (*ExecutionContext, error) {

	ec := &ExecutionContext{
		// Core fields
		CorrelationID:         headers["correlation_id"],
		OrchestrationID:       headers["orchestration_id"],
		OrchestrationName:     headers["orchestration_name"],
		ParentOrchestrationID: headers["parent_orchestration_id"],
		ClientID:              headers["client_id"],

		// Message identity
		MessageID:   headers["message_id"],
		RequestID:   headers["request_id"],
		MessageType: headers["message_type"],

		// Processing info
		ProcessingNode: headers["processing_node"],
		ToAgentID:      headers["to_agent_id"],
		ToAgentType:    headers["to_agent_type"],
		FromAgentID:    headers["from_agent_id"],
		FromAgentType:  headers["from_agent_type"],

		// Step info
		StepID:   headers["step_id"],
		StepName: headers["step_name"],
		Action:   headers["action"],

		// Status (for responses)
		Status: headers["status"],

		// Routing
		RequestsTopic:    headers["requests_topic"],
		ResponsesTopic:   headers["responses_topic"], // whose responses topic is this?
		ReplyToRequestID: headers["reply_to_request_id"],

		// Version
		Version: headers["version"],
	}

	// Parse sender if present
	if senderType := headers["sender_agent_type"]; senderType != "" {
		ec.Sender = AgentIdentity{
			AgentType:    senderType,
			AgentID:      headers["sender_agent_id"],
			PodName:      headers["sender_pod_name"],
			AgentVersion: headers["sender_agent_version"],
			Role:         headers["sender_role"],
		}
	}

	if responsesTopic := headers["responses_topic"]; responsesTopic != "" {
		ec.ResponsesTopic = responsesTopic
	}

	if replyToTopic := headers["reply_to_topic"]; replyToTopic != "" {
		ec.ReplyToTopic = replyToTopic
	} else {
		if parentResponsesTopic := headers["parent_responses_topic"]; parentResponsesTopic != "" {
			ec.ReplyToTopic = parentResponsesTopic
		}
	}

	// Handle response context
	if ec.MessageType == "response" {
		ec.InResponseTo = &ResponseContext{
			RequestID:               headers["in_response_to_request_id"],
			StepID:                  headers["in_response_to_step_id"],
			StepName:                headers["in_response_to_step_name"],
			MessageID:               headers["in_response_to_message_id"],
			Action:                  headers["in_response_to_action"],
			ParentOrchestrationID:   headers["in_response_to_parent_orchestration_id"],
			ParentOrchestrationName: headers["in_response_to_parent_orchestration_name"],
		}

		// Parse status flags
		ec.IsComplete = headers["is_complete"] == "true"
		ec.IsError = headers["is_error"] == "true"
		ec.IsMultipartResponse = headers["is_multipart_response"] == "true"

		// Parse retry version
		if retryStr := headers["retry_version"]; retryStr != "" {
			fmt.Sscanf(retryStr, "%d", &ec.RetryVersion)
		}
	}

	ec.FunctionalRole = headers["functional_role"]

	// Parse fuel budget
	if fuel := headers["fuel_budget"]; fuel != "" {
		fmt.Sscanf(fuel, "%d", &ec.FuelBudget)
	} else {
		ec.FuelBudget = 1000 // Default
	}

	// Parse timeout
	if timeout := headers["timeout_seconds"]; timeout != "" {
		fmt.Sscanf(timeout, "%d", &ec.TimeoutSeconds)
	} else {
		ec.TimeoutSeconds = 30 // Default
	}

	// Parse timestamp
	if ts := headers["timestamp"]; ts != "" {
		ec.Timestamp, _ = time.Parse(time.RFC3339, ts)
	} else {
		ec.Timestamp = time.Now().UTC()
	}

	// Set defaults for missing required fields
	if ec.MessageID == "" {
		ec.MessageID = uuid.New().String()
	}
	if ec.MessageType == "" {
		ec.MessageType = "request"
	}
	if ec.Version == "" {
		ec.Version = "2.0"
	}

	return ec, nil
}

// ToHeaders converts ExecutionContext to map[string]string for Kafka headers
func (ec *ExecutionContext) ToHeaders() map[string]string {
	headers := map[string]string{
		// Core identity
		"correlation_id":   ec.CorrelationID,
		"orchestration_id": ec.OrchestrationID,
		"client_id":        ec.ClientID,

		// Message identity
		"message_id":   ec.MessageID,
		"request_id":   ec.RequestID,
		"message_type": ec.MessageType,

		// Processing info
		"processing_node":     ec.ProcessingNode,
		"to_agent_id":         ec.ToAgentID,
		"to_agent_type":       ec.ToAgentType,
		"from_agent_id":       ec.FromAgentID,
		"from_agent_type":     ec.FromAgentType,
		"responses_topic":     ec.ResponsesTopic,
		"requests_topic":      ec.RequestsTopic,
		"reply_to_request_id": ec.ReplyToRequestID,

		// Step info
		"step_id":   ec.StepID,
		"step_name": ec.StepName,
		"action":    ec.Action,

		// Resources
		"fuel_budget":     fmt.Sprintf("%d", ec.FuelBudget),
		"timeout_seconds": fmt.Sprintf("%d", ec.TimeoutSeconds),

		// Metadata
		"timestamp": ec.Timestamp.Format(time.RFC3339),
		"version":   ec.Version,
	}

	// Add parent info if present
	if ec.ParentOrchestrationID != "" {
		headers["parent_orchestration_id"] = ec.ParentOrchestrationID
		headers["parent_orchestration_name"] = ec.ParentOrchestrationName
		headers["reply_to_request_id"] = ec.ReplyToRequestID
	}

	// Add sender info
	headers["sender_agent_type"] = ec.Sender.AgentType
	headers["sender_agent_id"] = ec.Sender.AgentID
	headers["sender_pod_name"] = ec.Sender.PodName
	headers["sender_agent_version"] = ec.Sender.AgentVersion
	headers["sender_role"] = ec.Sender.Role

	// Add response context if this is a response
	if ec.InResponseTo != nil {
		headers["in_response_to_request_id"] = ec.InResponseTo.RequestID
		headers["in_response_to_step_id"] = ec.InResponseTo.StepID
		headers["in_response_to_step_name"] = ec.InResponseTo.StepName
		headers["in_response_to_message_id"] = ec.InResponseTo.MessageID
		headers["in_response_to_action"] = ec.InResponseTo.Action
		headers["in_response_to_parent_orchestration_id"] = ec.InResponseTo.ParentOrchestrationID
		headers["in_response_to_parent_orchestration_name"] = ec.InResponseTo.ParentOrchestrationName

		// Add response status flags
		headers["status"] = ec.Status
		headers["is_complete"] = fmt.Sprintf("%v", ec.IsComplete)
		headers["is_error"] = fmt.Sprintf("%v", ec.IsError)
		headers["retry_version"] = fmt.Sprintf("%d", ec.RetryVersion)
	}

	// Add functional role if present (for requests)
	if ec.FunctionalRole != "" {
		headers["functional_role"] = ec.FunctionalRole
	}

	return headers
}

// IsChildOrchestration returns true if this is a child orchestration
func (ec *ExecutionContext) IsChildOrchestration() bool {
	return ec.ParentOrchestrationID != ""
}

// IsResponse returns true if this is a response message
func (ec *ExecutionContext) IsResponse() bool {
	return ec.MessageType == "response" || ec.InResponseTo != nil
}

// Validate ensures all required fields are present
func (ec *ExecutionContext) Validate() error {
	// Basic validation for all messages
	if ec.CorrelationID == "" {
		return fmt.Errorf("correlation_id is required")
	}
	if ec.MessageType == "" {
		return fmt.Errorf("message_type is required")
	}

	// Different validation based on message type
	switch ec.MessageType {
	case "request":
		// Requests MUST have orchestration_id and client_id
		if ec.OrchestrationID == "" {
			return fmt.Errorf("orchestration_id is required for request messages")
		}
		if ec.ClientID == "" {
			return fmt.Errorf("client_id is required for request messages")
		}

	case "response":
		// Responses should have orchestration_id but client_id is optional
		if ec.OrchestrationID == "" {
			// For responses, this is less critical but still log it
			// Don't fail validation - responses can come from agents without orchestration
			// return fmt.Errorf("orchestration_id is required for response messages")
		}
		// Client ID is NOT required for responses
		// Responses must have InResponseTo context
		if ec.InResponseTo == nil {
			return fmt.Errorf("in_response_to is required for response messages")
		}

	case "error":
		// Error messages have relaxed validation
		// They might not have all context if something went wrong

	default:
		return fmt.Errorf("unknown message_type: %s", ec.MessageType)
	}

	return nil
}

// GetAgentIdentity creates an AgentIdentity for the current agent
func GetAgentIdentity(agentType string) *AgentIdentity {
	podName := os.Getenv("HOSTNAME") // In K8s, hostname = pod name
	if podName == "" {
		podName = fmt.Sprintf("%s-local-%d", agentType, os.Getpid())
	}

	return &AgentIdentity{
		AgentType:    agentType,
		AgentID:      podName, // AgentID = PodName for stateless design
		PodName:      podName,
		AgentVersion: os.Getenv("AGENT_VERSION"),
	}
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

// ToMap converts ResponseHeaders to a map for Kafka headers
func (rh *ResponseHeaders) ToMap() map[string]string {
	headers := make(map[string]string)

	// Response tracking
	headers["in_response_to_request_id"] = rh.InResponseToRequestID
	headers["in_response_to_step_id"] = rh.InResponseToStepID
	headers["in_response_to_step_name"] = rh.InResponseToStepName
	headers["in_response_to_parent_orchestration_id"] = rh.InResponseToParentOrchID
	headers["in_response_to_parent_orchestration_name"] = rh.InResponseToParentOrchName
	headers["in_response_to_message_id"] = rh.InResponseToMessageID
	headers["in_response_to_action"] = rh.InResponseToAction
	headers["retry_count"] = fmt.Sprintf("%d", rh.RetryCount)

	// Include the orchestration ID
	headers["orchestration_id"] = rh.OrchestrationID
	headers["orchestration_name"] = rh.OrchestrationName

	// My context
	headers["my_orchestration_id"] = rh.MyOrchestrationID
	headers["my_orchestration_name"] = rh.MyOrchestrationName
	headers["my_requests_topic"] = rh.MyRequestsTopic
	headers["my_responses_topic"] = rh.MyResponsesTopic

	// Identity
	headers["correlation_id"] = rh.CorrelationID
	headers["correlation_name"] = rh.CorrelationName
	headers["client_id"] = rh.ClientID
	headers["message_type"] = rh.MessageType
	headers["from_agent"] = rh.FromAgent
	headers["to_agent"] = rh.ToAgent
	headers["to_agent_type"] = rh.ToAgentType

	// Status flags
	headers["is_complete"] = fmt.Sprintf("%v", rh.IsComplete)
	headers["is_error"] = fmt.Sprintf("%v", rh.IsError)
	headers["is_multipart_response"] = fmt.Sprintf("%v", rh.IsMultipartResponse)
	headers["part_count"] = fmt.Sprintf("%d", rh.PartCount)
	headers["status"] = rh.Status

	// Sender info
	headers["sender_agent_type"] = rh.Sender.AgentType
	headers["sender_agent_id"] = rh.Sender.AgentID
	headers["sender_pod_name"] = rh.Sender.PodName
	headers["sender_agent_version"] = rh.Sender.AgentVersion
	headers["sender_role"] = rh.Sender.Role

	// Timing & Resources
	headers["time_sent"] = rh.TimeSent.Format(time.RFC3339)
	headers["time_spent"] = rh.TimeSpent.String()
	headers["overall_time_budget_remaining"] = fmt.Sprintf("%d", rh.OverallTimeBudgetRemaining)
	headers["topic_sent_to"] = rh.TopicSentTo
	headers["fuel_used"] = fmt.Sprintf("%d", rh.FuelUsed)
	headers["remaining_fuel_budget"] = fmt.Sprintf("%d", rh.RemainingFuelBudget)

	return headers
}

// ToMap converts RequestHeaders to a map for Kafka headers
func (rh *RequestHeaders) ToMap() map[string]string {
	headers := make(map[string]string)

	// Identity
	headers["correlation_id"] = rh.CorrelationID
	headers["client_id"] = rh.ClientID
	headers["correlation_name"] = rh.CorrelationName
	headers["functional_role"] = rh.FunctionalRole

	// Sender
	headers["sender_agent_type"] = rh.Sender.AgentType
	headers["sender_agent_id"] = rh.Sender.AgentID
	headers["sender_pod_name"] = rh.Sender.PodName
	headers["sender_agent_version"] = rh.Sender.AgentVersion
	headers["sender_role"] = rh.Sender.Role

	if rh.FunctionalRole != "" {
		headers["functional_role"] = rh.FunctionalRole
	}

	// Orchestration Context
	headers["orchestration_id"] = rh.OrchestrationID
	headers["orchestration_name"] = rh.OrchestrationName
	headers["step_id"] = rh.StepID
	headers["step_name"] = rh.StepName
	headers["request_id"] = rh.RequestID
	headers["retry_version"] = fmt.Sprintf("%d", rh.RetryVersion)
	headers["parent_orchestration_id"] = rh.ParentOrchestrationID
	headers["parent_orchestration_name"] = rh.ParentOrchestrationName
	headers["reply_to_request_id"] = rh.ReplyToRequestID

	// Message Metadata
	headers["message_id"] = rh.MessageID
	headers["message_type"] = rh.MessageType
	headers["from_agent"] = rh.FromAgent
	headers["to_agent"] = rh.ToAgent
	headers["to_agent_type"] = rh.ToAgentType
	headers["action"] = rh.Action
	headers["timestamp"] = rh.Timestamp.Format(time.RFC3339)

	// Resource Management
	headers["fuel_budget"] = fmt.Sprintf("%d", rh.FuelBudget)
	headers["timeout_seconds"] = fmt.Sprintf("%d", rh.TimeoutSeconds)

	// Routing
	headers["requests_topic"] = rh.RequestsTopic
	headers["responses_topic"] = rh.ResponsesTopic

	return headers
}

// In types/context.go, add these methods:

// AdjustToReceiverPerspective transforms the context from sender's to receiver's view
func (ec *ExecutionContext) AdjustToReceiverPerspectiveOLD(receiverIdentity AgentIdentity) *ExecutionContext {
	if ec.MessageType == "response" {
		// For responses, flip the perspective
		return &ExecutionContext{
			OrchestrationID:   ec.InResponseTo.ParentOrchestrationID,
			OrchestrationName: ec.InResponseTo.ParentOrchestrationName,
			Sender:            receiverIdentity,
			InResponseTo:      ec.InResponseTo,
			CorrelationID:     ec.CorrelationID,
			ClientID:          ec.ClientID,
			Status:            ec.Status,
			MessageType:       "response",
			Timestamp:         time.Now(),
		}
	} else {
		// For requests, create new orchestration
		return &ExecutionContext{
			OrchestrationID:         uuid.New().String(),
			OrchestrationName:       GenerateReadableName(receiverIdentity.AgentType, "orchestration"),
			Sender:                  receiverIdentity,
			ParentOrchestrationID:   ec.OrchestrationID,
			ParentOrchestrationName: ec.OrchestrationName,
			ReplyToRequestID:        ec.RequestID,
			Action:                  ec.Action,
			RequestID:               ec.RequestID,
			ResponsesTopic:          ec.ResponsesTopic,
			CorrelationID:           ec.CorrelationID,
			ClientID:                ec.ClientID,
			FuelBudget:              ec.FuelBudget,
			FunctionalRole:          ec.FunctionalRole,
			MessageType:             "request",
			Timestamp:               time.Now(),
		}
	}
}

// DetermineResponseOrchestrationTarget explicitly determines which orchestration
// should handle a response based on the relationship between sender and receiver
func (ec *ExecutionContext) DetermineResponseOrchestrationTarget() (targetOrchestrationID, targetOrchestrationName string) {
	// If this is a child responding to its parent
	if ec.ParentOrchestrationID != "" {
		// Response goes to the parent orchestration
		return ec.ParentOrchestrationID, ec.ParentOrchestrationName
	}

	// If this is a response to a request (peer-to-peer or otherwise)
	if ec.InResponseTo != nil && ec.InResponseTo.ParentOrchestrationID != "" {
		// Response goes to whoever sent the original request
		return ec.InResponseTo.ParentOrchestrationID, ec.InResponseTo.ParentOrchestrationName
	}

	// If we have a ResponsesTopic set, this implies someone is waiting for our response
	// but we need to figure out who. This might need additional context.
	// For now, log a warning
	if ec.ResponsesTopic != "" {
		fmt.Print("DEBUG error: no responses topic set in DetermineResponseOrchestrationTarget...\n")
		// This is a problem - we have a topic but don't know the orchestration ID
		// Return empty and let the caller handle it
		return "", ""
	}

	// Default case - shouldn't happen in well-formed messages
	return "", ""
}

// ValidateResponseRouting ensures a response has proper routing information
func (rh *ResponseHeaders) ValidateResponseRouting() error {
	if rh.OrchestrationID == "" {
		return fmt.Errorf("response has no target orchestration_id - cannot route response")
	}

	if rh.MyOrchestrationID == "" {
		return fmt.Errorf("response has no my_orchestration_id - sender identity missing")
	}

	// Ensure we're not routing to ourselves (unless intentional)
	if rh.OrchestrationID == rh.MyOrchestrationID {
		// This might be valid in some cases, but worth logging
		return fmt.Errorf("response would route to self (orchestration_id == my_orchestration_id)")
	}

	return nil
}
