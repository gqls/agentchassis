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
// needs_human_review item naming the page and both remedies (demote
// nav_order, or set eligible_as_cta_target=false). It repairs nothing: the
// fossil may be deliberate — a flagship tool ranked first on purpose — and
// that judgement needs a human who knows the site's premise. When the
// condition clears (demotion, opt-out, or the page's retirement), the next
// run positively observes a healthy rank-1 and retracts through the shared
// Resolved seam. That silencing is correct, not blinded: the alarm is "your
// primary button is fossil-ranked", and after either remedy it no longer is.
package discovery_checks

import (
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
			"its siblings — the fossil shape of bugs_closed/391. If deliberate, dismiss this. " +
			"If not: demote pages.nav_order to join the pack, or set " +
			"pages.eligible_as_cta_target=false to remove it from CTA candidacy entirely " +
			"(nav placement and listings are unaffected). Note nav_order also orders the " +
			"visible menu where in_header=true, so a demotion moves the menu too.",
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
		// Keyed page + nav_order: items dedup in ANY status (bugs_open/326),
		// so the same page recurring at the SAME value after a human dismissed
		// it stays dismissed, while a NEW fossil value on the same page mints
		// a fresh item.
		ItemKey: fmt.Sprintf("cta_rank_anomaly_%s_%d_%s", winner.Name, winner.NavOrder, dctx.SiteID),
		BatchID: dctx.BatchID,
	})
	return result, nil
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
