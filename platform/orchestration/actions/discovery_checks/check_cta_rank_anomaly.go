// FILE: platform/orchestration/actions/discovery_checks/check_cta_rank_anomaly.go
//
// cta_rank_anomaly — the alarm half of bugs_open/436's owner-approved pairing
// (decision 3, 2026-08-25: "candidate 1 paired with candidate 4 — lever plus
// alarm"). The lever is `pages.eligible_as_cta_target`, read by the shared
// ranking; this check notices when a site's PRIMARY CTA has been won on a
// fossil `nav_order` — the shape that put a password toy on three
// consultancies' front pages for months (bugs_closed/391: the value was set
// at page creation, 2026-03-13, and nothing ever looked at it again).
//
// WHAT "ANOMALOUS" MEANS HERE, precisely, because a threshold nobody can
// defend is a threshold nobody trusts. The check fires when the site-level
// rank-1 CTA target (the header form of the ranking: no page excluded) is an
// interactive page that:
//   - holds the UNIQUE minimum nav_order among >= 3 eligible interactive
//     candidates (a shared minimum is the all-default state, which is
//     arbitrary but not anomalous — that shape is candidate 3's business,
//     deliberately out of scope per the 2026-08-25 decisions);
//   - sits BELOW the schema default of 100 (a fossil is a value someone once
//     set; the default is what nobody set); and
//   - leads the runner-up by >= 50 (a curated ladder — 10/20/30 — is a
//     deliberate ordering; a lone 1 against a pack at 100 is not).
//
// 391's fossil (1 vs 100) fires; the demoted state (pack at 100, fossil at
// 900) is silent; an all-default alphabetical winner is silent.
//
// THE RANKING IS THE REAL ONE, NOT A SIMULATION. Supply SQL and ordering come
// from datahelpers/cta_positional.go — the same constants and the same
// RankCTAPositionalCandidates the build-time resolver, the rerender recompute
// and the site header fallback use. The 391 lane's first fleet simulation
// omitted the linkability predicate and named winners the code would skip;
// its runbook's rule — "mirror the code exactly or the simulation proves
// nothing" — is why this check cannot carry its own copy of any of it.
//
// A DETECTOR, NOT A FIX (391 §Fix candidates, #4). It files ONE
// needs_human_review item naming the page and its remedies. It repairs
// nothing: the fossil may be deliberate — a flagship tool ranked first on
// purpose — and that judgement needs a human who knows the site's premise.
// When the condition clears (demotion, opt-out, acknowledgement, or the
// page's retirement), the next run positively observes a healthy rank-1 and
// retracts through the shared Resolved seam. That silencing is correct, not
// blinded: the alarm is "your primary button is fossil-ranked", and after any
// remedy it no longer is — or a person has said it is fine.
//
// THE THIRD REMEDY, AND WHY THE FIRST TWO WERE NOT ENOUGH (owner ruling,
// 2026-09-03). Until today this check offered exactly two remedies, and on a
// site whose fossil-shaped rank-1 is the RIGHT button both of them make the
// site worse. The worked case is boxingonline.com: `tool-fight-calendar` at
// nav_order 3 against a pack at 200 — the fossil shape exactly, and the
// correct primary button for a boxing site. Opting it out hands the button to
// a trivia quiz; demoting nav_order does the same AND moves the visible menu.
// The estate could say "never use this page as a CTA destination"
// (eligible_as_cta_target, 714) but had no way to say "this page SHOULD win,
// I have looked, stop asking". `pages.cta_rank_deliberate_nav_order` (750) is
// that sentence, and ctaRankAcknowledged below is the only thing that reads it.
//
// ⚠ AND "JUST DISMISS THE WORK ITEM" IS NOT A THIRD REMEDY — THIS FILE USED TO
// CLAIM IT WAS, AND THE CLAIM WAS FALSE. The comment on ItemKey below said
// "items dedup in ANY status … so the same page recurring at the SAME value
// after a human dismissed it stays dismissed". idx_swi_dedup is
//
//	UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL
//	  AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix',
//	                          'failed','unresolved','cancelled'])
//
// — the dedup slot is held ONLY by items in OPEN statuses. Closing an item as
// wont_fix or rejected RELEASES the key, and the next pass files an identical
// one. Measured on cv1.co.uk 2026-09-03: resolved to `complete` 10:00:35Z, a
// fresh identical item inserted 10:02:24Z. So dismissal is not durable, and
// perversely it is leaving the item OPEN that suppresses duplicates. Corrected
// at both sites, and 750 exists because of it.
package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&CTARankAnomalyCheck{}) }

type CTARankAnomalyCheck struct{}

func (c *CTARankAnomalyCheck) Name() string { return "cta_rank_anomaly" }

