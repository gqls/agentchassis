// FILE: platform/orchestration/actions/discovery_checks/check_componentless_pages.go
//
// Discovery check: a page that is ACTIVE, has SHIPPED, and has sections planned
// for it, but has ZERO page_components rows. It serves its stale deployed
// artefact — in practice site chrome and nothing else — for as long as nobody
// looks.
//
// WHY THIS EXISTS: THREE NEAR-MISSES, THREE DIFFERENT BLINDNESSES
//
// Found 2026-07-30 because an owner opened robot-hands.com and clicked "Tools"
// in the nav. `/tools.html` was status='active', in_header=true, nav_label
// 'Tools', sections ["hero","tool-list","call-to-action"], page_components 0,
// build_status 'needs_rebuild' — serving the artefact deployed 2026-05-10, 81
// days stale. Three checks look like they should have caught it and none can:
//
//   - check_empty_sections   iterates FROM page_components. There are no rows to
//                            scan, so it reports nothing. Replayed against that
//                            site while the defect was live: 0 rows.
//   - check_sectionless_pages requires pages.sections to be NULL or [].
//                            This page has three.
//   - check_unresolved_sections requires build_status='deployed'.
//                            This page is 'needs_rebuild'.
//
// The generalisable point, and the reason a fourth check is the right answer
// rather than widening one of the three: **a component-driven detector cannot
// see component ABSENCE.** An empty set is not an empty section. Widening
// check_empty_sections would mean changing its FROM clause to pages and
// left-joining components, which inverts what that check is about and would
// re-home its slot-level findings; widening either of the others means deleting
// the predicate that defines it.
//
// Empirical confirmation rather than argument: all three discovery lanes were
// fired at that site on 2026-07-30 with all 57 configured checks enabled, and
// the number of work items they filed against that page was zero.
//
// SCOPE: `deployed_at IS NOT NULL`, and why that is the honest line.
//
// Measured fleet-wide when this was written: 15 active pages had sections and no
// components, of which 9 carried deployed_at. The other 6 are build_status
// 'planned' or were never deployed — a page mid-pipeline, or one whose first
// build never finished. That is a real but *different* and weaker defect: no
// visitor is being served anything wrong, because nothing is being served.
// Including them would put mid-build pages in the queue and make the check churn
// against the pipeline it shares a site with.
//
// So the predicate is the platform's existing "did this page ship / will it
// serve" line (queryresolve.DeployedPageEligibilitySQL, and the same reasoning as
// bugs_open/052): once deployed_at is stamped a page keeps serving its old
// artefact even while flagged needs_rebuild, which is exactly the state that
// makes this defect invisible and user-visible at the same time.
//
// Deliberately NOT gated on current-plan membership. check_sectionless_pages is,
// because its repair depends on borrowing a same-role sibling's layout and only
// planned pages have a sibling to borrow from. This check needs no sibling: the
// page's own sections array is intact, so page-build-handler can build it
// directly. Of the 15 pages measured, only 3 were in their site's current plan —
// a plan gate would have dropped six genuine, live, chrome-only pages.
//
// ITEM TYPE: reuses needs_content_page rather than adding a 78th type. Already
// emitted by check_sectionless_pages and check_component_standards, already
// classified in verifier_coverage_test.go (catCreation, "page existence"), and
// already handled by page-build-handler. A new type would have to be classified
// or verified there before the build would go green, for no gain.
//
// NO `mode` KEY, deliberately. load_existing_content_action.go:66 treats
// mode="recreate" as "load this page's existing content and preserve it" — the
// adoption path. These pages have no components and therefore no existing
// content to preserve, so setting it would ask that action to load nothing.
// Omitting it leaves load_existing_content a no-op and lets the normal
// build-from-sections path run, which is precisely the repair wanted.
//
// Registration: automatic via init() -> Register(&ComponentlessPagesCheck{})
// Enable: add "componentless_pages" to a discovery agent's
//   default_config {workflow,steps,run_checks,config,checks} array.

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&ComponentlessPagesCheck{}) }

type ComponentlessPagesCheck struct{}

func (c *ComponentlessPagesCheck) Name() string { return "componentless_pages" }

// componentlessMaxPerPass bounds a single pass. A site that has just had a bad
// regenerate could present many of these at once; rebuilding them all in one
// batch would swamp the build pipeline it shares with everything else. The
// remainder is picked up next pass — the population is stable, not a stream.
const componentlessMaxPerPass = 10

