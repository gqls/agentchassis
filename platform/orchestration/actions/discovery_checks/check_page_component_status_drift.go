// FILE: platform/orchestration/actions/discovery_checks/check_page_component_status_drift.go
//
// Discovery check: page_components whose build_status excludes them from the
// audit surface while their page is live.
//
// WHY THIS EXISTS (the meta-bug it guards against)
//
// Essentially every discovery check filters `pc.build_status = 'deployed'` —
// check_empty_sections, check_image_url_404, check_undeployed_assets,
// check_placeholder_image_in_use, check_component_standards. A page_component
// carrying any other status is therefore invisible to ALL of them: it ships to
// the live site, but no check will ever look at it again.
//
// page_components.build_status has no CHECK constraint. It is free text. So a
// writer can invent a value and silently remove a live section from the entire
// audit surface. That is exactly what happened on 2026-07-09: apply_section_edit
// hardcoded build_status = 'approved' after successfully deploying, leaving
// vonc.com's provocation-card the only 'approved' row in 578 and invisible to
// every check above. Fixed in UpdatePageStatusAction (page_component_id_field),
// but nothing DETECTED it — this check is that detector, and a regression guard
// for the next writer that invents a status.
//
// TWO OUTCOMES, DELIBERATELY DIFFERENT
//
//  1. Unknown status (not in the known vocabulary below) on a deployed page —
//     EMIT a work item. The HTML is already live; the row's status is simply
//     wrong, so the repair is a metadata flip, routed to component-template-fixer
//     with fix_type=repair_page_component_status (which sits beside the existing
//     align_slot_name fix — the same class: page_component metadata repair).
//     Expected volume: zero. This is a regression guard.
//
//  2. 'pending' on a deployed page — FINDING ONLY, no work item. 'pending' is a
//     legitimate "component regenerated, awaiting content rebuild" state. Its
//     repair is a rebuild, NOT a status flip: marking such a row 'deployed' would
//     hide genuinely stale HTML. Emitting a rebuild here would also churn the
//     fleet (25 such rows on 2026-07-10, 18 of them tool-* slots, none with an
//     open work item). Surfacing them without acting is the honest move; closing
//     that backlog is its own task, not this check's job.
//
// Silence is the failure mode this codebase keeps paying for, so case 2 is
// reported rather than filtered away.
//
// Registration: automatic via init() -> Register(&PageComponentStatusDriftCheck{}).
// Enable: add "page_component_status_drift" to a discovery agent's
//   default_config {workflow,steps,run_checks,config,checks} array
//   (completeness-discovery-agent is the natural home).

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&PageComponentStatusDriftCheck{}) }

type PageComponentStatusDriftCheck struct{}

func (c *PageComponentStatusDriftCheck) Name() string { return "page_component_status_drift" }

// knownComponentStatuses is the vocabulary any reader in the codebase honours.
// Sourced from the Go writers: save_page_sections / deploy_tool ('deployed'),
// store_generated_component + update_component_html ('pending'),
// v3_site_actions' filters ('removed'), and the reconciler ('needs_rebuild').
// Anything outside this set is a status nobody reads.
var knownComponentStatuses = map[string]bool{
	"deployed":      true,
	"pending":       true,
	"removed":       true,
	"needs_rebuild": true,
}

type driftRow struct {
	pcID, slotName, status string
	pageID, pageName       string
	hasHTML                bool
}

func (c *PageComponentStatusDriftCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.id::text, COALESCE(pc.slot_name, ''), COALESCE(pc.build_status, ''),
		       p.id::text, p.name,
		       (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '') AS has_html
		  FROM page_components pc
		  JOIN pages p ON p.id = pc.page_id
		 WHERE p.site_id = $1
		   AND p.build_status = 'deployed'
		   AND COALESCE(pc.build_status, '') <> 'deployed'
		 ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("page_component_status_drift query failed: %w", err)
	}
	defer rows.Close()

	var found []driftRow
	for rows.Next() {
		var r driftRow
		if scanErr := rows.Scan(&r.pcID, &r.slotName, &r.status, &r.pageID, &r.pageName, &r.hasHTML); scanErr != nil {
			dctx.Logger.Warn("page_component_status_drift: row scan failed", zap.Error(scanErr))
			continue
		}
		found = append(found, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}
	pendingCount := 0

	for _, r := range found {
		// Case 2: a known-but-not-deployed status. Report, never auto-flip.
		if knownComponentStatuses[r.status] {
			pendingCount++
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":             "page_component_status_drift",
				"page":              r.pageName,
				"slot_name":         r.slotName,
				"page_component_id": r.pcID,
				"build_status":      r.status,
				"skipped":           true,
				"reason": fmt.Sprintf(
					"%q on a deployed page — awaiting rebuild, not a status defect; invisible to every discovery check until rebuilt",
					r.status),
			})
			continue
		}

		// Case 1: a status no reader honours. The HTML is already live, so the
		// row's status is simply wrong. Repair it.
		//
		// Positive-evidence guard: only claim a row is deployed when it actually
		// carries rendered HTML. Mirrors pageHasComponents' refusal to mark an
		// empty page deployed.
		if !r.hasHTML {
			dctx.Logger.Warn("page_component_status_drift: unknown status but no rendered_html — not auto-repairable",
				zap.String("page", r.pageName),
				zap.String("slot", r.slotName),
				zap.String("build_status", r.status))
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":             "page_component_status_drift",
				"page":              r.pageName,
				"slot_name":         r.slotName,
				"page_component_id": r.pcID,
				"build_status":      r.status,
				"skipped":           true,
				"reason":            "unknown build_status and empty rendered_html — needs a rebuild, not a status repair",
			})
			continue
		}

		dctx.Logger.Warn("page_component_status_drift: component invisible to discovery checks",
			zap.String("page", r.pageName),
			zap.String("slot", r.slotName),
			zap.String("build_status", r.status))

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":             "page_component_status_drift",
			"page":              r.pageName,
			"slot_name":         r.slotName,
			"page_component_id": r.pcID,
			"build_status":      r.status,
			"reason":            "build_status is not a value any reader honours — the section is live but excluded from every discovery check",
		})

		specJSON, marshalErr := json.Marshal(map[string]interface{}{
			"check":             "page_component_status_drift",
			"fix_type":          "repair_page_component_status",
			"page_component_id": r.pcID,
			"slot_name":         r.slotName,
			"observed_status":   r.status,
			"expected_status":   "deployed",
		})
		if marshalErr != nil {
			continue
		}

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(r.pageID); perr == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   pageIDPtr,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "page_component_status_drift",
			Severity: "high",
			Summary: fmt.Sprintf(
				"Section %q on page %q has build_status %q — live but invisible to every discovery check",
				r.slotName, r.pageName, r.status),
			SpecJSON:     string(specJSON),
			Priority:     10,
			HandlerAgent: "component-template-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("page_component_status_drift:%s", r.pcID),
			BatchID:      dctx.BatchID,
		})
	}

	if len(result.WorkItems) > 0 || pendingCount > 0 {
		dctx.Logger.Info("page_component_status_drift: pass complete",
			zap.Int("emitted", len(result.WorkItems)),
			zap.Int("awaiting_rebuild_reported", pendingCount),
			zap.Int("found_total", len(found)))
	}
	return result, nil
}
