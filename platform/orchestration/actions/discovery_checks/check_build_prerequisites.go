// FILE: platform/orchestration/actions/discovery_checks/check_build_prerequisites.go
//
// Discovery check: build_prerequisites — the PREREQUISITES seat of the site
// acceptance council (RFC_056, loanzy_uk_example_site lane, 2026-08-25).
//
// WHAT IT ASKS, per site: were the things a good site is BUILT FROM ever
// produced? Not "is the site good" — that is the other seats' question — but
// "did the research, the claims register and the feed enrolment that a good
// site presupposes actually exist when the build ran, and do they exist now?"
//
// WHY IT EXISTS. Every one of the six broken routes found on homegarden.uk on
// 2026-08-25 was an ABSENCE with status 'complete': the vertical exemplar
// research never ran, no page ever requested per-page research, the claims
// register was never minted, the site was never enrolled for feeds — and every
// work item and orchestration on the way through read green, because the
// estate has no row for "research never ran". A thing that was never produced
// leaves no failure to sweep, no error to triage and no artefact to verify, so
// nothing anywhere has ever filed it. This seat is the row that absence gets.
//
// FOUR KINDS, each one SQL predicate with $1 = site_id, issued verbatim and
// recorded verbatim in the spec so a reader can re-run it by hand:
//
//	vertical_landscape  site_specs aspect='vertical_landscape', is_current
//	                    — the best-of-niche exemplar research, written by
//	                    vertical-exemplar-researcher (migration 341 tells the
//	                    strategist to weigh it heavily; a site without it was
//	                    positioned blind).
//	page_research       any research_results row for the site — per-page
//	                    research, written by research-agent. [MEASURED
//	                    2026-08-25] research-agent has made 0 LLM calls in its
//	                    life on the greenfield route, because 0 pages request
//	                    it — so this absence is the route's, not the site's.
//	evidence_base       site_specs aspect='evidence_base', is_current, with a
//	                    non-empty JSON OBJECT in data — the claims register,
//	                    first written by verify_and_register_citations
//	                    (created_by 'evidence-researcher'). An empty '{}'
//	                    register is counted ABSENT: it satisfies "a row exists"
//	                    and gates nothing.
//	feed_sources        any active content_sources row — news-feed enrolment,
//	                    seeded by content-feed-orchestrator's seed_content_sources
//	                    step. bugs_open/316: an unenrolled site can never
//	                    become enrolled on its own, so this absence is a route
//	                    defect, not a site defect, and is recorded as such.
//
// FLEET BASELINE, so the first run is read against a known picture and not as
// an alarm. [MEASURED 2026-08-25] of 31 sites with status IN
// ('active','deployed'): vertical_landscape present on 23, research_results
// rows on 10, evidence_base on 19, active content_sources on 9. So on day one
// this seat files roughly 8 + 21 + 12 + 22 rows fleet-wide, most of them for
// page_research and feed_sources — which is the shape of the two route defects
// above, not 43 independent site problems.
//
// WHAT IT DOES NOT DO — three refusals, each deliberate:
//
//   - It NEVER MINTS. bugs_open/380 owner decision D1: no register minting and
//     no backfill — absence IS the cold posture, and a seat that manufactured
//     an empty evidence_base to satisfy itself would be the exact thing D1
//     forbids. This check only OBSERVES absence; the spec says so in as many
//     words so nobody promotes the row into a "make one" task.
//   - It NEVER DISPATCHES. Every item is a flag-only VERDICT row: empty handler
//     agent, status 'detected', never promoted (the promoter and triage both
//     require a handler). Nothing can supply a missing prerequisite
//     automatically without re-running the route that should have produced it,
//     and deciding to re-run a route is a human's call, not a fixer's.
//   - It NEVER INFERS FROM ABSENCE. A retraction (CheckResult.Resolved, RFC_010)
//     fires only when a kind is POSITIVELY observed present on this run; and if
//     ANY of the four predicates errors the check returns the error and files
//     nothing — no partial verdicts, no retractions — because a blinded check's
//     empty result is indistinguishable from a healthy site (016b §9).
//
// KEYING. One row per (kind, site): item_key "prerequisite_missing:<kind>:<site>",
// so the four kinds dedupe independently under idx_swi_dedup and each one
// self-clears under the SAME key on the next pass that observes it present.
// The item type is the key's prefix — the "{item_type}:{target}" contract the
// package's flag-only checks share (check_site_unreachable.go).
//
// Enablement: registered here; add `build_prerequisites` to a discovery agent's
// checks array by migration once this file's image has rolled — the runner
// hard-fails on a name the binary does not register.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func init() { Register(&BuildPrerequisitesCheck{}) }

type BuildPrerequisitesCheck struct{}

func (c *BuildPrerequisitesCheck) Name() string { return "build_prerequisites" }

// buildPrerequisiteKind is one thing a site is built from, with the predicate
// that observes it and the agent that normally supplies it.
type buildPrerequisiteKind struct {
	// Kind is the stable name used in the item key and the spec.
	Kind string
	// Supplier is the agent type that normally writes this prerequisite. Named
	// in the summary so a reader knows which route to look at — NOT a handler,
	// and never routed at (see the header's second refusal).
	Supplier string
	// Predicate is the SQL issued, verbatim, with $1 = site_id. It must return
	// exactly one boolean column.
	Predicate string
	// Absence is the one-line reading of a false predicate.
	Absence string
}

