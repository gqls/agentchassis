// FILE: platform/errors/errors.go
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrorCode represents standardized error codes across the platform
type ErrorCode string

const (
	// General errors
	ErrInternal     ErrorCode = "INTERNAL_ERROR"
	ErrValidation   ErrorCode = "VALIDATION_ERROR"
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
	ErrConflict     ErrorCode = "CONFLICT"
	ErrRateLimited  ErrorCode = "RATE_LIMITED"

	// Workflow errors
	ErrWorkflowInvalid  ErrorCode = "WORKFLOW_INVALID"
	ErrWorkflowTimeout  ErrorCode = "WORKFLOW_TIMEOUT"
	ErrWorkflowFailed   ErrorCode = "WORKFLOW_FAILED"
	ErrInsufficientFuel ErrorCode = "INSUFFICIENT_FUEL"

	// Agent errors
	ErrAgentNotFound   ErrorCode = "AGENT_NOT_FOUND"
	ErrAgentTimeout    ErrorCode = "AGENT_TIMEOUT"
	ErrAgentOverloaded ErrorCode = "AGENT_OVERLOADED"

	// External service errors
	ErrExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrAIServiceError  ErrorCode = "AI_SERVICE_ERROR"
)

// DomainError represents a standardized error in the platform
type DomainError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	Timestamp  time.Time              `json:"timestamp"`
	TraceID    string                 `json:"trace_id,omitempty"`
	Retryable  bool                   `json:"retryable"`
	RetryAfter *time.Duration         `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap allows for error chain inspection
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *DomainError) HTTPStatus() int {
	switch e.Code {
	case ErrValidation:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrAgentOverloaded:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// MarshalJSON customizes JSON serialization
func (e *DomainError) MarshalJSON() ([]byte, error) {
	type Alias DomainError
	return json.Marshal(&struct {
		*Alias
		HTTPStatus int `json:"http_status"`
	}{
		Alias:      (*Alias)(e),
		HTTPStatus: e.HTTPStatus(),
	})
}

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder struct {
	err *DomainError
}

// New creates a new error builder
func New(code ErrorCode, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &DomainError{
			Code:      code,
			Message:   message,
			Timestamp: time.Now().UTC(),
			Retryable: false,
		},
	}
}

// WithCause adds an underlying cause
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	return b
}

// WithDetails adds additional details
func (b *ErrorBuilder) WithDetails(details map[string]interface{}) *ErrorBuilder {
	b.err.Details = details
	return b
}

// WithDetail adds a single detail
func (b *ErrorBuilder) WithDetail(key string, value interface{}) *ErrorBuilder {
	if b.err.Details == nil {
		b.err.Details = make(map[string]interface{})
	}
	b.err.Details[key] = value
	return b
}

// WithTraceID adds a trace ID for correlation
func (b *ErrorBuilder) WithTraceID(traceID string) *ErrorBuilder {
	b.err.TraceID = traceID
	return b
}

// AsRetryable marks the error as retryable
func (b *ErrorBuilder) AsRetryable(retryAfter *time.Duration) *ErrorBuilder {
	b.err.Retryable = true
	b.err.RetryAfter = retryAfter
	return b
}

// Build returns the constructed error
func (b *ErrorBuilder) Build() *DomainError {
	return b.err
}

// Helper functions for common errors

// NotFound creates a not found error
func NotFound(resource string, id string) *DomainError {
	return New(ErrNotFound, fmt.Sprintf("%s not found", resource)).
		WithDetail("resource", resource).
		WithDetail("id", id).
		Build()
}

// ValidationError creates a validation error
func ValidationError(field string, issue string) *DomainError {
	return New(ErrValidation, "Validation failed").
		WithDetail("field", field).
		WithDetail("issue", issue).
		Build()
}

// InternalError creates an internal error
func InternalError(message string, cause error) *DomainError {
	return New(ErrInternal, message).
		WithCause(cause).
		Build()
}

// InsufficientFuel creates an insufficient fuel error
func InsufficientFuel(required, available int, action string) *DomainError {
	return New(ErrInsufficientFuel, "Insufficient fuel for operation").
		WithDetail("required", required).
		WithDetail("available", available).
		WithDetail("action", action).
		Build()
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Retryable
	}
	return false
}

// GetRetryAfter gets the retry after duration if available
func GetRetryAfter(err error) *time.Duration {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.RetryAfter
	}
	return nil
}

// AgentError represents an error with agent context
type AgentError struct {
	Err             error  // The underlying error
	Message         string // Human-readable message
	AgentType       string // Type of agent where error occurred
	AgentID         string // ID of agent where error occurred
	OrchestrationID string // Orchestration context
	StepName        string // Workflow step name
	Action          string // Action being executed
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v (agent=%s, orch=%s, step=%s, action=%s)",
			e.Message,
			e.Err,
			e.AgentType,
			e.OrchestrationID,
			e.StepName,
			e.Action)
	}
	return fmt.Sprintf("%v (agent=%s, orch=%s, step=%s, action=%s)",
		e.Err,
		e.AgentType,
		e.OrchestrationID,
		e.StepName,
		e.Action)
}

// Unwrap returns the underlying error (for errors.Is and errors.As)
func (e *AgentError) Unwrap() error {
	return e.Err
}

// WrapWithAgentContext wraps an error with agent execution context
func WrapWithAgentContext(
	err error,
	message string,
	agentType string,
	agentID string,
	orchestrationID string,
	stepName string,
	action string,
) error {
	if err == nil {
		return nil
	}

	return &AgentError{
		Err:             err,
		Message:         message,
		AgentType:       agentType,
		AgentID:         agentID,
		OrchestrationID: orchestrationID,
		StepName:        stepName,
		Action:          action,
	}
}

// Wrap wraps an error with additional context message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// GetAgentContext extracts agent context from an error if available
func GetAgentContext(err error) (agentType, agentID, orchestrationID, stepName, action string, ok bool) {
	if ae, isAgentError := err.(*AgentError); isAgentError {
		return ae.AgentType, ae.AgentID, ae.OrchestrationID, ae.StepName, ae.Action, true
	}
	return "", "", "", "", "", false
}

// IsRecoverable determines if an error is recoverable
func IsRecoverable(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Recoverable error patterns
	recoverablePatterns := []string{
		"timeout",
		"connection refused",
		"temporary failure",
		"network",
		"dial tcp",
		"i/o timeout",
		"context deadline exceeded",
		"too many requests",
		"rate limit",
		"503",
		"502",
		"504",
	}

	for _, pattern := range recoverablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
