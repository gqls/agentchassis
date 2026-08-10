// FILE: platform/errors/errors.go
package errors

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
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

	// Dispatch-resolution errors (bugs_open/239). An orchestration-action
	// request whose target workflow cannot be resolved must fail CLOSED. Before
	// these existed, every such failure fell through to the CONSUMING pod's own
	// default workflow — on the shared chassis that is the `generic` agent's
	// single no-op step — and the run was stamped COMPLETED with an empty
	// execution_path, so a dispatch that did nothing at all read as a fast
	// success.
	//
	// The pair is the point: the two failures need OPPOSITE dispositions.
	// ErrDispatchUnresolvable is terminal (an unparseable body or an unknown
	// agent type cannot heal by being retried); ErrDispatchLookupUnavailable is
	// transient (the same message against a healthy database resolves fine), so
	// it is built AsRetryable and left for the intake pool's re-attempt.
	ErrDispatchUnresolvable      ErrorCode = "DISPATCH_UNRESOLVABLE"
	ErrDispatchLookupUnavailable ErrorCode = "DISPATCH_LOOKUP_UNAVAILABLE"
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

// AsDomainError finds a *DomainError anywhere in err's chain.
//
// bugs_open/195. This exists so callers stop writing `err.(*DomainError)`: a
// bare type assertion sees only the outermost error, so it starts returning
// false the moment anything wraps — including a `%w` wrap that deliberately
// preserves the chain. Classification that silently degrades when someone adds
// a wrapper is how a permanent failure gets retried for ever.
func AsDomainError(err error) (*DomainError, bool) {
	var de *DomainError
	if stderrors.As(err, &de) {
		return de, true
	}
	return nil, false
}

// CodeOf returns the ErrorCode carried anywhere in err's chain, or "" if the
// chain holds no *DomainError. The code is exact where the message text is not:
// it cannot be defeated by rewording or capitalisation (bugs_open/195).
func CodeOf(err error) ErrorCode {
	if de, ok := AsDomainError(err); ok {
		return de.Code
	}
	return ""
}

// IsDispatchUnresolvable reports whether err is a TERMINAL bugs_open/239
// dispatch-resolution refusal: the request named a workflow the fleet cannot
// resolve, so nothing ran and nothing will.
//
// Keyed on the code, not the message, for the reason CodeOf documents: the
// intake pool switches an event to `failed` on this, and a classification that
// can be defeated by rewording would put a message back on the retry path that
// can never succeed.
func IsDispatchUnresolvable(err error) bool {
	return CodeOf(err) == ErrDispatchUnresolvable
}

// IsDispatchLookupUnavailable reports whether err is a TRANSIENT bugs_open/239
// dispatch-resolution failure — the agent-definition lookup itself faulted, so
// the same message is worth re-attempting.
func IsDispatchLookupUnavailable(err error) bool {
	return CodeOf(err) == ErrDispatchLookupUnavailable
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

// IsRetryable checks if an error is retryable.
//
// Uses AsDomainError, not a bare type assertion: a %w-wrapped DomainError is
// still a DomainError, and err.(*DomainError) silently answers false for one.
// That is the same defect bugs_open/195 fixed in the permanent classifier
// (code-review F8, 2026-08-05). Both this and GetRetryAfter had zero callers
// fleet-wide when this was corrected, so the fix changes no behaviour today —
// it stops the first caller inheriting the bug.
func IsRetryable(err error) bool {
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr.Retryable
	}
	return false
}

// GetRetryAfter gets the retry after duration if available. See IsRetryable on
// why this uses AsDomainError rather than a bare type assertion.
func GetRetryAfter(err error) *time.Duration {
	if domainErr, ok := AsDomainError(err); ok {
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

// IsRecoverable was deleted by bugs_open/197. It was a twelve-pattern,
// case-folded substring classifier with ZERO callers fleet-wide (verified
// 2026-08-06) that disagreed with the live agentbase classifier on case —
// two lists for one judgement, the drift bugs_closed/034 closed on the
// permanent side. The judgement now lives in ONE place: the pattern set was
// rebuilt from a live census with each entry individually argued, including
// the patterns from this function's list that were REJECTED and why. Do not
// recreate a SECOND recoverability list anywhere.
//
// > **CORRECTED 2026-08-08 (bugs_open/217).** This note used to say the one
// > place was messaging.MatchedTransientFailure and that this package "is the
// > wrong altitude for operational retry policy". The single implementation
// > now lives HERE (transient_failure.go / permanent_failure.go), because the
// > coordinator's failure sender needed RetryDisposition and messaging imports
// > orchestration — a cycle. What made the old IsRecoverable rot was not its
// > package but its ZERO callers beside a live twin; the relocated classifier
// > is the live twin, called by messaging (via re-exports), agentbase and the
// > coordinator. The rule that survives: ONE list, and change it only through
// > its pinning tests (platform/messaging).