const (
	// ctaRankAnomalyMinSiblings — below this many eligible interactive
	// candidates, "anomalous against its siblings" has no siblings to mean.
	ctaRankAnomalyMinSiblings = 3
	// ctaRankAnomalyDefaultNavOrder mirrors the pages.nav_order column default
	// (COALESCE(nav_order,100) in the supply SQL): a fossil is a value someone
	// once SET, so only values below what nobody set qualify.
	ctaRankAnomalyDefaultNavOrder = 100
	// ctaRankAnomalyMinLead — the rank-1 must lead the runner-up by at least
	// this much. Separates a lone ancient outlier (1 vs 100) from a curated
	// ladder (10/20/30), which is a deliberate ordering this check must not
	// second-guess.
	ctaRankAnomalyMinLead = 50
)

func (c *CTARankAnomalyCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	interactive, err := datahelpers.LoadCTAPositionalCandidates(
		dctx.Ctx, dctx.DB, dctx.SiteID, datahelpers.CTAPositionalInteractiveSQL)
	if err != nil {
		return nil, fmt.Errorf("cta_rank_anomaly: %w", err)
	}
	// Site-level ranking: pageName "" is the header fallback's own call shape
	// (render_site_components_action.go:190) — no page excluded, so this IS
	// "the site's primary button" and not one page's view of it.
	ranked := datahelpers.RankCTAPositionalCandidates("", interactive)

	anomalous, detail := ctaRankAnomaly(ranked)

	// The acknowledgement is consulted ONLY once the shape has already been
	// judged anomalous, and only for the page that actually won. Asking first
	// would let a stale acknowledgement suppress a DIFFERENT page's fossil on
	// the same site; asking here means the column can only ever silence the
	// exact finding a person looked at.
	if anomalous {
		if ackNav, ok, err := ctaRankAcknowledged(dctx, ranked[0].Name); err != nil {
			// Do NOT swallow: a failed lookup must not read as "not
			// acknowledged", which would file an item a human already retired.
			// The runner fails the step, which is loud and correct — the
			// alternative is a silently re-opened finding nobody can close.
			return nil, fmt.Errorf("cta_rank_anomaly: acknowledgement lookup for %q: %w", ranked[0].Name, err)
		} else if ok {
			anomalous = false
			detail = fmt.Sprintf("rank-1 '%s' at nav_order %d is acknowledged as deliberate "+
				"(pages.cta_rank_deliberate_nav_order = %d)", ranked[0].Name, ranked[0].NavOrder, ackNav)
		}
	}

	if !anomalous {
		// Positive observation, not absence: the ranking RAN over the real
		// supply and its rank-1 is not fossil-shaped (or there is no rank-1).
		// AllOfType is stated rather than implied — a healthy site-level
		// rank-1 answers the whole item type for this site at once, whichever
		// page an earlier item named.
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType:  "cta_rank_anomaly",
			AllOfType: true,
			Reason:    "site rank-1 CTA target is no longer nav_order-anomalous: " + detail,
		})
		return result, nil
	}

	winner := ranked[0]
	dctx.Logger.Info("cta_rank_anomaly: fossil-shaped rank-1 CTA target",
		zap.String("page", winner.Name), zap.Int("nav_order", winner.NavOrder))

	spec := map[string]interface{}{
		"page_name": winner.Name,
		"page_url":  winner.URL,
		"nav_order": winner.NavOrder,
		"detail":    detail,
		"source":    "cta_rank_anomaly",
		"fix": "This page wins the site's primary CTA everywhere on a nav_order far below " +
			"its siblings — the fossil shape of bugs_closed/391. THREE remedies, and the " +
			"first is the right one whenever the button is actually correct: (1) if this IS " +
			"the button you want, record that — " +
			fmt.Sprintf("UPDATE pages SET cta_rank_deliberate_nav_order = %d WHERE site_id = '%s' AND name = '%s';",
				winner.NavOrder, dctx.SiteID, winner.Name) +
			" this check then retracts and stays silent until the page's " +
			"nav_order changes. (2) set pages.eligible_as_cta_target=false to remove it from " +
			"CTA candidacy entirely (nav placement and listings unaffected) — but note this " +
			"hands the button to the NEXT page in the same nav_order ordering, which may be " +
			"worse. (3) demote pages.nav_order to join the pack; this also moves the visible " +
			"menu where in_header=true. Do NOT simply close this item: the dedup index " +
			"releases the key on a terminal status, so it will be re-filed next pass.",
	}
	specJSON, _ := json.Marshal(spec)

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":     "cta_rank_anomaly",
		"page_name": winner.Name,
		"nav_order": winner.NavOrder,
		"detail":    detail,
	})
	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "cta_rank_anomaly",
		Severity: "medium",
		Summary: fmt.Sprintf("Primary CTA target '%s' wins on an anomalous nav_order (%s)",
			winner.Name, detail),
		SpecJSON: string(specJSON),
		Priority: 40,
		// review-only: no HandlerAgent, deliberately — whether a fossil is a
		// fossil needs a human who knows the site's premise (391's own
		// "one human glance" precedent for the borderline case).
		Status:    "needs_human_review",
		CreatedBy: dctx.AgentType,
		// Keyed page + nav_order so a NEW fossil value on the same page mints a
		// fresh item rather than hiding behind the old one.
		//
		// ⚠ CORRECTED 2026-09-03 — this comment used to say "items dedup in ANY
		// status … so the same page recurring at the SAME value after a human
		// dismissed it stays dismissed". THAT IS FALSE. idx_swi_dedup's WHERE
		// clause excludes seven terminal statuses (complete, verified, rejected,
		// wont_fix, failed, unresolved, cancelled), so a closed item releases
		// the key and the next pass re-files an identical one — measured on
		// cv1.co.uk 2026-09-03 (complete 10:00:35Z, fresh item 10:02:24Z). The
		// durable "I have accepted this" is migration 750's column, read by
		// ctaRankAcknowledged above; closing the item is not a substitute and
		// the fix text now says so.
		ItemKey: fmt.Sprintf("cta_rank_anomaly_%s_%d_%s", winner.Name, winner.NavOrder, dctx.SiteID),
		BatchID: dctx.BatchID,
	})
	return result, nil
}

