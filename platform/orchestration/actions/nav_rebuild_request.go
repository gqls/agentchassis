// FILE: platform/orchestration/actions/nav_rebuild_request.go
//
// RequestNavRebuild is how an action that has just given a page nav membership
// asks for that membership to become real.
//
// WHY THIS AND NOT A NAV ROW (bugs_open/149 A6, and it corrects A6's own
// suggested fix). `site_nav_items` is DERIVED from `pages`. Until 2026-07-31 it
// had two writers: `populate_nav_tables` (the derivation — authoritative, and it
// DELETEs the site's rows and rebuilds) and `addToolToNav` (hand-written,
// incremental, into a bespoke `tools` group). The authoritative writer could not
// express what the incremental one wrote, so it destroyed it — 7 active rows
// fleet-wide were destroy-and-not-recreate as of 2026-07-31. Making two writers
// agree is a coordination invariant, and bugs_open/149 is the record of that
// invariant drifting. So there is one writer, and creators declare intent on the
// page row (in_header / in_footer) and request the rebuild.
//
// A creation-time nav row would also be WORSE THAN NOTHING here, which is the
// non-obvious part. Chrome is a stored artefact (bugs_open/117, /118), so a nav
// row changes no served page on its own — while `check_orphan_pages` treats the
// presence of a nav row as reachability. Writing the row without re-rendering
// would leave the page exactly as unreachable AND silence the one check that
// would have noticed: your own write blinding your own detector.
//
// What makes the link real is nav-updater, whose live workflow is
//
//	populate_nav_tables -> render_site_components -> create_rerender_items
//	  -> get_pages_for_rerender
//
// i.e. derive, re-render chrome, propagate to every deployed page. This function
// asks for exactly that, once per site, through the existing work-item path.
//
// ---------------------------------------------------------------------------
// COUNCIL ROUND 4486f1a9 — APPROVED (12 reviewers, 0 unreadable, 2 medium
// objections, none high). The two medium objections are answered here rather than
// in a doc, because both are questions a future reader of this file will have.
//
// 1. `bug_historian`, medium: "the fix's actual delivery depends on a triage step
//    this plan does not control … a work item that sits in 'triaged' with nothing
//    promoting it is functionally the same silent non-delivery, just moved one hop
//    downstream". MEASURED 2026-07-31, and the answer inverts the worry:
//      - `nav_drift` all-history: **17 raised, 17 claimed, 17 complete.** This
//        function uses that exact item_type and handler.
//      - `build-pipeline-trigger` is ENABLED, last fired 2026-07-31 09:18:21Z.
//      - the `pipeline='build'` lane claimed 1,580 of 1,664 items over 7 days,
//        with claims on every one of those days.
//    The 263-detected/0-triaged backlog cited in the submission is the
//    **`detected` → `triaged` promotion** step, and this request is born `triaged`
//    precisely to skip it. So the hop downstream is a lane with a perfect record
//    for this item type, not a second queue. **Re-measure before trusting this
//    paragraph** — the standing query is RUNBOOK R12 in
//    docs024_key_docs_latest/bugfix_149_nav_membership/.
//
// 2. `guardian`, medium: "I can't confirm from SQL that `recurrenceExpected`
//    exists on the workItem struct with the semantics claimed; if it doesn't …
//    the third rebuild request goes terminal silently". Correct to ask, and it is
//    a code fact rather than a database one: the field is declared at
//    `load_work_item_actions.go:1069` and the two-strike block it disables is at
//    `:1082-1118` (`if terminalCount >= 2 { … item.status = "unresolved" }`,
//    guarded by `if item.itemKey != "" && !item.recurrenceExpected`). The dedup is
//    NOT waived by the flag — `ON CONFLICT (site_id, item_key) WHERE item_key IS
//    NOT NULL AND status NOT IN (<terminal>) DO NOTHING` at `:1148` still refuses a
//    second OPEN request. `nav_rebuild_request_test.go` pins both halves.
//
// Also from the round, worth keeping:
//   - `architecture` (low): this helper is exported but has only the two callers
//     in this package. **A third caller from OUTSIDE
//     platform/orchestration/actions is an RFC moment, not a silent adoption** —
//     at that point it is a general on-ramp to the work-item queue and has never
//     been reviewed as one.
//   - `guardian` (low): deleting addToolToNav rested on a grep, and "grep is not
//     reliable evidence of absence". True, and the stronger proof is that the Go
//     COMPILER now enforces it: the package builds with the function gone, which
//     no grep could establish.
// ---------------------------------------------------------------------------

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NavRebuildRequest describes why a nav rebuild is being asked for.
type NavRebuildRequest struct {
	SiteID uuid.UUID

	// PageID and PageName identify the page whose nav membership prompted the
	// request. Recorded in the spec so the trail says which page caused it; the
	// rebuild itself is site-wide, so these do not narrow the work.
	PageID   uuid.UUID
	PageName string
	PageURL  string

	// InHeader / InFooter are the flags as WRITTEN to the page row, not as
	// intended. They go in the spec so a reader can see the declaration the
	// rebuild is meant to honour without joining back to `pages`.
	InHeader bool
	InFooter bool

	// RequestedBy names the emitting action (e.g. "tool-deployer"). Becomes both
	// `source` and `created_by`, matching the needs_content_page items these
	// actions already emit.
	RequestedBy string
}

