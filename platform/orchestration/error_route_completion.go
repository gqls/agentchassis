// FILE: platform/orchestration/error_route_completion.go
package orchestration

import (
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// bugs_open/354 — a run that ends at its ERROR TERMINAL used to be recorded with
// EXACTLY the same two columns as a run that executed every step: status
// 'COMPLETED' and error NULL. completeWorkflow's whole status decision was one
// unconditional line, so every consumer keying on status or error read a failed
// run as a success, permanently. The only surviving distinction was current_step
// (which nothing reads) and collected_data (deleted at 24h), while the real
// message sat in agent_error_log where nobody joins.
//
// This file holds the discriminator. It records the failure on the row; it does
// NOT change the status and does NOT change what the parent is told. That half is
// seam authority and belongs to 354's architecture RFC — bugs_closed/344 deferred
// it twice on exactly that ground, and RFC_023 records a hard guardian veto on
// SCOPE for a change that re-typed COMPLETED to FAILED on this same table.
//
// ── WHY A DECLARATION AND NOT A STRUCTURAL RULE ─────────────────────────────
// Three structural rules were tried and each was refuted by measurement against
// the 36 real rows in the live table on 2026-08-22:
//
//  1. "the error_step target is terminal" — misses page-build-handler, the top
//     producer at 22 of 36. Its routes point at a MID-GRAPH handler:
//     load_page_record --error_step--> mark_item_failed (which writes the work
//     item's failure, next_step: complete_error) --> complete_error.
//  2. "the terminal is unreachable from the start on normal edges" — misses it
//     too: check_page_found is a `conditional` whose else_step IS complete_error,
//     so the terminal has a perfectly ordinary in-edge.
//  3. "the run ends on the step the error route landed it on" — fires on 13 of
//     36, 36% coverage, missing that same top producer.
//
// They fail for one reason: the error handler is a REAL STEP THAT DOES WORK and
// only then flows into the terminal. So "did the run execute anything after the
// error route?" cannot separate the populations, because both do — the HONEST
// build-dispatch-loop runs the same kind of step (mark_failed). The only thing
// that differs is which terminal the run ends at, and that is AUTHORED, not
// structural. Hence a declaration.
//
// ── THE RULE, AND WHY BOTH HALVES ARE LOAD-BEARING ──────────────────────────
// The terminal step declares config.outcome = "error", AND __step_error.message
// is non-empty. Measured against every population in the live table:
//
//	36 dishonest (declared terminal + marker)  -> recorded    100%
//	13 honest    (recovered, end at 'complete') -> untouched
//	 2 skips     (declared terminal, NO marker) -> untouched
//
// Drop the declaration and the 13 honest runs are mislabelled — the trap named
// verbatim in UpdateWorkItemStatusAction (v3_site_actions.go): "__step_error is
// never cleared once set, so a workflow that recovers from a routed error and
// then stamps the item complete would otherwise be given a stale failure".
// Drop the marker and the 2 skip rows are mislabelled: those reach complete_error
// via check_page_found's else branch — a page genuinely not found, which is a
// SKIP, not a failure (bugs_closed/299: "a skip is a third state, neither clean
// nor failed"). The bug's own §2(b) control turned out to contain the evidence
// for that distinction.
//
// ── WHY config.outcome AND NOT A STEP-LEVEL FIELD ───────────────────────────
// convertToWorkflowPlan (platform/messaging/processor.go) builds models.Step from
// an explicit field whitelist and silently drops anything not listed — that is
// precisely the bugs_closed/086 mechanism, which left every step-level error_step
// inert fleet-wide. It copies `config` WHOLE (step.Config = config), and so does
// every other copier (substep_decode, loop-expansion cloneConfig). A config key
// therefore rides through with no converter change and no new drop hazard. If
// 354's RFC later grants `outcome` control-flow authority, promote it to a typed
// field THEN, keeping the config fallback — which is error_step's own pattern.
//
// ── RESIDUALS, NAMED SO NOBODY REDISCOVERS THEM ─────────────────────────────
//   - A run that routes to an error handler, RECOVERS onto the main path, and
//     then skip-exits into a declared terminal will carry a stale message: both
//     conjuncts hold. Zero instances in the measured population, and the cost is
//     bounded to prose in the error column.
//   - A terminal nobody has declared is missed, and the miss looks exactly like a
//     clean run — this bug's own failure mode, one level up. The standing audit
//     for it is in bugs_open/354: COMPLETED rows carrying __step_error at
//     UNDECLARED terminals. That query measures precisely the seeding gap.
//   - Runs whose action REFUSES without returning an error never reach
//     routeToErrorStep at all, so they carry no __step_error and nothing here can
//     see them (LANDMINES.md records two such shapes). Different defect.

// errorTerminalOutcome is the only value of config.outcome that activates the
// record. Read strictly: any other value is inert, so a typo or a future
// vocabulary fails toward today's behaviour rather than toward stamping.
const errorTerminalOutcome = "error"

// errorRouteTermination reports whether the run now completing ended at a
// terminal its author declared to be an error terminal, having actually suffered
// a routed step failure — and returns the message to record on the row.
//
// Evaluated at completion time deliberately. Plan-time terminality is a
// PREDICTION, not a fact: getNextStepFromResult (coordinator.go) lets any action
// return next_step / next_step_override at runtime and it wins over the plan, so
// a step that looks terminal can continue and one that looks mid-graph can end
// the run. Here, "this run is ending" is already established by the caller.
func errorRouteTermination(state *OrchestrationState) (string, bool) {
	if state == nil {
		return "", false
	}

	// The declaration lives on the step the run is ending ON. A step absent from
	// the plan (loop substeps are injected under prefixed names, but a plan can
	// still be sparse) simply does not declare, so this is inert rather than an
	// error.
	step, exists := state.WorkflowPlan.Steps[state.CurrentStep]
	if !exists {
		return "", false
	}
	outcome, _ := step.Config["outcome"].(string)
	if outcome != errorTerminalOutcome {
		return "", false
	}

	// The evidence half. ExtractNestedFieldString is already hardened against
	// __step_error being a bare string, a number or an array (see
	// update_work_item_status_owned_refusal_test.go) — reused rather than
	// re-implemented so the two readers cannot drift.
	msg := datahelpers.ExtractNestedFieldString(state.CollectedData, "__step_error.message")
	if msg == "" {
		return "", false
	}

	// Same shaping convention as UpdateWorkItemStatusAction: name the failing step
	// unless the message already does. The routed message has two shapes —
	// "step X failed: …" for action errors, and a bare "Request <id> timed out
	// after N retries" for awaited-request timeouts, which alone does not say WHAT
	// timed out. Converge on the prefix rather than inventing a second format.
	if failedStep := datahelpers.ExtractNestedFieldString(state.CollectedData, "__step_error.failed_step"); failedStep != "" &&
		!strings.HasPrefix(msg, "step ") {
		msg = "step " + failedStep + " failed: " + msg
	}

	return msg, true
}
