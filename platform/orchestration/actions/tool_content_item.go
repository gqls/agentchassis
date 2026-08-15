// FILE: platform/orchestration/actions/tool_content_item.go
//
// The single seam through which a tool build asks for prose around its widget:
// a `needs_content_page` item keyed `tool_content:<function>:<site>`, raised
// only when the page declares a section a content writer could actually write.
//
// WHY THE GUARD EXISTS (bugs_open/177). The two actions that create tool pages
// emitted a byte-identical work item on top of two different page shapes:
//
//   - deploy_tool_action.go (source `tool-deployer`) declares
//     pages.sections = ["hero-tool","tool-guide-intro","<function>","tool-cta"]
//     and then raises the item. Not one of the failures came from this path.
//   - create_tool_component_action.go (source `tool-generator`) copied the emit
//     and NOT the declaration: its `pages` INSERT names no sections column, so
//     the row is born `[]`. Every failure came from this path.
//
// page-build-handler resolves a page's sections from site_plan_sections, then
// site_specs.site_plan, then pages.sections. A freshly generated tool page sits
// in none of them, so the handler is handed nothing, no-ops, and the WDS-004/149
// routing parks the item in needs_human_review — correctly, given what it was
// handed. The item is UNSATISFIABLE AT THE MOMENT IT IS MINTED: 9 of 9
// `tool_content:%` items ever created died that way (8 of 8 when the bug was
// filed on 2026-08-02, a ninth minted the day after, so the class was still
// growing), spanning 2026-07-14 to 2026-08-02 across four sites, every one at
// attempt_count = 1 carrying the byte-identical error. Never one completion.
//
// The control that puts the cause in the declaration rather than in the emitter
// or the handler: the companion-guide items raised by the SAME two files, whose
// page DOES declare ["hero","article-body","call-to-action"], run 4 complete to
// 1 review. Same emitter, same handler, same window.
//
// These are not inert rows. A dependency is released only by complete/verified
// (bugs_closed/176), so one of them became a permanent blocker and, together
// with the selector defect 176 records, stalled the entire fleet twice on
// 2026-08-02.
//
// THE CONTRACT WITH THE HANDLER. The resolution is declaredPageSections
// (page_section_satisfiability.go), which mirrors
// load_page_sections_from_spec_action.go's fallbacks 1 to 3, in its order, first
// non-empty source winning — because the question asked here is exactly the
// question the handler asks later, and an answer taken from a different source
// order would be an answer to a different question. It lived in this file until
// bugs_open/187 found the same shape under two more needs_page emitters and
// extracted it; the extraction was pure, and the tests in
// tool_content_item_test.go passing unchanged is the proof of that.
//
// Two deliberate departures from the loader, both narrowing:
//
//   - Sibling synthesis (the loader's fallback 4) is excluded. It borrows a
//     layout from a same-role sibling IN THE CURRENT PLAN, so it can only serve
//     a page the plan already contains, and such a page has its own sections,
//     which fallback 1 has already found. A tool page minted seconds ago is in
//     no plan at all. THIS CALL SITE DOES NOT USE the shared resolver's
//     pageInCurrentPlan arm for that reason — the needs_page emitters do, and
//     their pages are not seconds old.
//   - Read-only. The loader SYNCS a served result back into pages.sections. This
//     runs at tool birth, judging a page whose shape is still being decided, and
//     a guard that writes to the table it is judging changes the answer for
//     everything downstream of it.
//
// THE EDGE WORTH STATING PLAINLY: a page whose only declared section IS the tool
// function skips as well. That is protective rather than an oversight. The one
// section the writer would be pointed at is the widget itself, and rewriting a
// tool widget as generic prose is the clobber mechanism (concept register
// TL-001) — the item would be satisfiable, and satisfying it would destroy the
// tool.
//
// THE WRITE goes through insertWorkItem (load_work_item_actions.go) rather than
// the local INSERT both call sites carried. The council has already ruled on
// that shape: gapPlanWorkItem (apply_gap_plan_action.go, correlation a5b70424)
// adopted the shared door after two seats objected that a hand-rolled
// `ON CONFLICT DO NOTHING` reimplements half of insertWorkItem's semantics with
// a bare conflict target that does not match idx_swi_dedup's partial predicate.
// `recurrenceExpected` is set for the same stated reason as the gap plan: this
// item is an ACTION REQUEST, so a completed predecessor means the request
// SUCCEEDED, and the two-strike rule must not brand a re-request `unresolved` at
// birth — bugs_open/024's regression on the tool fix loop. Dedup is NOT waived
// by the flag: idx_swi_dedup still refuses a second OPEN row for the same
// (site_id, item_key).
//
// Out of scope, deliberately: neither path gains a sections declaration here.
// Whether a generated tool page should carry prose around its widget at all is
// TL-009's open design question, and answering it in a bug fix would change the
// shape of every tool page this platform generates from now on.

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// toolContentItemRequest describes one freshly created tool page that may want
// prose written around its widget.
//
// pageName is the key both plan sources are indexed by, so it must be the name
// as WRITTEN to the pages row — the two creation paths derive it differently
// (CanonicalisePage on one arm, a `tool-` prefix strip on the other) and a name
// that disagrees resolves no sections and skips a page that wanted the item.
type toolContentItemRequest struct {
	siteID       uuid.UUID
	pageID       uuid.UUID
	pageName     string
	toolFunction string
	displayName  string
	description  string
	pageURL      string
	// source names the emitting actor, "tool-generator" or "tool-deployer". It
	// fills source, created_by and the spec's own `source` field, all of which
	// the two call sites already set to the same value.
	source string
}

