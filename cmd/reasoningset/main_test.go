package main

import "testing"

// These tests pin the corpus-admission rules that decide which historical LLM
// calls may enter a reasoning eval/training set (bugs_open/366).
//
// The bug: both truncation guards were written with nil-checks that SKIPPED the
// check when usage was never reported, so "the provider told us nothing about this
// call" was silently handled as "this call finished normally". Those are different
// claims and must not share an outcome.

func ip(v int) *int   { return &v }
func bp(v bool) *bool { return &v }

// The behaviour that already worked, pinned so the fix cannot cost it.
func TestJudgeTruncation_AtCeilingIsExcluded(t *testing.T) {
	ok, why := judgeTruncation(Provenance{OutputTokens: ip(2000), MaxTokens: ip(2000)})
	if ok || why != "truncated" {
		t.Fatalf("output at the ceiling is a CUT completion: got ok=%v why=%q", ok, why)
	}
}

func TestJudgeTruncation_BelowCeilingPasses(t *testing.T) {
	if ok, why := judgeTruncation(Provenance{OutputTokens: ip(500), MaxTokens: ip(2000)}); !ok || why != "" {
		t.Fatalf("a completion that finished below its ceiling is admissible: got ok=%v why=%q", ok, why)
	}
}

// THE FILED DEFECT. A ceiling was recorded — so the call went through the
// configured path and usage should have come back — and the provider reported
// none. Before the fix the nil-check skipped the guard and this row was admitted
// as a complete answer.
func TestJudgeTruncation_UsageUnreportedDespiteACeilingIsExcluded(t *testing.T) {
	for _, p := range []Provenance{
		{OutputTokens: nil, MaxTokens: ip(2000)},   // never reported
		{OutputTokens: ip(0), MaxTokens: ip(2000)}, // reported as zero, which is not a distinguishable answer
	} {
		ok, why := judgeTruncation(p)
		if ok {
			t.Fatalf("usage unreported despite a recorded ceiling must NOT read as a complete "+
				"completion: got ok=true for %+v", p)
		}
		if why != "usage_unreported" {
			t.Fatalf("the reason must name what is actually wrong, not borrow 'truncated' — "+
				"we do not know that it was cut: got %q", why)
		}
	}
}

// THE OTHER HALF OF THE JUDGEMENT, and the one a blanket "exclude Unknown" would
// have got wrong. No ceiling recorded at all means the question was never
// answerable and there is no truncation signal either way.
// [MEASURED 2026-09-03] 161 corpus rows are in this state — 8% of the 2,013
// successful rows, every one a score_relevance call from the Mar-May logging
// regime, averaging 3,092 output tokens with no empty responses. Excluding them
// would have shrunk the corpus to guard a population that is currently ZERO.
func TestJudgeTruncation_NoCeilingRecordedIsNotExcluded(t *testing.T) {
	for _, p := range []Provenance{
		{OutputTokens: ip(3092), MaxTokens: nil},   // the measured shape
		{OutputTokens: ip(3092), MaxTokens: ip(0)}, // recorded as zero, same meaning
	} {
		if ok, why := judgeTruncation(p); !ok {
			t.Fatalf("a row whose ceiling was never recorded carries no truncation signal and must "+
				"stay admissible: got ok=false why=%q for %+v", why, p)
		}
	}
}

// judgeInput must not carry its own copy of the shared rules. It used to, which is
// why the nil-check defect had to be fixed twice to be fixed at all.
func TestJudgeInputDelegatesToCommonExclusions(t *testing.T) {
	cases := []inRow{
		{InputState: "clean", Provenance: Provenance{OutputTokens: ip(2000), MaxTokens: ip(2000)}},
		{InputState: "clean", Provenance: Provenance{OutputTokens: nil, MaxTokens: ip(2000)}},
		{InputState: "clean", Provenance: Provenance{OutputTokens: ip(10), MaxTokens: ip(2000)}},
		{InputState: "clean", Provenance: Provenance{Success: bp(false), ErrorMessage: "boom"}},
		{InputState: "clean", Provenance: Provenance{Success: bp(false), ErrorMessage: "response truncated at 2000"}},
		{InputState: "clean", Provenance: Provenance{OutputTokens: ip(3092), MaxTokens: nil}},
	}
	for _, c := range cases {
		gotOK, gotWhy := judgeInput(c)
		wantOK, wantWhy := commonExclusions(c.InputState, c.Provenance)
		if gotOK != wantOK || gotWhy != wantWhy {
			t.Fatalf("judgeInput must BE commonExclusions for every shared rule, not a copy of it: "+
				"provenance %+v gave (%v,%q), commonExclusions gave (%v,%q)",
				c.Provenance, gotOK, gotWhy, wantOK, wantWhy)
		}
	}
}

// judgeInput's one genuinely local rule must survive the convergence, and must
// still outrank the shared ones: a step whose prompt rendered "<no value>"
// reasoned without seeing its inputs (bugs_open/016), which is a stronger
// disqualification than anything the usage counts say.
func TestJudgeInput_NoValueInjectionStillOutranksTheSharedRules(t *testing.T) {
	s := inRow{
		InputState: "the plan is <no value> and so on",
		// also truncated, so the two rules compete
		Provenance: Provenance{OutputTokens: ip(2000), MaxTokens: ip(2000)},
	}
	ok, why := judgeInput(s)
	if ok {
		t.Fatal("a <no value> row is never admissible")
	}
	if why != "no_value_injection" {
		t.Fatalf("the local rule must still win over the shared ones: got %q", why)
	}
}
