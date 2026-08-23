// FILE: platform/orchestration/datahelpers/cta_label_universe.go
//
// ONE definition of "which pages may a CTA label name" (bugs_open/308, Phase B,
// register LNK-036).
//
// THE DEFECT THIS CLOSES. The `misdirected_cta` detector and the two CTA
// WRITERS both ask that question and, until this file, answered it from
// different sources — the detector from every page on the site, the writers
// from `candidatesFromHubs` (page_type section-index/tool/game, utility areas
// dropped). So `BestLabelMatch` inside the check could return `/contact.html`
// as the repair target while the same function inside `applyCTARecompute`
// could not: the candidate list handed to it did not contain one. The repair
// ran, kept the stored tool link, reported success, and the next discovery
// pass filed the same finding again. Measured 2026-08-23: **188** findings
// naming a utility destination, **99** of them on work items the platform had
// marked `complete`.
//
// bugs_open/203's extraction shared the CLASSIFIER (BestLabelMatch) and left
// the CANDIDATE UNIVERSE unshared, which is why the churn survived it. This
// file is the other half. A future consumer must call this rather than build
// its own list — that is the whole point, and it is why the loader lives in
// datahelpers rather than beside either caller.
//
// ⚠ WHAT THIS FUNCTION IS **NOT**. It is not the POSITIONAL pick's supply.
// `chooseCTATargets` (interactive-then-hub, `nav_order` ascending) still reads
// `loadContentHubs`/`loadInteractivePages` and still refuses every utility
// area in its own `rank()`. Two reasons that separation is load-bearing:
//
//   - A fresh positional pick landing on /contact.html has no evidence behind
//     it at all — nobody's copy asked for it. `bugs_open/308`'s verification
//     bar #3 states this directly: "a fresh POSITIONAL pick must still never
//     land on a utility page, whatever else changes."
//   - `loadContentHubs`/`loadInteractivePages` have a THIRD non-test caller
//     that no CTA bug file named until 2026-08-22:
//     `render_site_components_action.go:182-190`, the site HEADER's CTA
//     fallback, which takes `ordered[0]`. Widening at the loaders silently
//     re-picks every site's header button — and `site_components` holds **0**
//     `cta_url` keys across all **24** header rows (measured 2026-08-22), so no
//     `content_data` diff could ever show it. Widen HERE; never there.
//
// THE `linkable` PREDICATE IS PART OF THE DEFINITION, and adding it fixed a
// second instance of this bug's own shape: the detector's universe had no
// build-state filter, so it could name a page the writer's `validPages` gate
// then refuses. Measured 2026-08-23: **43** of 764 live pages are planned and
// never deployed, and **10** live findings name one — a suggestion no writer
// could ever have performed, for the same reason as the other 188.
//
// VALIDITY IS DELIBERATELY NOT SHARED. "May this label name that page?" and
// "may a link point at that page?" are different questions with different
// answers: the detector's validity set includes the homepage (a phantom test)
// while its candidate set excludes it (link copy rarely names the homepage),
// and the resolver's validity set (`loadResolverPageSet`) admits not-yet-
// deployed pages on purpose. Sharing those would be a merge of two axes, which
// is the mistake LNK-033 already records.
package datahelpers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// CTALabelUniverseSQL is the query behind LoadCTALabelUniverse, exported so a
// caller that must fold it into a larger read (the detector builds its
// validity set from the same table in the same pass) can do so without
// re-typing the predicates. $1 is the site id.
const CTALabelUniverseSQL = `
	SELECT p.id::text, p.name, COALESCE(p.title, p.name),
	       COALESCE(p.nav_label, ''), p.url,
	       COALESCE(p.page_type, 'content')
	FROM pages p
	WHERE p.site_id = $1
	  AND p.status NOT IN ('deleted', 'archived')
	  AND p.url IS NOT NULL AND p.url <> ''
	  AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status, '') = 'planned')
`

// CTALabelCandidateRow converts one row of CTALabelUniverseSQL into a
// candidate, applying the two membership rules that are NOT expressible in the
// query: the homepage is never a candidate (link copy rarely names it), and a
// page whose name/title/nav_label yield no distinctive token can never be
// matched at all. ok=false means "not a candidate", not "error".
func CTALabelCandidateRow(id, name, title, navLabel, url, pageType string) (LabelMatchCandidate, bool) {
	if name == "index" || name == "home" {
		return LabelMatchCandidate{}, false
	}
	return NewLabelMatchCandidate(id, name, title, url, pageType == "tool" || pageType == "game", navLabel)
}

// LoadCTALabelUniverse returns every page on the site a CTA label may name.
//
// A load failure is returned, never swallowed: both writers already degrade to
// "no candidates → leave the field alone", and the caller decides. Returning an
// empty list silently would make a DB blip indistinguishable from "this site
// has no pages worth naming", which is precisely how a repair path reports
// success while changing nothing.
func LoadCTALabelUniverse(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]LabelMatchCandidate, error) {
	rows, err := db.QueryContext(ctx, CTALabelUniverseSQL, siteID)
	if err != nil {
		return nil, fmt.Errorf("cta label universe query failed: %w", err)
	}
	defer rows.Close()

	var out []LabelMatchCandidate
	for rows.Next() {
		var id, name, title, navLabel, url, pageType string
		if err := rows.Scan(&id, &name, &title, &navLabel, &url, &pageType); err != nil {
			continue // one unreadable row must not cost the whole universe
		}
		if c, ok := CTALabelCandidateRow(id, name, title, navLabel, url, pageType); ok {
			out = append(out, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cta label universe iteration failed: %w", err)
	}
	return out, nil
}