// raiseToolContentItem files the tool page's content item when, and only when,
// the page declares a prose section beyond the widget.
//
// The disposition is returned rather than logged alone so the calling action can
// put it in its output map: a guard that silently declines is indistinguishable
// from a guard that never ran, and this one declines on the common path.
// Values: "raised", "skipped_no_prose_sections", "deduped_open_item",
// "insert_failed".
//
// Non-fatal throughout, matching the behaviour it replaces: a tool that has been
// built and deployed must not be failed because the follow-on content request
// could not be written.
func raiseToolContentItem(ctx context.Context, params ActionParams, logger *zap.Logger, req toolContentItemRequest) string {
	logger = logger.With(
		zap.String("tool_function", req.toolFunction),
		zap.String("page_name", req.pageName),
		zap.String("emitted_by", req.source),
	)

	// Both call sites refuse to run without a database, so this cannot fire in
	// production. It is here because the alternative to a coarse label is a
	// panic, and no item was raised either way.
	if params.DB == nil {
		logger.Warn("raiseToolContentItem: no database connection, tool content item not raised")
		return "insert_failed"
	}

	declared, sectionSource := declaredPageSections(ctx, params.DB, logger, req.siteID, req.pageName)

	prose := make([]string, 0, len(declared))
	for _, name := range declared {
		if name != req.toolFunction {
			prose = append(prose, name)
		}
	}

	if len(prose) == 0 {
		logger.Info("raiseToolContentItem: tool page declares no prose sections beyond the widget — needs_content_page not raised; see bugs_open/177",
			zap.String("sections_source", sectionSource),
			zap.Strings("declared_sections", declared))
		return "skipped_no_prose_sections"
	}

	spec, _ := json.Marshal(map[string]interface{}{
		"page_name":         req.pageName,
		"page_id":           req.pageID.String(),
		"page_type":         "tool",
		"tool_function":     req.toolFunction,
		"tool_display_name": req.displayName,
		"tool_description":  req.description,
		"tool_page_url":     req.pageURL,
		"source":            req.source,
		// bugs_open/271: "suggestion" is the key page-build-handler reads and
		// forwards to the writer prompt; "content_guidance" had no reader.
		"suggestion": fmt.Sprintf(
			"This is a tool page for '%s'. Generate: (1) a hero section with the tool name and a one-line benefit statement, "+
				"(2) an educational guide section explaining the concept behind the tool — what it calculates, why it matters, "+
				"how users should interpret the results — written for the site's target audience, "+
				"(3) a CTA section encouraging users to try the tool and linking to related content. "+
				"Do NOT regenerate the tool widget itself — it is already deployed at position 2.",
			req.displayName,
		),
	})

	// withWorkItemTx is the shape gapPlanWorkItem's three call sites use: one
	// short transaction per item, because insertWorkItem takes a *sql.Tx for
	// callers that batch, and a failure here must never roll back the tool build
	// that has already succeeded.
	pageID := req.pageID
	inserted, err := withWorkItemTx(ctx, params.DB, logger, workItem{
		siteID:             req.siteID,
		source:             req.source,
		pipeline:           "build",
		itemType:           "needs_content_page",
		severity:           "medium",
		summary:            fmt.Sprintf("Write content for tool page: %s", req.displayName),
		spec:               string(spec),
		pageID:             &pageID,
		priority:           50,
		handlerAgent:       "page-build-handler",
		status:             "triaged",
		createdBy:          req.source,
		itemKey:            fmt.Sprintf("tool_content:%s:%s", req.toolFunction, req.siteID),
		recurrenceExpected: true,
	})
	if err != nil {
		logger.Warn("raiseToolContentItem: failed to create tool content work item (non-fatal)",
			zap.Error(err))
		return "insert_failed"
	}
	if !inserted {
		logger.Info("raiseToolContentItem: an open tool content item already holds this key",
			zap.String("sections_source", sectionSource))
		return "deduped_open_item"
	}

	logger.Info("raiseToolContentItem: tool content item raised",
		zap.String("sections_source", sectionSource),
		zap.Strings("prose_sections", prose))
	return "raised"
}
