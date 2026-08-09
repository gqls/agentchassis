// FILE: platform/orchestration/actions/discovery_checks/check_premise_incomplete.go
//
// vigilant_designer_offer_analysis Programme B, phase B3 (PLAN 2026-08-02;
// owner decision 2026-08-08 promoted B1+B2, this is the first check behind them).
//
// A site's offer can only be judged against its recorded premise, and on
// 2026-08-08 the estate held sites with no premise at all: loancalculator.co.uk
// was live with 27 deployed pages and NO strategy aspect until a hand-fired
// refresh wrote its first one, and gaswholesalers.com carries the pre-2026-05
// shape with no revenue_models block. A missing premise is invisible to every
// other check — the strategic review (B1) says "recorded strategy is empty" in
// its summary and judges from the domain, which is honest but blind, and
// check_revenue_shape (this check's sibling) cannot run at all without
// revenue_models.primary_model.
//
// PREDICATE — a deployed site is premise-incomplete when EITHER:
//   - it has no current `strategy` row in site_specs, OR
//   - the row's revenue_models.primary_model is absent or empty (the old-shape
//     case; the field's PATH matters — it is nested under revenue_models, and
//     reading it at the top level produced a false "0/17 sites have it" claim,
//     WRONG_CALLS 2026-08-08).
//
// SHIPPED-ONLY, via datahelpers.PageHasShippedPredicateFor — never
// `build_status = 'deployed'` bare (a needs_rebuild page HAS deployed and is
// still serving, bugs_closed/037 — the commit-time pattern check caught this
// file re-typing that predicate by hand), and never sites.status (which reads
// 'active' on a site with 41 deployed pages, measured 2026-08-08): a greenfield
// site mid-build gets its strategy from the build chain
// (vertical-exemplar-researcher → needs_strategy → domain-strategist →
// briefing → site plan), and filing here would double-produce against that
// chain's own item.
//
// ROUTING: needs_strategy → domain-strategist. SAFE ONLY BECAUSE B2 SHIPPED
// FIRST: until migration 341 (slug: domain_strategist_refresh_safe...,
// 2026-08-08 — the number is ambiguous, bugfix_220 took the same pair),
// domain-strategist unconditionally chained needs_briefing → needs_site_plan,
// so this check would have queued a re-plan of every premise-poor LIVE site it
// noticed. The B2 gate completes without chaining when deployed pages exist —
// which is exactly this check's firing condition. Do not enable this check on
// any estate whose domain-strategist predates that gate.
//
// SECOND PRODUCER (owner ruling 2026-08-02 / RFC_010 §1): needs_strategy's
// existing producer is vertical-exemplar-researcher (greenfield builds; 3 rows
// all-history at 2026-08-08). This check is the second, for the deployed
// estate. The shared item_key shape is `strategy_<domain>` — matched exactly so
// the dedup index holds ONE open strategy request per site whichever producer
// files first. Producer set + key shape are stated in the concept-register
// entry (this commit) per the ruling.
//
// RETRACTION (RFC_010, opt-in): when this run POSITIVELY observed the premise
// complete (strategy row present AND primary_model non-empty), the site's open
// needs_strategy item is resolved by key. Never derived from an empty result —
// a query error returns error, not retraction.
//
// Verifier: none. needs_strategy is classified catJudgement in
// verifier_coverage_test.go ("strategy judgement") — completion is the
// strategist's own write, and "is the strategy any GOOD" is deliberately not
// this check's question (features_open/030 §6 keeps premise-quality grading
// out of scope). This check asserts existence + shape only.
//
// Registration: automatic via init(). Enable by adding "premise_incomplete" to
// quality-discovery-agent's checks array AFTER the image carrying this rolls
// (an unregistered name in the array is FATAL since bugfix_149 B4) — IMP-016
// order: observe-only first.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&PremiseIncompleteCheck{}) }

type PremiseIncompleteCheck struct{}

func (c *PremiseIncompleteCheck) Name() string { return "premise_incomplete" }

// premiseState is one site's premise, read in a single row.
type premiseState struct {
	Domain        string
	DeployedPages int
	HasStrategy   bool
	PrimaryModel  string // '' when absent — the incomplete signal
}

func readPremiseState(dctx DiscoveryCheckContext) (*premiseState, error) {
	var st premiseState
	var primary sql.NullString
	var hasStrategy sql.NullBool
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT s.domain,
		       (SELECT COUNT(*) FROM pages p
		         WHERE p.site_id = s.id AND `+datahelpers.PageHasShippedPredicateFor("p")+`),
		       EXISTS (SELECT 1 FROM site_specs ss
		         WHERE ss.site_id = s.id AND ss.aspect = 'strategy' AND ss.is_current = true),
		       (SELECT ss.data->'revenue_models'->>'primary_model' FROM site_specs ss
		         WHERE ss.site_id = s.id AND ss.aspect = 'strategy' AND ss.is_current = true
		         ORDER BY ss.created_at DESC LIMIT 1)
		FROM sites s WHERE s.id = $1
	`, dctx.SiteID).Scan(&st.Domain, &st.DeployedPages, &hasStrategy, &primary)
	if err != nil {
		return nil, err
	}
	st.HasStrategy = hasStrategy.Valid && hasStrategy.Bool
	if primary.Valid {
		st.PrimaryModel = primary.String
	}
	return &st, nil
}

func (c *PremiseIncompleteCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	st, err := readPremiseState(dctx)
	if err != nil {
		// Error, not empty result: an unreadable site must never look premise-complete
		// (and must never trigger retraction — the runner skips Resolved on error).
		return nil, fmt.Errorf("premise_incomplete: site state read failed: %w", err)
	}

	itemKey := fmt.Sprintf("strategy_%s", st.Domain)

	if st.DeployedPages == 0 {
		// Greenfield / mid-build: the build chain owns the premise. Not a finding,
		// and NOT a retraction either — this run did not observe the premise.
		return &CheckResult{}, nil
	}

	if st.HasStrategy && st.PrimaryModel != "" {
		// Positively observed complete — close any open request this or the other
		// producer filed (RFC_010 narrow retraction, by key).
		return &CheckResult{
			Resolved: []ResolvedFinding{{
				ItemType: "needs_strategy",
				ItemKey:  itemKey,
				Reason: fmt.Sprintf(
					"premise positively observed complete: current strategy row with revenue_models.primary_model=%q",
					st.PrimaryModel),
			}},
		}, nil
	}

	reason := "strategy row exists but revenue_models.primary_model is absent or empty (pre-2026-05 shape)"
	if !st.HasStrategy {
		reason = "no current strategy row in site_specs"
	}

	dctx.Logger.Info("PremiseIncompleteCheck: deployed site with incomplete premise",
		zap.String("site_id", dctx.SiteID.String()),
		zap.String("domain", st.Domain),
		zap.String("reason", reason))

	spec, _ := json.Marshal(map[string]interface{}{
		"check":          "premise_incomplete",
		"reason":         reason,
		"deployed_pages": st.DeployedPages,
		"refresh_safety": "domain-strategist's deployed gate (B2, 2026-08-08) completes without chaining needs_briefing on deployed sites",
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":  "premise_incomplete",
			"domain": st.Domain,
			"reason": reason,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "needs_strategy",
			Severity: "medium",
			Summary: fmt.Sprintf(
				"Deployed site has an incomplete premise (%s) — the offer track cannot judge it until domain-strategist writes one",
				reason),
			SpecJSON:     string(spec),
			Priority:     40,
			HandlerAgent: "domain-strategist",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      itemKey,
			BatchID:      dctx.BatchID,
		}},
	}, nil
}
