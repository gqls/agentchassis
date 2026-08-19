// FILE: platform/orchestration/actions/create_tool_component_regenerate.go
//
// The REPLACE arm of create_tool_component (bugs_open/331, register TL-047).
//
// create_tool_component has always had a CREATE path and no REPLACE path: its
// per-site "already exists" probe returns {already_exists:true} and writes
// nothing, and a second generation for the same (function, site) would in any
// case die on content_components_name_key (UNIQUE(name), name =
// <function>-<domainSlug>, no predicate — is_active=false does not free it).
// So every re-fix of a native tool was a hand-assembled sequence of DB edits
// around the action: deactivate the old row, rename it, build, then retire the
// old slot before the generator's own rerender claimed (a race measured at
// 2–96 minutes with no floor; lost once, and the page served BOTH tools).
// agent_error_log holds exactly one duplicate-key row per gate on three
// consecutive days — each found by walking into it.
//
// This arm regenerates IN PLACE — the section writer's convention (CTS-009,
// store_generated_component's regeneration branch; lookupBaseComponent omits
// is_active precisely because the alternative is the name collision above).
// Same component_id, so every FK keeps resolving; forked_from / name /
// function untouched, so no uniqueness gate is met and RFC_036 §9.3's
// library-claim lookup cannot record the replacement as a fork of the row it
// replaces; the live placement's rendered_html is overwritten in the same
// transaction, so the page never holds two tool slots and there is no retire
// race; and that rendered_html UPDATE fires trg_page_component_artefact_archive_upd,
// so page_component_history receives the old bytes — the revert handle a
// status flip never produced (the lane had to record md5s by hand).
//
// Authority is PER ITEM, not per step: the optional input `replace_existing`
// (mapped strictly from input_data.spec.replace_existing by the tool-generator
// seed) must be true. Absent ⇒ the probe's already_exists short-circuit stands,
// because it is the per-site throttle that stops a duplicate add_tool (the
// suggester re-suggesting a tool the site has) from becoming a live
// regeneration — LANDMINES "why not just fix the query" is right about that.
//
// What is deliberately NOT done here: the creation-only tail (nav rebuild
// request, tool-content item, cross-link items, companion guide upsert). Those
// exist from the first build; a regeneration is a content change, which is
// what lets it be one transaction. The result map still carries every key the
// downstream tool-generator steps read (compose_plan / write_plan / index_plan
// / enqueue_rerender: page_id, page_url, function), plus regenerated:true.

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type toolRegenerateRequest struct {
	siteID      uuid.UUID
	incumbentID string // content_components.id of the site's live tool component for this function
	function    string
	displayName string
	description string
	category    string
	htmlContent string
}

