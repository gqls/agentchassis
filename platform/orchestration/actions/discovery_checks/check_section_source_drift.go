// FILE: platform/orchestration/actions/discovery_checks/check_section_source_drift.go
//
// Detects pages whose section list disagrees across the THREE stores that
// load_page_sections_from_spec reads in priority order:
//   1. site_plan_sections table (site_plans family) — AUTHORITATIVE
//   2. site_specs.site_plan aspect JSON (older planner generation)
//   3. pages.sections (materialised cache / what assembly & most edits touch)
//
// The build resolves the highest-priority source that has the page and SYNCS
// it DOWN over pages.sections. So if someone edits only pages.sections (or the
// aspect) while a higher source still holds a different list, the next rebuild
// silently reverts the edit and resurrects the old layout. That exact trap bit
// the robot-hands product-detail component swap (2026-07-15): migration 153
// updated pages.sections + the aspect, but not the authoritative table, so the
// rebuild brought the deleted components back.
//
// This check computes, per page, the EFFECTIVE authoritative list (table if
// present, else aspect if present) and compares it to pages.sections. A
// persistent mismatch means the page has not been rebuilt since the sources
// diverged — a latent revert waiting for the next build. Flag-only at
// needs_human_review (no handler auto-fixes a planning inconsistency; the
// resolution is a human aligning the sources, as migration 154 did).
//
// Ordered comparison: "sections" is a layout, so [a,b] != [b,a] is real drift.
//
// Registration: automatic via init() -> Register(&SectionSourceDriftCheck{})
// Enable: add "section_source_drift" to completeness-discovery-agent's
//   {workflow,steps,run_checks,config,checks} array.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func init() { Register(&SectionSourceDriftCheck{}) }

type SectionSourceDriftCheck struct{}

func (c *SectionSourceDriftCheck) Name() string { return "section_source_drift" }

// maxSectionDriftFlagsPerPass bounds noise on a badly-drifted site.
const maxSectionDriftFlagsPerPass = 25

func (c *SectionSourceDriftCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	tableSections, err := loadPlanTableSections(dctx)
	if err != nil {
		return nil, err
	}
	aspectSections, err := loadAspectSections(dctx)
	if err != nil {
		return nil, err
	}
	cacheSections, err := loadPagesCacheSections(dctx)
	if err != nil {
		return nil, err
	}

	result := &CheckResult{}
	emitted := 0

	// Iterate the cache set — every deployed page has a pages row. A page with
	// no cache entry but a plan entry is the sectionless case (owned by
	// check_sectionless_pages), not drift.
	for pageName, cache := range cacheSections {
		var authoritative []string
		var authSource string
		if t, ok := tableSections[pageName]; ok {
			authoritative, authSource = t, "site_plan_sections"
		} else if a, ok := aspectSections[pageName]; ok {
			authoritative, authSource = a, "site_specs.site_plan"
		} else {
			// No higher source — pages.sections is authoritative for this page;
			// nothing can silently override it. Not drift.
			continue
		}

		if orderedListsEqual(authoritative, cache) {
			continue
		}

		if emitted >= maxSectionDriftFlagsPerPass {
			dctx.Logger.Info("section_source_drift: per-pass cap reached",
				zap.Int("cap", maxSectionDriftFlagsPerPass))
			break
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":                "section_source_drift",
			"page":                 pageName,
			"authoritative_source": authSource,
			"authoritative":        authoritative,
			"pages_sections":       cache,
		})

		spec, mErr := json.Marshal(map[string]interface{}{
			"check":                "section_source_drift",
			"page_name":            pageName,
			"authoritative_source": authSource,
			"authoritative":        authoritative,
			"pages_sections":       cache,
			"reason": "the authoritative section source disagrees with pages.sections; " +
				"the next rebuild will overwrite pages.sections with the authoritative list, " +
				"reverting any edit made only to pages.sections (or only to the aspect)",
		})
		if mErr != nil {
			continue
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "section_source_drift",
			Severity: "medium",
			Summary: fmt.Sprintf("Section-list drift on page '%s': %s has [%s] but pages.sections has [%s]",
				pageName, authSource, strings.Join(authoritative, ", "), strings.Join(cache, ", ")),
			SpecJSON: string(spec),
			Priority: 130,
			// Flag-only: no handler auto-aligns planning sources; a human picks
			// the intended layout and updates all sources (cf. migration 154).
			HandlerAgent: "",
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("section_source_drift:%s", pageName),
			BatchID:      dctx.BatchID,
		})
		emitted++
	}

	if emitted > 0 {
		dctx.Logger.Info("section_source_drift: flagged pages with divergent section sources",
			zap.Int("count", emitted))
	}
	return result, nil
}

// loadPlanTableSections returns page_name -> ordered component list from the
// current plan's site_plan_sections table.
func loadPlanTableSections(dctx DiscoveryCheckContext) (map[string][]string, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sps.page_name, sps.component_name
		FROM site_plan_sections sps
		JOIN site_plans sp ON sp.id = sps.plan_id
		WHERE sp.site_id = $1 AND sp.is_current = true
		ORDER BY sps.page_name, sps.ordering
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("section_source_drift: plan table query failed: %w", err)
	}
	defer rows.Close()

	out := map[string][]string{}
	for rows.Next() {
		var page, comp string
		if scanErr := rows.Scan(&page, &comp); scanErr != nil {
			dctx.Logger.Warn("section_source_drift: table scan failed", zap.Error(scanErr))
			continue
		}
		out[page] = append(out[page], comp)
	}
	return out, rows.Err()
}

// loadAspectSections returns page_name -> ordered section list from the current
// site_specs.site_plan aspect JSON (pages[].sections).
func loadAspectSections(dctx DiscoveryCheckContext) (map[string][]string, error) {
	var planJSON []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true
	`, dctx.SiteID).Scan(&planJSON)
	if err != nil || planJSON == nil {
		// No aspect for this site (most sites) — not an error.
		return map[string][]string{}, nil
	}

	var plan struct {
		Pages []struct {
			Name     string   `json:"name"`
			Sections []string `json:"sections"`
		} `json:"pages"`
	}
	if json.Unmarshal(planJSON, &plan) != nil {
		return map[string][]string{}, nil
	}

	out := map[string][]string{}
	for _, p := range plan.Pages {
		if p.Name != "" && p.Sections != nil {
			out[p.Name] = p.Sections
		}
	}
	return out, nil
}

// loadPagesCacheSections returns page_name -> ordered section list from
// pages.sections (the materialised cache), for non-deleted pages.
func loadPagesCacheSections(dctx DiscoveryCheckContext) (map[string][]string, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT name, COALESCE(sections, '[]'::jsonb)::text
		FROM pages
		WHERE site_id = $1 AND COALESCE(status, '') <> 'deleted'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("section_source_drift: pages query failed: %w", err)
	}
	defer rows.Close()

	out := map[string][]string{}
	for rows.Next() {
		var name, sectionsJSON string
		if scanErr := rows.Scan(&name, &sectionsJSON); scanErr != nil {
			dctx.Logger.Warn("section_source_drift: pages scan failed", zap.Error(scanErr))
			continue
		}
		var sections []string
		if json.Unmarshal([]byte(sectionsJSON), &sections) != nil {
			continue
		}
		// Only pages that actually have a cached list can drift; an empty cache
		// with a plan entry is the sectionless case, owned elsewhere.
		if len(sections) > 0 {
			out[name] = sections
		}
	}
	return out, rows.Err()
}

// orderedListsEqual reports whether two section lists are identical in content
// and order. "sections" is a layout, so order is significant.
func orderedListsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
