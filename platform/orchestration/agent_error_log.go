// FILE: platform/orchestration/agent_error_log.go
//
// Persistent error logging for agent workflows. Writes to agent_error_log
// table so errors are queryable without kubectl log access.
//
// Called from two places in coordinator.go:
//   - routeToErrorStep (step failed, routing to error handler)
//   - notifyParentOfFailure (workflow failed entirely)

package orchestration

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// AgentErrorEntry represents a row in agent_error_log
type AgentErrorEntry struct {
	SiteID          string
	Domain          string
	WorkItemID      string
	OrchestrationID string
	AgentType       string
	AgentID         string
	PodName         string
	StepName        string
	Action          string
	ErrorMessage    string
	ErrorCode       string
	Severity        string
	Context         map[string]interface{}
}

// logAgentError writes an error to the agent_error_log table.
// Best-effort — failures are logged but don't affect the workflow.
func (s *SagaCoordinator) logAgentError(ctx context.Context, entry AgentErrorEntry) {
	if s.db == nil {
		return
	}

	if entry.Severity == "" {
		entry.Severity = "error"
	}
	if entry.ErrorCode == "" {
		entry.ErrorCode = classifyError(entry.ErrorMessage)
	}

	contextJSON, _ := json.Marshal(entry.Context)
	if contextJSON == nil {
		contextJSON = []byte("{}")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO agent_error_log (
			site_id, domain, work_item_id, orchestration_id,
			agent_type, agent_id, pod_name, step_name, action,
			error_message, error_code, severity, context
		) VALUES (
			NULLIF($1, '')::uuid, $2, NULLIF($3, '')::uuid, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13::jsonb
		)
	`,
		entry.SiteID, entry.Domain, entry.WorkItemID, entry.OrchestrationID,
		entry.AgentType, entry.AgentID, entry.PodName, entry.StepName, entry.Action,
		entry.ErrorMessage, entry.ErrorCode, entry.Severity, string(contextJSON),
	)
	if err != nil {
		s.logger.Warn("Failed to write to agent_error_log",
			zap.Error(err),
			zap.String("error_message", entry.ErrorMessage))
	}
}

// buildErrorEntry extracts an AgentErrorEntry from OrchestrationState.
// Used by both routeToErrorStep and notifyParentOfFailure.
func (s *SagaCoordinator) buildErrorEntry(state *OrchestrationState, stepName string, errorMsg string) AgentErrorEntry {
	entry := AgentErrorEntry{
		OrchestrationID: state.OrchestrationID,
		AgentType:       state.OwnerAgentType,
		AgentID:         state.OwnerAgentID,
		PodName:         s.podName,
		StepName:        stepName,
		ErrorMessage:    errorMsg,
	}

	// Extract site context from collected_data
	if state.CollectedData != nil {
		entry.SiteID = datahelpers.ExtractNestedFieldString(state.CollectedData, "site_record.site_id")
		entry.Domain = datahelpers.ExtractNestedFieldString(state.CollectedData, "site_record.domain")
		if entry.Domain == "" {
			entry.Domain = datahelpers.ExtractNestedFieldString(state.CollectedData, "input_data.domain")
		}
		entry.WorkItemID = datahelpers.ExtractNestedFieldString(state.CollectedData, "input_data.work_item_id")

		// Extract the action from the step that failed
		if stepName != "" {
			if step, exists := state.WorkflowPlan.Steps[stepName]; exists {
				entry.Action = step.Action
			}
		}

		// Build a compact context snapshot
		entry.Context = buildErrorContext(state.CollectedData, stepName)
	}

	return entry
}

// buildErrorContext creates a compact snapshot of relevant data for debugging.
// Avoids storing the entire collected_data (which can be megabytes).
func buildErrorContext(collectedData map[string]interface{}, stepName string) map[string]interface{} {
	ctx := map[string]interface{}{}

	// Item type and spec
	if itemType := datahelpers.ExtractNestedFieldString(collectedData, "input_data.item_type"); itemType != "" {
		ctx["item_type"] = itemType
	}
	if pageName := datahelpers.ExtractNestedFieldString(collectedData, "input_data.page_name"); pageName != "" {
		ctx["page_name"] = pageName
	}
	if spec, ok := collectedData["input_data"].(map[string]interface{}); ok {
		if s, ok := spec["spec"].(map[string]interface{}); ok {
			// Include category and a truncated description
			if cat, ok := s["category"].(string); ok {
				ctx["category"] = cat
			}
			if desc, ok := s["description"].(string); ok {
				if len(desc) > 150 {
					desc = desc[:150] + "..."
				}
				ctx["description"] = desc
			}
		}
	}

	// Handler agent
	if handler := datahelpers.ExtractNestedFieldString(collectedData, "input_data.handler_agent"); handler != "" {
		ctx["handler_agent"] = handler
	}

	// Step error context (from routeToErrorStep)
	if stepErr, ok := collectedData["__step_error"].(map[string]interface{}); ok {
		ctx["failed_step"] = stepErr["failed_step"]
	}

	return ctx
}

// classifyError derives an error_code from the error message.
// Used for grouping and filtering in dashboards.
func classifyError(msg string) string {
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "template") && strings.Contains(lower, "can't evaluate field"):
		return "TEMPLATE_FIELD_ERROR"
	case strings.Contains(lower, "validate_page_content") || strings.Contains(lower, "content validation"):
		return "CONTENT_VALIDATION_FAILED"
	case strings.Contains(lower, "fix_type is required"):
		return "MISSING_FIX_TYPE"
	case strings.Contains(lower, "context deadline exceeded") || strings.Contains(lower, "timeout"):
		return "TIMEOUT"
	case strings.Contains(lower, "connection refused") || strings.Contains(lower, "no such host"):
		return "CONNECTION_ERROR"
	case strings.Contains(lower, "anthropic") || strings.Contains(lower, "api"):
		return "LLM_API_ERROR"
	case strings.Contains(lower, "child_orchestration_failed"):
		return "CHILD_ORCHESTRATION_FAILED"
	case strings.Contains(lower, "failed to parse") || strings.Contains(lower, "unmarshal"):
		return "PARSE_ERROR"
	default:
		return "UNKNOWN"
	}
}
