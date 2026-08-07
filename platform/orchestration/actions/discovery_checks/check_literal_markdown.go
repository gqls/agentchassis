// FILE: platform/orchestration/actions/discovery_checks/check_literal_markdown.go
//
// CHANGES:
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
	"regexp"
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

var (
	mdBoldRe     = regexp.MustCompile(`\*\*[A-Za-z][^*\n]{0,80}\*\*`)
	mdCodeSpanRe = regexp.MustCompile("`[A-Za-z0-9][^`\n]{0,80}`")
	mdHeadingRe  = regexp.MustCompile(`(?m)^#{1,6} \S`)
	// A value carrying markup or script is not a text-typed field — the
	// code-span pattern is suppressed there (backticks are code, not prose).
	htmlMarkupRe = regexp.MustCompile(`<[A-Za-z/!]`)
)

type literalMarkdownFinding struct {
	SlotName string `json:"slot_name"`
	Field    string `json:"field,omitempty"` // content_data path; empty for rendered_html
	Pattern  string `json:"pattern"`         // bold | code_span | heading
	Matched  string `json:"matched"`
	Source   string `json:"source"` // content_data | rendered_html
}

// scanPlainTextMarkdown applies the guarded patterns to one plain-text
// string. One finding per pattern per value — the repair is a page rewrite,
// so occurrence counts change nothing about what to do.
func scanPlainTextMarkdown(text, slot, field, source string, includeCodeSpan bool) []literalMarkdownFinding {
	var out []literalMarkdownFinding
	add := func(pattern, matched string) {
		if len(matched) > 120 {
			matched = matched[:120]
		}
		out = append(out, literalMarkdownFinding{
			SlotName: slot, Field: field, Pattern: pattern,
			Matched: matched, Source: source,
		})
	}
	if m := mdBoldRe.FindString(text); m != "" {
		add("bold", m)
	}
	if includeCodeSpan {
		if m := mdCodeSpanRe.FindString(text); m != "" {
			add("code_span", m)
		}
	}
	if m := mdHeadingRe.FindString(text); m != "" {
		add("heading", m)
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
	// reporting Resolved on the second would stamp 'complete' over a destroyed page.
	// Refuse to answer instead: an error means "could not verify", which the caller
	// records rather than silently treating as success.
	if scanned == 0 {
		return VerifyResult{}, fmt.Errorf(
			"literal_markdown verifier: page %s has no scannable components — "+
				"cannot distinguish 'repaired' from 'content lost', refusing to certify",
			target.PageID)
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
		       COALESCE(pc.rendered_html, ''), COALESCE(pc.content_data::text, '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
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

	for rows.Next() {
		var pageID, pageName, pageURL, slot, html, contentJSON string
		if err := rows.Scan(&pageID, &pageName, &pageURL, &slot, &html, &contentJSON); err != nil {
			dctx.Logger.Warn("literal_markdown: scan error", zap.Error(err))
			continue
		}
		scannedPages[pageID] = true

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
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "literal_markdown",
			"page_id":   pf.PageID,
			"page_name": pf.PageName,
			"page_url":  pf.PageURL,
			"findings":  pf.Findings,
			"fix": "Rewrite the affected fields WITHOUT markdown syntax: text-typed " +
				"fields are rendered verbatim (STRICT RULE 9), so **bold**, `code spans` " +
				"and # headings reach the visitor as literal characters. Do not merely " +
				"delete the markers — if the writer wanted emphasis, re-word so the words " +
				"carry it. Fix page_components.content_data (the durable copy), not only " +
				"the rendered HTML: a rerender reprints content_data.",
		})
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
			HandlerAgent: "page-build-handler", // NOT page-content-writer — see header, bugs_open/201
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
