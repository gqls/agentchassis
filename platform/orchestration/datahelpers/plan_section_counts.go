// FILE: platform/orchestration/datahelpers/plan_section_counts.go
//
// How many instances of each component does the PLAN say a page has?
//
// Exists for the content-duplication repair (council trail da3f2d9b, round-2/3
// bug_historian objection; owner decision 2026-07-31 to build rather than defer):
// remove_duplicate_page_sections deletes page_components rows, but page_components
// is DOWNSTREAM of the plan stores — a full (non-assemble) rebuild re-creates the
// page from the effective plan source, so deleting a row the plan still specifies
// is either wasted (the rebuild resurrects it) or a loss (on a page whose plan
// deliberately repeats a component — real: webdesign.co.uk's current plan lists
// info-card-grid twice on index). Same shape as bugs_closed/058 (rebuild path vs
// page_component locks) and bugs_closed/069 (site_components writers vs chrome
// locks): a destructive action on a downstream table that never consulted the
// authoritative one.
//
// THE PRIORITY WALK MIRRORS THE BUILD, NOT JUST THE TABLE. The build resolves a
// page's section list as (load_page_sections_from_spec_action.go:5-16):
//
//  1. site_plan_sections for the is_current plan   (authoritative when present)
//  2. site_specs.site_plan aspect JSON              (older planner generation)
//  3. pages.sections                                (materialised cache)
//
// Only 310 of 1,026 slot-named page_components rows resolve to a current-plan
// TABLE entry (measured 2026-07-31) — most sites live on the aspect or cache
// path, so a table-only guard would be blind on most of the fleet. Whichever
// source the build would read is the one whose repetition a rebuild will
// re-create, so it is the one the guard must honour.
//
// COUNTS, NEVER POSITIONS. page_components.position does not track
// site_plan_sections.ordering (verified on webdesign.co.uk: plan orders 0..3,
// component positions 1..4). The question the caller asks is "may this page drop
// below N rows of component X" — a per-(page, component) count answers it; a
// positional join does not.
//
// FAILS CLOSED. Any error reading any store is returned as an error. The caller
// is a destructive path: an unreadable plan must mean "refuse", never "proceed
// as if unplanned". (The one exception is a genuinely ABSENT source — no current
// plan row, no aspect row — which is a normal state, not an error; the walk
// falls through.)
package datahelpers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// PlanSectionQuerier is the narrow read surface this helper needs — same
// pattern as DocRekeyer above it in this package. *sql.DB and *sql.Tx both
// satisfy it.
type PlanSectionQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// PlanSpecifiedSectionCounts returns component_name -> number of instances the
// EFFECTIVE plan source specifies for one page, plus which source answered
// ("site_plan_sections", "site_specs.site_plan", "pages.sections", or "" when
// no source covers the page at all — an empty map with source "" means the plan
// stores are silent and impose no constraint).
func PlanSpecifiedSectionCounts(ctx context.Context, q PlanSectionQuerier, siteID uuid.UUID, pageName string) (map[string]int, string, error) {
	// 1. site_plan_sections for the current plan — the same query the build runs.
	counts := map[string]int{}
	rows, err := q.QueryContext(ctx, `
		SELECT sps.component_name
		FROM site_plan_sections sps
		JOIN site_plans sp ON sp.id = sps.plan_id
		WHERE sp.site_id = $1 AND sp.is_current = true AND sps.page_name = $2
	`, siteID, pageName)
	if err != nil {
		return nil, "", fmt.Errorf("plan guard: site_plan_sections read failed: %w", err)
	}
	for rows.Next() {
		var comp string
		if err := rows.Scan(&comp); err != nil {
			rows.Close()
			return nil, "", fmt.Errorf("plan guard: site_plan_sections scan failed: %w", err)
		}
		if comp != "" {
			counts[comp]++
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("plan guard: site_plan_sections iterate failed: %w", err)
	}
	if len(counts) > 0 {
		return counts, "site_plan_sections", nil
	}

	// 2. site_specs.site_plan aspect (pages[].name + pages[].sections[]) —
	// shape per check_section_source_drift.loadAspectSections.
	var planJSON []byte
	err = q.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true
	`, siteID).Scan(&planJSON)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// No aspect for this site — normal, fall through.
	case err != nil:
		return nil, "", fmt.Errorf("plan guard: site_specs aspect read failed: %w", err)
	case len(planJSON) > 0:
		var plan struct {
			Pages []struct {
				Name     string   `json:"name"`
				Sections []string `json:"sections"`
			} `json:"pages"`
		}
		// An aspect that exists but does not parse is a broken store, not an
		// absent one — fail closed rather than fall through to a lower source
		// the build would not have reached either.
		if uerr := json.Unmarshal(planJSON, &plan); uerr != nil {
			return nil, "", fmt.Errorf("plan guard: site_specs aspect unparseable: %w", uerr)
		}
		for _, p := range plan.Pages {
			if p.Name != pageName {
				continue
			}
			for _, s := range p.Sections {
				if s != "" {
					counts[s]++
				}
			}
		}
		if len(counts) > 0 {
			return counts, "site_specs.site_plan", nil
		}
	}

	// 3. pages.sections — the materialised cache, a jsonb string array. Lowest
	// priority, but a full rebuild of a page with no higher source reads it, so
	// its repetition is re-created too.
	var cacheJSON []byte
	err = q.QueryRowContext(ctx, `
		SELECT sections FROM pages
		WHERE site_id = $1 AND name = $2 AND sections IS NOT NULL
	`, siteID, pageName).Scan(&cacheJSON)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return map[string]int{}, "", nil // no source covers this page
	case err != nil:
		return nil, "", fmt.Errorf("plan guard: pages.sections read failed: %w", err)
	}
	var cache []string
	if uerr := json.Unmarshal(cacheJSON, &cache); uerr != nil {
		return nil, "", fmt.Errorf("plan guard: pages.sections unparseable: %w", uerr)
	}
	for _, s := range cache {
		if s != "" {
			counts[s]++
		}
	}
	if len(counts) == 0 {
		return counts, "", nil
	}
	return counts, "pages.sections", nil
}
