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
	       COALESCE(p.page_type, 'content'),
	       p.eligible_as_cta_target
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
//
// eligibleAsCTATarget is the pages column verbatim. An OPTED-OUT page is
// deliberately still a candidate (bugs_open/436): the match must be able to
// FIND it and then REFUSE it — dropping it from the pool instead would let a
// weak-token runner-up win, the exact failure measured for the self-link rule
// (10 of 35 wrote somewhere else, most wrong; see BestLabelMatchForPage).
func CTALabelCandidateRow(id, name, title, navLabel, url, pageType string, eligibleAsCTATarget bool) (LabelMatchCandidate, bool) {
	if name == "index" || name == "home" {
		return LabelMatchCandidate{}, false
	}
	c, ok := NewLabelMatchCandidate(id, name, title, url, pageType == "tool" || pageType == "game", navLabel)
	if !ok {
		return c, false
	}
	c.IneligibleAsCTATarget = !eligibleAsCTATarget
	return c, true
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
		var eligible bool
		if err := rows.Scan(&id, &name, &title, &navLabel, &url, &pageType, &eligible); err != nil {
			continue // one unreadable row must not cost the whole universe
		}
		if c, ok := CTALabelCandidateRow(id, name, title, navLabel, url, pageType, eligible); ok {
			out = append(out, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cta label universe iteration failed: %w", err)
	}
	return out, nil
}

// BestLabelMatchForPage is BestLabelMatch with one extra rule: a label that
// names THE PAGE IT SITS ON names nothing. It reports ok=false, which for both
// CTA writers means their keep branches hold the stored value, and for the
// detector means no finding is filed.
//
// WHY THIS EXISTS — measured, not anticipated. The owner asked for the Phase B
// widening to be hand-audited before it rolled (2026-08-23). **35 of the 291
// writes it would have performed — 12% — pointed a button at the page it was
// already on**: "Read the full grip styles guide" on /blog/grip-styles.html
// resolving to /blog/grip-styles.html; "Read the policy" on /privacy.html
// resolving to /privacy.html; 33 more. The widening is what manufactures them,
// because a page's own copy is usually the best token match for that page and
// before Phase B most pages were not candidates at all.
//
// THE PLATFORM ALREADY AGREES A SELF-LINK IS A DEFECT, in three places, which is
// why this is a fix rather than a preference:
//   - chooseCTATargets' rank() has dropped `h.Name == pageName` since
//     2026-07-14 ("don't point a page's hero at itself");
//   - applyCTARecompute's keeps refuse a stored value equal to pageURL;
//   - check_misdirected_cta files "links back to its own page" as a finding —
//     so suggesting one as the REPAIR was the same defect from the other side.
//
// ⚠ REFUSING IS NOT THE SAME AS DROPPING THE PAGE FROM THE POOL, and the
// difference was measured too. Filtering the page out first and letting the
// runner-up win was tried: **25 of the 35 then matched nothing (correct — the
// button is left alone), but 10 wrote somewhere else, and most of those were
// wrong** — "Compare flight shapes" landing on the barrel-shapes guide,
// "Catch up on this week's darts news" on the dartboard setup guide,
// "Talk to FineTuning About Your Automation Audit" on a blog post about
// employment law. Once the best candidate is removed, a single shared token is
// enough for noise to win. So the answer is NO OPINION, exactly as for an
// ambiguous tie: copy that names its own page is a CONTENT defect (the button
// should not be there), not a destination the matcher should guess at.
//
// Match on NAME or URL because the two diverge and callers hold different ones:
// the build path knows `page_name` only, the repair path knows both, and a
// page's name is frequently not its URL stem (`grip-styles` lives at
// /blog/grip-styles.html). Pass "" for whichever you lack.
func BestLabelMatchForPage(label string, candidates []LabelMatchCandidate,
	pageName, pageURL string) (best LabelMatchCandidate, ok bool, ambiguous bool) {
	best, ok, ambiguous = BestLabelMatch(label, candidates)
	if !ok {
		return best, ok, ambiguous
	}
	if pageName != "" && best.Name == pageName {
		return LabelMatchCandidate{}, false, false
	}
	if pageURL != "" && NormalizePagePath(best.URL) == NormalizePagePath(pageURL) {
		return LabelMatchCandidate{}, false, false
	}
	// A page opted out of CTA targethood (pages.eligible_as_cta_target=false,
	// bugs_open/436) is REFUSED, not dropped from the pool: the label match
	// runs AHEAD of the positional pick at both writers, so without this the
	// opt-out has a hole exactly the shape of 391's lock-in — a page the
	// ranking refuses still wins through its own copy. Refusal, like the
	// self-link rule above, means NO OPINION: the keeps decide, then the
	// positional pick (which also refuses the page). Dropping the page from
	// the candidate list instead was measured wrong for self-links — once the
	// best candidate is removed, a single shared token lets noise win.
	if best.IneligibleAsCTATarget {
		return LabelMatchCandidate{}, false, false
	}
	return best, true, false
}
