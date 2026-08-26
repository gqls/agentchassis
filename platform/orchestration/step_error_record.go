// FILE: platform/orchestration/step_error_record.go
package orchestration

import (
	"fmt"
	"time"
)

// bugs_open/243 — `collected_data.__step_error` is a SINGLE key that
// routeToErrorStep OVERWRITES on every routed failure, so in a workflow where two
// steps fail the first is erased and no downstream step can ask "did MY
// predecessor fail?".
//
// The council gate needs exactly that question answered. A review seat whose call
// failed leaves its result field ABSENT — indistinguishable from a seat the
// relevance filter skipped — and diagnose_council_decide must not conflate "an
// opinion we were owed and lost" (unreadable, which blocks an approval) with "not
// applicable" (an abstention, which does not). `__step_errors` is what lets it
// tell them apart.
//
// ── WHY THIS IS A FUNCTION AND NOT STILL INLINE (2026-08-26) ────────────────
// It was inline in routeToErrorStep, which needs a DB and a live state
// repository, so the WRITER had no test of its own while the READER
// (reviewStepFailed, in the actions package) had four. That asymmetry is the
// dangerous kind: the two sides agree on a contract — the map is keyed by PLAIN
// STEP NAME — that nothing asserted. A refactor of either could have silently
// returned the council to counting a lost seat as an abstention, which is the
// precise silent-approval hazard the design exists to prevent, and every test
// would still have passed.
//
// Extracted as a pure function on the map so it can be pinned directly. This
// mirrors errorRouteTermination in error_route_completion.go, which was pulled
// out of completeWorkflow for the same reason.
//
// NO BEHAVIOUR CHANGE: same cap, same marker, same key shape, same values.

// maxStepErrors bounds the accumulated record.
//
// BOUNDED because this path is fleet-wide and NOT council-scoped (the guardian
// seat's objection on council correlation 82f07fa6): routeToErrorStep is hit by
// every routed step failure in every workflow, and a loop that expands into many
// failing iterations produces a DISTINCT step name per iteration — so an
// unbounded map would grow collected_data without limit on exactly the runs that
// are already going badly.
//
// 50 is far above any real council round (17 seats), which is why the council
// case can never be truncated — a property reviewStepFailed's doc comment relies
// on. If a future panel approaches this, that comment must be revisited.
const maxStepErrors = 50

// stepErrorTruncatedKey marks a record that stopped admitting new steps. It is
// deliberately NOT a step name, so it cannot collide with one.
const stepErrorTruncatedKey = "__truncated"

// recordStepError accumulates one routed step failure into
// collected["__step_errors"], keyed by the failing step's name.
//
// It leaves `__step_error` completely alone — that key is read by 33 Go sites and
// 6 live agent configs (measured 2026-08-24), so its shape is not negotiable.
//
// At the cap it stops admitting NEW steps and records that it did. Re-failures of
// a step already present still update in place, so a retrying step reports its
// LATEST error rather than its first. Never silent: without the marker, a seat
// whose failure fell off the end would read as an abstention — the exact
// conflation this record exists to prevent.
func recordStepError(collected map[string]interface{}, failedStepName, errorMsg string, at time.Time) {
	if collected == nil || failedStepName == "" {
		return
	}

	stepErrors, _ := collected["__step_errors"].(map[string]interface{})
	if stepErrors == nil {
		stepErrors = map[string]interface{}{}
	}

	_, alreadyPresent := stepErrors[failedStepName]
	if alreadyPresent || len(stepErrors) < maxStepErrors {
		stepErrors[failedStepName] = map[string]interface{}{
			"message": errorMsg,
			"at":      at.UTC().Format(time.RFC3339),
		}
	} else if _, noted := stepErrors[stepErrorTruncatedKey]; !noted {
		stepErrors[stepErrorTruncatedKey] = map[string]interface{}{
			"message": fmt.Sprintf("step-error record capped at %d entries; later distinct steps are not listed", maxStepErrors),
			"at":      at.UTC().Format(time.RFC3339),
		}
	}

	collected["__step_errors"] = stepErrors
}
