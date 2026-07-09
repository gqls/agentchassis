package diagnose

import (
	"strings"
	"testing"
)

// --- test doubles ----------------------------------------------------------

type fakeGather struct{ n int }

func (f *fakeGather) Gather(_ string, _ Scope) (string, error) {
	f.n++
	return "bundle-" + itoa(f.n) + ".md", nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

// scriptVerdicter returns a pre-scripted verdict per iteration.
type scriptVerdicter struct {
	verdicts []Verdict
	i        int
}

func (s *scriptVerdicter) Assess(_, _ string) (Verdict, error) {
	v := s.verdicts[s.i]
	s.i++
	return v, nil
}

// fakeCG returns fixed neighbours, to test call-graph following.
type fakeCG map[string][]string

func (f fakeCG) Neighbourhood(sym string) []string { return f[sym] }

func cite(where, quote string, t Tier) Citation {
	return Citation{Tier: t, Where: where, Quote: quote}
}

// --- tests -----------------------------------------------------------------

func TestConfirmedHappyPath(t *testing.T) {
	v := &scriptVerdicter{verdicts: []Verdict{
		// tier-covered: mechanism (static) + occurrence (runtime); a single-tier
		// confirm no longer stands (see coerceVerdict).
		{Outcome: Confirmed, Citations: []Citation{
			cite("save_page_sections_action.go:SavePageSectionsAction", "content regression blocked", TierRuntime),
			cite("save_page_sections_action.go:SavePageSectionsAction", "if newLen < existingLen/2 { block }", TierStatic)}},
	}}
	res, err := Run("hypothesis A", Scope{Symbols: []string{"a.go"}}, &fakeGather{}, v, nil, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Confirmed || res.StoppedBy != "confirmed" {
		t.Fatalf("expected confirmed, got %v / %s", res.Status, res.StoppedBy)
	}
	if !strings.Contains(res.Conclusion, "CONFIRMED") || !strings.Contains(res.Conclusion, "regression blocked") {
		t.Fatalf("conclusion missing cause/citation: %s", res.Conclusion)
	}
	if len(res.Trail) != 1 {
		t.Fatalf("expected 1 step, got %d", len(res.Trail))
	}
}

func TestRefuteThenConfirm_FollowsCallGraph(t *testing.T) {
	// iter1 REFUTED, names spawn_content_writer as next scope; call graph expands
	// it to the content-writer action; iter2 CONFIRMED there.
	v := &scriptVerdicter{verdicts: []Verdict{
		{Outcome: Refuted,
			Citations:         []Citation{cite("save_page_sections_action.go", "sections DO reach save", TierRuntime)},
			RevisedHypothesis: "the generation is short upstream",
			NextScope:         []string{"plan_sections_action.go:spawn_content_writer"}},
		{Outcome: Confirmed,
			Citations: []Citation{
				cite("content_writer.go:generate_content", "max_tokens 2000", TierStatic),
				cite("agent_error_log", "content generated: 1998 tokens (cap hit)", TierRuntime)}},
	}}
	cg := fakeCG{"plan_sections_action.go:spawn_content_writer": {"content_writer.go:generate_content"}}
	res, err := Run("sections never reach save", Scope{Symbols: []string{"save_page_sections_action.go:SavePageSectionsAction"}}, &fakeGather{}, v, cg, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Confirmed {
		t.Fatalf("expected confirmed after refute, got %v (%s)", res.Status, res.StoppedBy)
	}
	if len(res.Trail) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(res.Trail))
	}
	// iter2's scope should contain the call-graph-followed symbol
	got := res.Trail[1].Scope.Symbols
	found := false
	for _, s := range got {
		if s == "content_writer.go:generate_content" {
			found = true
		}
	}
	if !found {
		t.Fatalf("re-scope did not follow call graph; iter2 scope = %v", got)
	}
}

func TestNoCitationCoercedToUnverifiable(t *testing.T) {
	// A "Confirmed" with no citation must NOT confirm — coerced to Unverifiable,
	// then the loop runs to the iteration cap (here the same empty verdict).
	empty := Verdict{Outcome: Confirmed} // no citations
	v := &scriptVerdicter{verdicts: []Verdict{empty, empty}}
	res, err := Run("h", Scope{Symbols: []string{"a.go"}}, &fakeGather{}, v, nil, Config{MaxIterations: 2})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status == Confirmed {
		t.Fatal("a citation-less verdict must NOT confirm")
	}
	if res.Trail[0].Verdict.Outcome != Unverifiable {
		t.Fatalf("expected coercion to Unverifiable, got %s", res.Trail[0].Verdict.Outcome)
	}
}

func TestIterationCap(t *testing.T) {
	// Always Unverifiable with fresh evidence → should stop at the cap, not loop forever.
	mk := func(i int) Verdict {
		return Verdict{Outcome: Unverifiable,
			Citations:      []Citation{cite("x.go", "clue "+itoa(i), TierStatic)},
			NeededEvidence: "more",
			NextScope:      []string{"x.go"}}
	}
	v := &scriptVerdicter{verdicts: []Verdict{mk(1), mk(2), mk(3)}}
	res, err := Run("h", Scope{Symbols: []string{"x.go"}}, &fakeGather{}, v, nil, Config{MaxIterations: 3})
	if err != nil {
		t.Fatal(err)
	}
	if res.StoppedBy != "iteration-cap" {
		t.Fatalf("expected iteration-cap, got %s", res.StoppedBy)
	}
	if len(res.Trail) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(res.Trail))
	}
}