// RequestNavRebuild files one work item asking nav-updater to rebuild the site's
// nav tables and re-render its chrome. Returns true if an item was created;
// false if an open request already covers this site (not a failure).
//
// COALESCING, spelled out because the `editquality` seat asked what happens to the
// second request (low): the item_key is scoped to the SITE, so while one request is
// still open a second returns false and writes nothing — including nothing from its
// own spec. That is intended: the rebuild is site-wide, so one pending request
// covers any number of new pages. The consequence to know is that `spec.page_name`
// then names only the FIRST page that asked, so read it as "what prompted this",
// never as "what this will fix".
//
// Best-effort by design: a nav request must never roll back a tool build that
// has already succeeded, so failures are logged and swallowed exactly as the
// cross-link emitter does.
func RequestNavRebuild(ctx context.Context, db *sql.DB, req NavRebuildRequest, logger *zap.Logger) bool {
	if db == nil || req.SiteID == uuid.Nil {
		return false
	}

	spec, _ := json.Marshal(map[string]interface{}{
		"reason":       "nav_membership_declared",
		"page_id":      req.PageID.String(),
		"page_name":    req.PageName,
		"page_url":     req.PageURL,
		"in_header":    req.InHeader,
		"in_footer":    req.InFooter,
		"requested_by": req.RequestedBy,
		"fix": "A page was created with nav membership declared (in_header/in_footer). " +
			"Rebuild site_nav_items from pages and re-render chrome so the link ships.",
	})

	var pageIDPtr *uuid.UUID
	if req.PageID != uuid.Nil {
		pageIDPtr = &req.PageID
	}

	inserted, err := withWorkItemTx(ctx, db, logger, workItem{
		siteID:   req.SiteID,
		source:   req.RequestedBy,
		pipeline: "build",

		// item_type nav_drift, deliberately: nav-updater is already its handler,
		// and 11 of the 17 nav_drift items in the fleet's whole history are
		// hand-fired requests of exactly this shape ("News page created — nav
		// tables need rebuilding again"). A new item_type would need a new route
		// to mean anything.
		itemType: "nav_drift",
		severity: "medium",
		summary: fmt.Sprintf("Nav membership declared for %s — rebuild nav tables and re-render chrome",
			req.PageName),
		spec:         string(spec),
		pageID:       pageIDPtr,
		priority:     30,
		handlerAgent: "nav-updater",

		// 'triaged' so it is dispatchable now. Matching the needs_content_page
		// items these same actions emit: nothing currently promotes 'detected' to
		// 'triaged' on a schedule (263 items sat detected and 0 triaged fleet-wide
		// on 2026-07-30), so a request born 'detected' would not be a request.
		status:    "triaged",
		createdBy: req.RequestedBy,

		// A DISTINCT key from the detector's `nav_drift:<site_id>`. Sharing it
		// would make this request inherit the detector's strike history and would
		// blur a signal worth keeping legible: a nav_drift that keeps RECURRING
		// means the repair is not working.
		itemKey: fmt.Sprintf("nav_rebuild:%s", req.SiteID),

		// The two-strike rule in insertWorkItem brands a third item with a
		// repeated item_key as `unresolved` — terminal, never dispatched. That is
		// right for a detected defect and wrong for an action request, where a
		// completed predecessor means the request SUCCEEDED. Without this flag the
		// third tool added to a site would silently stop reaching the nav, which
		// is bugs_open/024's mechanism and the writing-side view of 149 A1's
		// "born unresolved".
		recurrenceExpected: true,
	})

	if err != nil {
		logger.Warn("RequestNavRebuild: failed to file nav rebuild request (non-fatal)",
			zap.String("site_id", req.SiteID.String()),
			zap.String("page_name", req.PageName),
			zap.Error(err))
		return false
	}
	if !inserted {
		logger.Info("RequestNavRebuild: an open nav rebuild request already covers this site",
			zap.String("site_id", req.SiteID.String()),
			zap.String("page_name", req.PageName))
		return false
	}

	logger.Info("RequestNavRebuild: filed nav rebuild request",
		zap.String("site_id", req.SiteID.String()),
		zap.String("page_name", req.PageName),
		zap.Bool("in_header", req.InHeader),
		zap.Bool("in_footer", req.InFooter))
	return true
}
