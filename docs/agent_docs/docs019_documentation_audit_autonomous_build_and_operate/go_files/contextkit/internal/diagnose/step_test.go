package diagnose

import "testing"

// DecideStep is what the chassis diagnose_run action calls per iteration (the
// verdict being a separate workflow step). These test it directly.

func TestDecideStep_ConfirmStops(t *testing.T) {
	d := DecideStep(StepInput{
		Iteration: 1, MaxIterations: 5,
		Hypothesis: "h", Scope: Scope{Symbols: []string{"a.go"}},
		Verdict:       Verdict{Outcome: Confirmed, Citations: []Citation{{Where: "a.go:F", Quote: "q", Tier: TierRuntime}}},
		PrevScopeSize: 2,
	})
	if d.Decision != "stop" || d.StopReason != "confirmed" || d.TerminalStatus != Confirmed {
		t.Fatalf("confirm should stop confirmed, got %+v", d)
	}
}

func TestDecideStep_RefuteContinuesWithNextScope(t *testing.T) {
	cg := fakeCG{"a.go:F": {"b.go:G"}}
	d := DecideStep(StepInput{
		Iteration: 1, MaxIterations: 5, FollowCallGraph: true,
		Hypothesis: "h1", Scope: Scope{Symbols: []string{"a.go:F"}},
		Verdict: Verdict{Outcome: Refuted,
			Citations:         []Citation{{Where: "log", Quote: "contradiction", Tier: TierRuntime}},
			RevisedHypothesis: "h2", NextScope: []string{"a.go:F"}},
		CallGraph: cg, SeenCitations: map[string]bool{}, PrevScopeSize: 2,
	})
	if d.Decision != "continue" || d.NextHypothesis != "h2" {
		t.Fatalf("refute should continue with revised hyp, got %+v", d)
	}
	// next scope should include the call-graph-followed symbol
	found := false
	for _, s := range d.NextScope.Symbols {
		if s == "b.go:G" {
			found = true
		}
	}
	if !found {
		t.Fatalf("re-scope should follow call graph, got %v", d.NextScope.Symbols)
	}
}

func TestDecideStep_NoCitationCoercedNoConfirm(t *testing.T) {
	d := DecideStep(StepInput{
		Iteration: 1, MaxIterations: 5,
		Hypothesis: "h", Scope: Scope{Symbols: []string{"a.go"}},
		Verdict:       Verdict{Outcome: Confirmed}, // no citation
		SeenCitations: map[string]bool{}, PrevScopeSize: 2,
	})
	if d.TerminalStatus == Confirmed && d.StopReason == "confirmed" {
		t.Fatal("a citation-less confirm must NOT yield a confirmed stop")
	}
}

func TestDecideStep_IterationCapStops(t *testing.T) {
	d := DecideStep(StepInput{
		Iteration: 5, MaxIterations: 5, // at the cap
		Hypothesis: "h", Scope: Scope{Symbols: []string{"a.go"}},
		Verdict: Verdict{Outcome: Unverifiable,
			Citations: []Citation{{Where: "x", Quote: "clue", Tier: TierStatic}}, NextScope: []string{"a.go"}},
		SeenCitations: map[string]bool{}, PrevScopeSize: 2,
	})
	if d.Decision != "stop" || d.StopReason != "iteration-cap" {
		t.Fatalf("at the cap should stop iteration-cap, got %+v", d)
	}
}

// NewCallGraphFromJSON is the chassis entry point (Output as bytes).
func TestNewCallGraphFromJSON(t *testing.T) {
	raw := []byte(`{"root":"/r","file_count":1,"files":[
	  {"path":"a.go","functions":[{"name":"foo","calls":["bar"]},{"name":"bar","calls":[]}]}
	]}`)
	g, err := NewCallGraphFromJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	nb := g.Neighbourhood("a.go:foo")
	if len(nb) != 1 || nb[0] != "a.go:bar" {
		t.Fatalf("expected [a.go:bar], got %v", nb)
	}
}
