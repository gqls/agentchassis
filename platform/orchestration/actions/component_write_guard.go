// FILE: platform/orchestration/actions/component_write_guard.go
//
// Last gate before a whole-component rewrite overwrites the durable source.
//
// A whole-component writer (tool-improver and friends) can return a
// mid-generation fragment which is then written straight over
// content_components.html_template. On 2026-07-18 tool-improver saved a
// 10,272-char working tool back as 1,253 chars of bare CSS — no <script>, no
// markup, ending mid-declaration — and reported success. Nothing in the
// pipeline noticed; the live page survived only because the render had not
// re-propagated. bugs_open/012 has the full case, 016b §9 the pattern.
//
// Two upstream fixes narrow the window but neither closes it:
//   - migration 168 raised improve_tool / generate_tool_html 8000 → 32000,
//     which makes truncation rarer, never impossible;
//   - f32b208e5 decodes stop_reason/done_reason and hard-errors on a capped
//     completion — but only for callers going through GenerateText, and only
//     when the provider reports it. A fragment arriving by any other route
//     still lands here.
//
// So this gate is about the write itself: whatever produced it, a replacement
// bearing the marks of a cut-off generation is not persisted.
//
// ---------------------------------------------------------------------------
// Calibration — every threshold below was simulated against the full live
// component_versions history (29 recorded transitions) on 2026-07-18 before
// being committed. Two earlier candidate checks were DROPPED because that
// simulation caught them misfiring on real, legitimate rewrites:
//
//   - "the replacement lost a </script>/</div> region the current row had"
//     fired on 3 transitions, 2 of them legitimate: provocation-card
//     (10,300 → 6,618) and tool-list (9,290 → 11,588) both deliberately
//     dropped their JavaScript and both end cleanly on </section>.
//   - the same check ungated by size would have refused tool-list, a rewrite
//     that GREW by 25%.
//
// That produced the organising principle: TRUNCATION CANNOT GROW AN ARTIFACT.
// Every false positive observed was a replacement that grew, so the structural
// checks below are gated on the replacement being no larger than what it
// replaces. Final calibration: 1 block across 29 transitions — the confirmed
// bugs_open/012 write — and 0 false positives.
//
// If you add a check here, re-run that simulation first. A guard that refuses
// good work gets switched off, and then it protects nothing.
// ---------------------------------------------------------------------------
//
// Every check is also COMPARATIVE: it fires only when the replacement is worse
// than the row it would replace, never merely because the result is imperfect.
// An absolute quality gate here would block legitimate repairs on exactly the
// components most likely to need them — the already-broken ones.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// componentCollapseRatio is the fraction of the current template's length a
// replacement must retain. Below it, the write is a truncated generation
// rather than an edit.
//
// Grounded in the live history: 9 of 29 transitions shrank the template and
// the most aggressive LEGITIMATE one retained 64% (provocation-card,
// 10,300 → 6,618); none has ever retained less than 50%. The bugs_open/012
// final write retained 12%. A 0.5 floor sits in clear air between the two.
const componentCollapseRatio = 0.5

// balancedPairs are open/close tokens whose balance is checked in the
// replacement. An unterminated <script> is the direct signature of a
// completion cut mid-stream — it is what catches the bugs_open/012
// intermediate write (6,765 chars, 66% retained, comfortably inside the
// legitimate size band, but with <script> left open).
var balancedPairs = []struct{ open, close string }{
	{"<script", "</script>"},
	{"<style", "</style>"},
	{"<section", "</section>"},
}

