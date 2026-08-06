// FILE: platform/orchestration/agenterrors/agenterrors.go
//
// The ONE writer against agent_error_log, in a leaf package importable from
// BOTH sides of the platform/orchestration -> actions import edge.
//
// WHY THIS PACKAGE EXISTS (RFC_012, owner ruling 2026-08-06, option B).
// orchestration.LogAgentError was already "the one INSERT against
// agent_error_log" — but it lives in the package that imports `actions`, so no
// action could call it without a cycle, and the actions package accumulated
// NINETEEN hand-copied INSERTs whose column lists drifted: 8/9/10/11/13
// columns, orchestration_id written last or absent entirely (a row that cannot
// be joined to its run), and three incompatible null-handling disciplines for
// the same uuid columns. Moving the writer one package down retires the copy
// class; orchestration.LogAgentError remains as a thin forwarder so its
// existing callers (agentbase, messaging, the coordinator) compile unchanged.
//
// THE FINDINGS-PLUS-AWAIT DOOR (RecordFindings) is the named escape hatch for
// the await-overwrite class RFC_012 records: an action that both computes
// findings and dispatches an awaited request loses every in-memory copy of
// those findings — the sibling collected_data key was REFUTED live
// (persistAwaitingStateWithRetry loads fresh state at park time and copies
// only awaited-request entries across; RFC_012 addendum 2). THIS TABLE IS THE
// ONLY SINK THAT SURVIVES AN AWAITED STEP. Rows must be written BEFORE the
// dispatch, so a failed send cannot unrecord what the action found — the
// pattern proven end to end by bugs_closed/098's retraction audit
// (RETRACTION_AUDIT / RETRACTION_REFUSED rows, live-probed 2026-08-05).
//
// Best-effort by design, and the failure is COUNTED, not swallowed: Write
// returns whether the row landed, and RecordFindings returns
// (attempted, recorded) so the caller can put the difference in its audit —
// recorded-but-lost must never read as recorded.
package agenterrors

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"go.uber.org/zap"
)

// Entry represents a row in agent_error_log. Field-for-field the shape
// orchestration.AgentErrorEntry always had; that name is now an alias of this
// type, so the two cannot drift.
type Entry struct {
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

// Write persists one row. Best-effort: a failure to record must never change
// the disposition the caller has already decided — it warns and returns false.
// The return value is the caller's evidence; ignore it only when nothing
// downstream claims the row exists.
//
// The INSERT below is deliberately byte-compatible with the historical
// orchestration.LogAgentError statement (same columns, same order, same
// NULLIF-uuid casts): every sqlmock expectation that pinned the 13-argument
// shape keeps passing, and the live table sees no new statement fingerprint.
func Write(ctx context.Context, db *sql.DB, logger *zap.Logger, entry Entry) bool {
	if db == nil {
		return false
	}

	if entry.Severity == "" {
		entry.Severity = "error"
	}
	if entry.ErrorCode == "" {
		entry.ErrorCode = ClassifyError(entry.ErrorMessage)
	}

	contextJSON, _ := json.Marshal(entry.Context)
	if contextJSON == nil {
		contextJSON = []byte("{}")
	}

	_, err := db.ExecContext(ctx, `
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
		if logger != nil {
			logger.Warn("Failed to write to agent_error_log — the disposition stands but its durable trace is lost",
				zap.Error(err),
				zap.String("error_message", entry.ErrorMessage))
		}
		return false
	}
	return true
}

// Finding is one condition an action computed and needs to survive an await:
// a refusal, a stranded target, an audit summary. Identity comes from the
// base Entry passed to RecordFindings; a Finding carries only what varies.
type Finding struct {
	// ErrorCode is the finding's queryable name (e.g. RETRACTION_REFUSED).
	// Required — a findings row with an inferred code is not findable.
	ErrorCode string
	// Severity: "error", "warning", or "info" ("info" is sanctioned here for
	// always-written audit rows — the row every real run writes, clean ones
	// included, so the absence of rows never has two readings).
	Severity string
	Message  string
	Context  map[string]interface{}
}

// RecordFindings persists each finding as its own row, carrying the base
// entry's identity. Call it BEFORE the dispatch the findings relate to —
// after park, the DB rows are the only copy (RFC_012 addendum 2).
//
// Returns (attempted, recorded). Put the difference in your audit rather than
// letting a lost row read as a recorded one:
//
//	attempted, recorded := agenterrors.RecordFindings(ctx, db, logger, base, findings)
//	audit["conditions_recorded"] = recorded
//	if attempted != recorded { audit["conditions_lost"] = attempted - recorded }
func RecordFindings(ctx context.Context, db *sql.DB, logger *zap.Logger, base Entry, findings []Finding) (attempted, recorded int) {
	for _, f := range findings {
		attempted++
		row := base
		row.ErrorCode = f.ErrorCode
		row.Severity = f.Severity
		row.ErrorMessage = f.Message
		row.Context = f.Context
		if Write(ctx, db, logger, row) {
			recorded++
		}
	}
	return attempted, recorded
}

// ClassifyError derives an error_code from the error message. Moved here with
// the writer (it is pure string classification); orchestration keeps calling
// it through the forwarder.
func ClassifyError(msg string) string {
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
