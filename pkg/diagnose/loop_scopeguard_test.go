// loop_scopeguard_test.go — guard-vs-expansion fix (run 17933a83): the
// narrowing guard measures the MODEL-NAMED scope; the engine's call-graph
// enrichment neither trips the guard nor grows unbounded.
package diagnose

import "testing"

type fatCG struct{ n map[string][]string }

func (f fatCG) Neighbourhood(sym string) []string { return f.n[sym] }

func seedScope(n int) Scope {
	s := Scope{}
	for i := 0; i < n; i++ {
		s.Symbols = append(s.Symbols, "seed.go:S"+string(rune('A'+i)))
	}
	return s
}

func fatGraph(named []string, fanout int) fatCG {
	g := fatCG{n: map[string][]string{}}
	for _, s := range named {
		for i := 0; i < fanout; i++ {
			g.n[s] = append(g.n[s], s+"_callee"+string(rune('a'+i)))
		}
	}
	return g
}

func TestExpansionDoesNotTripScopeGuard(t *testing.T) {
	named := []string{"a.go:A", "b.go:B", "c.go:C", "d.go:D", "e.go:E", "f.go:F"}
	in := StepInput{
		Iteration:       1,
		MaxIterations:   5,
		Hypothesis:      "h",
		Scope:           seedScope(12),
		PrevScopeSize:   13, // seed.size()+1, as InitLoopState sets
		FollowCallGraph: true,
		CallGraph:       fatGraph(named, 10), // 6 named × 10 callees = 66 raw
		Verdict: Verdict{
			Outcome:      Unverifiable,
			NextScope:    named,
			DataRequests: []DataRequest{{SQL: "SELECT 1", Why: "progress"}},
		},
	}
	d := DecideStep(in)
	if d.Decision != "continue" {
		t.Fatalf("expansion must not trip the guard: got %s (%s)", d.Decision, d.StopReason)
	}
	if d.NamedScopeSize != 6 {
		t.Fatalf("NamedScopeSize must be the model-named count, got %d", d.NamedScopeSize)
	}
	if got := d.NextScope.size(); got != defaultMaxExpandedScope {
		t.Fatalf("expansion must cap at the default (%d), got %d", defaultMaxExpandedScope, got)
	}
	for _, s := range named {
		found := false
		for _, x := range d.NextScope.Symbols {
			if x == s {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("named entry %s must survive capping", s)
		}
	}
}

func TestModelFlailStillTripsGuard(t *testing.T) {
	var wide []string
	for i := 0; i < 16; i++ { // 16 > 13+2
		wide = append(wide, "w.go:W"+string(rune('A'+i)))
	}
	in := StepInput{
		Iteration: 1, MaxIterations: 5, Hypothesis: "h",
		Scope: seedScope(12), PrevScopeSize: 13,
		FollowCallGraph: true, CallGraph: fatCG{n: map[string][]string{}},
		Verdict: Verdict{Outcome: Unverifiable, NextScope: wide,
			DataRequests: []DataRequest{{SQL: "SELECT 2", Why: "x"}}},
	}
	d := DecideStep(in)
	if d.Decision != "stop" || d.StopReason != "scope-not-narrowing" {
		t.Fatalf("model naming prev+3 entries must still stop, got %s/%s", d.Decision, d.StopReason)
	}
}

func TestUnlimitedExpansionOptOut(t *testing.T) {
	named := []string{"a.go:A", "b.go:B"}
	in := StepInput{
		Iteration: 1, MaxIterations: 5, Hypothesis: "h",
		Scope: seedScope(12), PrevScopeSize: 13,
		FollowCallGraph: true, MaxExpandedScope: -1,
		CallGraph: fatGraph(named, 15),
		Verdict: Verdict{Outcome: Unverifiable, NextScope: named,
			DataRequests: []DataRequest{{SQL: "SELECT 3", Why: "x"}}},
	}
	d := DecideStep(in)
	if d.Decision != "continue" {
		t.Fatalf("expected continue, got %s (%s)", d.Decision, d.StopReason)
	}
	if got := d.NextScope.size(); got != 2+30 {
		t.Fatalf("MaxExpandedScope<0 must be unlimited: want 32, got %d", got)
	}
}

func TestAdvanceThreadsNamedSizeAsPrev(t *testing.T) {
	named := []string{"a.go:A", "b.go:B", "c.go:C"}
	st := LoopState{
		Iteration: 0, MaxIterations: 5, Hypothesis: "h",
		Scope: seedScope(12), PrevScopeSize: 13, Follow: true,
	}
	res := Advance(&st, Verdict{Outcome: Unverifiable, NextScope: named,
		DataRequests: []DataRequest{{SQL: "SELECT 4", Why: "x"}}},
		fatGraph(named, 10))
	if res.Stop {
		t.Fatalf("expected continue, stopped by %s", res.StoppedBy)
	}
	if st.PrevScopeSize != 3 {
		t.Fatalf("PrevScopeSize must thread the NAMED size (3), got %d", st.PrevScopeSize)
	}
	if st.Scope.size() != defaultMaxExpandedScope {
		t.Fatalf("adopted scope must be the capped expansion (%d), got %d", defaultMaxExpandedScope, st.Scope.size())
	}
}
