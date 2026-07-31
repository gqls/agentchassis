// FILE: platform/orchestration/actions/prune_floor_test.go
//
// bugs_open/135. The verification rule the case file names is
// verify-the-failing-branch: "a green run over a healthy repo proves only that the
// guard is inert". So these tests assert the REFUSAL fires — with the numbers, and
// with the remedy in its text — as well as that a healthy run is untouched.
//
// Pure-function tests, no DB: the rule was deliberately split from the SQL that
// measures it (prune_floor.go), which is what makes every branch reachable here.

package actions

import (
	"strings"
	"testing"
)

func TestEvaluatePruneFloorAllowsAHealthyRun(t *testing.T) {
	v := evaluatePruneFloor(0.5, []pruneCohort{
		{Label: "kind=func", Confirmed: 3048, Stored: 3050},
		{Label: "kind=struct", Confirmed: 857, Stored: 857},
		{Label: "distinct paths", Confirmed: 530, Stored: 531},
	})
	if !v.Allowed {
		t.Fatalf("healthy run refused: %s", v.Reason("op", "subject", "prune_floor_ratio"))
	}
	if len(v.Failing) != 0 {
		t.Fatalf("expected no failing cohorts, got %d", len(v.Failing))
	}
	// The detail must be reported on a PASS too — candidate (3): the denominator
	// travels with the count, or `pruned: N` is an alarm presented as output.
	if len(v.Detail()) != 3 {
		t.Fatalf("expected 3 cohorts in Detail(), got %d", len(v.Detail()))
	}
}

func TestEvaluatePruneFloorRefusesATruncatedRun(t *testing.T) {
	// The motivating shape: a partial tarball. Every cohort shrinks together, so
	// the run looks successful and returns no error.
	v := evaluatePruneFloor(0.5, []pruneCohort{
		{Label: "kind=func", Confirmed: 1160, Stored: 3048},
		{Label: "kind=method", Confirmed: 400, Stored: 1025},
		{Label: "distinct paths", Confirmed: 218, Stored: 530},
	})
	if v.Allowed {
		t.Fatal("a run that confirmed ~38% of the index was ALLOWED to prune")
	}
	if len(v.Failing) != 3 {
		t.Fatalf("expected all 3 cohorts failing, got %d", len(v.Failing))
	}
	// Worst first, so a truncated log line still names the most alarming cohort.
	// (kind=func 1160/3048 = 38.1%, kind=method 400/1025 = 39.0%, paths 218/530 = 41.1%.)
	if v.Failing[0].Label != "kind=func" {
		t.Fatalf("expected the worst cohort (kind=func, 38%%) first, got %q", v.Failing[0].Label)
	}
	r := v.Reason("index_code_symbols: prune", "gqls/agentchassis at abc12345", "prune_floor_ratio")
	for _, want := range []string{"REFUSED", "gqls/agentchassis at abc12345", "1160 of 3048",
		"NOTHING was deleted", "prune_floor_ratio", "0 disables"} {
		if !strings.Contains(r, want) {
			t.Errorf("refusal text is missing %q; got:\n%s", want, r)
		}
	}
}

func TestEvaluatePruneFloorRefusesAWipedKindThatTheTotalWouldHide(t *testing.T) {
	// This is why cohorts exist. 4,930 of 4,963 rows re-confirmed — 99% — and one
	// whole kind gone. A single whole-corpus ratio passes this run happily.
	cohorts := []pruneCohort{
		{Label: "kind=func", Confirmed: 3048, Stored: 3048},
		{Label: "kind=method", Confirmed: 1025, Stored: 1025},
		{Label: "kind=struct", Confirmed: 857, Stored: 857},
		{Label: "kind=interface", Confirmed: 0, Stored: 33},
		{Label: "distinct paths", Confirmed: 530, Stored: 530},
	}
	var confirmed, stored int
	for _, c := range cohorts {
		confirmed += c.Confirmed
		stored += c.Stored
	}
	if got := float64(confirmed) / float64(stored); got < 0.95 {
		t.Fatalf("test premise broken: whole-corpus ratio is %.2f, not the ~0.99 this case is about", got)
	}
	v := evaluatePruneFloor(0.5, cohorts)
	if v.Allowed {
		t.Fatal("a run that lost EVERY interface row was allowed to prune (the total hid it)")
	}
	if len(v.Failing) != 1 || v.Failing[0].Label != "kind=interface" {
		t.Fatalf("expected exactly kind=interface to fail, got %+v", v.Failing)
	}
	if !strings.Contains(v.Reason("op", "s", "k"), "0 of 33") {
		t.Errorf("refusal must name the numbers; got %s", v.Reason("op", "s", "k"))
	}
}

func TestEvaluatePruneFloorNewCohortCannotRefuse(t *testing.T) {
	// A class stored for the first time (Stored 0) has nothing to lose. If 0/0 read
	// as 0% it would refuse every prune for ever the moment a new kind appeared.
	v := evaluatePruneFloor(0.5, []pruneCohort{
		{Label: "kind=func", Confirmed: 100, Stored: 100},
		{Label: "kind=doc", Confirmed: 0, Stored: 0},
	})
	if !v.Allowed {
		t.Fatalf("an empty cohort refused the prune: %s", v.Reason("op", "s", "k"))
	}
}

