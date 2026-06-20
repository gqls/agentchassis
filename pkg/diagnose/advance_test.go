package diagnose

import (
	"testing"
)

// driveAdvance simulates the chassis loop: InitLoopState, then per iteration feed
// the scripted verdict through Advance — AND round-trip the state through
// encode/decode each iteration, exactly as diagnose_route does via collected_data.
// Returns the terminal (status, stoppedBy, iterations).
func driveAdvance(seedHyp string, seed Scope, verdicts []Verdict, cg CallGraph, maxIter int) (Outcome, string, int) {
	st := InitLoopState(seedHyp, seed, maxIter, cg != nil)
	for _, v := range verdicts {
		// round-trip state through the JSON map shape (the cross-step persistence)
		encoded := EncodeLoopState(&st)
		var rt LoopState
		if err := DecodeLoopState(encoded, &rt); err != nil {
			panic(err)
		}
		st = rt

		res := Advance(&st, v, cg)
		if res.Stop {
			return res.Status, res.StoppedBy, st.Iteration
		}
	}
	// scripts exhausted without a stop → mirror Run()'s iteration-cap fall-through
	return Unverifiable, "iteration-cap", st.Iteration
}

// For each scenario already covered for Run(), assert Advance threading matches.

func TestAdvance_ConfirmedMatchesRun(t *testing.T) {
	v := []Verdict{
		{Outcome: Confirmed, Citations: []Citation{cite("a.go:F", "proof", TierRuntime)}},
	}
	status, by, iters := driveAdvance("h", Scope{Symbols: []string{"a.go"}}, v, nil, DefaultConfig().MaxIterations)
	if status != Confirmed || by != "confirmed" || iters != 1 {
		t.Fatalf("Advance confirmed mismatch: status=%s by=%s iters=%d", status, by, iters)
	}
}

func TestAdvance_RefuteThenConfirm_FollowsCallGraph(t *testing.T) {
	v := []Verdict{
		{Outcome: Refuted,
			Citations:         []Citation{cite("save.go", "reaches save", TierRuntime)},
			RevisedHypothesis: "short upstream",
			NextScope:         []string{"plan.go:spawn"}},
		{Outcome: Confirmed,
			Citations: []Citation{cite("cw.go:gen", "cap 2000", TierStatic)}},
	}
	cg := fakeCG{"plan.go:spawn": {"cw.go:gen"}}
	status, by, iters := driveAdvance("sections never reach save",
		Scope{Symbols: []string{"save.go:Save"}}, v, cg, DefaultConfig().MaxIterations)
	if status != Confirmed || by != "confirmed" || iters != 2 {
		t.Fatalf("Advance refute→confirm mismatch: status=%s by=%s iters=%d", status, by, iters)
	}
}

func TestAdvance_EvidenceNotGrowingMatchesRun(t *testing.T) {
	same := Verdict{Outcome: Refuted,
		Citations:         []Citation{cite("x.go", "same clue", TierStatic)},
		RevisedHypothesis: "h2", NextScope: []string{"x.go"}}
	status, by, _ := driveAdvance("h1", Scope{Symbols: []string{"x.go"}}, []Verdict{same, same}, nil, 5)
	if status == Confirmed || by != "evidence-not-growing" {
		t.Fatalf("Advance evidence-not-growing mismatch: status=%s by=%s", status, by)
	}
}

func TestAdvance_IterationCapMatchesRun(t *testing.T) {
	mk := func(i int) Verdict {
		return Verdict{Outcome: Unverifiable,
			Citations:      []Citation{cite("x.go", "clue "+itoa(i), TierStatic)},
			NeededEvidence: "more", NextScope: []string{"x.go"}}
	}
	status, by, iters := driveAdvance("h", Scope{Symbols: []string{"x.go"}},
		[]Verdict{mk(1), mk(2), mk(3)}, nil, 3)
	if by != "iteration-cap" || iters != 3 || status == Confirmed {
		t.Fatalf("Advance iteration-cap mismatch: status=%s by=%s iters=%d", status, by, iters)
	}
}

func TestAdvance_NoCitationCoercedMatchesRun(t *testing.T) {
	empty := Verdict{Outcome: Confirmed} // no citation
	status, _, _ := driveAdvance("h", Scope{Symbols: []string{"a.go"}}, []Verdict{empty, empty}, nil, 2)
	if status == Confirmed {
		t.Fatal("Advance must not confirm a citation-less verdict")
	}
}

func TestParseVerdictValue_FromMap(t *testing.T) {
	// the shape execute_llm_prompt yields under .result
	m := map[string]interface{}{
		"outcome": "REFUTED",
		"citations": []interface{}{
			map[string]interface{}{"tier": "runtime", "where": "agent_error_log", "quote": "regression blocked"},
		},
		"revised_hypothesis": "short upstream",
		"next_scope":         []interface{}{"plan.go:Plan"},
	}
	v, err := ParseVerdictValue(m)
	if err != nil {
		t.Fatal(err)
	}
	if v.Outcome != Refuted || len(v.Citations) != 1 || v.Citations[0].Tier != TierRuntime || len(v.NextScope) != 1 {
		t.Fatalf("ParseVerdictValue mapped wrong: %+v", v)
	}
}