// buildPrerequisiteKinds is the fixed, ordered set this seat observes. The
// predicates are package-level values rather than inlined so a test can assert
// on the SQL actually issued — a sqlmock happy-path test returns whatever it is
// handed regardless of the WHERE clause, so the emptiness guard on evidence_base
// in particular can only be pinned by reading the text (the same reasoning as
// check_missing_structure_test.go's header).
var buildPrerequisiteKinds = []buildPrerequisiteKind{
	{
		Kind:      "vertical_landscape",
		Supplier:  "vertical-exemplar-researcher",
		Predicate: `SELECT EXISTS (SELECT 1 FROM site_specs WHERE site_id = $1 AND aspect = 'vertical_landscape' AND is_current)`,
		Absence:   "no current site_specs row for aspect 'vertical_landscape' — the best-of-niche exemplar research never ran",
	},
	{
		Kind:      "page_research",
		Supplier:  "research-agent",
		Predicate: `SELECT EXISTS (SELECT 1 FROM research_results WHERE site_id = $1)`,
		Absence:   "no research_results row at all — no page ever requested per-page research",
	},
	{
		Kind:      "evidence_base",
		Supplier:  "evidence-researcher",
		Predicate: `SELECT EXISTS (SELECT 1 FROM site_specs WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current AND jsonb_typeof(data) = 'object' AND data <> '{}'::jsonb)`,
		Absence:   "no current, non-empty evidence_base register — every claim on the site is ungated (bugs_open/380 D1: observed, never minted)",
	},
	{
		Kind:      "feed_sources",
		Supplier:  "content-feed-orchestrator",
		Predicate: `SELECT EXISTS (SELECT 1 FROM content_sources WHERE site_id = $1 AND is_active)`,
		Absence:   "no active content_sources row — the site was never enrolled for feeds (bugs_open/316: a route defect, not a site defect)",
	},
}

// buildPrerequisitesNotDispatchable is the spec note that stops a reader
// promoting the row. Same wording family as remit.go's capability_gap note,
// adapted for a verdict that is 'detected' rather than 'deferred'.
const buildPrerequisitesNotDispatchable = "flag-only VERDICT row: status 'detected' + empty handler_agent — deliberate; " +
	"never promote this row. Nothing can supply a missing prerequisite automatically, and " +
	"bugs_open/380 D1 forbids minting or backfilling one to satisfy the check (RFC_056)"

func (c *BuildPrerequisitesCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	observedAt := time.Now().UTC().Format(time.RFC3339)

	// Observe every kind BEFORE filing anything. If any predicate cannot be
	// judged the whole run is abandoned: partial verdicts would let a blinded
	// half of the check read as a clean bill, and a retraction issued beside
	// an error is exactly what RFC_010's safety property forbids.
	present := make([]bool, len(buildPrerequisiteKinds))
	for i, k := range buildPrerequisiteKinds {
		var ok bool
		if err := dctx.DB.QueryRowContext(dctx.Ctx, k.Predicate, dctx.SiteID).Scan(&ok); err != nil {
			return nil, fmt.Errorf("build_prerequisites: %s predicate failed: %w", k.Kind, err)
		}
		present[i] = ok
	}

	result := &CheckResult{}
	var missing []string

	for i, k := range buildPrerequisiteKinds {
		itemKey := fmt.Sprintf("prerequisite_missing:%s:%s", k.Kind, dctx.SiteID)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":    "build_prerequisites",
			"kind":     k.Kind,
			"present":  present[i],
			"supplier": k.Supplier,
		})

		if present[i] {
			// A positive observation — the only thing that may retract. Narrow
			// ItemKey, not AllOfType: each kind owns its own key and the other
			// three may still be absent on this very run.
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "prerequisite_missing",
				ItemKey:  itemKey,
				Reason: fmt.Sprintf("%s positively observed present on %s (discovery run %s)",
					k.Kind, observedAt, dctx.BatchID),
			})
			continue
		}

		missing = append(missing, k.Kind)

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":            "build_prerequisites",
			"kind":             k.Kind,
			"predicate":        k.Predicate,
			"supplier":         k.Supplier,
			"seat":             "prerequisites",
			"rfc":              "RFC_056",
			"observed_at":      observedAt,
			"absence":          k.Absence,
			"not_dispatchable": buildPrerequisitesNotDispatchable,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "prerequisite_missing",
			Severity: "medium",
			Summary: fmt.Sprintf("Prerequisite missing: %s — %s (normally supplied by %s)",
				k.Kind, k.Absence, k.Supplier),
			SpecJSON:     string(specJSON),
			Priority:     120,
			HandlerAgent: "", // verdict only; nothing can supply a prerequisite, and D1 forbids minting one
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      itemKey,
			BatchID:      dctx.BatchID,
		})
	}

	if len(missing) > 0 {
		dctx.Logger.Warn("build_prerequisites: site was built without prerequisites",
			zap.Strings("missing", missing),
			zap.Int("present", len(buildPrerequisiteKinds)-len(missing)))
	}

	return result, nil
}