// regenerateToolComponentInPlace is called by CreateToolComponentAction when
// replace_existing is set and the per-site probe found an incumbent. It
// returns the action's result map or an error; on error nothing has changed
// (the only write outside the transaction is the best-effort version snapshot,
// which is additive history).
func regenerateToolComponentInPlace(ctx context.Context, params ActionParams, logger *zap.Logger, req toolRegenerateRequest) (map[string]interface{}, error) {
	db := params.DB
	logger = logger.With(zap.String("arm", "replace_existing"), zap.String("incumbent_id", req.incumbentID))

	// --- Read the incumbent as it stands (the bytes the snapshot preserves) ---
	var currentHTML, currentSchema string
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(html_template, ''), COALESCE(input_schema::text, '{}')
		FROM content_components
		WHERE id = $1::uuid AND is_active = true AND component_level = 'tool'
	`, req.incumbentID).Scan(&currentHTML, &currentSchema); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("replace_existing: incumbent tool component %s is no longer an active tool row — re-read before retrying", req.incumbentID)
		}
		return nil, fmt.Errorf("replace_existing: read incumbent %s: %w", req.incumbentID, err)
	}

	// --- Snapshot to component_versions (best-effort, same contract as the
	// section writer and update_component_html; the load-bearing revert handle
	// is page_component_history, written by the rendered_html trigger below) ---
	var maxVersion int
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version_number), 0)
		FROM component_versions
		WHERE component_id = $1::uuid
	`, req.incumbentID).Scan(&maxVersion); err != nil {
		logger.Warn("CreateToolComponentAction: could not read max version_number, defaulting to 0", zap.Error(err))
		maxVersion = 0
	}
	previousVersion := maxVersion + 1
	if currentHTML != "" {
		if err := snapshotComponentVersion(
			ctx, db, req.incumbentID, previousVersion,
			currentHTML, currentSchema, "",
			"Regenerated in place by tool-generator (replace_existing, bugs_open/331)",
			"tool-generator:replace_existing",
			nullStringOrEmpty(params.ExecutionContext.OrchestrationID),
			logger,
		); err != nil {
			logger.Warn("CreateToolComponentAction: version snapshot failed, continuing (page_component_history still archives the slot)", zap.Error(err))
			previousVersion = 0
		}
	} else {
		previousVersion = 0
	}

	// --- Shared-template fence (bugs_open/285) ---
	// A write keyed by component_id reaches EVERY placement of that row. This
	// arm is site-scoped by intent (the requesting site's own tool), so an
	// incumbent placed on more than one site is a shared/library row and is
	// refused here, not regenerated under the other sites. Unlike
	// update_component_html (which logs and proceeds for tools, preserving its
	// historical behaviour) a NEW arm has no legacy to fail towards: an
	// unreadable census also refuses. Nothing has been written at this point.
	fence, ferr := sharedComponentWriteCheck(ctx, db, req.incumbentID, "tool", false)
	if ferr != nil {
		return nil, fmt.Errorf("replace_existing refused: placement census for %s unreadable (%v) — nothing written; retry", req.incumbentID, ferr)
	}
	if fence.PlacementSites > 1 {
		return nil, fmt.Errorf("replace_existing refused: tool component %s (%s) is placed on %d pages across %d sites — a site-scoped regeneration must not rewrite a row other sites serve; fork it per site (deploy_tool_to_site) or regenerate the site's own fork",
			req.incumbentID, req.function, fence.PlacementPages, fence.PlacementSites)
	}

	// --- One transaction: lock the live placements, rewrite row + slots ---
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("replace_existing: begin tx: %w", err)
	}
	defer tx.Rollback()

	// The incumbent's LIVE, agent-writable placements on this site's pages.
	// build_status='removed' is the assembly-excluded tombstone and is not a
	// live slot; a human lock (pageComponentAgentWritableSQL) is honoured —
	// automation must not overwrite it, so a locked tool refuses here rather
	// than being regenerated around.
	rows, err := tx.QueryContext(ctx, `
		SELECT pc.id::text, pc.page_id::text, p.name, p.url, COALESCE(pc.rendered_html, '') <> ''
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.component_id = $1::uuid
		  AND p.site_id = $2
		  AND pc.build_status IS DISTINCT FROM 'removed'
		  AND `+pageComponentAgentWritableSQL("pc.")+`
		ORDER BY pc.position ASC, pc.id
		FOR UPDATE OF pc
	`, req.incumbentID, req.siteID)
	if err != nil {
		return nil, fmt.Errorf("replace_existing: lock placements: %w", err)
	}
	var slotIDs []string
	var pageID, pageName, pageURL string
	for rows.Next() {
		var sid, pid, pname, purl string
		var hadHTML bool
		if err := rows.Scan(&sid, &pid, &pname, &purl, &hadHTML); err != nil {
			rows.Close()
			return nil, fmt.Errorf("replace_existing: scan placement: %w", err)
		}
		slotIDs = append(slotIDs, sid)
		if pageID == "" {
			pageID, pageName, pageURL = pid, pname, purl
		}
		if !hadHTML {
			// The archive trigger's WHEN clause requires a non-empty old
			// rendered_html; an empty slot has nothing to archive and that is
			// fine — say so, so a missing history row is not read as a bug.
			logger.Info("CreateToolComponentAction: placement had no rendered_html; no archive row will be written for it", zap.String("slot_id", sid))
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("replace_existing: placements: %w", err)
	}
	rows.Close()
	if len(slotIDs) == 0 {
		return nil, fmt.Errorf("replace_existing refused: tool component %s (%s) has no live, agent-writable placement on site %s — the slot is removed (withdrawn) or human-locked; a withdrawn tool is re-filed WITHOUT replace_existing after deactivating the old row (LANDMINES: the already-exists probe), and a locked one is a human's to unlock",
			req.incumbentID, req.function, req.siteID)
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE content_components
		SET html_template = $1,
		    display_name = $2,
		    description = $3,
		    category = $4,
		    source_agent_type = $5,
		    source_orchestration_id = $6,
		    updated_at = NOW()
		WHERE id = $7::uuid AND is_active = true AND component_level = 'tool'
	`, req.htmlContent, req.displayName, req.description, req.category,
		nullIfEmpty(params.ExecutionContext.Sender.AgentType),
		nullIfEmpty(params.ExecutionContext.OrchestrationID),
		req.incumbentID)
	if err != nil {
		return nil, fmt.Errorf("replace_existing: update component: %w", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return nil, fmt.Errorf("replace_existing: component %s changed under us (%d rows updated) — nothing written", req.incumbentID, n)
	}

	// Tools carry the template verbatim as rendered_html (what creation writes
	// at its link INSERT and what deploy_tool writes for a fork); the assembler
	// reads rendered_html and never re-renders a slot from the template, so a
	// 'pending' flip alone would leave the old tool serving. This UPDATE is
	// what fires the archive trigger.
	res, err = tx.ExecContext(ctx, `
		UPDATE page_components
		SET rendered_html = $1, build_status = 'deployed', updated_at = NOW()
		WHERE id::text = ANY($2::text[])
	`, req.htmlContent, toPGTextArrayLiteral(slotIDs))
	if err != nil {
		return nil, fmt.Errorf("replace_existing: update placements: %w", err)
	}
	if n, _ := res.RowsAffected(); int(n) != len(slotIDs) {
		return nil, fmt.Errorf("replace_existing: expected %d placement rows updated, got %d — nothing written", len(slotIDs), n)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("replace_existing: commit: %w", err)
	}

	guideURL := ""
	if _, gu, gerr := companionGuideIdentity(pageName); gerr == nil {
		guideURL = gu
	}

	logger.Info("CreateToolComponentAction: tool regenerated IN PLACE",
		zap.String("component_id", req.incumbentID),
		zap.String("page_id", pageID),
		zap.String("page_url", pageURL),
		zap.Int("slots_updated", len(slotIDs)),
		zap.Int("previous_version", previousVersion))

	return map[string]interface{}{
		"component_id":     req.incumbentID,
		"page_id":          pageID,
		"page_url":         pageURL,
		"function":         req.function,
		"display_name":     req.displayName,
		"guide_url":        guideURL,
		"needs_rerender":   true,
		"generated":        true,
		"regenerated":      true,
		"page_adopted":     true, // the page pre-existed and was used as it stands — the lane's grading query reads this key
		"slots_updated":    len(slotIDs),
		"previous_version": previousVersion,
	}, nil
}

// nullStringOrEmpty is the change_source argument shape snapshotComponentVersion
// wants: "" means NULL.
func nullStringOrEmpty(s string) string { return s }

// replaceExistingRequested reads the per-item flag. A JSON boolean is the
// declared shape; the string "true" is accepted because a work-item spec
// written by hand (the lane's RUNBOOK INSERT) may carry it quoted. Anything
// else — absent, false, "", "false" — is OFF.
func replaceExistingRequested(inputs interface {
	GetBool(key string, def bool) bool
	Get(key string) string
}) bool {
	if inputs.GetBool("replace_existing", false) {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(inputs.Get("replace_existing")), "true")
}
