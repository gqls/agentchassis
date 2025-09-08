package types

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
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
	ParentRequestID         string `json:"parent_request_id,omitempty"`

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
	Sender         AgentIdentity `json:"sender"` // NEW: Who sent this message
	ToAgentType    string        `json:"to_agent_type"`
	RequestsTopic  string        `json:"requests_topic,omitempty"`
	ResponsesTopic string        `json:"responses_topic,omitempty"`

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
	ParentRequestID         string `json:"parent_request_id"` // If spawned

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
	InResponseToStepID         string `json:"in_response_to_step_id"`
	InResponseToStepName       string `json:"in_response_to_step_name"`
	InResponseToParentOrchID   string `json:"in_response_to_parent_orchestration_id"`
	InResponseToParentOrchName string `json:"in_response_to_parent_orchestration_name"`
	InResponseToMessageID      string `json:"in_response_to_message_id"`
	InResponseToAction         string `json:"in_response_to_action"`
	RetryCount                 int    `json:"retry_count"`

	// My Context
	MyOrchestrationID   string `json:"my_orchestration_id"`
	MyOrchestrationName string `json:"my_orchestration_name"`
	MyRequestsTopic     string `json:"my_requests_topic"`
	MyResponsesTopic    string `json:"my_responses_topic"`

	// Identity
	CorrelationID   string `json:"correlation_id"`
	CorrelationName string `json:"correlation_name"`
	ClientID        string `json:"client_id"`
	MessageType     string `json:"message_type"` // "response"
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

type ResponseBody struct {
	Success bool                   `json:"success"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Body    struct {
		Result      interface{} `json:"result"`
		Calculation interface{} `json:"calculation,omitempty"`
		Error       *ErrorInfo  `json:"error,omitempty"`
	} `json:"body"`
}

type ErrorInfo struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
	RetryAfter  int    `json:"retry_after_seconds,omitempty"`
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
		ParentRequestID:         ec.ParentRequestID,

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
		RequestsTopic:  ec.RequestsTopic,
		ResponsesTopic: ec.ResponsesTopic,
	}
}

// ToResponseHeaders converts ExecutionContext to ResponseHeaders for sending responses
func (ec *ExecutionContext) ToResponseHeaders() ResponseHeaders {
	headers := ResponseHeaders{
		// Who Am I
		Sender: ec.Sender,

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
		ToAgent:         "", // Will be set based on InResponseTo

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
		ParentRequestID:         headers.ParentRequestID,

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
		ParentRequestID:         ec.RequestID,

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
			RequestID:               ec.RequestID,
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
		CorrelationID:   correlationID,
		ClientID:        clientID,
		OrchestrationID: uuid.New().String(),
		MessageID:       uuid.New().String(),
		RequestID:       uuid.New().String(),
		MessageType:     "request",
		FuelBudget:      1000,
		Timestamp:       time.Now().UTC(),
		Version:         "2.0",
	}
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
		// InResponseTo handled below
		Version: headers["version"],
	}

	// FIXED: Parse InResponseTo if it's a JSON string
	if inResponseToStr := headers["in_response_to"]; inResponseToStr != "" {
		var responseCtx ResponseContext
		if err := json.Unmarshal([]byte(inResponseToStr), &responseCtx); err == nil {
			ec.InResponseTo = &responseCtx
		}
		// If it fails to parse as JSON, might be legacy string format
		// In that case, create a minimal ResponseContext
		if ec.InResponseTo == nil {
			ec.InResponseTo = &ResponseContext{
				RequestID: inResponseToStr,
			}
		}
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

	return ec, nil
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
	// FIXED: Check for InResponseTo properly
	if ec.MessageType == "response" && ec.InResponseTo == nil {
		return fmt.Errorf("in_response_to is required for response messages")
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
