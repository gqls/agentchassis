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
	"fmt"
	"regexp"
	"strings"

	"github.com/gqls/agentchassis/platform/content"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
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

// The structural pair list and its counter live in platform/content
// (content.StructuralTagPairs / content.StructuralTagCounts), the leaf package
// both this package and discovery_checks can import — which retired the
// hand-maintained mirror in check_truncated_component.go. An unterminated
// <script> is the direct signature of a completion cut mid-stream — it is what
// catches the bugs_open/012 intermediate write (6,765 chars, 66% retained,
// comfortably inside the legitimate size band, but with <script> left open).
// <div> and <fieldset> were added after the council gate's edit-quality seat
// noted the wrecked artifact was missing those too — simulated before adding,
// zero additional blocks across the recorded transitions.
//
// Since bugs_open/303 the counting is MARKUP-CONTEXT, not substring: mentions
// of a tag inside JS/CSS bodies, comments or regexes do not count, so a tool
// that manipulates HTML can exist. The markup_balance.go header carries the
// exact semantics; the 2026-08-18 recalibration against the full live
// population is recorded there and in bugs_open/303.

// hasUnbalancedStructuralTags reports whether any structural pair is left open
// (more opens than closes) in html, counting only markup-context tags. This is
// the ABSOLUTE form of the guard's tag-balance signal — no prior row, no size
// gate, no mid-token check — for use at BIRTH writes, where a whole LLM
// artifact is persisted with nothing to compare it against (bugs_open/021
// INSTANCE 1).
//
// Calibration (bugs_open/046, 2026-07-21; re-run for bugs_open/303,
// 2026-08-18): across every active component fleet-wide the 5-pair predicate
// flags exactly the known truncation casualties. The companion "ends
// mid-token" signal (endsCleanly) is deliberately NOT folded in here: it is
// tool-safe but adds ~36 false positives fleet-wide on non-tool components
// that legitimately end on text or a closed non-tag. So a general birth gate
// (store_generated_component's section path) uses tag-imbalance ALONE; only
// the tool path (toolTemplateValid) additionally applies endsCleanly, against
// the tool population it was calibrated on.
func hasUnbalancedStructuralTags(html string) bool {
	return len(content.UnbalancedStructuralTags(html)) > 0
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

	// 2. Unterminated tags in the replacement, but only where the current row
	//    was balanced. Comparative on purpose: if the component is ALREADY
	//    unbalanced, blocking here would trap it permanently — no rewrite
	//    could ever land to repair it. Markup-context counts (bugs_open/303):
	//    a fix that ADDS a comment or regex mentioning a tag must not read as
	//    a truncation — the rewrite path was the bug's worst case, because
	//    acting on a reported defect is when precise tag mentions appear.
	curCounts := content.StructuralTagCounts(currentHTML)
	newCounts := content.StructuralTagCounts(newHTML)
	for k := range newCounts {
		if newCounts[k].Opens <= newCounts[k].Closes {
			continue
		}
		if curCounts[k].Opens <= curCounts[k].Closes {
			issues = append(issues, fmt.Sprintf(
				"replacement leaves %s unterminated (%d open vs %d close, counting markup only) where the current template is balanced — the completion was cut mid-stream",
				newCounts[k].Open, newCounts[k].Opens, newCounts[k].Closes))
		}
	}

	// 3. The replacement stops mid-token where the current row ended cleanly.
	//    A completed fragment ends on a closed tag; a cut one ends inside an
	//    attribute, a CSS declaration or a JS literal. This is what separates
	//    the bugs_open/012 writes (ending "'Epic" and "font-weight: bold;")
	//    from the legitimate rewrites above (both ending "</section>").
	// 3a. The replacement INTRODUCES a fabricated business fact (bugs_open/140,
	//     RFC_009 option B). Comparative like everything else here: a template that
	//     already fabricates must stay repairable, so only a NEW fabrication is
	//     refused. Full writer census and the reasoning in
	//     component_fallback_guard.go's fabricatedFallbackRegression.
	if issue := fabricatedFallbackRegression(currentHTML, newHTML); issue != "" {
		issues = append(issues, issue)
	}

	// 4. Script elements KEPT but their bodies STUBBED (council round 8,
	//    bugs_open/324's judged-gate control generalised): an LLM whole-template
	//    rewrite can keep every <script> tag and return near-empty bodies — the
	//    page keeps its markup, loses its program, and no check above sees it
	//    (the control retained 58% overall and passed this guard clean).
	//    Deliberately NARROWER than "the script shrank": removing the script
	//    ELEMENT entirely is the legitimate redesign/extraction shape (6 such
	//    transitions in history: provocation-card and tool-list dropping JS,
	//    the js-extraction pattern moving a body behind src=), and a shrunken-
	//    but-substantial script is a rework (tool-arena-interface v3→v4 kept a
	//    7KB clean program at 30% of its predecessor). The stub shape is: the
	//    current row has ≥1000 bytes of inline (non-src) script, the
	//    replacement still carries inline script elements, and EVERY one of
	//    their bodies is under 200 bytes.
	//    Calibration (2026-08-19, per this file's own rule): simulated over all
	//    235 distinct consecutive component_versions transitions + current
	//    rows — 0 legitimate transitions match; the gutted-script control does.
	if stubbedScriptIssue := scriptStubRegression(currentHTML, newHTML); stubbedScriptIssue != "" {
		issues = append(issues, stubbedScriptIssue)
	}

	if endsCleanly(currentHTML) && !endsCleanly(newHTML) {
		issues = append(issues, fmt.Sprintf(
			"replacement ends mid-token (%q) where the current template ends on a closed tag — the completion was cut mid-stream",
			tailForMessage(newHTML)))
	}

	return issues
}

