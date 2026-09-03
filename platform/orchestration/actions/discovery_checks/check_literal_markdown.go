// FILE: platform/orchestration/actions/discovery_checks/check_literal_markdown.go
//
// CHANGES:
//   - Per-page ROUTING between two handlers (was: every item to page-rerender),
//     2026-08-20, bugs_open/277 §5 (owner ruling, same day). page-rerender's
//     repair regenerates rendered_html from content_data — which is INAPPLICABLE
//     BY CONSTRUCTION to a component whose content_data cannot reproduce its own
//     rendered_html (the worked case: Ported Page, 100 of 115 instances hold
//     NONE of their template's fields; all 100 sit on owned pages, so the
//     owned-page guard took the blame for three wrongly-costed routes — see the
//     LANDMINES entry). Migration 499 made "read the finding's source, then ask
//     whether content_data can reproduce rendered_html" the HUMAN routing test;
//     transformRouteSlot below is that test automated at filing time. A page
//     routes to section-editor (edit_type rendered_html_transform, transform
//     code_span_to_code_tag — datahelpers.ConvertLiteralCodeSpansInHTML, opt-in
//     via migration 513) IFF every finding is source=rendered_html AND
//     pattern=code_span AND all sit on ONE slot occurring ONCE on the page AND
//     that slot's component cannot regenerate
//     (datahelpers.ContentDataCanFillTemplate). Everything else keeps the
//     page-rerender route unchanged. item_type, item_key and the whole-page
//     verifier are untouched — the verifier is keyed on item_type and its remit
//     (zero findings on either surface, whole page) is exactly what the
//     transform must achieve for the item to complete. NB the new pair
//     (literal_markdown → section-editor) starts with zero lifetime completes,
//     so the 444 promoter holds it until one canary run completes — that canary
//     is deliberate, not a bug (bugs_open/277 records the bootstrap).
//   - HandlerAgent: "page-rerender" (was "page-build-handler"), 2026-08-18,
//     bugs_open/184 fix-2. The page-build-handler pair was 1 complete / 28 failed
//     lifetime — the worst in the fleet, held by the migration-444 promoter floor —
//     because repair-by-LLM-regeneration was tried and the regenerating writer has
//     the same habit: proven at the artefact 2026-08-07, a full regeneration wrote
//     18 markdown findings back into the very field it was dispatched to clean,
//     three days AFTER the prompt-rule hardening (304) was verified live. The
//     repair is now mechanical: the item spec carries reason:"literal_markdown",
//     page-rerender's check_rerender_mode accepts it (migration 473, the
//     cta_links_stale precedent), and the rerender strips markdown from stored
//     content_data (datahelpers.StripLiteralMarkdown) before re-rendering — no
//     LLM in the loop, so the repair cannot reintroduce the defect. The verifier
//     below is keyed on item_type and survives the re-route; its whole-page remit
//     still matches (rerender_sections rewrites every unlocked section). The old
//     pair stays held — no 444 rollback; it simply stops receiving items.
//   - md_link pattern added, 2026-08-18, same fix. The live symptom widened past
//     the filing: 9 md-link components fleet-wide and "## [title](url)" composites
//     in open items. Patterns are now single-sourced in
//     datahelpers/literal_markdown.go so this check, the verifier and the stripper
//     cannot drift (property test: scan(strip(x)) == nothing).
//   - HandlerAgent: "page-build-handler" (was "page-content-writer"), 2026-08-05,
//     bugs_open/201. page-content-writer must not be dispatched DIRECTLY: it plans
//     its own sections from `input_data.current_page.sections` (bugs_closed/087's
//     self-plan branch), and a discovery item's spec has no `sections` key at all —
//     measured 2026-08-05, all 14 page-content-writer items fleet-wide carry only
//     {check, findings, fix, original_pipeline, page_id, page_name, page_url}. So
//     plan_sections early-returns `ready_count: 0, reason: "no sections to plan"`
//     (plan_sections_action.go:867-875) and the run hard-fails at
//     fail_no_ready_sections — 11 of 11 attempts, 2026-08-04, before writing
//     anything. page-build-handler instead sources sections from
//     site_specs.site_plan via load_page_sections_from_spec (authoritative, with a
//     pages.sections fallback), so it never depends on the caller's spec shape.
//     It is proven on ALREADY-BUILT pages: content_rewrite 19 complete /
//     empty_section 12 complete, measured 2026-08-05. Same migration
//     check_empty_sections.go already made.
//
// bugs_open/184 (llm_markdown_reaches_the_page_as_literal_asterisks): LLM
// content writers sometimes emit markdown syntax (**bold**, `code spans`,
// # headings) into content_data fields whose schema type is `text`. The render
// path (RenderComponentAction -> RenderTemplate -> text/template) does no
// markdown interpretation and no HTML escaping, so the syntax reaches the
// visitor verbatim. Three live rows on three unrelated sites when filed
// (2026-08-03: mortgagecalculator.co.uk hero, gaswholesalers.com pricing,
// webdesign.co.uk news-listing), silent to every existing check.
//
// TWO SURFACES, per check_unverified_claims (bugs_open/093): stored
// content_data AND rendered_html. content_data is the durable copy — the
// no-LLM rerender path (rerender_page_sections_action.go) republishes it with
// no writer in the loop, so a poisoned row reproduces on every future
// rerender. The row predicate is an OR: content with no rendered_html yet is
// exactly what the next rerender WILL publish, unchecked.
//
// FALSE-POSITIVE DISCIPLINE (the filer's own letter-guard, carried through):
//
//	bold      \*\*[A-Za-z]...  — `3 * 4 = 12`, `2**10`, `a ** b` never fire
//	code span `[A-Za-z0-9]...  — `${x}` interpolations and `/api` paths never
//	          fire; and on content_data the pattern is only applied to values
//	          carrying no HTML markup (a markup-bearing value is not a
//	          text-typed field; backticks there are code, not prose)
//	heading   ^#{1,6} \S       — #fff, "#1 rated", href="#", "issue #12"
//	          never fire
//
// rendered_html is scanned via datahelpers.ExtractAssertionText — text nodes
// only, script/style/code subtrees skipped — so JS template literals inside
// inline scripts are structurally invisible to the backtick pattern.
//
// Routing: a definite mechanical defect (markdown syntax is never correct in
// a plain-text field), so it routes for auto-repair rather than HITL.
// ⚠ CORRECTED 2026-08-05 (bugs_open/201): this paragraph used to read "routes at
// page-content-writer … the check_placeholder_contact precedent". BOTH halves were
// wrong. The writer cannot be dispatched directly (it self-plans from a `sections`
// key this check's spec does not have, and hard-fails), and the cited precedent had
// NEVER RUN — placeholder_contact's items had reached neither complete nor failed.
// It now routes at page-build-handler; see the CHANGES block at the top. One item per
// page, item_key 'literal_markdown:<page_id>'. Producer set for this
// item_type: this check alone (owner ruling 2026-08-02 point 1 — stated in
// the concept-register entry, CQ-019).
//
// Retraction (RFC_010): follows check_required_fields_missing's estabished
// shape, not a hand-rolled status filter — resolveWorkItems alone owns the
// open/closed predicate (workItemClosedStatuses). This function queries EVERY
// site_work_items row of this item_type regardless of status (no `status NOT
// IN (...)` here — a second, drifting copy of that list in a function whose
// job is CLOSING items is exactly the hazard required_fields_missing's header
// warns about) and only emits a ResolvedFinding for a page that was
// POSITIVELY re-scanned this run (appears in this run's row set) and now
// carries zero findings on either surface. A page absent from this run's
// results (e.g. deleted, or the query missed it) is never resolved.
//
// Locked components (locked_at IS NOT NULL) skipped by precedent.
// Registration: automatic via init(). Enable by adding "literal_markdown" to
// a discovery agent's checks array (quality-discovery-agent, migration 302).

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() {
	Register(&LiteralMarkdownCheck{})
	// bugs_open/201 SYMPTOM 2: without this, a handler saga that reports success
	// having written nothing gets its item stamped 'complete' on its own word.
	// Live instance, gaswholesalers.com 2026-08-06: the existing literal_markdown
	// item read status='complete' while the markdown was still in content_data on
	// the page it named.
	RegisterVerifier("literal_markdown", VerifyLiteralMarkdownResolved)
}

