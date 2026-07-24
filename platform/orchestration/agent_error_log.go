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
	"github.com/gqls/agentchassis/platform/orchestration/types"
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

// ── Adapter-reported conditions ──────────────────────────────────────────────
//
// A remote service can flag a non-fatal condition it detected while
// SUCCEEDING — e.g. the image adapter's unrouted-kind guard (bugs_open/011
// §4 residual, council e996bf0a): the request was served, but from a weaker
// provider nobody chose. The detecting service has no DB handle, and this
// coordinator must not re-derive the check (it would be predicting config
// compiled into a different binary — the dedup-index↔Go-list drift class).
// So the contract is: the adapter puts a `reported_conditions` list in its
// response body; every complete response crosses handleCompleteResponse,
// which calls persistReportedConditions; each condition becomes an
// agent_error_log row (resolved=false), where dashboards and the
// immune-system sweep already look. Detection lives with the deployed truth,
// persistence with the DB.

// maxReportedConditionsPerResponse bounds how many condition rows one
// response may create — a misbehaving adapter must not flood the table.
const maxReportedConditionsPerResponse = 10

// conditionReportingAgentTypes is the set of senders whose
// reported_conditions the chassis will persist. This gate exists by
// guardian veto (council e996bf0a, round 5): an unconditional parse in
// handleCompleteResponse would be a fleet-wide, ungoverned reporting
// channel that any future adapter could start feeding by emitting the key
// — "intentionally or by copy-paste coincidence" — with zero contract
// review. Adding a sender here IS the review: one line, one council
// submission, per adapter. An unsanctioned sender emitting the field is
// warned about loudly and persisted never. This list is owned by the
// consumer and describes whom it believes — it does not predict any other
// binary's config, so the routing-table drift class does not apply.
var conditionReportingAgentTypes = map[string]bool{
	"image-generator": true,
}

// senderMayReportConditions reports whether a sender's conditions are
// sanctioned for persistence. Split out for testability.
func senderMayReportConditions(agentType string) bool {
	return conditionReportingAgentTypes[agentType]
}

// reportedCondition is one entry of a response body's `reported_conditions`
// list, after the Kafka JSON round-trip.
type reportedCondition struct {
	Code     string
	Severity string
	Message  string
	Context  map[string]interface{}
}

// parseReportedConditions extracts well-formed conditions from a normalised
// response body. Malformed entries are skipped, not fatal: a bad condition
// report must never break response processing. Pure — unit-tested without a
// coordinator.
//
// Three returns beyond the conditions, all there because silence about lost
// data is the very failure this mechanism exists to cure — and the cure
// must not reproduce it at its own edges (bug_historian, council rounds 5,
// 7 and 8):
//
//	malformed — the field was PRESENT but unusable as a whole (not a list,
//	  or a list yielding nothing). Distinct from absent, which is healthy
//	  and silent. A contract reshape upstream must surface, not read as "no
//	  conditions to report".
//	skipped — how many individual entries were dropped as junk while others
//	  parsed fine. A mixed list is the subtler case: without this count, some
//	  conditions vanish and the surviving ones make the response look wholly
//	  healthy. The caller warns on any non-zero value.
//	truncated — how many entries were never examined because the
//	  per-response cap fired. Distinct from skipped: these were dropped for
//	  CAPACITY, not because they were junk, so the remedy differs (raise the
//	  cap or report less, versus fix the emitter) and folding them into
//	  skipped would misdiagnose a healthy-but-chatty reporter as a broken
//	  one. The caller warns on any non-zero value.
func parseReportedConditions(normalisedData map[string]interface{}) (conditions []reportedCondition, malformed bool, skipped, truncated int) {
	raw, present := normalisedData["reported_conditions"]
	if !present {
		return nil, false, 0, 0
	}
	list, ok := raw.([]interface{})
	if !ok {
		return nil, true, 0, 0 // present but not a list — the contract broke
	}

	for i, item := range list {
		if len(conditions) >= maxReportedConditionsPerResponse {
			// Everything from here on is dropped unexamined. Count it —
			// an over-cap loss with skipped=0 must not read as a clean
			// parse of a short list (bug_historian, round 8).
			truncated = len(list) - i
			break
		}
		m, ok := item.(map[string]interface{})
		if !ok {
			skipped++
			continue
		}
		c := reportedCondition{}
		c.Code, _ = m["code"].(string)
		c.Severity, _ = m["severity"].(string)
		c.Message, _ = m["message"].(string)
		c.Context, _ = m["context"].(map[string]interface{})
		if c.Code == "" && c.Message == "" {
			skipped++ // nothing worth a row, but it was MEANT to be one
			continue
		}
		if c.Severity == "" {
			c.Severity = "warning"
		}
		if c.Message == "" {
			c.Message = c.Code
		}
		conditions = append(conditions, c)
	}
	// A non-empty list that yielded nothing usable is a broken contract,
	// not a healthy absence — reported as malformed rather than as N
	// skipped entries, because there is no surviving signal at all.
	if len(list) > 0 && len(conditions) == 0 {
		return nil, true, skipped, truncated
	}
	return conditions, false, skipped, truncated
}

