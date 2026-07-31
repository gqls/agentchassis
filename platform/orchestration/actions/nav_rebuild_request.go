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