// componentRegressionIssues compares a proposed html_template against the row
// it would overwrite and returns the reasons it must not be persisted. An
// empty result means the write is safe to proceed.
//
// Pure function: same inputs → same outputs, no DB access, so the thresholds
// stay directly testable against real component_versions rows.
func componentRegressionIssues(currentHTML, newHTML string) []string {
	// Nothing to regress against. A component with no current template is the
	// birth path's problem, and store_generated_component already gates that.
	if strings.TrimSpace(currentHTML) == "" {
		return nil
	}

	var issues []string

	// 1. Size collapse — the replacement is a fraction of what it replaces.
	retained := float64(len(newHTML)) / float64(len(currentHTML))
	if retained < componentCollapseRatio {
		issues = append(issues, fmt.Sprintf(
			"replacement is %d chars against the current %d (%.0f%% retained, floor %.0f%%) — a collapse this large is a truncated generation, not an edit",
			len(newHTML), len(currentHTML), 100*retained, 100*componentCollapseRatio))
	}

	// Truncation cannot grow an artifact. A replacement larger than the row it
	// replaces is a deliberate rewrite, so the structural checks below do not
	// apply to it — see the calibration note in the file header for the real
	// rewrites this exempts.
	if len(newHTML) > len(currentHTML) {
		return issues
	}

	// Case-folded copies for token matching only (<SCRIPT> must not slip past
	// a lowercase token). Lengths and messages use the originals.
	currentFolded := strings.ToLower(currentHTML)
	newFolded := strings.ToLower(newHTML)

	// 2. Unterminated tags in the replacement, but only where the current row
	//    was balanced. Comparative on purpose: if the component is ALREADY
	//    unbalanced, blocking here would trap it permanently — no rewrite
	//    could ever land to repair it.
	for _, pair := range balancedPairs {
		newOpen, newClose := strings.Count(newFolded, pair.open), strings.Count(newFolded, pair.close)
		if newOpen <= newClose {
			continue
		}
		curOpen, curClose := strings.Count(currentFolded, pair.open), strings.Count(currentFolded, pair.close)
		if curOpen <= curClose {
			issues = append(issues, fmt.Sprintf(
				"replacement leaves %s unterminated (%d open vs %d close) where the current template is balanced — the completion was cut mid-stream",
				pair.open, newOpen, newClose))
		}
	}

	// 3. The replacement stops mid-token where the current row ended cleanly.
	//    A completed fragment ends on a closed tag; a cut one ends inside an
	//    attribute, a CSS declaration or a JS literal. This is what separates
	//    the bugs_open/012 writes (ending "'Epic" and "font-weight: bold;")
	//    from the legitimate rewrites above (both ending "</section>").
	if endsCleanly(currentHTML) && !endsCleanly(newHTML) {
		issues = append(issues, fmt.Sprintf(
			"replacement ends mid-token (%q) where the current template ends on a closed tag — the completion was cut mid-stream",
			tailForMessage(newHTML)))
	}

	return issues
}

// endsCleanly reports whether s finishes on a closed tag, ignoring trailing
// whitespace.
func endsCleanly(s string) bool {
	return strings.HasSuffix(strings.TrimSpace(s), ">")
}

// tailForMessage returns the last few characters of s, whitespace-collapsed,
// so a rejection message shows where the generation stopped.
func tailForMessage(s string) string {
	const tailLen = 40
	t := strings.Join(strings.Fields(s), " ")
	if len(t) > tailLen {
		t = t[len(t)-tailLen:]
	}
	return t
}

// ============================================================================
// Rejection recording
// ============================================================================

// actionProvenance identifies which writer a rejection came from. It exists
// because the rejection recorder is shared by more than one write path, and a
// row in agent_error_log that misattributes the writer is worse than no row:
// it sends the next investigation to the wrong file.
type actionProvenance struct {
	AgentType string
	StepName  string
	Action    string
}

// recordComponentWriteRejection persists a structured rejection to
// agent_error_log so blocked writes are queryable across the fleet rather than
// living only in pod logs. Best-effort — the caller's refusal is the real
// outcome; a failure to log must never become a failure to protect.
func recordComponentWriteRejection(
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
	params ActionParams,
	prov actionProvenance,
	errorMessage string,
	errorCode string,
	severity string,
	contextPayload map[string]interface{},
) {
	if db == nil {
		return
	}

	contextJSON, _ := json.Marshal(contextPayload)
	if contextJSON == nil {
		contextJSON = []byte("{}")
	}

	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id")
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain")
	workItemID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id")

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
		siteID,
		domain,
		workItemID,
		params.ExecutionContext.OrchestrationID,
		prov.AgentType,
		params.ExecutionContext.Sender.AgentID,
		params.ExecutionContext.Sender.PodName,
		prov.StepName,
		prov.Action,
		errorMessage,
		errorCode,
		severity,
		string(contextJSON),
	)
	if err != nil {
		logger.Warn("recordComponentWriteRejection: failed to write to agent_error_log",
			zap.Error(err),
			zap.String("error_code", errorCode))
	}
}