func (c *ComponentlessPagesCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	pages, err := findComponentlessPages(dctx)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check": c.Name(),
			"count": len(pages),
			"pages": pages,
		}},
	}

	for i, pg := range pages {
		if i >= componentlessMaxPerPass {
			// Say what was dropped. A silent cap reads as "that was all of
			// them" to the next person to look at the findings.
			dctx.Logger.Info("componentless_pages: per-pass cap reached; remainder next pass",
				zap.Int("cap", componentlessMaxPerPass),
				zap.Int("found", len(pages)),
				zap.Int("emitted", componentlessMaxPerPass))
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":     c.Name(),
				"skipped":   len(pages) - componentlessMaxPerPass,
				"reason":    "per-pass cap; remainder emitted next pass",
				"cap":       componentlessMaxPerPass,
			})
			break
		}

		specJSON, err := json.Marshal(map[string]interface{}{
			"check":     c.Name(),
			"page_name": pg.PageName,
			"page_type": pg.PageType,
			"page_url":  pg.URL,
			// The two facts that make this a defect rather than a plan state.
			"planned_sections": pg.SectionCount,
			"reason":           "page has sections planned but zero page_components; it is serving its stale deployed artefact",
		})
		if err != nil {
			return nil, fmt.Errorf("componentless_pages: marshal spec for %s: %w", pg.PageName, err)
		}

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(pg.PageID); perr == nil {
			pageIDPtr = &parsed
		}

		// in_header raises severity: a componentless page reachable from the
		// site nav is one a visitor finds by clicking, not by deep link. That is
		// how this class was found at all.
		severity, priority := "medium", 85
		if pg.InHeader {
			severity, priority = "high", 95
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   pageIDPtr,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "needs_content_page",
			Severity: severity,
			Summary: fmt.Sprintf(
				"Page %q is live with %d section(s) planned but no components — serving chrome only; rebuild it",
				pg.PageName, pg.SectionCount),
			SpecJSON:     string(specJSON),
			Priority:     priority,
			HandlerAgent: "page-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			// Distinct prefix from check_sectionless_pages' "sectionless_page:"
			// so the two cannot collide in idx_swi_dedup, and per-page so a
			// second pass re-files nothing while one is open.
			ItemKey: fmt.Sprintf("componentless_page:%s", pg.PageID),
			BatchID: dctx.BatchID,
		})
	}

	return result, nil
}

type componentlessPageFinding struct {
	PageID       string `json:"page_id"`
	PageName     string `json:"page_name"`
	PageType     string `json:"page_type"`
	URL          string `json:"url"`
	SectionCount int    `json:"section_count"`
	BuildStatus  string `json:"build_status"`
	InHeader     bool   `json:"in_header"`
}

// componentlessPagesQuery is a package-level const rather than an inline string
// so check_componentless_pages_test.go can assert the two guards below are still
// in it. Both are one-line deletions that leave every behavioural test green —
// sqlmock returns the rows a test hands it whatever the WHERE clause says — so
// the query text is the only place the guards can be pinned.
const componentlessPagesQuery = `
		SELECT p.id::text, p.name, COALESCE(p.page_type, ''), p.url,
		       jsonb_array_length(COALESCE(p.sections, '[]'::jsonb)),
		       COALESCE(p.build_status, ''), COALESCE(p.in_header, false)
		  FROM pages p
		 WHERE p.site_id = $1
		   AND p.status = 'active'
		   -- has shipped, so it is serving something to visitors right now
		   AND p.deployed_at IS NOT NULL
		   -- sections planned for it
		   AND jsonb_array_length(COALESCE(p.sections, '[]'::jsonb)) > 0
		   -- and nothing built to fill them
		   AND NOT EXISTS (
		       SELECT 1 FROM page_components pc WHERE pc.page_id = p.id
		   )
		 ORDER BY p.in_header DESC, p.name
	`

func findComponentlessPages(dctx DiscoveryCheckContext) ([]componentlessPageFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, componentlessPagesQuery, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("componentless_pages query failed: %w", err)
	}
	defer rows.Close()

	var findings []componentlessPageFinding
	for rows.Next() {
		var f componentlessPageFinding
		if err := rows.Scan(&f.PageID, &f.PageName, &f.PageType, &f.URL,
			&f.SectionCount, &f.BuildStatus, &f.InHeader); err != nil {
			dctx.Logger.Warn("componentless_pages: scan failed", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("componentless_pages rows iter failed: %w", err)
	}
	return findings, nil
}
