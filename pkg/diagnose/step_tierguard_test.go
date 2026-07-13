// step_tierguard_test.go — F0.4e: a CONFIRMED verdict must span both evidence
// families (static mechanism + state/runtime occurrence) or it degrades to
// Unverifiable and the loop keeps gathering. Benchmark run 4d43d002 (2026-07-09)
// showed the three-tier doctrine was unenforced: the only confirm guard was
// citations ≥ 1, so a plausible single-family confirm stopped the loop.
package diagnose

import "testing"

func confirmWith(cs ...Citation) StepInput {
	return StepInput{
		Iteration: 1, MaxIterations: 5,
		Hypothesis: "h", Scope: Scope{Symbols: []string{"a.go"}},
		// symptom coverage supplied so these fixtures isolate the TIER variable
		// (the closure gate has its own tests in step_closuregate_test.go).
		Verdict:       Verdict{Outcome: Confirmed, Citations: cs, SymptomCheck: covered("h")},
		PrevScopeSize: 2,
	}
}

func TestTierGuard_SingleFamilyConfirmDegrades(t *testing.T) {
	cases := []struct {
		name string
		cs   []Citation
	}{
		{"static only", []Citation{cite("a.go:F", "mechanism", TierStatic)}},
		{"state only", []Citation{cite("pages", "row shows effect", TierState)}},
		{"runtime only", []Citation{cite("agent_error_log", "step failed", TierRuntime)}},
		{"state+runtime, no static", []Citation{
			cite("pages", "row", TierState), cite("agent_error_log", "log", TierRuntime)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := DecideStep(confirmWith(tc.cs...))
			if d.Decision == "stop" && d.StopReason == "confirmed" {
				t.Fatalf("single-family confirm must not stand: %+v", d)
			}
		})
	}
}

func TestTierGuard_CoveredConfirmStops(t *testing.T) {
	for _, observed := range []Tier{TierState, TierRuntime} {
		d := DecideStep(confirmWith(
			cite("a.go:F", "the mechanism", TierStatic),
			cite("evidence", "the occurrence", observed),
		))
		if d.Decision != "stop" || d.StopReason != "confirmed" || d.TerminalStatus != Confirmed {
			t.Fatalf("static+%v confirm should stop confirmed, got %+v", observed, d)
		}
	}
}

// Refutation is exempt by design: one contradicting log line legitimately breaks
// a hypothesis (the verdict prompt's rule-3 asymmetry).
func TestTierGuard_SingleTierRefuteStands(t *testing.T) {
	in := confirmWith()
	in.Verdict = Verdict{
		Outcome:           Refuted,
		Citations:         []Citation{cite("agent_error_log", "contradicts", TierRuntime)},
		RevisedHypothesis: "something else",
		NextScope:         []string{"b.go:G"},
	}
	d := DecideStep(in)
	if d.NextHypothesis != "something else" {
		t.Fatalf("single-tier refute must stand and re-scope, got %+v", d)
	}
}

// The trail must record the coerced verdict — what was decided on, not what the
// model said — via the same shared coercion (no drift between decision and trail).
func TestTierGuard_TrailRecordsCoercion(t *testing.T) {
	st := InitLoopState("h", Scope{Symbols: []string{"a.go"}}, 5, false)
	res := Advance(&st, Verdict{Outcome: Confirmed,
		Citations:    []Citation{cite("pages", "single family", TierState)},
		SymptomCheck: covered("h")}, nil)
	if res.Stop && res.StoppedBy == "confirmed" {
		t.Fatalf("Advance let a single-family confirm stand")
	}
	if got := st.Trail[len(st.Trail)-1].Verdict.Outcome; got != Unverifiable {
		t.Fatalf("trail must record the coerced outcome, got %v", got)
	}
}
