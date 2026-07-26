// FILE: platform/orchestration/actions/discovery_checks/remit.go
//
// The shared seam for the defect class filed as bugs_open/077: a DETECTOR whose
// predicate is wider than its HANDLER's remit files work items no handler run
// can ever clear. They park as 'unresolved' under the two-strike rule
// (load_work_item_actions.go:1041) with the label "[unresolved after 2
// attempts]" — which asserts the handler FAILED, when in truth it was never able
// to succeed. That poisons 'unresolved', which is the fleet-wide "needs
// investigation" signal for every other item type.
//
// The detector's breadth is not the bug and must not be narrowed away — the wide
// predicate is the only thing that sees light hexes, 3-digit forms and inline
// style="" attributes at all. What was missing is that the detector never SPLIT
// what it saw. So:
//
//	population = in-remit ∪ residue
//	  in-remit  → the check's normal dispatchable item, counted honestly
//	  residue   → ONE capability_gap item: "found work I have no handler for"
//
// capability_gap is not a new item type invented here. It is the platform's
// existing durable record of exactly this, emitted since long before by
// WriteBuildItemsAction (load_work_item_actions.go:250-280) for page types whose
// builders do not exist yet; already classified in the build-enforced coverage
// guard (verifier_coverage_test.go); and already read as a ROADMAP view by
// diagnose_triage_action.go:355-378 and fixloop_digest_action.go:358. Reusing it
// means the residue is queued for a handler to be built, by machinery that
// already exists, without adding a 78th item type.
//
// WHY THE RESIDUE ITEM CANNOT BE DISPATCHED, twice over:
//
//	status 'deferred'   — claim_work_item_action.go:102 and
//	                      load_work_item_actions.go:559 both filter
//	                      status IN ('triaged','approved').
//	handler_agent ''    — the canonical spelling of "no handler" since migration
//	                      217; a row that somehow reached claim would be marked
//	                      blocked rather than handed to a fixer that cannot fix it.
//
// 'deferred' is NOT in workItemTerminalStatuses (work_items_common.go:29-44), so
// idx_swi_dedup holds one open capability_gap per site per check. Bounded
// population, no re-file churn, no repeat spend. stale-work-item-reaper only
// touches status='triaged' AND pipeline='build', so it cannot churn these either.
//
// Companion to 016b §9 "A 'did the fix work?' check must assert the HANDLER's
// remit, not the DETECTOR's predicate" — that entry scoped the VERIFIER to the
// handler and said in terms that it did nothing about the detector re-filing
// out-of-remit matches for ever. This is that half.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
)

// RemitCandidate is one member of a detector's population, carried with enough
// identity to name it in a summary. Body is the artefact the HANDLER transforms —
// which is not always the artefact the detector queried (check_broken_nav_links
// detects on rendered site_components and its handler rewrites the backing
// content_components.html_template; partitioning on the rendered html there would
// silently answer the wrong question).
type RemitCandidate struct {
	Key  string // human identity, e.g. "home/hero" or "header"
	Body string // what the handler's transform is applied to
}

// PartitionByRemit splits a detector's population by whether the handler's own
// transform would change it. inRemit is what a handler run can clear; residue is
// what no handler run can ever clear, however many times it is dispatched.
//
// Pass the handler's LITERAL transform, not a re-implementation of it — a mirror
// drifts, and the drift is invisible until it has filed a few hundred items. The
// three transforms shared this way today are ReplaceHardcodedColors (used by the
// fix_hardcoded_colors action, this check and its verifier) and
// ApplyNavLinkPatterns (used by fix_nav_link_templates and its check).
func PartitionByRemit(candidates []RemitCandidate, transform func(string) string) (inRemit, residue []RemitCandidate) {
	for _, c := range candidates {
		if transform(c.Body) != c.Body {
			inRemit = append(inRemit, c)
		} else {
			residue = append(residue, c)
		}
	}
	return inRemit, residue
}

// Gap kinds. The distinction matters to whoever picks the item up: a remit gap is
// a change to code that exists, a missing handler is a thing to build or seed.
const (
	// GapHandlerRemit — the handler exists and runs, but its transform provably
	// cannot touch this part of the detector's population.
	GapHandlerRemit = "handler_remit"

	// GapHandlerMissing — the routing names an agent that is not registered, so
	// every item would be marked blocked at claim
	// (claim_work_item_action.go:126-168). The whole population is residue.
	GapHandlerMissing = "handler_missing"
)

// CapabilityGap describes work the platform can SEE and cannot ACT on. It is the
// intake shape for the feature builder (designer → council → implementer → PR),
// whose spec gate wants code_pointers alongside the owner's approval — so a check
// that knows which function is too narrow should say so here rather than leave
// the next reader to find it.
type CapabilityGap struct {
	Check         string // the check that found it, e.g. "hardcoded_section_colors"
	Pipeline      string // "design" / "build" — the check's own pipeline
	BuilderNeeded string // agent that would have to exist, or widen, e.g. "color-variable-fixer"
	GapKind       string // GapHandlerRemit | GapHandlerMissing
	Capability    string // one line: what the handler would have to learn to do
	Population    int    // everything the detector matched
	Residue       int    // the part outside the handler's remit
	Examples      []RemitCandidate
	CodePointers  []map[string]string // {"path": ..., "why": ...}
}