func TestEvidenceNotGrowing(t *testing.T) {
	// Same citation twice with no new evidence → evidence-not-growing stop.
	same := Verdict{Outcome: Refuted,
		Citations:         []Citation{cite("x.go", "same clue", TierStatic)},
		RevisedHypothesis: "h2", NextScope: []string{"x.go"}}
	v := &scriptVerdicter{verdicts: []Verdict{same, same}}
	res, err := Run("h1", Scope{Symbols: []string{"x.go"}}, &fakeGather{}, v, nil, Config{MaxIterations: 5})
	if err != nil {
		t.Fatal(err)
	}
	if res.StoppedBy != "evidence-not-growing" {
		t.Fatalf("expected evidence-not-growing, got %s", res.StoppedBy)
	}
}

func TestScopeNotNarrowing(t *testing.T) {
	// A verdict whose NextScope balloons well beyond the previous size → stop.
	big := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		big = append(big, "f"+itoa(i)+".go")
	}
	v := &scriptVerdicter{verdicts: []Verdict{
		{Outcome: Refuted,
			Citations:         []Citation{cite("x.go", "clue", TierStatic)},
			RevisedHypothesis: "h2", NextScope: big},
	}}
	// start with a small scope so the balloon trips the guard
	res, err := Run("h1", Scope{Symbols: []string{"x.go"}}, &fakeGather{}, v, nil, Config{MaxIterations: 5, FollowCallGraph: false})
	if err != nil {
		t.Fatal(err)
	}
	if res.StoppedBy != "scope-not-narrowing" {
		t.Fatalf("expected scope-not-narrowing, got %s", res.StoppedBy)
	}
}

func TestHypothesisThrash(t *testing.T) {
	// Oscillate h1->h2->h1 with no NEW evidence on the repeat → thrash stop.
	// Need fresh evidence on steps 1 and 2 (so they don't trip evidence-not-growing),
	// then a repeat hypothesis with stale evidence on step 3.
	v := &scriptVerdicter{verdicts: []Verdict{
		{Outcome: Refuted, Citations: []Citation{cite("a.go", "c1", TierStatic)}, RevisedHypothesis: "h2", NextScope: []string{"a.go"}},
		{Outcome: Refuted, Citations: []Citation{cite("b.go", "c2", TierStatic)}, RevisedHypothesis: "h1", NextScope: []string{"b.go"}},
		{Outcome: Refuted, Citations: []Citation{cite("a.go", "c1", TierStatic)}, RevisedHypothesis: "h2", NextScope: []string{"a.go"}}, // repeat h2, stale evidence
	}}
	res, err := Run("h1", Scope{Symbols: []string{"a.go"}}, &fakeGather{}, v, nil, Config{MaxIterations: 5})
	if err != nil {
		t.Fatal(err)
	}
	// Either thrash or evidence-not-growing is an acceptable halt here (both are
	// "stop spinning" guards); assert it did NOT run away and did NOT confirm.
	if res.Status == Confirmed {
		t.Fatal("must not confirm while thrashing")
	}
	if res.StoppedBy != "hypothesis-thrash" && res.StoppedBy != "evidence-not-growing" {
		t.Fatalf("expected a spin guard to fire, got %s", res.StoppedBy)
	}
}