type LiteralMarkdownCheck struct{}

func (c *LiteralMarkdownCheck) Name() string { return "literal_markdown" }

// Patterns are single-sourced in datahelpers/literal_markdown.go (bug 184
// fix-2, 2026-08-18) so this check, its completion verifier and the write-seam
// stripper (StripLiteralMarkdown) cannot drift: the repair contract is
// scan(strip(x)) == nothing, enforced by datahelpers' property test. The
// markup-suppression heuristic (a value carrying markup is not a text-typed
// field) moved with them: datahelpers.HTMLMarkupRe.
var htmlMarkupRe = datahelpers.HTMLMarkupRe

// maxLiteralMarkdownSpecFindings bounds the finding list carried in the work
// item's spec. The count is always exact (spec.findings_total); only the
// examples are capped — check_asset_reference_404.go's maxEmptyRefSamples shape.
// 25 rather than that check's 5 because a repairing agent reads these to
// understand the page, where the 404 check's samples are illustrative only.
const maxLiteralMarkdownSpecFindings = 25

// capSpecFindings bounds what the spec carries. Extracted so the cap has a
// direct test — the same reason create_work_item_action.go:468 extracted its
// gate. It must NEVER be used to decide anything: routing reads the full slice.
func capSpecFindings(findings []literalMarkdownFinding, pageName string, logger *zap.Logger) []literalMarkdownFinding {
	if len(findings) <= maxLiteralMarkdownSpecFindings {
		return findings
	}
	if logger != nil {
		logger.Info("literal_markdown: spec findings capped",
			zap.String("page", pageName),
			zap.Int("found", len(findings)),
			zap.Int("carried", maxLiteralMarkdownSpecFindings))
	}
	return findings[:maxLiteralMarkdownSpecFindings]
}

