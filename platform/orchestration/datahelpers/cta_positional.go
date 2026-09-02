// FILE: platform/orchestration/datahelpers/cta_positional.go
//
// ONE definition of the POSITIONAL CTA pick's candidate supply and ordering
// (bugs_open/436, register LNK-041) — the sibling of cta_label_universe.go
// (LNK-036), which is the same move for the LABEL match.
//
// THE DEFECT THIS CLOSES. `chooseCTATargets` ranks every tool/game page by
// `COALESCE(nav_order,100)` then `name` and takes `[0]` — no topic, tag or
// semantic input at all — so whatever sorts first wins the primary button on
// every page of the site, and a fossil `nav_order` set years earlier picks an
// off-topic tool fleet-wide (bugs_closed/391: a password toy as the primary
// CTA on three consultancies, label-locked on 20 of 80 fields). The owner-
// approved lever (391 decision 3) is `pages.eligible_as_cta_target`: an
// explicit opt-out, read HERE at the ranking, so "this page must never be a
// CTA destination" is finally sayable.
//
// WHY THIS LIVES IN datahelpers AND NOT BESIDE ITS CALLERS. The paired
// detector (`cta_rank_anomaly`, discovery_checks) must compute the site's
// rank-1 CTA target EXACTLY as the writers do — "mirror the code exactly or
// the simulation proves nothing" (the 391 lane's runbook, proven when its
// first fleet simulation omitted the linkability predicate and named winners
// the code would skip) — and discovery_checks cannot import actions (actions
// imports it). check_misdirected_cta.go:66 already carries a hand-mirrored
// copy of the excluded-areas set for the same reason; this file is where both
// halves stop drifting.
//
// THE RANKING, NOT THE LOADERS, READS ELIGIBILITY — the 391 review's
// constraint 1, and it is load-bearing: the loaders have a third caller, the
// site HEADER's CTA fallback (`render_site_components_action.go:182-190`),
// whose output is never persisted (`site_components` holds 0 `cta_url` keys,
// measured 2026-08-22). Policy applied inside a loader's WHERE clause moves
// every site's header button with no diff anywhere to show it; policy applied
// in RankCTAPositionalCandidates is a named, unit-testable decision that all
// three callers share on purpose. The SQL below therefore SELECTS the flag
// and decides nothing with it.
package datahelpers

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// CTAPositionalCandidate is one page the positional CTA pick may consider.
// actions' contentHub is an alias of this type.
type CTAPositionalCandidate struct {
	Name     string
	Title    string
	URL      string
	Area     string // first path segment of URL ("/tools/x.html" -> "tools")
	NavOrder int
	// IneligibleAsCTATarget is `NOT pages.eligible_as_cta_target` — true only
	// when a row EXPLICITLY opts out (bugs_open/436). INVERTED from the column
	// on purpose: the zero value of this struct must mean "eligible", because
	// candidates are also built literally in tests and by callers that predate
	// the flag, and a zero value meaning "ineligible" would silently switch
	// every such candidate off — the exact failure shape LANDMINES records for
	// pin-vs-pool predicates. Set it only from the column; never infer it.
	IneligibleAsCTATarget bool
}

// CTAExcludedAreas names the utility areas a FRESH positional pick must never
// land in. Moved verbatim from actions' areasExcludedFromCTA (which now
// aliases this) so the discovery checks can read the real set instead of the
// hand-mirrored copy check_misdirected_cta carried. It governs the POSITIONAL
// PICK ONLY — the label universe deliberately offers utility pages so copy
// saying "Contact our supply team" can reach the contact page (bugs_open/308
// Phase B); see storedCTADestinationIsAuthored for why judging STORED values
// with this set was bugs_open/248's clobber.
var CTAExcludedAreas = map[string]bool{
	"about": true, "contact": true, "privacy": true, "terms": true, "legal": true,
}

// CTAExcludedDestination reports whether a URL lands in an area a fresh CTA
// pick should never point at. firstPathSegment cannot express this for
// top-level pages ("/contact.html" has no second slash), so it normalises
// first and strips the .html suffix: "/contact.html" -> "contact",
// "/legal/privacy.html" -> "legal".
func CTAExcludedDestination(url string) bool {
	p := strings.TrimPrefix(NormalizePagePath(url), "/")
	if i := strings.IndexByte(p, '/'); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimSuffix(p, ".html")
	return CTAExcludedAreas[p]
}

// FirstPathSegment("/tools/index.html") -> "tools"; "/index.html" -> "".
func FirstPathSegment(url string) string {
	trimmed := strings.TrimPrefix(url, "/")
	if i := strings.IndexByte(trimmed, '/'); i >= 0 {
		return trimmed[:i]
	}
	return ""
}

