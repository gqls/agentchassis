// FILE: platform/orchestration/actions/provocation_gate_live_calibration_test.go
//
// THE LIVE CALIBRATION — the run PLAN §10.6 makes a precondition for wiring the
// gate to anything that publishes:
//
//	"With a human backstop that was good practice; without one it is the only
//	 evidence the gate works at all. It must pass all 9 and reject a set of
//	 deliberately bad samples ... before it is wired to anything that publishes."
//
// The committed calibration in provocation_gate_action_test.go stubs the judge, so
// it proves the deterministic layers and the fail-closed wiring and says NOTHING
// about whether a real model judges provocations well. This file closes that gap.
// It is the file the sibling's header promises.
//
// WHY IT IS OPT-IN
// It spends real tokens and needs a real key, so it must not run in an ordinary
// `go test ./...`. It is gated on PROVOCATION_LIVE_CALIBRATION=1.
//
// WHY A SKIP HERE IS NOT A PASS, AND WHY THAT DISTINCTION IS THE POINT
// This lane has already been bitten by exactly the failure this guards against: a
// live driver printed "SKIP PIL unavailable" and then "ALL LIVE CHECKS PASSED",
// so three of nine checks never ran and the run still read as green. So:
//
//	env unset          -> SKIP, and the message says the calibration is UNSCORED
//	env set, no key    -> FAIL, loudly. Asking for the run and not getting it is a
//	                      failure, never a skip. A missing dependency is a failed
//	                      check.
//	env set, key present -> the real thing, and every case must come out right.
//
// HOW TO RUN
//
//	ANTHROPIC_API_KEY=... PROVOCATION_LIVE_CALIBRATION=1 \
//	  go test ./platform/orchestration/actions/ -run TestLiveCalibration -v -timeout 20m
//
// Optionally CALIBRATION_MODEL=claude-sonnet-5 (default below). The model this
// passes with is the model the gate must be configured to use — a calibration is
// evidence about ONE model, not about the gate in general, and swapping the model
// afterwards invalidates the run.

package actions

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

const defaultCalibrationModel = "claude-sonnet-5"

// liveJudge builds a judgeFn backed by a real provider, or explains why it cannot.
func liveJudge(ctx context.Context, t *testing.T) (judgeFn, string) {
	t.Helper()

	model := os.Getenv("CALIBRATION_MODEL")
	if model == "" {
		model = defaultCalibrationModel
	}

	// api_key_env_var is the platform's convention; the client reads the named
	// variable itself rather than taking the secret through config.
	cfg := map[string]interface{}{
		"provider":        "anthropic",
		"model":           model,
		"api_key_env_var": "ANTHROPIC_API_KEY",
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Fatalf("PROVOCATION_LIVE_CALIBRATION=1 was set but ANTHROPIC_API_KEY is empty.\n"+
			"This is a FAILURE, not a skip: the calibration was asked for and did not happen, "+
			"and §10.6 makes this run the only evidence the gate works.\n"+
			"Set the key, or run this inside a pod that already has it.\n"+
			"(model requested: %s)", model)
	}

	client, err := createAIClient(ctx, cfg)
	if err != nil {
		t.Fatalf("could not build the %s client for calibration: %v", model, err)
	}
	return func(c context.Context, prompt string) (string, error) {
		// A generous ceiling: the judge must have room to finish its JSON, and a
		// truncation is a rejection, so a mean budget would show up as the gate
		// rejecting everything for a reason that is our fault, not the model's.
		return client.GenerateText(c, prompt, map[string]interface{}{"max_tokens": 2048})
	}, model
}

