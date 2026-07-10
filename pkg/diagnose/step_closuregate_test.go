// step_closuregate_test.go — F0.4d: the symptom-closure gate. A CONFIRMED
// verdict must map every observation of the ORIGINAL symptom to the confirmed
// mechanism, or it degrades to Unverifiable and the loop keeps working the
// residue. Benchmark run dd1186b9 (2026-07-09): a tier-covered, well-cited
// confirm explained the blank-page half of the symptom and dismissed the
// nav-link half as "not a nav issue" — the exact half-answer this gate refuses.
package diagnose

import (
	"strings"
	"testing"
)

// tier-covered citations so these fixtures isolate the CLOSURE variable.
func bothTiers() []Citation {
	return []Citation{
		cite("a.go:F", "the mechanism", TierStatic),
		cite("agent_error_log", "the occurrence", TierRuntime),
	}
}

func TestClosureGate_MissingSymptomCheckDegrades(t *testing.T) {
	d := DecideStep(StepInput{
		Iteration: 1, MaxIterations: 5,
		Hypothesis: "h", Scope: Scope{Symbols: []string{"a.go"}},
		Verdict:       Verdict{Outcome: Confirmed, Citations: bothTiers()},
		PrevScopeSize: 2,
	})
	if d.Decision == "stop" && d.StopReason == "confirmed" {
		t.Fatalf("confirm without symptom_check must not stand: %+v", d)
	}
}

func TestClosureGate_UnexplainedObservationDegrades(t *testing.T) {
	v := Verdict{Outcome: Confirmed, Citations: bothTiers(), SymptomCheck: []SymptomCheck{
		{Observation: "the page is blank", Explained: true, How: "never built"},
		{Observation: "a nav link to it was published", Explained: false, How: "not investigated"},
	}}
	coerced := coerceVerdict(v)
	if coerced.Outcome != Unverifiable {
		t.Fatalf("partially-explained confirm must degrade, got %v", coerced.Outcome)
	}
	// The residue must be NAMED so the next iteration can chase it.
	if !strings.Contains(coerced.NeededEvidence, "a nav link to it was published") {
		t.Fatalf("the unexplained observation must be named in NeededEvidence: %q", coerced.NeededEvidence)
	}
}

func TestClosureGate_FullCoverageStands(t *testing.T) {
	d := DecideStep(StepInput{
		Iteration: 1, MaxIterations: 5,
		Hypothesis: "h", Scope: Scope{Symbols: []string{"a.go"}},
		Verdict: Verdict{Outcome: Confirmed, Citations: bothTiers(), SymptomCheck: []SymptomCheck{
			{Observation: "the page is blank", Explained: true, How: "never built", Cites: []int{0}},
			{Observation: "a nav link to it was published", Explained: true, How: "nav reads pages.status not build_status", Cites: []int{1}},
		}},
		PrevScopeSize: 2,
	})
	if d.Decision != "stop" || d.StopReason != "confirmed" || d.TerminalStatus != Confirmed {
		t.Fatalf("fully-covered, tier-covered confirm should stand, got %+v", d)
	}
	// The human-facing conclusion renders the coverage the gate enforced.
	if !strings.Contains(d.Conclusion, "Symptom coverage:") ||
		!strings.Contains(d.Conclusion, "nav link") {
		t.Fatalf("conclusion should render symptom coverage: %s", d.Conclusion)
	}
}

// Refuted and Unverifiable verdicts are exempt: the gate exists to stop a
// premature VICTORY declaration, not to burden the investigative outcomes.
func TestClosureGate_NonConfirmExempt(t *testing.T) {
	for _, oc := range []Outcome{Refuted, Unverifiable} {
		v := coerceVerdict(Verdict{Outcome: oc,
			Citations:         []Citation{cite("x", "q", TierRuntime)},
			RevisedHypothesis: "elsewhere", NextScope: []string{"b.go"}})
		if v.Outcome != oc {
			t.Fatalf("%v must be exempt from the closure gate, got %v", oc, v.Outcome)
		}
	}
}

// The wire format round-trips the new field — the seam the prompt's schema
// must stay in lockstep with.
func TestClosureGate_WireParsesSymptomCheck(t *testing.T) {
	raw := []byte(`{"outcome":"CONFIRMED",
	  "citations":[{"tier":"static","where":"a.go:F","quote":"mechanism"},
	               {"tier":"runtime","where":"log","quote":"occurrence"}],
	  "symptom_check":[{"observation":"page blank","explained":true,"how":"never built"},
	                   {"observation":"nav link published","explained":false}]}`)
	v, err := ParseVerdict(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.SymptomCheck) != 2 || v.SymptomCheck[0].Explained != true || v.SymptomCheck[1].Explained != false {
		t.Fatalf("symptom_check parsed wrong: %+v", v.SymptomCheck)
	}
	if coerceVerdict(v).Outcome != Unverifiable {
		t.Fatal("wire-parsed partial coverage must degrade")
	}
}

// ── F0.6: context disposition + citation-backed explanations ────────────────
// Benchmark run 5179a2ea: the model marked comparative control clauses
// ("gamesdesign works on the same platform") explained:true while its own text
// said "unverifiable from this bundle" — grade inflation a structural gate on
// explained/unexplained alone cannot see.

func TestClosureGate_ContextEntriesExempt(t *testing.T) {
	// Control clauses marked context (not explained) must NOT block the confirm.
	d := DecideStep(StepInput{
		Iteration: 1, MaxIterations: 5,
		Hypothesis: "h", Scope: Scope{Symbols: []string{"a.go"}},
		Verdict: Verdict{Outcome: Confirmed, Citations: bothTiers(), SymptomCheck: []SymptomCheck{
			{Observation: "the page is blank", Explained: true, How: "never built", Cites: []int{0}},
			{Observation: "gamesdesign works on the same platform", Context: true, How: "comparative control; out of this run's evidence scope"},
		}},
		PrevScopeSize: 2,
	})
	if d.Decision != "stop" || d.StopReason != "confirmed" {
		t.Fatalf("context entries must be exempt from the accounting, got %+v", d)
	}
	if !strings.Contains(d.Conclusion, "[context]") {
		t.Fatalf("conclusion should render the context mark: %s", d.Conclusion)
	}
}

func TestClosureGate_ExplainedWithoutCitesDegrades(t *testing.T) {
	// The run-4 case: explained:true with no grounding.
	v := coerceVerdict(Verdict{Outcome: Confirmed, Citations: bothTiers(), SymptomCheck: []SymptomCheck{
		{Observation: "the page is blank", Explained: true, How: "never built", Cites: []int{0}},
		{Observation: "gaswholesalers has a working news feed", Explained: true, How: "unverifiable from this bundle"},
	}})
	if v.Outcome != Unverifiable {
		t.Fatalf("ungrounded explanation must degrade, got %v", v.Outcome)
	}
	if !strings.Contains(v.NeededEvidence, "gaswholesalers") {
		t.Fatalf("the ungrounded observation must be named: %q", v.NeededEvidence)
	}
}

func TestClosureGate_OutOfRangeCiteIsNotGrounding(t *testing.T) {
	v := coerceVerdict(Verdict{Outcome: Confirmed, Citations: bothTiers(), SymptomCheck: []SymptomCheck{
		{Observation: "the page is blank", Explained: true, How: "never built", Cites: []int{7}},
	}})
	if v.Outcome != Unverifiable {
		t.Fatalf("an out-of-range cites index must not count as grounding, got %v", v.Outcome)
	}
}