// reInlineScriptElem captures (open tag, body) of each script element so src=
// elements can be excluded — an empty body behind src= is the js-extraction
// pattern, not a stub.
var reInlineScriptElem = regexp.MustCompile(`(?is)(<script\b[^>]*>)(.*?)</script>`)

// scriptStubRegression implements check 4 above. Pure, like everything here.
func scriptStubRegression(currentHTML, newHTML string) string {
	if len(newHTML) > len(currentHTML) {
		return "" // truncation cannot grow an artifact — same exemption as checks 2/3
	}
	inlineBodies := func(t string) []string {
		var out []string
		for _, m := range reInlineScriptElem.FindAllStringSubmatch(t, -1) {
			if reScriptSrcAttr.MatchString(m[1]) {
				continue
			}
			out = append(out, m[2])
		}
		return out
	}
	curMass := 0
	for _, b := range inlineBodies(currentHTML) {
		curMass += len(b)
	}
	if curMass < 1000 {
		return ""
	}
	newBodies := inlineBodies(newHTML)
	if len(newBodies) == 0 {
		return "" // element removed or extracted — the deliberate class
	}
	maxLen := 0
	for _, b := range newBodies {
		if l := len(strings.TrimSpace(b)); l > maxLen {
			maxLen = l
		}
	}
	if maxLen >= 200 {
		return ""
	}
	return fmt.Sprintf(
		"replacement keeps %d inline script element(s) but every body is a stub (largest %d bytes) where the current template carries %d bytes of inline script — the program was deleted while the markup survived",
		len(newBodies), maxLen, curMass)
}

var reScriptSrcAttr = regexp.MustCompile(`(?i)\ssrc\s*=`)

// reTrailingTemplateEnd matches ONE trailing `{{end}}` action, tolerating
// whitespace and Go's trim markers (`{{- end -}}`). Case-sensitive on purpose:
// `end` is a template keyword and has no case variants, so a case-insensitive
// match would widen the tolerance for nothing.
var reTrailingTemplateEnd = regexp.MustCompile(`\s*\{\{-?\s*end\s*-?\}\}\s*$`)

