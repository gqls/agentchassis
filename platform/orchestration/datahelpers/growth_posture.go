// FILE: platform/orchestration/datahelpers/growth_posture.go
//
// Per-site GROWTH POSTURE — the owner's "this site is ready" switch
// (owner decision 5 of 2026-08-31, design in the loanzy lane's
// PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md,
// ADDITION 2026-08-31).
//
// The key: sites.settings->'maintenance_profile'->>'growth_posture'.
//   - absent or "open"  -> today's behaviour: growth items dispatch normally.
//   - "hold"            -> growth items are FILED but not DISPATCHED: they are
//     born status='deferred' with handler_agent='' — the exact shape the
//     record-mode council verdicts use, which the detected-item-promoter
//     refuses BY CONSTRUCTION (a handler-less row is excluded from its scored
//     CTE before any door is evaluated; a 'deferred' row is never selected at
//     all). Release is a human verb: set status/handler back by hand, recipe
//     stamped in the spec. Filing-not-skipping is deliberate — a skipped
//     suggestion leaves no record that the machinery wanted to grow the site,
//     and the owner's decision was to hold growth, not to blind it.
//
// The maintenance_profile family is the established home for per-site owner
// posture (the structure floor from migration 618 lives beside this key), so
// the vocabulary rhymes with the estate instead of adding a peer namespace.
//
// WHERE THE CHECK RUNS: the growth door in writeWorkItem
// (growth_posture_door.go), beside the owned-page policy door — the one seam
// EVERY filing passes (insertWorkItem is a thin wrapper over writeWorkItem,
// so discovery sweeps, config-driven create_work_item steps and direct Go
// callers all cross it). That placement was the council's round-1 point
// (corr 1e735fa2): a per-producer guard covers the producers you found, a
// door covers the ones you didn't — and no census over a rolling-window
// table has to stay true for the hold to hold.
//
// WHY THE TYPE SET IS SMALL [MEASURED 2026-08-31, 30-day census in the PLAN]:
// the audit seats' growth types (needs_content_page / needs_content_planning
// from the five model seats) have filed as record-mode verdicts since
// migration 624 — that half is already held. The still-dispatching growth is
// the TOOL CHAIN, whose heads are `evaluate_tools` (check_missing_tools) and
// `add_tool` (tool-suggester via create_work_item). Everything further down
// (tool-generator / tool-deployer filing the guide-page needs_content_page
// rows) runs only as a consequence of an add_tool executing, so holding the
// heads holds the chain — and because the door is at the shared seam, a
// third producer of either type is held too, wherever it files from.
//
// `source == "owner-request"` BYPASSES the hold: the owner asking for a tool
// is not growth to refuse (the webdesign-tool-rebuilds lane files add_tool
// with that source; the census shows no other producer uses it).

package datahelpers

import (
	"context"
	"database/sql"
	"fmt"
)

// GrowthPostureHold is the settings value that holds growth; any other value
// (or an absent key) reads as open. Deliberately a single unsafe-side-OFF
// opt-in string rather than a boolean so a future posture ("review", say) is
// a value, not a schema change.
const GrowthPostureHold = "hold"

// GrowthPostureSourceBypass names the one source whose growth items are never
// held: an explicit owner request.
const GrowthPostureSourceBypass = "owner-request"

// GrowthGatedItemTypes is the set of item types that consult the posture at
// filing time. Exactly the two HEADS of the tool chain — see the file header
// for why the downstream types are deliberately absent.
var GrowthGatedItemTypes = map[string]bool{
	"evaluate_tools": true,
	"add_tool":       true,
}

// GrowthGateApplies is the pure half of the decision: does this filing consult
// the posture at all? Split from the DB read so the bypass and the type set
// are provable without a database — a mocked "query never ran" assertion is
// vacuous when the caller fails open on an unexpected-query error.
func GrowthGateApplies(itemType, source string) bool {
	return GrowthGatedItemTypes[itemType] && source != GrowthPostureSourceBypass
}

// GrowthPostureQuerier is the narrow read surface SiteGrowthPosture needs —
// same pattern as PlanSectionQuerier in this package. *sql.DB and *sql.Tx
// both satisfy it, which matters because the caller is a door inside
// writeWorkItem's transaction.
type GrowthPostureQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// SiteGrowthPosture reads the site's growth posture. Absent settings, an
// absent key, or any value other than GrowthPostureHold all return "open" —
// the fail-open direction is deliberate (the switch is an opt-in hold; a read
// error must not silently stop fleet-wide tool growth).
func SiteGrowthPosture(ctx context.Context, db GrowthPostureQuerier, siteID string) (string, error) {
	var posture string
	err := db.QueryRowContext(ctx,
		`SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture', 'open')
		   FROM sites WHERE id = $1`, siteID).Scan(&posture)
	if err == sql.ErrNoRows {
		return "open", nil
	}
	if err != nil {
		return "open", fmt.Errorf("read growth_posture for site %s: %w", siteID, err)
	}
	if posture != GrowthPostureHold {
		return "open", nil
	}
	return GrowthPostureHold, nil
}
