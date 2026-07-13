// FILE: platform/orchestration/timeout_helpers.go
// Timeout conversion utilities for HITL and other long-running actions
package datahelpers

import (
	"fmt"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

const (
	// Timeout limits for validation
	MinApprovalTimeout     = 60     // 1 minute
	MaxApprovalTimeout     = 604800 // 7 days
	DefaultApprovalTimeout = 86400  // 24 hours (1 day)
	DefaultRequestTimeout  = 180    // 3 minutes (existing default)
)

// ConvertStepTimeout converts config.timeout_seconds to step.Timeout
// This ensures workflow config timeouts are properly used by the system
func ConvertStepTimeout(step *models.Step, logger *zap.Logger) {
	// If step already has a timeout set, respect it
	if step.Timeout > 0 {
		logger.Debug("Step timeout already set",
			zap.String("action", step.Action),
			zap.Duration("timeout", step.Timeout))
		return
	}

	// Check if config has timeout_seconds
	if step.Config == nil {
		return
	}

	// Try to extract timeout_seconds from config
	var timeoutSeconds int

	switch v := step.Config["timeout_seconds"].(type) {
	case int:
		timeoutSeconds = v
	case float64:
		timeoutSeconds = int(v)
	case int64:
		timeoutSeconds = int(v)
	default:
		// No timeout_seconds in config, leave step.Timeout as 0
		return
	}

	// Validate timeout is reasonable
	if timeoutSeconds < 0 {
		logger.Warn("Negative timeout_seconds in config, ignoring",
			zap.String("action", step.Action),
			zap.Int("timeout_seconds", timeoutSeconds))
		return
	}

	// For approval actions, apply stricter validation
	if isApprovalAction(step.Action) {
		timeoutSeconds = validateApprovalTimeout(timeoutSeconds, logger)
	}

	// Convert to time.Duration and set on step
	step.Timeout = time.Duration(timeoutSeconds) * time.Second

	logger.Info("Converted timeout_seconds to step.Timeout",
		zap.String("action", step.Action),
		zap.Int("timeout_seconds", timeoutSeconds),
		zap.Duration("step_timeout", step.Timeout))
}

// validateApprovalTimeout ensures approval timeouts are within acceptable range
func validateApprovalTimeout(timeout int, logger *zap.Logger) int {
	if timeout < MinApprovalTimeout {
		logger.Warn("Approval timeout below minimum, using minimum",
			zap.Int("requested", timeout),
			zap.Int("minimum", MinApprovalTimeout))
		return MinApprovalTimeout
	}

	if timeout > MaxApprovalTimeout {
		logger.Warn("Approval timeout exceeds maximum, using maximum",
			zap.Int("requested", timeout),
			zap.Int("maximum", MaxApprovalTimeout))
		return MaxApprovalTimeout
	}

	return timeout
}

// isApprovalAction checks if an action is an approval-related action
func isApprovalAction(action string) bool {
	approvalActions := map[string]bool{
		"await_approval":             true,
		"create_approval_request":    true,
		"wait_for_approval_response": true,
		"process_approval_decision":  false, // This processes response, doesn't wait
	}

	return approvalActions[action]
}

// GetStepTimeout returns the timeout for a step, with fallback logic
func GetStepTimeout(step models.Step) time.Duration {
	if step.Timeout > 0 {
		return step.Timeout
	}

	// Check if it's an approval action with default timeout
	if isApprovalAction(step.Action) {
		return DefaultApprovalTimeout * time.Second
	}

	// Use system default
	return DefaultRequestTimeout * time.Second
}

// FormatTimeout formats a timeout duration for logging/display
func FormatTimeout(d time.Duration) string {
	hours := d.Hours()
	if hours >= 24 {
		days := hours / 24
		return fmt.Sprintf("%.1f days", days)
	}
	if hours >= 1 {
		return fmt.Sprintf("%.1f hours", hours)
	}
	minutes := d.Minutes()
	if minutes >= 1 {
		return fmt.Sprintf("%.1f minutes", minutes)
	}
	return fmt.Sprintf("%.0f seconds", d.Seconds())
}