type literalMarkdownFinding struct {
	SlotName string `json:"slot_name"`
	Field    string `json:"field,omitempty"` // content_data path; empty for rendered_html
	Pattern  string `json:"pattern"`         // the LiteralMarkdownPatterns names — read them there, not here
	Matched  string `json:"matched"`
	Source   string `json:"source"` // content_data | rendered_html
}

// scanPlainTextMarkdown applies the guarded patterns to one plain-text
// string. One finding per pattern per value — the repair is a page rewrite,
// so occurrence counts change nothing about what to do. The pattern set is
// datahelpers.LiteralMarkdownPatterns (includeCodeSpan gates code_span AND
// md_link, the two patterns suppressed on markup-bearing values).
func scanPlainTextMarkdown(text, slot, field, source string, includeCodeSpan bool) []literalMarkdownFinding {
	var out []literalMarkdownFinding
	for _, pm := range datahelpers.LiteralMarkdownPatterns(text, includeCodeSpan) {
		matched := pm[1]
		if len(matched) > 120 {
			matched = matched[:120]
		}
		out = append(out, literalMarkdownFinding{
			SlotName: slot, Field: field, Pattern: pm[0],
			Matched: matched, Source: source,
		})
	}
	return out
}

// walkContentDataStrings visits every string leaf of a decoded content_data
// object. Keys beginning "_" are platform metadata (_built_at, ...), never
// writer output — skipped.
func walkContentDataStrings(prefix string, v interface{}, visit func(path, s string)) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if strings.HasPrefix(k, "_") {
				continue
			}
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			walkContentDataStrings(p, val, visit)
		}
	case []interface{}:
		for i, val := range t {
			walkContentDataStrings(fmt.Sprintf("%s[%d]", prefix, i), val, visit)
		}
	case string:
		visit(prefix, t)
	}
}