func TestEvaluatePruneFloorDisabledIsAllowedAndSaysSo(t *testing.T) {
	// 0 is the documented escape hatch. It must permit the delete AND be legible:
	// "no refusal" must never be mistaken for "the guard looked and was content".
	v := evaluatePruneFloor(0, []pruneCohort{{Label: "kind=func", Confirmed: 1, Stored: 5000}})
	if !v.Allowed {
		t.Fatal("prune_floor_ratio=0 must disable the floor, not refuse")
	}
	if !v.Disabled {
		t.Fatal("a disabled floor must report itself as disabled")
	}
	r := v.Reason("index_code_symbols: prune", "repo", "prune_floor_ratio")
	if !strings.Contains(r, "DISABLED") || !strings.Contains(r, "unchecked") {
		t.Errorf("a disabled guard must say so loudly; got %s", r)
	}
}

func TestEvaluatePruneFloorNegativeDisablesAndFlagsClamped(t *testing.T) {
	v := evaluatePruneFloor(-1, []pruneCohort{{Label: "kind=func", Confirmed: 1, Stored: 5000}})
	if !v.Allowed || !v.Disabled || !v.Clamped {
		t.Fatalf("a negative floor should read as disabled+clamped, got %+v", v)
	}
}

func TestEvaluatePruneFloorAboveOneIsClampedNotUnsatisfiable(t *testing.T) {
	// An unclamped 1.5 refuses every prune for ever while looking like a working
	// guard — the failure mode where the wrong result is indistinguishable from the
	// right one. Clamped to 1.0: strict, but satisfiable by a run that saw it all.
	v := evaluatePruneFloor(1.5, []pruneCohort{{Label: "kind=func", Confirmed: 3048, Stored: 3048}})
	if !v.Clamped || v.Floor != 1 {
		t.Fatalf("expected clamp to 1.0, got floor=%v clamped=%v", v.Floor, v.Clamped)
	}
	if !v.Allowed {
		t.Fatal("a fully-confirmed cohort must satisfy a clamped floor of 1.0")
	}
}

func TestEvaluatePruneFloorBoundaryIsInclusive(t *testing.T) {
	// Exactly at the floor passes (Below is ratio < floor). Asserted because the
	// off-by-one here is invisible in production: it only ever shows up as a
	// refusal nobody expected, on a run that was fine.
	at := evaluatePruneFloor(0.5, []pruneCohort{{Label: "kind=func", Confirmed: 50, Stored: 100}})
	if !at.Allowed {
		t.Error("a cohort exactly at the floor must pass")
	}
	just := evaluatePruneFloor(0.5, []pruneCohort{{Label: "kind=func", Confirmed: 49, Stored: 100}})
	if just.Allowed {
		t.Error("a cohort one row below the floor must refuse")
	}
}

func TestEvaluatePruneFloorNoCohortsAllows(t *testing.T) {
	// Nothing stored at all — the first-ever run. There is nothing to protect and
	// the guard must not block bootstrapping.
	v := evaluatePruneFloor(0.5, nil)
	if !v.Allowed {
		t.Fatal("an empty corpus refused its first prune")
	}
	if !strings.Contains(v.Reason("op", "s", "k"), "nothing stored") {
		t.Errorf("expected the empty case to say so; got %s", v.Reason("op", "s", "k"))
	}
}

func TestPruneFloorCohortListCaps(t *testing.T) {
	var cohorts []pruneCohort
	for i := 0; i < 9; i++ {
		cohorts = append(cohorts, pruneCohort{Label: "kind=k" + string(rune('a'+i)), Confirmed: 0, Stored: 10})
	}
	r := evaluatePruneFloor(0.5, cohorts).Reason("op", "s", "k")
	if !strings.Contains(r, "and 3 more") {
		t.Errorf("expected the cohort list to cap at 6 and say how many were elided; got %s", r)
	}
}

func TestPruneFloorFromConfig(t *testing.T) {
	cases := []struct {
		name    string
		config  map[string]interface{}
		want    float64
		present bool
	}{
		{"absent", map[string]interface{}{}, defaultPruneFloorRatio, false},
		{"nil value", map[string]interface{}{"prune_floor_ratio": nil}, defaultPruneFloorRatio, false},
		{"json float", map[string]interface{}{"prune_floor_ratio": 0.75}, 0.75, true},
		{"json zero", map[string]interface{}{"prune_floor_ratio": 0.0}, 0, true},
		{"go int", map[string]interface{}{"prune_floor_ratio": 1}, 1, true},
		{"quoted", map[string]interface{}{"prune_floor_ratio": "0.25"}, 0.25, true},
		{"quoted blank", map[string]interface{}{"prune_floor_ratio": "  "}, defaultPruneFloorRatio, false},
		// A typo must fall back to the DEFAULT, never to 0. Falling back to 0 would
		// silently disable a destructive-operation guard — the wrong outcome looking
		// exactly like the right one.
		{"garbage", map[string]interface{}{"prune_floor_ratio": "half"}, defaultPruneFloorRatio, false},
		{"wrong type", map[string]interface{}{"prune_floor_ratio": []string{"0.5"}}, defaultPruneFloorRatio, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, present := pruneFloorFromConfig(tc.config, "prune_floor_ratio", defaultPruneFloorRatio)
			if got != tc.want || present != tc.present {
				t.Errorf("got (%v, %v), want (%v, %v)", got, present, tc.want, tc.present)
			}
		})
	}
}
