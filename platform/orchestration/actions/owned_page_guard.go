// FILE: platform/orchestration/actions/owned_page_guard.go
//
// Page-ownership guard for the GENERIC composition loops (bugs_open/208).
//
// pages.rebuild_policy='owned' marks a page that belongs to a tool/widget or is a
// runtime-fill shell — Experience Loop guard rail 1, migration 164, mechanising
// TL-001 after the vonc arena clobber. Migration 164 named two refusals:
// ReconcileSitePlanAction emits owned_page_review instead of needs_page, and
// SavePageSectionsAction hard-refuses a generic section save.
//
// Those two were sufficient for the route 164 modelled (needs_page ->
// page-build-handler, whose live order is save_sections -> update_status ->
// deploy_page, so the refusal precedes any commit). They are NOT sufficient for
// the three loops that compose and commit in the opposite order — measured
// 2026-08-06, page-rebuild, pageflow-builder and site-work-orchestrator all run:
//
//	assemble_page -> deploy_page (git_commit) -> save_sections -> update_page_status
//
// There the regenerated HTML is committed to the site repo, which the site
// deploys from, one step BEFORE the guard refuses the database write. The live
// tool is destroyed and page_components still describes it.
//
// This file holds the shared parts of the fix so the three touch points cannot
// drift apart: page identity resolution, the policy read, and the review-item
// emission. Keeping the precondition in a helper rather than in each caller is
// deliberate — a precondition parked in a caller is one port away from gone
// (the lesson recorded on bugs_closed/145).
//
// WHY assemble_page and not git_commit: git_commit is also how owned pages
// LEGITIMATELY deploy. page-rerender (rerender_single_page) and section-editor
// (apply_section_edit) both commit pages, and migration 164 says in terms that
// re-assembly of existing page_components "is deliberately NOT gated — it is how
// owned pages deploy". assemble_page is used by exactly the three loops above and
// by nothing else, and is fed freshly LLM-written HTML
// (content_field: page_content.response.page_html) in all three. It is therefore
// the one seam that means "generic composition about to be committed".
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ownedRebuildPolicy is the pages.rebuild_policy value meaning "not the generic
// pipeline's to rebuild" (migration 164's CHECK allows only 'generic'|'owned').
const ownedRebuildPolicy = "owned"

// ownedPageSkipReasonPrefix marks a skip caused by this guard. It is matched by
// nothing — the skip protocol is a boolean — but it is the string to pod-grep to
// prove the guard is live on a running binary, and it is what tells an operator
// reading a skipped iteration WHY the page was left alone.
const ownedPageSkipReasonPrefix = "OWNED_PAGE_GUARD"

// resolveGuardedPage finds the page a composition step is working on.
//
// The three consumers name it differently, so all four shapes are tried in
// order of authority (an id beats a name lookup):
//
//	page-rebuild / pageflow-builder : current_page.id, then current_page.name
//	site-work-orchestrator          : current_item.spec.id, then .spec.name
//
// current_item.spec is the full page record for build items — WriteBuildItemsAction
// marshals exactly what queryPagesForBuild returned — but a fix item minted by a
// discovery check carries only its own spec shape, which is why the name path
// exists and why an unresolved page must fail OPEN rather than block the loop.
//
// Returns ok=false when the page cannot be identified, which callers must treat
// as "carry on as before".
func resolveGuardedPage(ctx context.Context, db *sql.DB, collectedData map[string]interface{}, logger *zap.Logger) (pageID uuid.UUID, pageName string, ok bool) {
	if db == nil {
		return uuid.Nil, "", false
	}

	// Direct id paths first — no query needed, and no ambiguity.
	for _, idPath := range []string{"current_page.id", "current_item.spec.id"} {
		if raw := datahelpers.ExtractNestedFieldString(collectedData, idPath); raw != "" {
			if parsed, err := uuid.Parse(raw); err == nil {
				name := datahelpers.ExtractNestedFieldString(collectedData, "current_page.name")
				if name == "" {
					name = datahelpers.ExtractNestedFieldString(collectedData, "current_item.spec.name")
				}
				return parsed, name, true
			}
		}
	}

	// Name paths need the site to disambiguate: page names are unique per site,
	// not globally.
	siteIDStr := datahelpers.ExtractNestedFieldString(collectedData, "site_record.site_id")
	if siteIDStr == "" {
		siteIDStr = datahelpers.ExtractNestedFieldString(collectedData, "site_id")
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return uuid.Nil, "", false
	}

	for _, namePath := range []string{"current_page.name", "current_item.spec.name", "current_item.spec.page_name"} {
		name := datahelpers.ExtractNestedFieldString(collectedData, namePath)
		if name == "" {
			continue
		}
		// Reuses save_page_sections' own lookup so the two guards resolve a page
		// the same way by construction.
		id, _, lookupErr := saveSectionsLookupPageID(ctx, db, siteID, name)
		if lookupErr != nil {
			logger.Debug("resolveGuardedPage: name lookup missed",
				zap.String("path", namePath), zap.String("page_name", name), zap.Error(lookupErr))
			continue
		}
		return id, name, true
	}

	return uuid.Nil, "", false
}