type pageMarkdownFindings struct {
	PageID, PageName, PageURL string
	Findings                  []literalMarkdownFinding
}

// slotRepro is what the transform route needs to know about one slot of one
// page: how many component rows share the slot name (the section-editor
// resolves its target by page_name + slot_name, so a duplicated slot is an
// ambiguous target and refuses the route), and whether the component could
// regenerate its rendered_html from content_data (if it can, the
// regenerate-from-source route stays correct and this route must not fire).
type slotRepro struct {
	occurrences   int
	canRegenerate bool
}

// transformRouteSlot is migration 499's routing test, automated: it returns
// the single slot on which a page's findings are repairable by
// section-editor's code_span_to_code_tag rendered_html transform, or ok=false
// to keep the default page-rerender route. ALL conditions must hold — each
// refusal direction lands on today's behaviour, so a wrong "false" costs an
// escalation to a human (the status quo) while a wrong "true" would aim an
// HTML-surface edit at a component whose content_data still carries the
// defect, to be reprinted by its next regeneration. Conservative by
// construction:
//   - every finding source=rendered_html (a content_data finding NEEDS the
//     regenerate route);
//   - every finding pattern=code_span (the only transform that exists —
//     bold/heading/md_link on this surface still go to a human);
//   - every finding on ONE named slot, occurring exactly once on the page;
//   - that slot's component CANNOT regenerate (ContentDataCanFillTemplate
//     false — template names fields, content_data holds none of them).
func transformRouteSlot(findings []literalMarkdownFinding, slots map[string]*slotRepro) (string, bool) {
	if len(findings) == 0 {
		return "", false
	}
	slot := findings[0].SlotName
	if slot == "" {
		return "", false
	}
	for _, f := range findings {
		if f.Source != "rendered_html" || f.Pattern != "code_span" || f.SlotName != slot {
			return "", false
		}
	}
	sr := slots[slot]
	if sr == nil || sr.occurrences != 1 || sr.canRegenerate {
		return "", false
	}
	return slot, true
}

// scanComponentRowForMarkdown applies this check's predicate to ONE
// page_components row, across both surfaces.
//
// Extracted 2026-08-06 (bugs_open/201 symptom 2) so the detector above and
// VerifyLiteralMarkdownResolved below CANNOT DRIFT. verifiers.go's contract is that a
// verifier re-runs "the SAME predicate the discovery check used to create the item";
// two hand-kept copies of a five-line scan is exactly how that stops being true, and a
// verifier that has drifted STRICTER strands correctly-handled items in 'failed' (the
// page_rerender cautionary tale in verifier_coverage_test.go's own header).
//
// The row SELECTION is still written twice — site-wide in Run, page-scoped in the
// verifier — because they genuinely differ in scope. Their WHERE clauses are otherwise
// identical and must stay so: `locked_at IS NULL` plus the both-surfaces-empty exclusion.
func scanComponentRowForMarkdown(slot, html, contentJSON string, logger *zap.Logger) []literalMarkdownFinding {
	var fs []literalMarkdownFinding
	if cd := decodeContentData(contentJSON, logger); cd != nil {
		walkContentDataStrings("", cd, func(path, s string) {
			fs = append(fs, scanPlainTextMarkdown(
				s, slot, path, "content_data", !htmlMarkupRe.MatchString(s))...)
		})
	}
	for _, block := range datahelpers.ExtractAssertionText(html) {
		fs = append(fs, scanPlainTextMarkdown(block, slot, "", "rendered_html", true)...)
	}
	return fs
}

