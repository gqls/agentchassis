// FILE: platform/orchestration/retry_window_declared_test.go
//
// bugs_open/029. A REPLAYED request must wait the window its step DECLARED.
//
// The code these tests replaced capped every retry at 5 minutes and dropped to
// 3 minutes for any step declaring more than 30 — an INVERSION, in which the
// longer a step declared the less it was given. Measured live 2026-08-18:
// call_dispatch declares 900s and showed a 15:00 window at retry_version 0 and
// 05:00 at versions 1-3; process_item_iter_0_call_handler declares 1200s and
// showed 20:00 then 05:00.
//
// Every assertion below is written so that RESTORING the old block fails it —
// see TestRetryWindowRejectsTheOldTruncation, which encodes the old arithmetic
// directly so the mutation is checked rather than merely described.

package orchestration

import (
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

func planWithStep(stepName string, timeoutSeconds int) *OrchestrationState {
	step := models.Step{Action: "call_agent"}
	if timeoutSeconds > 0 {
		step.Config = map[string]interface{}{"timeout_seconds": timeoutSeconds}
	}
	return &OrchestrationState{
		WorkflowPlan: models.WorkflowPlan{Steps: map[string]models.Step{stepName: step}},
	}
}

func awaitedFor(stepName string, rowWindow time.Duration) *AwaitedRequest {
	now := time.Now()
	return &AwaitedRequest{StepName: stepName, SentAt: now, TimeoutAt: now.Add(rowWindow)}
}

// The declared value comes back UNDIMINISHED, at every magnitude the fleet
// actually uses. The three marked cases are what the old code returned.
func TestRetryWindowHonoursTheDeclaredTimeout(t *testing.T) {
	cases := []struct {
		name     string
		declared int
		want     time.Duration
		oldGave  time.Duration // what the truncation returned — the regression to catch
	}{
		{"600s — the commonest declaration above the cap", 600, 10 * time.Minute, 5 * time.Minute},
		{"900s — call_dispatch, the measured case", 900, 15 * time.Minute, 5 * time.Minute},
		{"1200s — process_item_iter_N_call_handler, measured", 1200, 20 * time.Minute, 5 * time.Minute},
		{"1800s — exactly the inversion boundary", 1800, 30 * time.Minute, 5 * time.Minute},
		{"2100s — first step PAST the boundary, so the old code gave 3m not 5m", 2100, 35 * time.Minute, 3 * time.Minute},
		{"86400s — the human-approval step", 86400, 24 * time.Hour, 3 * time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := retryWindow(planWithStep("s", tc.declared), awaitedFor("s", 90*time.Second), zap.NewNop())
			if got != tc.want {
				t.Fatalf("declared %ds: retry window = %s, want %s (the old truncation gave %s — if that is what you see, the inversion is back)",
					tc.declared, got, tc.want, tc.oldGave)
			}
		})
	}
}

// Monotonicity is the property the old code broke, and it is worth asserting
// separately: a step that declares MORE must never be given LESS.
func TestRetryWindowIsMonotoneInTheDeclaration(t *testing.T) {
	// The row window is set to the DECLARED value, as it is on attempt 0 in
	// production. That matters: with a constant row window this test cannot
	// fail against the old truncation (every case returns the same capped
	// number, which is trivially monotone), so it would assert nothing. Driving
	// the row from the declaration is what makes the 30-minute cliff — 5m below
	// it, 3m above — show up as the non-monotonicity it is.
	declarations := []int{60, 180, 300, 600, 900, 1200, 1800, 2100, 3600, 21600, 43200, 86400}
	var prev time.Duration
	for _, d := range declarations {
		row := time.Duration(d) * time.Second
		got := retryWindow(planWithStep("s", d), awaitedFor("s", row), zap.NewNop())
		if got < prev {
			t.Fatalf("declaring %ds yielded %s, which is LESS than the %s given to a smaller declaration — "+
				"this is the bugs_open/029 inversion", d, got, prev)
		}
		prev = got
	}
}

// The mutation test proper: the old arithmetic, encoded. If anyone restores it,
// this states exactly what it does to the fleet's real declarations.
func TestRetryWindowRejectsTheOldTruncation(t *testing.T) {
	oldTruncation := func(rowWindow time.Duration) time.Duration {
		n := rowWindow
		if n <= 0 || n > 30*time.Minute {
			n = 3 * time.Minute
		}
		if n > 5*time.Minute {
			n = 5 * time.Minute
		}
		return n
	}
	for _, declared := range []int{600, 900, 1200, 86400} {
		d := time.Duration(declared) * time.Second
		if got := retryWindow(planWithStep("s", declared), awaitedFor("s", d), zap.NewNop()); got == oldTruncation(d) {
			t.Fatalf("declared %ds: retryWindow returned %s, identical to the old truncation. "+
				"The declared window is not being honoured.", declared, got)
		}
	}
}

// Loop-expanded steps are the population the bug was MEASURED on. The stored
// plan carries the suffixed keys, so the lookup must hit them by their expanded
// name — a helper that only knew the template name would silently miss every
// step in the bug while passing every other test here.
func TestRetryWindowResolvesLoopExpandedStepNames(t *testing.T) {
	const expanded = "process_item_iter_1_call_handler"
	got := retryWindow(planWithStep(expanded, 1200), awaitedFor(expanded, 5*time.Minute), zap.NewNop())
	if got != 20*time.Minute {
		t.Fatalf("loop-expanded step %q: got %s, want 20m — the live plan stores these suffixed keys with "+
			"config.timeout_seconds intact, and they are exactly the steps 029 wedged on", expanded, got)
	}
}

// Fallbacks: never below the system default, and never a division by the row's
// own already-truncated history when the plan can answer.
func TestRetryWindowFallbacks(t *testing.T) {
	sysDefault := time.Duration(180) * time.Second

	t.Run("step absent from the plan: use the row when it is longer than the default", func(t *testing.T) {
		st := &OrchestrationState{WorkflowPlan: models.WorkflowPlan{Steps: map[string]models.Step{}}}
		if got := retryWindow(st, awaitedFor("missing", 12*time.Minute), zap.NewNop()); got != 12*time.Minute {
			t.Fatalf("got %s, want 12m from the row", got)
		}
	})

	t.Run("step absent and the row is SHORTER than the default: floor at the default", func(t *testing.T) {
		st := &OrchestrationState{WorkflowPlan: models.WorkflowPlan{Steps: map[string]models.Step{}}}
		if got := retryWindow(st, awaitedFor("missing", 20*time.Second), zap.NewNop()); got != sysDefault {
			t.Fatalf("got %s, want the %s system default — a retry shorter than the default is the "+
				"same defect in miniature", got, sysDefault)
		}
	})

	t.Run("a zero-length row cannot produce a zero window", func(t *testing.T) {
		st := &OrchestrationState{WorkflowPlan: models.WorkflowPlan{Steps: map[string]models.Step{}}}
		if got := retryWindow(st, awaitedFor("missing", 0), zap.NewNop()); got != sysDefault {
			t.Fatalf("got %s, want %s", got, sysDefault)
		}
	})

	t.Run("nil inputs are survivable", func(t *testing.T) {
		if got := retryWindow(nil, nil, zap.NewNop()); got != sysDefault {
			t.Fatalf("got %s, want %s", got, sysDefault)
		}
	})
}