// persistReportedConditions writes each adapter-reported condition to
// agent_error_log. Best-effort, like logAgentError: a failed insert is
// logged and never affects the workflow. Only sanctioned senders persist
// (see conditionReportingAgentTypes) — for everyone else this is a no-op
// beyond one map lookup, and an unsanctioned sender actually emitting the
// field is warned about by name so the attempt cannot pass silently.
func (s *SagaCoordinator) persistReportedConditions(ctx context.Context, state *OrchestrationState, stepName string, response types.ResponseMessage, normalisedData map[string]interface{}) {
	if _, present := normalisedData["reported_conditions"]; !present {
		return // the fleet-wide healthy path: one lookup, nothing else
	}

	senderType := response.Headers.Sender.AgentType
	if !senderMayReportConditions(senderType) {
		s.logger.Warn("reported_conditions from an UNSANCTIONED sender — not persisted; add the agent type to conditionReportingAgentTypes (with review) if this reporter is intended",
			zap.String("sender_agent_type", senderType),
			zap.String("step_name", stepName))
		return
	}

	conditions, malformed, skipped, truncated := parseReportedConditions(normalisedData)
	if malformed {
		s.logger.Warn("reported_conditions present but MALFORMED — the reporting contract broke; conditions were lost, fix the emitter",
			zap.String("sender_agent_type", senderType),
			zap.String("step_name", stepName))
		return
	}
	// A partially-bad list is the dangerous case: the entries that DID
	// parse make the response look healthy while others vanish. Say so.
	if skipped > 0 {
		s.logger.Warn("reported_conditions partly UNPARSEABLE — some entries were dropped and are lost; the surviving ones are not the whole picture, fix the emitter",
			zap.Int("dropped", skipped),
			zap.Int("parsed", len(conditions)),
			zap.String("sender_agent_type", senderType),
			zap.String("step_name", stepName))
	}
	// Over-cap loss is a different failure from junk: well-formed entries
	// were dropped for CAPACITY. Without its own count a chatty reporter's
	// losses read as a clean parse of a short list.
	if truncated > 0 {
		s.logger.Warn("reported_conditions TRUNCATED at the per-response cap — well-formed entries were dropped unexamined; raise the cap or report less",
			zap.Int("truncated", truncated),
			zap.Int("persisted", len(conditions)),
			zap.Int("cap", maxReportedConditionsPerResponse),
			zap.String("sender_agent_type", senderType),
			zap.String("step_name", stepName))
	}
	if len(conditions) == 0 {
		return
	}

	for _, c := range conditions {
		entry := s.buildErrorEntry(state, stepName, c.Message)
		// The row's subject is the REPORTING service, not this coordinator —
		// its identity travels in the response headers it set itself.
		if response.Headers.Sender.AgentType != "" {
			entry.AgentType = response.Headers.Sender.AgentType
		}
		if response.Headers.Sender.AgentID != "" {
			entry.AgentID = response.Headers.Sender.AgentID
		}
		if response.Headers.Sender.PodName != "" {
			entry.PodName = response.Headers.Sender.PodName
		}
		entry.ErrorCode = c.Code
		entry.Severity = c.Severity
		if len(c.Context) > 0 {
			if entry.Context == nil {
				entry.Context = map[string]interface{}{}
			}
			// Namespaced so a condition's keys cannot collide with the
			// step-context snapshot buildErrorEntry already assembled.
			entry.Context["reported"] = c.Context
		}
		s.logAgentError(ctx, entry)
	}

	s.logger.Info("Persisted adapter-reported conditions to agent_error_log",
		zap.Int("count", len(conditions)),
		zap.String("step_name", stepName),
		zap.String("sender_agent_type", response.Headers.Sender.AgentType))
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