// capabilityGapExampleCap bounds the spec payload — the item is a signal, not an
// inventory, and the population is re-derivable from the check at any time.
const capabilityGapExampleCap = 5

// CapabilityGapItem builds the residue's work item. Field choices mirror
// WriteBuildItemsAction's capability_gap (item_type, status 'deferred', priority
// 200, severity 'low', spec.builder_needed) so the two producers group together
// under triageGatherCapabilityGaps' COALESCE(spec->>'builder_needed', ...).
//
// It deliberately differs on ONE field: handler_agent is left empty where
// WriteBuildItemsAction sets the needed builder. That producer's builders do not
// exist at all; ours may exist and merely be too narrow, and naming a real agent
// on an undispatchable row is an invitation for someone to promote it — which
// re-creates 077 exactly. The grouping is unaffected because spec.builder_needed
// takes precedence in that COALESCE.
func CapabilityGapItem(dctx DiscoveryCheckContext, g CapabilityGap) WorkItemSpec {
	examples := make([]string, 0, capabilityGapExampleCap)
	for _, c := range g.Examples {
		if len(examples) >= capabilityGapExampleCap {
			break
		}
		examples = append(examples, c.Key)
	}
	sort.Strings(examples)

	spec := map[string]interface{}{
		"check":          g.Check,
		"gap_kind":       g.GapKind,
		"builder_needed": g.BuilderNeeded,
		"capability":     g.Capability,
		"population":     g.Population,
		"residue":        g.Residue,
		"examples":       examples,
		// Read by whoever picks this up, and by nobody automatically: this row is
		// a roadmap entry, not a dispatch. Said out loud so it is not mistaken for
		// an item the machinery dropped (bugs_open/083's class, which this is not).
		"not_dispatchable": "status 'deferred' + empty handler_agent — deliberate; " +
			"promoting this row dispatches work no handler can do (bugs_open/077)",
	}
	if len(g.CodePointers) > 0 {
		spec["code_pointers"] = g.CodePointers
	}
	specJSON, _ := json.Marshal(spec)

	return WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     g.Pipeline,
		ItemType:     "capability_gap",
		Severity:     "low",
		Summary:      capabilityGapSummary(g),
		SpecJSON:     string(specJSON),
		Priority:     200,
		HandlerAgent: "",
		Status:       "deferred",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("capability_gap:%s", g.Check),
		BatchID:      dctx.BatchID,
	}
}

// HandlerStepConfig reads the LIVE step config a handler agent would run for a
// given action, so a check can partition against the handler's actual remit
// rather than a plausible reading of the Go source.
//
// This exists because a handler's remit is not always all in Go. nav-link-fixer's
// find/replace patterns are seeded in agent_definitions and OVERRIDE the action's
// Go defaults entirely; measured 2026-07-26 the live row carries THREE patterns
// where the Go default list has four. A check partitioning on the Go defaults
// would credit the handler with a pattern it does not apply and file an item it
// cannot fix — bugs_open/077 again, one level down, and invisible to any test
// that only reads this repository.
//
// Returns (config, agentExists, error). The existence half deliberately uses the
// same predicate as claim_work_item_action.go's registration check — if that
// would mark an item 'blocked', this must call the handler missing, or the two
// disagree about what "has a handler" means.
func HandlerStepConfig(ctx context.Context, db *sql.DB, agentType, action string) (map[string]interface{}, bool, error) {
	var exists bool
	var raw []byte
	err := db.QueryRowContext(ctx, `
		SELECT
			EXISTS(SELECT 1 FROM agent_definitions
			        WHERE type = $1 AND deleted_at IS NULL),
			(SELECT step.value->'config'
			   FROM agent_definitions ad,
			        LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS step
			  WHERE ad.type = $1
			    AND ad.deleted_at IS NULL
			    AND COALESCE(ad.is_snapshot, false) = false
			    AND step.value->>'action' = $2
			  LIMIT 1)
	`, agentType, action).Scan(&exists, &raw)
	if err != nil {
		return nil, false, fmt.Errorf("handler step config lookup failed (%s/%s): %w", agentType, action, err)
	}
	if !exists || len(raw) == 0 {
		return nil, exists, nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, exists, fmt.Errorf("handler step config for %s/%s is not an object: %w", agentType, action, err)
	}
	return cfg, exists, nil
}

// capabilityGapSummary is the line a human or an agent reads in the backlog. It
// has one job: make it impossible to read this row as "a fixer keeps failing".
func capabilityGapSummary(g CapabilityGap) string {
	switch g.GapKind {
	case GapHandlerMissing:
		return fmt.Sprintf(
			"%d finding(s) from %s need agent %s, which is not registered — nothing can be dispatched until it exists",
			g.Residue, g.Check, g.BuilderNeeded)
	default:
		return fmt.Sprintf(
			"%d of %d finding(s) from %s are outside %s's remit — no handler run can clear them (capability gap, not a handler failure)",
			g.Residue, g.Population, g.Check, g.BuilderNeeded)
	}
}
