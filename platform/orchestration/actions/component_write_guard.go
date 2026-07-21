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
// replaces. Calibration: 1 block across the recorded transitions — the
// confirmed bugs_open/012 write — and 0 false positives.
//
// If you add or change a check here, RE-RUN THAT SIMULATION FIRST. The evidence
// base grows, and it has already invalidated one threshold within a single day:
// a transition retaining 39% and ending cleanly appeared hours after the
// collapse floor was set at 50%, and would have been refused (see
// componentCollapseRatio). A guard that refuses good work gets switched off,
// and then it protects nothing.
//
// The simulation, for reuse — compare consecutive versions of each component:
//
//	WITH v AS (SELECT component_id, version_number, html_template AS cur,
//	       lead(html_template) OVER (PARTITION BY component_id ORDER BY version_number) AS nxt
//	     FROM component_versions)
//	SELECT c.name, length(cur), length(nxt),
//	       round(100.0*length(nxt)/length(cur)) AS pct,
//	       (right(rtrim(nxt),1)='>') AS ends_cleanly
//	FROM v JOIN content_components c ON c.id=v.component_id
//	WHERE nxt IS NOT NULL AND length(nxt) < length(cur)
//	ORDER BY 1.0*length(nxt)/length(cur) ASC;
//
// A row that shrinks hard AND ends cleanly is a legitimate rewrite; one that
// ends mid-token is the shape this guard exists to refuse.
// ---------------------------------------------------------------------------
//
// Every check is also COMPARATIVE: it fires only when the replacement is worse
// than the row it would replace, never merely because the result is imperfect.
// An absolute quality gate here would block legitimate repairs on exactly the
// components most likely to need them — the already-broken ones.
//
// ---------------------------------------------------------------------------
// Why this is not folded into the existing quality pipeline.
// content_components already carries quality_score / quality_issues /
// quality_checked_at, populated by compute_component_quality.go's
// scoreComponent(), and store_generated_component gates the BIRTH path on it.
// That machinery was read first and deliberately not extended here (the
// question was put by the council gate's reuse seat):
//
//   - scoreComponent is ABSOLUTE and single-artifact — it scores one template
//     against a contract. Every check here needs the row being REPLACED, which
//     that signature has no access to. "Is this component good?" and "is this
//     replacement worse than what it overwrites?" are different questions, and
//     only the second one can be answered without blocking repairs.
//   - Its structural check, TemplateClosed, requires open>0 balanced <section>
//     tags. The component this bug destroyed opens on <style> and has no
//     <section> at all, so TemplateClosed is false for it in BOTH the healthy
//     and the wrecked state — it cannot separate them.
//   - Its remaining checks are schema-shaped (placeholder counts, schema/
//     template sync, stranded fields), which truncation does not disturb in
//     any consistent direction.
//
// So this is not a second opinion on component quality; it is a narrower
// question the quality pipeline does not ask. If scoreComponent ever grows a
// comparative form, fold checks 2 and 3 into it and delete them here.
// ---------------------------------------------------------------------------

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
// Grounded in the live component_versions history, and REVISED once already:
//
//	2026-07-18, 29 transitions: lowest legitimate shrink retained 64%
//	  (provocation-card, 10,300 → 6,618) → floor set at 0.5.
//	2026-07-18, later the same day, 30 transitions: a new transition appeared —
//	  tool-list 11,588 → 4,535, retaining 39% and ending cleanly on
//	  "{{end}}</a></div></div></section>". A complete template, not a
//	  truncation. The 0.5 floor would have REFUSED it.
//
// So the floor is 0.3: below the lowest observed legitimate shrink (39%), above
// the bugs_open/012 wreck (12%, 1,253/10,272). Note this check is the WEAKEST of
// the three and is defence-in-depth only — the 012 wreck is independently caught
// by the mid-token check, and the 012 intermediate write (66% retained) is
// caught by tag balance. So when in doubt bias this floor DOWN: a missed
// collapse is usually caught structurally, whereas a false refusal is not
// caught by anything and teaches people to switch the guard off.
//
// The lesson, for whoever revises it next: this threshold is empirical and the
// evidence base grows. Re-run the simulation in the header before trusting it.
const componentCollapseRatio = 0.3

// balancedPairs are open/close tokens whose balance is checked in the
// replacement. An unterminated <script> is the direct signature of a
// completion cut mid-stream — it is what catches the bugs_open/012
// intermediate write (6,765 chars, 66% retained, comfortably inside the
// legitimate size band, but with <script> left open).
// <div> and <fieldset> were added after the council gate's edit-quality seat
// noted the wrecked artifact was missing those too, so covering only
// script/style/section narrowed the guard's generality for no reason. Simulated
// before adding: both contribute ZERO additional blocks across the recorded
// transitions, so they cost nothing and widen what a mid-stream cut can be
// caught by.
var balancedPairs = []struct{ open, close string }{
	{"<script", "</script>"},
	{"<style", "</style>"},
	{"<section", "</section>"},
	{"<div", "</div>"},
	{"<fieldset", "</fieldset>"},
}

// hasUnbalancedStructuralTags reports whether any structural pair in
// balancedPairs is left open (more opens than closes) in html. This is the
// ABSOLUTE form of the guard's tag-balance signal — no prior row, no size gate,
// no mid-token check — for use at BIRTH writes, where a whole LLM artifact is
// persisted with nothing to compare it against (bugs_open/021 INSTANCE 1).
//
// Calibration (bugs_open/046, 2026-07-21): across every active component
// fleet-wide this 5-pair predicate flagged EXACTLY the 9 known truncation
// casualties and nothing else — 0 over-fire. The companion "ends mid-token"
// signal (endsCleanly) is deliberately NOT folded in here: it is tool-safe but
// adds ~36 false positives fleet-wide on non-tool components that legitimately
// end on text or a closed non-tag. So a general birth gate (store_generated_component's
// section path) uses tag-imbalance ALONE; only the tool path (toolTemplateValid)
// additionally applies endsCleanly, against the tool population it was calibrated
// on.
func hasUnbalancedStructuralTags(html string) bool {
	folded := strings.ToLower(html)
	for _, pair := range balancedPairs {
		if strings.Count(folded, pair.open) > strings.Count(folded, pair.close) {
			return true
		}
	}
	return false
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