// TestLiveCalibration is §10.6's run.
func TestLiveCalibration(t *testing.T) {
	if os.Getenv("PROVOCATION_LIVE_CALIBRATION") != "1" {
		t.Skip("UNSCORED — live calibration not run (set PROVOCATION_LIVE_CALIBRATION=1). " +
			"This is NOT evidence the gate works; §10.6 requires this run against a real " +
			"model before the gate is wired to anything that publishes.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	judge, model := liveJudge(ctx, t)
	t.Logf("live calibration against model %s", model)

	type result struct {
		name     string
		approved bool
		reasons  string
	}
	var passes, failures []result

	// ---- Half one: every real provocation must be APPROVED. ----
	// A rejection here is the false-positive direction, which silently starves the
	// site: the gate would refuse the very entries the owner published.
	t.Run("the nine real provocations", func(t *testing.T) {
		for _, c := range realProvocations {
			c := c
			t.Run(c.Slug, func(t *testing.T) {
				v := gateCandidate(ctx, c, judge, model)
				r := result{name: c.Slug, approved: v.Approved, reasons: ruleList(v)}
				if !v.Approved {
					failures = append(failures, r)
					t.Errorf("REJECTED a real provocation: %s\n  advisory: interesting=%d current=%d %q",
						ruleList(v), v.Advisory.Interesting, v.Advisory.Current, v.Advisory.Note)
					return
				}
				passes = append(passes, r)
				t.Logf("approved (advisory interesting=%d current=%d)",
					v.Advisory.Interesting, v.Advisory.Current)
			})
		}
	})

	// ---- Half two: every deliberately bad sample must be REJECTED. ----
	// A pass here is the false-negative direction: a false statement on a live
	// homepage with nobody in the loop.
	t.Run("the deliberately bad set", func(t *testing.T) {
		for _, tc := range badProvocations {
			tc := tc
			t.Run(strings.ReplaceAll(tc.name, " ", "_"), func(t *testing.T) {
				v := gateCandidate(ctx, tc.c, judge, model)
				r := result{name: tc.name, approved: v.Approved, reasons: ruleList(v)}
				if v.Approved {
					failures = append(failures, r)
					t.Errorf("APPROVED %q — the model did not catch it and no human will", tc.name)
					return
				}
				passes = append(passes, r)
				t.Logf("rejected: %s", ruleList(v))
			})
		}
	})

	// ---- The read-out. ----
	// Printed unconditionally so a partial run is legible as a partial run. A
	// calibration whose summary only appears on success is a calibration that
	// reads as green when it dies halfway.
	t.Logf("\n=== LIVE CALIBRATION SUMMARY (model %s) ===", model)
	t.Logf("correct: %d   incorrect: %d   (expected %d correct)",
		len(passes), len(failures), len(realProvocations)+len(badProvocations))
	for _, f := range failures {
		t.Logf("  WRONG  %-40s approved=%v  %s", f.name, f.approved, f.reasons)
	}
	if len(failures) > 0 {
		t.Fatalf("CALIBRATION FAILED: %d of %d cases judged wrongly. "+
			"§10.6 forbids wiring the gate to anything that publishes until this passes.",
			len(failures), len(passes)+len(failures))
	}
	if want := len(realProvocations) + len(badProvocations); len(passes) != want {
		t.Fatalf("CALIBRATION INCOMPLETE: %d of %d cases actually ran. "+
			"An incomplete run is not a pass — this is the exact shape that printed "+
			"ALL LIVE CHECKS PASSED with three checks skipped.", len(passes), want)
	}
	t.Logf("CALIBRATION PASSED against %s — all %d cases judged correctly.", model, len(passes))
	t.Logf("RECORD THIS: the gate is calibrated for %s ONLY. Configure ai_service.model "+
		"to that value when wiring, and re-run this if the model changes.", model)
}

// A guard on the guard: if the corpora ever shrink, the live run above would
// "pass" while proving less. §10.6 names the sizes, so they are asserted.
func TestLiveCalibrationCorporaAreTheStatedSize(t *testing.T) {
	if len(realProvocations) != 9 {
		t.Errorf("the real corpus is %d, not the 9 §10.6 names", len(realProvocations))
	}
	if len(badProvocations) != 4 {
		t.Errorf("the bad set is %d, not the 4 §10.6 names (insult, factual-claim-as-opinion, "+
			"one-sided political, trending slop)", len(badProvocations))
	}
	// The bad set must cover the four NAMED kinds, not four variations of one.
	seen := map[string]bool{}
	for _, b := range badProvocations {
		seen[b.name] = true
	}
	for _, want := range []string{
		"a bare insult",
		"a factual claim dressed as opinion",
		"a one-sided political take",
		"trending slop",
	} {
		if !seen[want] {
			t.Errorf("the bad set no longer covers %q, which §10.6 names explicitly", want)
		}
	}
	if t.Failed() {
		t.Log(fmt.Sprintf("corpora: %d real, %d bad", len(realProvocations), len(badProvocations)))
	}
}