// endsCleanly reports whether s finishes on a closed tag, ignoring trailing
// whitespace and any trailing `{{end}}` actions.
//
// ── WHY {{end}} IS STRIPPED FIRST (bugs_open/351) ──────────────────────────
//
// A component whose whole body is wrapped in a conditional or a range
// legitimately ends on `</section>{{end}}` or `</script>{{end}}` — the markup
// closed, then the template action closed. The original rule required the LAST
// character to be `>`, so it called those templates truncated. That is a false
// positive with real consequences at both call sites: sectionTemplateValid
// dropped the component silently at load, and the write-time regression check
// below refused the replacement at BIRTH ("ends mid-token"), so a generator
// converting a section to a conditional wrapper could not save its work.
//
// Measured over every active section- and tool-level template (150 + 124
// respectively, as of 2026-08-22): exactly TWO rows change verdict, and both
// were hand-checked as complete and correctly conditional-wrapped —
// `about-commercial-block` (`</section>{{end}}`) and `case-studies-grid`
// (`</script>{{end}}`, re-wrapped that morning, i.e. the shape is being
// authored NOW, not a legacy artefact).
//
// ── WHY NOT SIMPLY ACCEPT A TRAILING `}}` ─────────────────────────────────
//
// Because a template cut immediately after any complete mid-template action
// ALSO ends `}}`, and a suffix rule would pass that truncation — the exact
// failure this guard exists to catch. Stripping only complete `{{end}}` tokens
// and then re-testing for `>` refuses it: after the strip, a mid-cut remainder
// ends on prose or an open tag. `{{end}}` is the only action that legitimately
// terminates a template; a tail of `{{if …}}`, `{{range …}}` or a bare
// placeholder is suspicious in every case and stays suspicious here.
//
// The strip repeats because wrappers nest (`{{if}}{{range}}…{{end}}{{end}}`).
// No row in the 2026-08-22 corpus needs more than one pass; the loop is
// defensive, not demand-proven.
func endsCleanly(s string) bool {
	t := strings.TrimSpace(s)
	for {
		stripped := reTrailingTemplateEnd.ReplaceAllString(t, "")
		if stripped == t {
			break
		}
		t = strings.TrimSpace(stripped)
	}
	return strings.HasSuffix(t, ">")
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
// Shared-component fence (bugs_open/281)
// ============================================================================

// WRITERS OF content_components.html_template, enumerated 2026-08-15 (grep
// "UPDATE content_components" over platform/ internal/ pkg/, non-test):
//
//   update_component_html_action.go      tool-improver's writeback. Subject is a
//                                        TOOL — i.e. one page's finding — so a
//                                        shared component here is always wrong.
//                                        Calls sharedComponentWriteCheck. ← fenced
//   fix_component_template_action.go     reads a page's rendered_html to
//                                        DIAGNOSE, but all THREE of its
//                                        html_template writes take the component
//                                        as subject: repair_template_slots
//                                        (mechanical slot repair keyed by
//                                        spec.component_id), chrome_overflow_fix
//                                        (CSS append to a chrome template), and
//                                        scope_component_instance (2026-08-17,
//                                        bugs_open/283 — fan-out to every
//                                        placement is the point, and the write
//                                        is gate-refused when the script is
//                                        unscoped). Not a per-page LLM rewrite;
//                                        not fenced. (Corrected 2026-08-16 — an
//                                        earlier version of this census called
//                                        it "page-aware, open".)
//   fix_harcoded_colours_action.go       component-scoped subjects: the fix IS
//   fix_forced_text_colours_action.go    meant to reach every placement.
//   fix_nav_link_templates_action.go     Not fenced.
//   store_generated_component_action.go  component-creator regen; the component
//                                        is the subject. Not fenced.
//   internal/core-manager/admin/…        human-driven admin. Not fenced.
//
// A new writer that takes a PAGE-scoped finding must call the check below —
// and since 2026-08-16 something DOES stop it: component_template_writer_coverage_test.go
// fails the build for any html_template rewriter in this package that neither
// calls sharedComponentWriteCheck nor is declared fan-out-intended with a cited
// reason (council d8668e1f). This header is the narrative; that test is the census.

// sharedComponentWriteCheck is the fence's decision, separated from the
// action so any writer of html_template can ask it. Refuse when the component
// is not a tool fork (component_level <> 'tool') AND is placed on more than one
// page, unless the caller passes allow=true (a HUMAN-authored step-config
// opt-in, never LLM output). Tool-level forks are never refused: a per-site
// fork on two pages is the established shape (tool-llm-cost-calculator, 2
// sites, five successful rewrites) — for those the caller gets the counts back
// to WARN on. The census is fail-closed ONLY on the non-tool path: a census
// error for a tool fork is returned as err with refuse=false, so the common
// path gains no new failure mode.
type sharedComponentVerdict struct {
	Refuse         bool
	PlacementPages int
	PlacementSites int
}

func sharedComponentWriteCheck(ctx context.Context, q dbQueryer, componentID interface{}, componentLevel string, allow bool) (sharedComponentVerdict, error) {
	var v sharedComponentVerdict
	err := q.QueryRowContext(ctx, `
		SELECT count(DISTINCT pc.page_id), count(DISTINCT p.site_id)
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.component_id = $1
	`, componentID).Scan(&v.PlacementPages, &v.PlacementSites)
	if err != nil {
		if componentLevel != "tool" {
			// A fence that cannot look must not wave a non-tool write through.
			return v, fmt.Errorf("shared-component fence: placement census failed: %w", err)
		}
		return v, err // tool fork: caller logs and proceeds
	}
	v.Refuse = v.PlacementPages > 1 && componentLevel != "tool" && !allow
	return v, nil
}

// dbQueryer is the one method the fence needs; *sql.DB and *sql.Tx both satisfy it.
type dbQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
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
// The `db` parameter this used to take was dropped when the write moved to the
// shared helper (RFC_012 B): all three callers passed params.DB, and a
// signature that lets a caller name a DIFFERENT handle than the one actually
// written to is the kind of seam that reads correct and is not.
func recordComponentWriteRejection(
	ctx context.Context,
	logger *zap.Logger,
	params ActionParams,
	prov actionProvenance,
	errorMessage string,
	errorCode string,
	severity string,
	contextPayload map[string]interface{},
) {
	// prov is set explicitly on the entry and never inherited — this recorder is
	// shared by three write paths, and a row that misattributes the writer sends
	// the next investigation to the wrong file (see the type doc above).
	LogActionEntry(ctx, params, agenterrors.Entry{
		SiteID:       datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id"),
		Domain:       datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain"),
		WorkItemID:   datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id"),
		AgentType:    prov.AgentType,
		StepName:     prov.StepName,
		Action:       prov.Action,
		ErrorMessage: errorMessage,
		ErrorCode:    errorCode,
		Severity:     severity,
		Context:      contextPayload,
	}, logger)
}