// VerifyLiteralMarkdownResolved is the completion-time guard for bugs_open/201
// SYMPTOM 2: a handler saga can report success having written nothing, and
// complete_work_item then stamps the item 'complete' on the saga's own word.
// Proven live on gaswholesalers.com — its literal_markdown item read 'complete'
// while the markdown was still in content_data on the page it named.
//
// SCOPE IS WHOLE-PAGE, AND THAT IS DELIBERATE. The remit test (verifier_coverage_test.go's
// page_rerender entry) is that a verifier must not be STRICTER than the handler it judges.
// This item's handler is page-build-handler (re-pointed by this same bug's fix-1), whose
// build_pages_loop plans the page's sections from the site spec and rewrites all of them —
// so "no literal markdown anywhere on this page" is exactly its remit, not more.
//
// FAILING VERIFICATION IS THE SAFE DIRECTION HERE, unlike page_rerender. A failed verify
// routes the item into the attempt machinery and, after max_attempts, to human review —
// which is the right destination for a markdown defect the writer did not clear. What we
// must never do is stamp 'complete' over a defect that is still being served.
func VerifyLiteralMarkdownResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	if target.PageID == nil {
		return VerifyResult{}, fmt.Errorf("literal_markdown verifier: item carries no page_id")
	}

	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(pc.slot_name, ''),
		       COALESCE(pc.rendered_html, ''), COALESCE(pc.content_data::text, '')
		FROM page_components pc
		WHERE pc.page_id = $1
		  AND pc.locked_at IS NULL
		  AND ( (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '')
		     OR (pc.content_data IS NOT NULL AND pc.content_data::text <> '{}') )
		ORDER BY pc.position
	`, *target.PageID)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("literal_markdown verifier: page load failed: %w", err)
	}
	defer rows.Close()

	scanned := 0
	var remaining []literalMarkdownFinding
	for rows.Next() {
		var slot, html, contentJSON string
		if err := rows.Scan(&slot, &html, &contentJSON); err != nil {
			return VerifyResult{}, fmt.Errorf("literal_markdown verifier: scan failed: %w", err)
		}
		scanned++
		remaining = append(remaining, scanComponentRowForMarkdown(slot, html, contentJSON, logger)...)
	}
	if err := rows.Err(); err != nil {
		return VerifyResult{}, fmt.Errorf("literal_markdown verifier: iteration failed: %w", err)
	}

	// A ZERO HERE HAS TWO CAUSES AND ONLY ONE OF THEM IS "FIXED".
	// No scannable rows means either the page was repaired into emptiness or its
	// components were lost (bugs_closed/194's damage class — 31 of 106 components on
	// one live site carry NULL content_data). Both present as "no markdown found", and
	// certifying the second would stamp 'complete' over a destroyed page.
	//
	// ⚠ RETURN Resolved:false, NOT AN ERROR. The first version of this returned an
	// error, reasoning that "could not verify" is the honest answer. It is — and it is
	// also USELESS HERE, because the registry's documented policy is to FAIL OPEN on a
	// verifier error (verifiers.go:60-63): CompleteWorkItemAction records the error and
	// stamps 'complete' anyway. So the error branch delivered exactly the outcome this
	// verifier exists to prevent, on the one input where the ambiguous case IS content
	// loss — bugs_closed/032's shape ("verifier reads a deleted target as a successful
	// fix"), and bugs_open/201's own symptom 2 reproduced through its new guard.
	// Caught by the council's bug_historian seat (gating, HIGH,
	// corr f14a8b64-4f71-4915-88d0-9587db845052). Noting the fail-open in prose, as the
	// first submission did, is not a control on it.
	//
	// Resolved:false blocks completion and routes the item into the attempt machinery
	// and then human review, which is the correct destination for a page that cannot be
	// shown to be repaired. An unverifiable page is not a repaired page.
	if scanned == 0 {
		return VerifyResult{
			Resolved: false,
			Detail: fmt.Sprintf("page %s has no scannable components — cannot distinguish "+
				"'repaired' from 'content lost' (bugs_closed/194's class), so completion is refused "+
				"rather than taken on trust", target.PageID),
		}, nil
	}

	if len(remaining) == 0 {
		return VerifyResult{
			Resolved: true,
			Detail: fmt.Sprintf("no literal markdown on either surface across %d component(s) of page %s",
				scanned, target.PageID),
		}, nil
	}

	first := remaining[0]
	return VerifyResult{
		Resolved: false,
		Detail: fmt.Sprintf("%d finding(s) still present across %d component(s); first: slot %q field %q pattern %s in %s — %q",
			len(remaining), scanned, first.SlotName, first.Field, first.Pattern, first.Source, first.Matched),
	}, nil
}

func (c *LiteralMarkdownCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.page_id::text, p.name, COALESCE(p.url, ''),
		       COALESCE(pc.slot_name, ''),
		       COALESCE(pc.rendered_html, ''), COALESCE(pc.content_data::text, ''),
		       COALESCE(cc.html_template, '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		LEFT JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND pc.locked_at IS NULL
		  AND ( (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '')
		     OR (pc.content_data IS NOT NULL AND pc.content_data::text <> '{}') )
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("literal_markdown query failed: %w", err)
	}
	defer rows.Close()

	byPage := map[string]*pageMarkdownFindings{}
	var pageOrder []string
	scannedPages := map[string]bool{} // pages POSITIVELY re-scanned this run — the retraction gate
	// Slot metadata for the transform route (2026-08-20): tracked for EVERY
	// scanned row, not only finding-bearing ones, because occurrence counting
	// must see a slot's duplicates wherever they sit in the page.
	slotsByPage := map[string]map[string]*slotRepro{}

	for rows.Next() {
		var pageID, pageName, pageURL, slot, html, contentJSON, htmlTemplate string
		if err := rows.Scan(&pageID, &pageName, &pageURL, &slot, &html, &contentJSON, &htmlTemplate); err != nil {
			dctx.Logger.Warn("literal_markdown: scan error", zap.Error(err))
			continue
		}
		scannedPages[pageID] = true

		if slotsByPage[pageID] == nil {
			slotsByPage[pageID] = map[string]*slotRepro{}
		}
		if sr, ok := slotsByPage[pageID][slot]; ok {
			sr.occurrences++
		} else {
			slotsByPage[pageID][slot] = &slotRepro{
				occurrences:   1,
				canRegenerate: datahelpers.ContentDataCanFillTemplate(htmlTemplate, decodeContentData(contentJSON, dctx.Logger)),
			}
		}

		fs := scanComponentRowForMarkdown(slot, html, contentJSON, dctx.Logger)
		if len(fs) == 0 {
			continue
		}
		pf, ok := byPage[pageID]
		if !ok {
			pf = &pageMarkdownFindings{PageID: pageID, PageName: pageName, PageURL: pageURL}
			byPage[pageID] = pf
			pageOrder = append(pageOrder, pageID)
		}
		pf.Findings = append(pf.Findings, fs...)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("literal_markdown iteration failed: %w", err)
	}

	result := &CheckResult{}
	if len(pageOrder) > 0 {
		total := 0
		for _, id := range pageOrder {
			total += len(byPage[id].Findings)
		}
		result.Findings = []map[string]interface{}{{
			"check": "literal_markdown", "count": total, "pages": len(pageOrder),
		}}
	}

	for _, id := range pageOrder {
		pf := byPage[id]
		// CAP THE SPEC, NOT THE DECISION (bugs_open/332, 2026-09-03). This check
		// was the only one of its family writing an unbounded finding list into
		// the spec — check_asset_reference_404.go:204-206, check_structure_floor,
		// check_componentless_pages and check_page_content_divergence all bound
		// theirs. A page whose whole news listing is dirty produces one finding
		// per value, and 332 widens the pattern set, so the cap lands first.
		//
		// Two properties, both load-bearing:
		//   - findings_total is ALWAYS exact; only the examples are capped. A
		//     silent cap is bugs_open/181's shape — the reader cannot tell a
		//     capped list from a complete one.
		//   - transformRouteSlot below is called on the FULL slice. Routing is a
		//     judgement about EVERY finding on the page ("all code_span, one
		//     slot"); routing on a capped view would let finding 26 change the
		//     right answer and never be seen.
		carried := capSpecFindings(pf.Findings, pf.PageName, dctx.Logger)
		spec := map[string]interface{}{
			"check":          "literal_markdown",
			"reason":         "literal_markdown", // page-rerender gates its sections branch on this (migration 473)
			"page_id":        pf.PageID,
			"page_name":      pf.PageName,
			"page_url":       pf.PageURL,
			"findings":       carried,
			"findings_total": len(pf.Findings),
			"fix": "A literal_markdown rerender strips markdown markers from plain-text " +
				"content_data fields deterministically (datahelpers.StripLiteralMarkdown, " +
				"gated by strip_literal_markdown on page-rerender's rerender_sections step) " +
				"and re-renders the page from the cleaned values — no LLM in the loop. " +
				"Both surfaces heal in one pass: content_data is stripped before it feeds " +
				"the render, and rendered_html is regenerated from it.",
		}
		// Route per page — see the header CHANGES entry (2026-08-20) and
		// transformRouteSlot's own comment for the conditions.
		handlerAgent := "page-rerender" // mechanical strip-on-rerender — see header CHANGES, 2026-08-18
		if slot, ok := transformRouteSlot(pf.Findings, slotsByPage[pf.PageID]); ok {
			handlerAgent = "section-editor"
			spec["edit_type"] = "rendered_html_transform"
			spec["transform_name"] = "code_span_to_code_tag"
			spec["slot_name"] = slot
			spec["fix"] = "This component's content_data cannot reproduce its rendered_html " +
				"(template fields absent from content_data — migration 499's test), so every " +
				"regenerate-from-source route is inapplicable BY CONSTRUCTION. The repair is " +
				"section-editor's rendered_html_transform: code_span_to_code_tag converts " +
				"`x` to <code>x</code> in assertion text only, byte-spliced " +
				"(datahelpers.ConvertLiteralCodeSpansInHTML), content_data untouched, floors " +
				"enforced, page reassembled and redeployed — no LLM in the loop."
		}
		specJSON, _ := json.Marshal(spec)
		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(pf.PageID); perr == nil {
			pageIDPtr = &parsed
		}
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "literal_markdown",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Literal markdown syntax on page %s (%d finding(s))", pf.PageName, len(pf.Findings)),
			SpecJSON:     string(specJSON),
			Priority:     40,
			HandlerAgent: handlerAgent,
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("literal_markdown:%s", pf.PageID),
			BatchID:      dctx.BatchID,
		})
	}

	// Retraction. Follows check_required_fields_missing's shape: read every
	// existing item of this item_type for the site with NO status filter
	// (resolveWorkItems alone owns workItemClosedStatuses), then only resolve
	// a page this run POSITIVELY re-scanned (scannedPages) and found clean
	// (absent from byPage). A page this run never touched is left alone.
	openRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT item_key FROM site_work_items
		WHERE site_id = $1 AND item_type = 'literal_markdown' AND item_key IS NOT NULL
	`, dctx.SiteID)
	if err != nil {
		// Findings and items already computed are this run's real output; a
		// failed retraction lookup must not discard them.
		dctx.Logger.Warn("literal_markdown: existing-items query failed — skipping retractions", zap.Error(err))
		return result, nil
	}
	defer openRows.Close()
	for openRows.Next() {
		var key string
		if err := openRows.Scan(&key); err != nil {
			continue
		}
		pageID := strings.TrimPrefix(key, "literal_markdown:")
		if pageID == key || !scannedPages[pageID] {
			continue // malformed key, or page not in this run's population — never resolve unseen
		}
		if _, stillDirty := byPage[pageID]; stillDirty {
			continue
		}
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType: "literal_markdown",
			ItemKey:  key,
			Reason:   "literal_markdown re-scan: page's unlocked components carry no markdown syntax on either surface",
		})
	}
	if err := openRows.Err(); err != nil {
		dctx.Logger.Warn("literal_markdown: existing-items iteration failed", zap.Error(err))
	}

	return result, nil
}
