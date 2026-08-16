// FILE: platform/orchestration/actions/triage_detect_items_action.go
//
// TriageDetectedItemsAction promotes discovery findings from status='detected'
// to status='triaged' with pipeline='build' so the dispatch loop picks them up.
//
// Discovery agents write items with different pipelines (content, design) and
// status='detected'. The dispatch loop filters item_pipeline='build' and
// status IN ('triaged', 'approved'). This action bridges the gap.
//
// Used by: THREE live agents, not one — `improvement-loop` (step
// `triage_findings`), `design-audit-agent` (`triage`) and `site-review-agent`
// (`triage`). The improvement loop calls the other two as children BEFORE
// running its own copy.
//
// ┌── READ THIS BEFORE BRANCHING ON THIS ACTION'S OUTPUT (bugs_open/150) ──────┐
// │ `has_items` is CALL-SCOPED: it means "this invocation promoted at least    │
// │ one row", nothing more. The promotion below is site-wide over TYPES        │
// │ (site_id + status, no type filter — it is filtered only on whether the row │
// │ can be routed at all, see the guard), so the FIRST copy to run takes every │
// │ row and every later copy honestly reports promoted: 0. A branch reading    │
// │ `has_items` therefore asks "did I personally promote something?" when what │
// │ it means to ask is "does this site have work?" — which is how one measured │
// │ run promoted 67 findings and terminated on "No issues found — site is      │
// │ clean", skipping its own closing rerender and dispatch.                    │
// │                                                                            │
// │ `site_dispatchable` / `site_dispatchable_count` are the SITE-SCOPED        │
// │ answer, and they are what a conditional step should read. They are correct │
// │ whoever promoted, in whatever order, including a fourth caller that does   │
// │ not exist yet.                                                             │
// │                                                                            │
// │ `has_items` is kept, unchanged, because it is a fleet-wide convention      │
// │ across actions ("my own result set was non-empty") with other live         │
// │ consumers — build-dispatch-loop and site-work-orchestrator both branch on  │
// │ their own loaders' has_items, correctly. Redefining it here would fix one  │
// │ branch by making a shared word mean two things.                            │
// └────────────────────────────────────────────────────────────────────────────┘
//
// Config (literals):
//   - target_pipeline: string — pipeline to set on promoted items (default: "build")
//   - batch_id:      NOT IMPLEMENTED. Advertised here since this file was written,
//     but no code reads it (grep: this comment is its only occurrence). Setting it
//     on a step does nothing and always has. Left documented-as-absent rather than
//     deleted so the next reader does not re-add it to the spec believing it live.
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — which site's items to promote

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var TriageDetectedItemsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{},
	// Exactly one key, because exactly one is read (:84). Opting in makes
	// bugs_open/136 visible: all three live callers set `target_domain`, which
	// nothing reads, while `target_pipeline` is set by ZERO definitions
	// fleet-wide — the same domain→pipeline rename that half-landed on
	// run_discovery_checks. It is invisible today only because every caller asks
	// for "build" and the default at :83 is already "build".
	//
	// `batch_id` is NOT declared even though the header comment above advertises
	// it, because no code reads it — it is documentation of an intention, not of
	// a behaviour. Declaring it would make a dead key RECOGNISED and silence the
	// detector for it, which is the recorded WRONG_CALLS.md 2026-07-28 mistake
	// (committed by bugs_closed/101's own fix). If batch_id is ever implemented,
	// add it here in the same commit that reads it.
	ConfigKeys: []string{"target_pipeline"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
	// Declaring the key (above) made bugs_open/136 visible here; this closes it.
	// `target_domain` is what the sole live caller writes, and it is now read.
	//
	// No value changes today: improvement-loop asks for "build", which is also
	// the default. That is the point — the key stops being a coincidence and
	// starts being a specification, so changing it to "content" will do what it
	// says instead of nothing. (The three carriers named in the comment above
	// are now one: migration 286 removed the two child triage steps under
	// RFC 006. Re-measured 2026-08-08.)
	DeprecatedConfigKeys: map[string]string{"target_domain": "target_pipeline"},

	// This action promotes every ROUTABLE `detected` row on the site — no type
	// filter, no ownership filter, and since bugs_open/284 one filter that is not
	// a choice about scope: a row nothing can route is left where it is, because
	// promoting it only ever ends in `blocked`. Its blast radius is still the
	// site, not the run that called it, so a second live agent carrying this step
	// is a defect:
	// whichever copy runs first takes everything and every later copy reports
	// an honest `promoted: 0`.
	//
	// That is not hypothetical. It was a step in improvement-loop
	// (`triage_findings`), design-audit-agent (`triage`) and site-review-agent
	// (`triage`); the parent calls both children before its own copy, so its
	// copy always reported zero and `check_has_findings` ended the run on "No
	// issues found — site is clean" over a site holding 67 promoted findings
	// (bugs_closed/150). The two child steps were removed by migration 286
	// under RFC 006's owner ruling of 2026-08-02 — improvement-loop is the one
	// owner. This declaration is what stops the fan-out coming back: without
	// it, the next agent to gain a triage step re-creates the bug and nothing
	// reports it until a site is wrongly called clean.
	//
	// Detector: scripts/audit-single-owner-actions.sh (offline, read-only).
	SingleOwner: true,
}