// ctaRankAcknowledged reports whether a human has accepted THIS page winning
// the site's primary CTA AT ITS CURRENT nav_order (migration 750).
//
// The equality against the page's own live nav_order is the whole design, not
// a sanity check: it makes the acknowledgement SELF-EXPIRING. A boolean would
// silence this check for the page for ever, including for a future shape
// nobody reviewed — an acknowledgement outliving what it acknowledged. Storing
// the reviewed nav_order means renumbering the page lapses the acknowledgement
// and the alarm speaks again, which is exactly the granularity the work item
// key already uses (cta_rank_anomaly_<page>_<nav>_<site>).
//
// The comparison is done in SQL against COALESCE(nav_order,100) — the same
// expression the positional supply and the ranking use — rather than against
// the ranked candidate's NavOrder, so a future divergence between the two
// cannot silently widen what an acknowledgement covers.
//
// A missing row returns (0, false, nil): a page that has vanished between the
// ranking and this lookup is not acknowledged, and the caller's ordinary
// anomalous path is right for it.
func ctaRankAcknowledged(dctx DiscoveryCheckContext, pageName string) (int, bool, error) {
	const q = `
		SELECT cta_rank_deliberate_nav_order
		FROM pages
		WHERE site_id = $1 AND name = $2
		  AND cta_rank_deliberate_nav_order IS NOT NULL
		  AND cta_rank_deliberate_nav_order = COALESCE(nav_order, 100)`
	var ackNav int
	err := dctx.DB.QueryRowContext(dctx.Ctx, q, dctx.SiteID, pageName).Scan(&ackNav)
	switch {
	case err == sql.ErrNoRows:
		return 0, false, nil
	case err != nil:
		return 0, false, err
	}
	return ackNav, true, nil
}

// ctaRankAnomaly is the pure predicate over an already-ranked eligible
// interactive candidate list. Split from Run so the threshold semantics are
// testable without a database, and so a false verdict names WHY in the detail
// string (which the Resolved reason and the work item both carry).
func ctaRankAnomaly(ranked []datahelpers.CTAPositionalCandidate) (bool, string) {
	if len(ranked) == 0 {
		return false, "no eligible interactive candidates"
	}
	if len(ranked) < ctaRankAnomalyMinSiblings {
		return false, fmt.Sprintf("only %d eligible interactive candidate(s) — no sibling population to compare against", len(ranked))
	}
	first := ranked[0]
	second := ranked[1]
	if first.NavOrder >= ctaRankAnomalyDefaultNavOrder {
		return false, fmt.Sprintf("rank-1 '%s' nav_order %d is not below the default (%d)",
			first.Name, first.NavOrder, ctaRankAnomalyDefaultNavOrder)
	}
	if second.NavOrder == first.NavOrder {
		return false, fmt.Sprintf("rank-1 '%s' shares nav_order %d with '%s' — not a unique minimum",
			first.Name, first.NavOrder, second.Name)
	}
	if second.NavOrder-first.NavOrder < ctaRankAnomalyMinLead {
		return false, fmt.Sprintf("rank-1 '%s' leads by %d (< %d) — a curated ladder, not a lone fossil",
			first.Name, second.NavOrder-first.NavOrder, ctaRankAnomalyMinLead)
	}
	return true, fmt.Sprintf("'%s' nav_order %d vs runner-up '%s' at %d, among %d candidates",
		first.Name, first.NavOrder, second.Name, second.NavOrder, len(ranked))
}
