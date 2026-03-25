// FILE: platform/orchestration/actions/ai_errors.go
//
// Error types for AI endpoint unavailability.
//
// When an LLM call fails due to infrastructure (connection refused, DNS failure,
// credit exhaustion), the work item should go back to triaged — not count as a
// failed attempt. This file provides the error type and detection helper.
//
// Usage in execute_llm_prompt:
//   if isAIUnavailable(err) {
//       return nil, &AIUnavailableError{Provider: ..., Cause: err}
//   }
//
// Usage in coordinator/dispatch:
//   var aiErr *AIUnavailableError
//   if errors.As(err, &aiErr) {
//       // release item to triaged, don't increment attempt_count
//   }

package actions

import (
	"fmt"
	"strings"
)

// AIUnavailableError indicates the AI endpoint is unreachable or has
// auth/credit issues. Work items that fail with this error should be
// released back to triaged without counting as a failed attempt.
type AIUnavailableError struct {
	Provider string
	Model    string
	Endpoint string
	Cause    error
}

func (e *AIUnavailableError) Error() string {
	return fmt.Sprintf("AI endpoint unavailable: provider=%s model=%s endpoint=%s: %v",
		e.Provider, e.Model, e.Endpoint, e.Cause)
}

func (e *AIUnavailableError) Unwrap() error {
	return e.Cause
}

// isAIUnavailable checks whether an error indicates the AI endpoint is
// unreachable or has auth/credit problems (as opposed to a model error
// or bad output, which are real failures).
func isAIUnavailable(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()

	// Connection-level failures (endpoint down, DNS, timeout)
	connectionErrors := []string{
		"connection refused",
		"no such host",
		"i/o timeout",
		"connection reset",
		"dial tcp",
		"context deadline exceeded",
		"EOF",
		"transport connection broken",
		"TLS handshake timeout",
	}
	for _, pattern := range connectionErrors {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Auth/credit failures (endpoint reachable but won't serve)
	creditErrors := []string{
		"status 401",
		"status 402",
		"credit",
		"insufficient_quota",
		"authentication",
		"api key",
	}
	for _, pattern := range creditErrors {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	// Ollama-specific: model not loaded / OOM
	ollamaErrors := []string{
		"requires more system memory",
		"model not found",
		"failed to load model",
	}
	for _, pattern := range ollamaErrors {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
