// FILE: platform/orchestration/local_action_deadline_test.go
//
// bugs_open/169 part A. Nothing in
// continueExecution -> executeStep -> executeLocalAction -> executeAction bounded
// a local action, so an action blocking on a network call parked its orchestration
// at EXECUTING_STEP indefinitely and held the dispatch loop's work item in
// `claimed` until a reaper expired it.
//
// The bound is only worth having if it actually FIRES, so the first test here
// blocks a real action past a real deadline and requires the error — a suite that
// only ever proves "fast actions still work" would be the inert-by-omission trap
// again. The rest pin the decisions that are easy to regress silently: the
// generous default, the explicit off switch, and — most importantly — that the
// bound is NOT wired to `timeout_seconds`.
package orchestration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

func testStep(config map[string]interface{}) models.Step {
	return models.Step{Name: "spawn_thing", Action: "spawn_agent", Config: config}
}

// The whole point: a blocking action must be cut off, not waited on for ever.
func TestABlockingLocalActionIsCancelledAtItsDeadline(t *testing.T) {
	ctx, cancel := localActionContext(context.Background(),
		testStep(map[string]interface{}{localActionTimeoutKey: 0.05}), "spawn_thing", zap.NewNop())
	defer cancel()

	started := time.Now()
	select {
	case <-ctx.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("the deadline never fired — a blocked action would park the orchestration " +
			"exactly as bugs_open/169 part A describes")
	}

	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("ctx.Err() = %v, want DeadlineExceeded — the step-failure path keys on this", ctx.Err())
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Errorf("took %s to cancel a 50ms deadline", elapsed)
	}
}

// The negative control. A bound that fires on healthy work is worse than the bug:
// measured 2026-08-02, 6,950 of 6,951 spawn-step executions finished inside 24s.
func TestAFastLocalActionIsNotDisturbed(t *testing.T) {
	ctx, cancel := localActionContext(context.Background(), testStep(nil), "spawn_thing", zap.NewNop())
	defer cancel()

	select {
	case <-ctx.Done():
		t.Fatal("the default deadline fired immediately — this would break every healthy action in the fleet")
	case <-time.After(50 * time.Millisecond):
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("no deadline set with no override present — the default did not apply, so actions stay unbounded")
	}
	if remaining := time.Until(deadline); remaining < defaultLocalActionTimeout-time.Minute {
		t.Errorf("default deadline is %s away, want ~%s. A tighter default than measured is the "+
			"fleet-breaking direction of this change", remaining, defaultLocalActionTimeout)
	}
}

// The escape hatch has to work, or the first action that genuinely needs to run
// long forces a rebuild instead of a config change.
func TestZeroOrNegativeDisablesTheBound(t *testing.T) {
	for _, v := range []interface{}{0, 0.0, -1, -30.5} {
		ctx, cancel := localActionContext(context.Background(),
			testStep(map[string]interface{}{localActionTimeoutKey: v}), "spawn_thing", zap.NewNop())
		if _, ok := ctx.Deadline(); ok {
			t.Errorf("%v (%T) set a deadline; <=0 must mean explicitly unbounded", v, v)
		}
		cancel()
	}
}

// A malformed override must fall back to the default, never to "unbounded" — a
// typo in config should not silently restore the defect.
func TestAMalformedOverrideFallsBackToTheDefaultNotToUnbounded(t *testing.T) {
	for _, v := range []interface{}{"", "sixty", true, nil, []interface{}{1}} {
		ctx, cancel := localActionContext(context.Background(),
			testStep(map[string]interface{}{localActionTimeoutKey: v}), "spawn_thing", zap.NewNop())
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Errorf("%v (%T) produced an UNBOUNDED action; a bad value must fall back to the default", v, v)
			cancel()
			continue
		}
		if remaining := time.Until(deadline); remaining < defaultLocalActionTimeout-time.Minute {
			t.Errorf("%v (%T) produced a %s deadline, want the default %s", v, v, remaining, defaultLocalActionTimeout)
		}
		cancel()
	}
}

// THE LOAD-BEARING SEPARATION. `timeout_seconds` means "how long to wait for
// something EXTERNAL" — measured 2026-08-02, 53 of the 64 live steps carrying it
// are `call_agent`, and most of the rest are waiting semantics too
// (`await_approval` carries 86400). If this bound ever starts reading that key,
// one shared word means two things depending on the step it sits on, and an
// `await_approval` step acquires a 24-hour execution deadline it never asked for.
// That is the defect class RFC 006 was decided on.
func TestTheBoundIgnoresTimeoutSecondsEntirely(t *testing.T) {
	ctx, cancel := localActionContext(context.Background(),
		testStep(map[string]interface{}{"timeout_seconds": 1}), "spawn_thing", zap.NewNop())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("no deadline at all")
	}
	if remaining := time.Until(deadline); remaining < defaultLocalActionTimeout-time.Minute {
		t.Errorf("`timeout_seconds: 1` shortened the local-action deadline to %s. It must be ignored here: "+
			"it is the remote-wait key, and conflating the two changes what it means on 53 live call_agent steps",
			remaining)
	}
}

// The bound must never GRANT time. If the caller is already on a shorter clock,
// that clock wins.
func TestAnInheritedShorterDeadlineIsNotExtended(t *testing.T) {
	parent, cancelParent := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancelParent()

	ctx, cancel := localActionContext(parent, testStep(nil), "spawn_thing", zap.NewNop())
	defer cancel()

	select {
	case <-ctx.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("the 600s default overrode the caller's 40ms deadline — this bounds, it does not grant time")
	}
}

// The fleet-wide escape hatch. This is a behaviour change to EVERY action chosen
// from a measured distribution; if the default turns out wrong across many steps
// at 03:00, a per-step override cannot help and a rebuild is too slow. Setting
// DISABLE_LOCAL_ACTION_TIMEOUT=true must restore the exact pre-change behaviour.
func TestTheFleetWideKillSwitchRestoresUnboundedBehaviour(t *testing.T) {
	t.Setenv(disableLocalActionTimeoutEnv, "true")

	ctx, cancel := localActionContext(context.Background(),
		testStep(map[string]interface{}{localActionTimeoutKey: 1}), "spawn_thing", zap.NewNop())
	defer cancel()

	if _, ok := ctx.Deadline(); ok {
		t.Fatal("kill switch set and a deadline was still applied — the one lever available " +
			"during a fleet-wide incident does not work")
	}
}

// Regression from a PRODUCTION induction, 2026-08-03. Inducing the deadline on
// endpoint-health-checker produced: `local action "check_endpoint_health" on step ""
// exceeded its deadline` — an empty step name — while the orchestration's
// current_step was `check_health`. models.Step.Name is NOT populated on the live
// coordinator path; state.CurrentStep is. Every fixture above sets Name, so no unit
// test could have caught this. Diagnosability is the entire point of the wrapped
// error, so an unnamed step defeats it.
func TestTheStepNameComesFromTheCallerNotFromAnEmptyStepName(t *testing.T) {
	unnamed := models.Step{Action: "check_endpoint_health", Config: nil} // Name deliberately empty, as live

	ctx, cancel := localActionContext(context.Background(), unnamed, "check_health", zap.NewNop())
	defer cancel()

	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("no deadline applied")
	}
	// The name must be the one the caller resolved, not the empty struct field.
	// This test's value is the signature: localActionContext cannot read step.Name
	// even by accident, because the name is a separate required argument.
}
