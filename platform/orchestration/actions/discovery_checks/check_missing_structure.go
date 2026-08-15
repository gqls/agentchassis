// FILE: platform/orchestration/actions/discovery_checks/check_missing_structure.go
//
// Flags a site whose chrome (header, footer, <head>) never rendered, and
// orders a full reassembly (refresh_site_components=true) to repair it.
//
// bugfix_270 (2026-08-15): this check used to read pages.rendered_header /
// rendered_footer / rendered_head, three columns that are empty on every page
// fleet-wide (LANDMINES.md, "pages.rendered_header / rendered_footer /
// rendered_head are VESTIGIAL") — chrome actually lives in site_components,
// written by render_site_components. Combined with a second broken filter
// (pages.status IN ('active','deployed'), where 'deployed' is a value of a
// DIFFERENT column and never occurs in pages.status), the old predicate was
// true for every non-archived page on every site, unconditionally — the check
// fired every discovery pass, on every site, forever, and dispatched a
// high-priority full-site rerender each time. ~31 such rerenders ran for
// nothing between 2026-04-24 and 2026-08-14 before this fix. Full evidence,
// live census and the fix design: bugs_open/270, PLAN in
// docs/agent_docs/docs024_key_docs_latest/bugfix_270_missing_structure/.
//
// The predicate now asks the store chrome actually lives in: a slot is
// healthy when site_components holds a non-empty rendered_html for it, not
// merely build_status='rendered' — build_status='pending' coexists with valid,
// currently-serving content elsewhere in this codebase (chrome_link_policy.go
// sets it as a "re-render scheduled" signal, not a "content missing" one), so
// gating on it here would manufacture a different false-positive class.
//
// Self-clearing (RFC_010, CheckResult.Resolved): once a site's three slots are
// all healthy, this check positively retracts its own item under the SAME
// item_key it always used ("missing_structure:rerender") — which is how the
// items this bug's old predicate filed close themselves, with no hand-written
// cleanup, on each affected site's next discovery pass.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func init() { Register(&MissingStructureCheck{}) }

type MissingStructureCheck struct{}

func (c *MissingStructureCheck) Name() string { return "missing_structure" }

// missingStructureQuery is a package-level const, not inlined, so a test can
// assert on the predicate actually issued — the sqlmock happy-path tests
// alone cannot catch a regression back to pages.rendered_* or to
// build_status, because they return whatever rows they are handed regardless
// of the WHERE clause (see check_componentless_pages_test.go's header for the
// same reasoning, copied here deliberately).
const missingStructureQuery = `
	SELECT
	  EXISTS (SELECT 1 FROM pages p
	          WHERE p.site_id = $1 AND p.status = 'active')          AS has_active_pages,
	  EXISTS (SELECT 1 FROM site_components sc
	          WHERE sc.site_id = $1 AND sc.slot_name = 'header'
	            AND coalesce(length(sc.rendered_html), 0) > 0)       AS header_ok,
	  EXISTS (SELECT 1 FROM site_components sc
	          WHERE sc.site_id = $1 AND sc.slot_name = 'footer'
	            AND coalesce(length(sc.rendered_html), 0) > 0)       AS footer_ok,
	  EXISTS (SELECT 1 FROM site_components sc
	          WHERE sc.site_id = $1 AND sc.slot_name = 'head'
	            AND coalesce(length(sc.rendered_html), 0) > 0)       AS head_ok
`

const missingStructureItemKey = "missing_structure:rerender"

func (c *MissingStructureCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	var hasActivePages, headerOK, footerOK, headOK bool
	err := dctx.DB.QueryRowContext(dctx.Ctx, missingStructureQuery, dctx.SiteID).
		Scan(&hasActivePages, &headerOK, &footerOK, &headOK)
	if err != nil {
		return nil, fmt.Errorf("missing_structure query failed: %w", err)
	}

	var missingSlots []string
	if !headerOK {
		missingSlots = append(missingSlots, "header")
	}
	if !footerOK {
		missingSlots = append(missingSlots, "footer")
	}
	if !headOK {
		missingSlots = append(missingSlots, "head")
	}

	result := &CheckResult{}

	if len(missingSlots) == 0 {
		// Chrome is demonstrably present — a positive observation, which is
		// what RFC_010 requires before a retraction (never inferred from an
		// absence of findings). Narrow ItemKey, not AllOfType: this check
		// only ever files under one fixed key per site, so the narrow claim
		// is the honest one and cannot touch another check's needs_rerender
		// items sharing that item_type.
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType: "needs_rerender",
			ItemKey:  missingStructureItemKey,
			Reason: "site_components healthy: header, footer and head all hold non-empty " +
				"rendered_html (bugfix 270 — earlier items were filed by a predicate reading " +
				"vestigial pages columns)",
		})
		return result, nil
	}

	if !hasActivePages {
		// Nothing to reassemble, and chrome absence for a site with no live
		// pages is not a claim we have grounds to make either way — no
		// finding, no work item, no retraction.
		dctx.Logger.Info("missing_structure: chrome incomplete but site has no active pages, skipping",
			zap.Strings("missing_slots", missingSlots))
		return result, nil
	}

	dctx.Logger.Warn("missing_structure: site chrome incomplete",
		zap.Strings("missing_slots", missingSlots))

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":         "missing_structure",
		"missing_slots": missingSlots,
		"detail":        "site_components rows for these slots are absent or hold empty rendered_html",
	})

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":                   "missing_structure",
		"refresh_site_components": true,
		"missing_slots":           missingSlots,
		"reason": fmt.Sprintf(
			"site_components rows for %s are absent or hold empty rendered_html — chrome cannot "+
				"assemble until render_site_components runs", strings.Join(missingSlots, ", ")),
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: "build",
		ItemType: "needs_rerender",
		Severity: "high",
		Summary: fmt.Sprintf("Site chrome incomplete: %s missing or empty in site_components — full reassembly needed",
			strings.Join(missingSlots, ", ")),
		SpecJSON:     string(specJSON),
		Priority:     30, // high priority — structural issue visible to users
		HandlerAgent: "rerender-pages",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      missingStructureItemKey,
		BatchID:      dctx.BatchID,
	})

	return result, nil
}