// The two positional supplies. Selection changes here reach the build-time
// resolver, the rerender recompute AND the site header fallback at once —
// which is why ELIGIBILITY IS SELECTED, NOT FILTERED: the WHERE clauses below
// answer "which pages of this type exist and may be linked", nothing more.
// The status arm is lifecycle; PageMayBeLinkedPredicateFor is the build-axis
// arm (a planned-and-never-deployed page is a live 404 — measured 2026-08-08,
// fundamentallyai.com served one for 18 days).
var (
	// CTAPositionalInteractiveSQL returns the site's tool/game pages — the
	// destinations a CTA prefers over a content hub when both exist. $1 = site.
	CTAPositionalInteractiveSQL = `
		SELECT name, COALESCE(title, name), url, COALESCE(nav_order, 100), eligible_as_cta_target
		FROM pages
		WHERE site_id = $1
		  AND page_type IN ('tool', 'game')
		  AND status IN ('active', 'deployed')
		  AND ` + PageMayBeLinkedPredicateFor("")

	// CTAPositionalHubsSQL returns the site's content hubs (section indexes).
	CTAPositionalHubsSQL = `
		SELECT name, COALESCE(title, name), url, COALESCE(nav_order, 100), eligible_as_cta_target
		FROM pages
		WHERE site_id = $1
		  AND page_type = 'section-index'
		  AND status IN ('active', 'deployed')
		  AND ` + PageMayBeLinkedPredicateFor("")
)

// LoadCTAPositionalCandidates runs one of the two supply queries above.
// A query or iteration failure is returned, never swallowed, and so is a
// SCAN shortfall (ScanShortfall, bugs_open/410): a thinned supply is worse
// here than an empty one, because dropping one row silently RE-RANKS the
// site — lose the rank-1 row to a projection drift and every button on the
// site moves with no error anywhere. All rows are attempted before the
// refusal so a mixed failure reports its full extent, not its first row.
func LoadCTAPositionalCandidates(ctx context.Context, db *sql.DB, siteID uuid.UUID, query string) ([]CTAPositionalCandidate, error) {
	rows, err := db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, fmt.Errorf("cta positional candidate query failed: %w", err)
	}
	defer rows.Close()

	var out []CTAPositionalCandidate
	offered := 0
	for rows.Next() {
		offered++
		var c CTAPositionalCandidate
		var eligible bool
		if err := rows.Scan(&c.Name, &c.Title, &c.URL, &c.NavOrder, &eligible); err != nil {
			// scan-loss:accepted: counted — ScanShortfall below refuses the
			// partial result. This continue is safe ONLY while that trailing
			// check survives; delete the ScanShortfall return and this branch
			// reverts to bugs_open/410's defect (a thinned supply silently
			// re-ranks the site). All rows are attempted first so a mixed
			// failure reports its full extent, not its first row.
			continue
		}
		c.Area = FirstPathSegment(c.URL)
		c.IneligibleAsCTATarget = !eligible
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cta positional candidate iteration failed: %w", err)
	}
	if err := ScanShortfall(offered, len(out), "cta positional candidates"); err != nil {
		return nil, err
	}
	return out, nil
}

// RankCTAPositionalCandidates is THE positional ordering: drop excluded
// areas/destinations, drop the page itself (a hero must not point at its own
// page), drop pages explicitly opted out of CTA targethood (bugs_open/436),
// then stable-sort by (NavOrder, Name). Callers take [0]/[1] of the
// interactive ranking followed by the hub ranking.
//
// The eligibility test is HERE and not in the SQL above — see the package
// comment. It binds every caller of the ranking at once, including the
// header fallback, and a page flipped back to eligible re-enters the very
// next ranking with no other state to clean up.
func RankCTAPositionalCandidates(pageName string, candidates []CTAPositionalCandidate) []CTAPositionalCandidate {
	ordered := make([]CTAPositionalCandidate, 0, len(candidates))
	for _, h := range candidates {
		if CTAExcludedAreas[h.Area] || CTAExcludedDestination(h.URL) {
			continue
		}
		if h.IneligibleAsCTATarget {
			continue
		}
		if pageName != "" && h.Name == pageName { // don't point a page's hero at itself
			continue
		}
		ordered = append(ordered, h)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].NavOrder != ordered[j].NavOrder {
			return ordered[i].NavOrder < ordered[j].NavOrder
		}
		return ordered[i].Name < ordered[j].Name
	})
	return ordered
}
