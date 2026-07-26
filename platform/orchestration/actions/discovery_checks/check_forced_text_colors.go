// FILE: platform/orchestration/actions/discovery_checks/check_forced_text_colors.go
//
// Finds sections with a dark background and no light text override — dark-on-dark
// text, which is unreadable. The detection is sound and stays.
//
// WHAT CHANGED 2026-07-26 (bugs_open/077, second instance)
// -------------------------------------------------------
// This check routed its items at HandlerAgent "forced-text-color-fixer". That
// agent HAS NEVER EXISTED. Measured on the live cluster that day, the only
// colour-related agent_definitions rows are color-variable-fixer and
// nav-link-fixer. So every item this check filed was destined to be marked
// 'blocked' at claim time ("Handler agent not registered",
// claim_work_item_action.go:126-168) — detection wired to nothing.
//
// The check IS enabled (design-discovery-agent's checks array), and there was one
// live match on 2026-07-26 (webdesign.co.uk), so this was not theoretical: the
// next discovery pass over that site would have filed it.
//
// It now files a capability_gap instead (remit.go) — the platform's durable
// "found work I have no handler for", which is read as a roadmap view by
// diagnose_triage_action.go and the fixloop digest, and cannot be dispatched.
// The finding is preserved; the false promise of a fixer is not.
//
// TWO THINGS THE NEXT THREAD NEEDS, and neither is obvious:
//
//  1. The gap is SMALLER than "build a fixer". The action fix_forced_text_colors
//     is already written and already registered (actions/registry.go:759). What is
//     missing is the agent_definitions row that names it. That is a seed, not a
//     build — which is why the code_pointers below say so.
//
//  2. Seeding it blind would re-create bugs_open/077 for this item type, so do not
//     just add the row. FixForcedTextColorsAction's remit is much narrower than
//     this detector's predicate: it only enters <style> blocks, only rewrites
//     rules whose SELECTOR matches a text element (h1-h6, p, li, blockquote,
//     strong, em, cite, span, dt, dd), explicitly returns container and link
//     selectors unchanged, and bails out of the whole strip step when its WCAG
//     pre-check fails. Whoever seeds it should partition this check the way
//     check_hardcoded_section_colors.go does — which needs the pure part of that
//     transform moved into this package first, as ReplaceHardcodedColors was.
//
// CHANGE: Added pc.locked_at IS NULL to skip locked components.

package discovery_checks

import (
	"fmt"
)

func init() { Register(&ForcedTextColorsCheck{}) }

type ForcedTextColorsCheck struct{}

func (c *ForcedTextColorsCheck) Name() string { return "forced_text_colors" }

func (c *ForcedTextColorsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	candidates, err := forcedTextColourPopulation(dctx)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return &CheckResult{}, nil
	}

	// No partition to make: with no registered handler the remit is empty by
	// definition, so the whole population is residue. When the agent is seeded,
	// this becomes a PartitionByRemit call like the hardcoded-colours check's.
	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":            "forced_text_colors",
			"components_found": len(candidates),
			"handler":          "none registered — filed as a capability gap",
		}},
		WorkItems: []WorkItemSpec{
			CapabilityGapItem(dctx, CapabilityGap{
				Check:         "forced_text_colors",
				Pipeline:      "design",
				BuilderNeeded: "forced-text-color-fixer",
				GapKind:       GapHandlerMissing,
				Capability: "give dark-background sections a light text override. The action " +
					"fix_forced_text_colors already exists and is registered; what is missing is the " +
					"agent_definitions row naming it — and a decision on remit, because that action " +
					"only rewrites text-element selectors inside <style> blocks and bails out entirely " +
					"below its WCAG contrast floor, so seeding it alone would file items it cannot fix.",
				Population:   len(candidates),
				Residue:      len(candidates),
				Examples:     candidates,
				CodePointers: forcedTextColourCodePointers,
			}),
		},
	}, nil
}

var forcedTextColourCodePointers = []map[string]string{
	{
		"path": "platform/orchestration/actions/registry.go",
		"why":  "fix_forced_text_colors is ALREADY registered here — the gap is a seed, not a new action",
	},
	{
		"path": "platform/orchestration/actions/fix_forced_text_colours_action.go",
		"why":  "the transform; its remit (text selectors only, <style> blocks only, WCAG bail-out) is narrower than this check's predicate",
	},
	{
		"path": "docs/agent_docs/sql_for_agents/056_colour_variable_fixer.sql",
		"why":  "the shape of the agent_definitions row to copy — one workflow step naming the action",
	},
	{
		"path": "platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go",
		"why":  "the partition pattern to follow once a handler exists, so this check does not repeat bugs_open/077",
	},
}

// forcedTextColourPopulation finds page_components that have dark background
// colors (in <style> blocks) but no corresponding light text color override.
// These sections render as dark-on-dark text — unreadable.
// Skips locked components — human-managed content is excluded.
func forcedTextColourPopulation(dctx DiscoveryCheckContext) ([]RemitCandidate, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, COALESCE(pc.slot_name, ''), COALESCE(pc.rendered_html, '')
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html LIKE '%<style%'
		  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-4][0-9a-fA-F]{5}'
		  AND pc.rendered_html NOT SIMILAR TO '%color:\s*(#[f-fF-F][0-9a-fA-F]{5}|#[e-fE-F][0-9a-fA-F]{5}|white|#fff)%'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("forced_text_colors query failed: %w", err)
	}
	defer rows.Close()

	var out []RemitCandidate
	for rows.Next() {
		var page, slot, html string
		if err := rows.Scan(&page, &slot, &html); err != nil {
			return nil, err
		}
		out = append(out, RemitCandidate{Key: page + "/" + slot, Body: html})
	}
	return out, rows.Err()
}
