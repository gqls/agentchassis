// FILE: platform/orchestration/actions/discovery_checks/check_sectionless_pages.go
//
// Detects pages that are in the current site plan but have no sections
// (pages.sections is NULL or []) AND for which a same-role sibling in the
// plan has sections. Emits a needs_content_page work item routed to
// page-build-handler so the build runs again. The build's
// load_page_sections_from_spec sibling fallback then synthesises the layout
// from a same-role sibling and the page builds normally.
//
// Why this exists (durability): the read-time sibling fallback only helps a
// page that actually gets a build attempt. A page can be left sectionless
// with NO further attempt when:
//   - adoption convergence unions an adopted page back into the plan carrying
//     an empty sections array (nothing later re-plans it), or
//   - an earlier build died (claim-timeout) before sections were established
//     and was marked complete, so nothing retried it.
// This check is the retrigger that closes that gap.
//
// Why scoped to "has a same-role sibling with sections": that is exactly the
// set the load_page_sections_from_spec fallback can repair, so detection +
// fallback form a closed, self-healing loop with no churn. A sectionless page
// with no usable sibling is out of scope here (it cannot be auto-repaired from
// a sibling); insertWorkItem's two-strike rule already flags a persistently
// failing item as unresolved rather than looping.
//
// HandlerAgent is page-build-handler (NOT page-content-writer): the build
// handler is the workflow that runs load_spec_sections -> plan_sections, which
// is where the sibling fallback lives. Routing straight to the writer would
// bypass it.
//
// Registration: automatic via init() -> Register(&SectionlessPagesCheck{})
// Enable: add "sectionless_pages" to the completeness-discovery-agent
//   default_config {workflow,steps,run_checks,config,checks} array.

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&SectionlessPagesCheck{}) }

type SectionlessPagesCheck struct{}

func (c *SectionlessPagesCheck) Name() string { return "sectionless_pages" }

func (c *SectionlessPagesCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	pages, err := findSectionlessPages(dctx)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check": "sectionless_pages",
			"count": len(pages),
			"pages": pages,
		}},
	}

	for _, pg := range pages {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "sectionless_pages",
			"mode":      "recreate",
			"source":    "adoption",
			"page_name": pg.PageName,
			"page_type": pg.PageType,
		})

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(pg.PageID); perr == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "needs_content_page",
			Severity:     "high",
			Summary:      fmt.Sprintf("Page '%s' is in the plan with no sections — re-trigger build (same-role sibling layout available)", pg.PageName),
			SpecJSON:     string(specJSON),
			Priority:     90,
			HandlerAgent: "page-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("sectionless_page:%s:%s", pg.PageName, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

type sectionlessPageFinding struct {
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
	PageType string `json:"page_type"`
}

func findSectionlessPages(dctx DiscoveryCheckContext) ([]sectionlessPageFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		WITH cur AS (
			SELECT id AS plan_id FROM site_plans
			WHERE site_id = $1 AND is_current = true
		)
		SELECT p.id::text, p.name, COALESCE(p.page_type, '')
		FROM pages p
		JOIN cur ON true
		JOIN site_plan_pages spp
		  ON spp.plan_id = cur.plan_id AND spp.name = p.name
		WHERE p.site_id = $1
		  AND (p.sections IS NULL OR p.sections = '[]'::jsonb)
		  AND COALESCE(p.status, '') <> 'deleted'
		  AND EXISTS (
		      SELECT 1
		      FROM site_plan_pages sib
		      JOIN site_plan_sections ss
		        ON ss.plan_id = sib.plan_id AND ss.page_name = sib.name
		      WHERE sib.plan_id = cur.plan_id
		        AND sib.role IS NOT DISTINCT FROM spp.role
		        AND sib.name <> p.name
		  )
		ORDER BY p.name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("sectionless_pages query failed: %w", err)
	}
	defer rows.Close()

	var findings []sectionlessPageFinding
	for rows.Next() {
		var f sectionlessPageFinding
		if scanErr := rows.Scan(&f.PageID, &f.PageName, &f.PageType); scanErr != nil {
			dctx.Logger.Warn("Failed to scan sectionless page", zap.Error(scanErr))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