// pageIsOwnedForGuard reads the ownership marker.
//
// Fails OPEN (false) on any error, matching SavePageSectionsAction's existing
// posture — "a scan error means an older schema, in which case the guard stands
// down". A guard that fails CLOSED here would stop generic page building
// fleet-wide the first time this query hiccupped, which is a worse outcome than
// the one it prevents, and the selection-level exclusion already carries the
// load for the two loops that own most of the traffic.
func pageIsOwnedForGuard(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) bool {
	if db == nil || pageID == uuid.Nil {
		return false
	}
	var policy string
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(rebuild_policy, 'generic') FROM pages WHERE id = $1
	`, pageID).Scan(&policy); err != nil {
		logger.Warn("ownedPageGuard: rebuild_policy lookup failed — guard standing down for this page",
			zap.String("page_id", pageID.String()), zap.Error(err))
		return false
	}
	return policy == ownedRebuildPolicy
}

// emitOwnedPageReviewItem records that a generic build was refused for an owned
// page, so the request does not vanish silently.
//
// Deliberately the SAME item_type and item_key namespace ReconcileSitePlanAction
// already uses (reconcile_site_plan_action.go:238-266): one deterministic key per
// page, arbitrated by the partial unique index on (site_id, item_key) over open
// statuses, with ON CONFLICT DO NOTHING. So repeated dispatches converge on one
// row rather than accumulating, and our emission dedups against reconcile's for
// free instead of competing with it.
//
// Errors are logged and swallowed: a guard must never fail because its reporting
// failed, and the refusal itself is what protects the page.
func emitOwnedPageReviewItem(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, source, reason string, logger *zap.Logger) {
	if db == nil || siteID == uuid.Nil || pageName == "" {
		return
	}

	spec := map[string]interface{}{
		"page_name":      pageName,
		"rebuild_policy": ownedRebuildPolicy,
		"reason":         reason,
		"refused_by":     source,
		"fix": "Owned/interactive page was excluded from a generic rebuild. Do NOT route it to " +
			"the generic page builder — it produces a prose page where an interactive tool " +
			"belongs. Rebuild via the tool pipeline (tool-generator/create_tool_component), " +
			"edit via apply_section_edit/section-editor, or change pages.rebuild_policy if the " +
			"page genuinely is the pipeline's to compose.",
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		logger.Warn("ownedPageGuard: could not marshal review spec", zap.Error(err))
		return
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, status, created_by, item_key
		) VALUES ($1, $2, 'build', 'owned_page_review',
		          'high', $3, $4::jsonb, 30,
		          'needs_human_review', $2, $5)
		ON CONFLICT DO NOTHING
	`, siteID, source,
		fmt.Sprintf("Owned page %s was excluded from a generic rebuild — needs owner-aware handling, not the generic builder", pageName),
		string(specJSON),
		"owned_page_review:"+pageName,
	); err != nil {
		logger.Warn("ownedPageGuard: could not emit owned_page_review item",
			zap.String("page_name", pageName), zap.Error(err))
		return
	}

	logger.Info("ownedPageGuard: owned_page_review recorded",
		zap.String("page_name", pageName), zap.String("source", source))
}

// censusExcludedOwnedPages names the owned pages that the selection WOULD have
// returned had the ownership exclusion not been applied, and records a review item
// for each.
//
// Without this the exclusion is silent: the page stays at needs_rebuild, is
// re-excluded on every subsequent run, and an operator who explicitly asked for it
// gets neither the rebuild nor a reason. The predicate is deliberately the exact
// inverse of ownedPageExclusionSQL — same status filter, same liveness filter, the
// ownership test flipped — so it cannot drift into reporting a different population
// from the one that was excluded.
//
// Returns the page names for the action result. Errors are logged and swallowed:
// reporting must never fail a selection.
func censusExcludedOwnedPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, statuses []string, includeAll bool, source string, logger *zap.Logger) []string {
	if db == nil || siteID == uuid.Nil {
		return nil
	}

	query := `SELECT name FROM pages
	          WHERE site_id = $1 AND status = 'active'
	            AND COALESCE(rebuild_policy, 'generic') = 'owned'
	          ORDER BY name`
	args := []interface{}{siteID}

	if !includeAll {
		placeholders := make([]string, len(statuses))
		for i, s := range statuses {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, s)
		}
		query = fmt.Sprintf(`SELECT name FROM pages
		          WHERE site_id = $1 AND status = 'active'
		            AND COALESCE(build_status, 'planned') IN (%s)
		            AND COALESCE(rebuild_policy, 'generic') = 'owned'
		          ORDER BY name`, strings.Join(placeholders, ", "))
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Warn("ownedPageGuard: excluded-page census failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if scanErr := rows.Scan(&name); scanErr != nil {
			logger.Warn("ownedPageGuard: census scan failed", zap.Error(scanErr))
			continue
		}
		names = append(names, name)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		logger.Warn("ownedPageGuard: census iteration failed", zap.Error(rowsErr))
	}
	// Close before the INSERTs: holding the result set open while writing on the
	// same *sql.DB can pin a connection unnecessarily.
	rows.Close()

	if len(names) == 0 {
		return nil
	}

	logger.Warn("ownedPageGuard: owned pages excluded from a generic build selection",
		zap.String("site_id", siteID.String()),
		zap.Strings("pages", names),
		zap.String("source", source),
	)

	for _, name := range names {
		emitOwnedPageReviewItem(ctx, db, siteID, name, source,
			ownedPageSkipReasonPrefix+": excluded at selection — an owned page is not the generic builder's to compose",
			logger)
	}

	return names
}

// upstreamAssemblySkipped reports whether the assemble step for THIS iteration
// declared a skip, and why.
//
// Reads the same collected_data shape checkUpstreamSkipped's first branch reads
// (assembled_page.skipped), which every live consumer populates: all three
// assemble_page steps declare output_field "assembled_page", measured 2026-08-06.
func upstreamAssemblySkipped(collectedData map[string]interface{}) (bool, string) {
	assembled, ok := collectedData["assembled_page"].(map[string]interface{})
	if !ok {
		return false, ""
	}
	skipped, _ := assembled["skipped"].(bool)
	if !skipped {
		return false, ""
	}
	reason, _ := assembled["skip_reason"].(string)
	return true, reason
}