func init() {
	datahelpers.RegisterActionInputSpec("triage_detected_items", TriageDetectedItemsInputSpec)
}

// ============================================================================
// ACTION: triage_detected_items
// ============================================================================

func TriageDetectedItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "triage_detected_items"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("TriageDetectedItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Resolve site_id ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		TriageDetectedItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// --- Config ---
	config := params.StepConfig.Config
	targetPipeline, _ := datahelpers.ResolveConfigSetting(
		config, TriageDetectedItemsInputSpec, "target_pipeline", "build", logger)

	// --- Promote detected → triaged ---
	// Sets pipeline to targetPipeline so the dispatch loop (which filters item_pipeline='build')
	// will pick them up. Preserves the original pipeline in spec.original_pipeline for auditing.
	//
	// ROUTABILITY GUARD (bugs_open/284). The promotion is still unconditional over
	// the site's TYPES — it promotes work, not a chosen list — but it will not
	// promote a row the claim path is going to refuse. workItemRoutableSQL renders
	// claim's own test, from the same function claim calls, so promotion and claim
	// cannot disagree about what is dispatchable.
	//
	// Without it, the platform's flag-only findings — the ones that name no handler
	// ON PURPOSE, because nothing on the platform can repaint a brand, restart a VM
	// or repoint an image reference — were promoted here, claimed by the dispatch
	// loop, and stamped `blocked` with "No handler_agent set". A correct finding
	// recorded as a routing failure, permanently: `blocked` is not terminal, so the
	// row goes on holding its dedup slot and the check that found it can never file
	// it again, and feasibility-recheck cannot release it because no agent type is
	// the empty string. 60 rows across 4 item_types and 15+ sites when this landed.
	updateSQL := fmt.Sprintf(`
		UPDATE site_work_items wi
		SET status = 'triaged',
		    triaged_at = now(),
		    spec = jsonb_set(
		        COALESCE(wi.spec, '{}'::jsonb),
		        '{original_pipeline}',
		        to_jsonb(wi.pipeline)
		    ),
		    pipeline = $2
		WHERE wi.site_id = $1
		  AND wi.status = 'detected'
		  AND %s
	`, workItemRoutableSQL("wi"))

	result, err := params.DB.ExecContext(ctx, updateSQL, siteID, targetPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to promote detected items: %w", err)
	}

	promoted, _ := result.RowsAffected()

	// --- Site-scoped state, for callers that need to know whether there is
	// work HERE rather than whether THIS call did any (see the header box).
	out := map[string]interface{}{
		"site_id":         siteIDStr,
		"promoted":        promoted,
		"target_pipeline": targetPipeline,
		"has_items":       promoted > 0,
	}

	// What the routability guard held back, NAMED. These rows are working as
	// designed — a flag-only finding waiting for a human — but a filter that
	// silently promotes fewer rows reads exactly like a site with less work, so
	// the count is reported and logged rather than left to be inferred. A failure
	// to count is not a failure to promote: log it and carry on.
	if held, byType, heldErr := countUnroutableDetected(ctx, params.DB, siteID); heldErr != nil {
		logger.Warn("TriageDetectedItemsAction: could not count unroutable detected items",
			zap.String("site_id", siteIDStr), zap.Error(heldErr))
	} else if held > 0 {
		out["not_promotable"] = held
		out["not_promotable_by_type"] = byType
		logger.Info("TriageDetectedItemsAction: held back detected items nothing can route",
			zap.String("site_id", siteIDStr),
			zap.Int64("not_promotable", held),
			zap.Any("by_item_type", byType),
		)
	}

	dispatchable, countErr := countDispatchableWorkItems(ctx, params.DB, siteID, targetPipeline)
	if countErr != nil {
		// FAIL TOWARD "NOT CLEAN". The count is an input to a branch whose
		// false side terminates the improvement loop on "site is clean" and
		// skips its closing rerender. If we cannot answer, the honest answer
		// is not "no work" — it is "we do not know", and of the two available
		// behaviours the loud one is right: a needless rerender costs a
		// render, a false clean costs the findings.
		logger.Error("TriageDetectedItemsAction: dispatchable count failed — reporting site_dispatchable=true so the caller does not terminate as clean",
			zap.String("site_id", siteIDStr),
			zap.String("target_pipeline", targetPipeline),
			zap.Error(countErr),
		)
		out["site_dispatchable"] = true
		out["site_dispatchable_count"] = int64(-1)
		out["site_dispatchable_error"] = countErr.Error()
	} else {
		out["site_dispatchable"] = dispatchable > 0
		out["site_dispatchable_count"] = dispatchable
	}

	logger.Info("TriageDetectedItemsAction: Complete",
		zap.String("site_id", siteIDStr),
		zap.Int64("promoted", promoted),
		zap.String("target_pipeline", targetPipeline),
		zap.Any("site_dispatchable_count", out["site_dispatchable_count"]),
	)

	return out, nil
}
