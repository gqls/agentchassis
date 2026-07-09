// loop_datarequest_test.go — regression tests for the data_request progress rule
// in guardAfter (DESIGN §3). A verdict that issues a NEW read-only data_request is
// making progress (its answer arrives in the NEXT gather), so evidence-not-growing
// must not stop the loop one iteration before that answer lands — the exact
// behaviour that truncated the live gamesdesign runs at iteration 3. A RE-issue of
// the SAME query is not new progress and must still trip the guard.
//
// These use the existing helpers in loop_test.go (scriptVerdicter, fakeGather,
// cite) — same package, so they are available here.
package diagnose

import "testing"

// A fresh data_request on a step whose citations are all stale keeps the loop
// alive to the iteration where the requested data arrives and a verdict lands.
func TestNewDataRequestDefersEvidenceNotGrowing(t *testing.T) {
	c1 := cite("x.go", "same clue", TierStatic)
	// step 1: new clue + a new request (work_items) — continues (clue is new).
	step1 := Verdict{
		Outcome: Unverifiable, Citations: []Citation{c1}, NextScope: []string{"x.go"},
		DataRequests: []DataRequest{{SQL: "SELECT 1 FROM work_items"}},
	}
	// step 2: SAME (now stale) clue, but a DIFFERENT request (pages) — without the
	// data_request rule this trips evidence-not-growing; with it, the loop continues.
	step2 := Verdict{
		Outcome: Unverifiable, Citations: []Citation{c1}, NextScope: []string{"x.go"},
		DataRequests: []DataRequest{{SQL: "SELECT 1 FROM pages WHERE page_type = 'index'"}},
	}
	// step 3: the requested data has now been gathered; a grounded, tier-covered
	// confirm lands (state row showing the effect + code showing the mechanism).
	confirm := Verdict{Outcome: Confirmed, Citations: []Citation{
		cite("pages", "content is 120 chars", TierState),
		cite("x.go:F", "truncates content to 120", TierStatic)}}

	v := &scriptVerdicter{verdicts: []Verdict{step1, step2, confirm}}
	res, err := Run("index page is a stub", Scope{Symbols: []string{"x.go"}}, &fakeGather{}, v, nil, Config{MaxIterations: 5})
	if err != nil {
		t.Fatal(err)
	}
	if res.StoppedBy == "evidence-not-growing" {
		t.Fatalf("a fresh data_request on step 2 must NOT trip evidence-not-growing; the loop stopped early")
	}
	if res.Status != Confirmed {
		t.Fatalf("expected the loop to reach the confirm once the requested data arrived, got %s (%s)", res.Status, res.StoppedBy)
	}
}

// Re-issuing the SAME data_request with stale citations is a genuine spin and must
// still trip evidence-not-growing (the rule rewards NEW requests, not repetition).
func TestReissuedDataRequestStillTripsEvidenceNotGrowing(t *testing.T) {
	same := Verdict{
		Outcome: Unverifiable, Citations: []Citation{cite("x.go", "same clue", TierStatic)},
		NextScope:    []string{"x.go"},
		DataRequests: []DataRequest{{SQL: "SELECT 1 FROM pages"}},
	}
	// iteration 1: clue + request both new → progress. iteration 2: both already
	// seen → no new evidence, no new request → spin guard fires.
	v := &scriptVerdicter{verdicts: []Verdict{same, same, same}}
	res, err := Run("h1", Scope{Symbols: []string{"x.go"}}, &fakeGather{}, v, nil, Config{MaxIterations: 5})
	if err != nil {
		t.Fatal(err)
	}
	if res.StoppedBy != "evidence-not-growing" {
		t.Fatalf("a re-issued data_request with stale citations should still trip evidence-not-growing, got %s", res.StoppedBy)
	}
}
