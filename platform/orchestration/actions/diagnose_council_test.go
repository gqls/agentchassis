package actions

import "testing"

// F2.1: the decision layer is deterministic — read the outcome off the rules.
func TestDecideCouncil(t *testing.T) {
	rv := func(name, verdict string) councilReview {
		return councilReview{Reviewer: name, Verdict: verdict}
	}
	guardianVeto := map[string]bool{"guardian": true}

	cases := []struct {
		name     string
		reviews  []councilReview
		hard     map[string]bool
		decision string
		byPrefix string
	}{
		{"all approve", []councilReview{rv("quality", "approve"), rv("guardian", "approve")}, guardianVeto, "approved", "all reviewers"},
		{"objection → revise", []councilReview{rv("quality", "object"), rv("guardian", "approve")}, guardianVeto, "revise", "objection from quality"},
		{"hard veto wins over objections", []councilReview{rv("quality", "object"), rv("guardian", "veto")}, guardianVeto, "rejected", "hard veto from guardian"},
		{"advisory veto still rejects", []councilReview{rv("quality", "veto"), rv("guardian", "approve")}, guardianVeto, "rejected", "veto from quality"},
		{"hard veto reported as hard", []councilReview{rv("guardian", "veto"), rv("quality", "veto")}, guardianVeto, "rejected", "hard veto from guardian"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, by := decideCouncil(tc.reviews, tc.hard)
			if d != tc.decision {
				t.Fatalf("decision: want %s got %s (by %s)", tc.decision, d, by)
			}
			if tc.byPrefix != "" && len(by) < len(tc.byPrefix) || by[:len(tc.byPrefix)] != tc.byPrefix {
				t.Fatalf("decided_by: want prefix %q got %q", tc.byPrefix, by)
			}
		})
	}
}

// F2.2 revise loop: the round/cap mapping. shouldRevise = revise AND round<max;
// a revise that runs out of rounds becomes 'exhausted' (terminal), not a
// silent approve.
func TestReviseCapMapping(t *testing.T) {
	// mirror the action's post-decision logic
	decide := func(decision string, round, maxRounds int) (string, bool) {
		shouldRevise := decision == "revise" && round < maxRounds
		if decision == "revise" && !shouldRevise {
			return "exhausted", false
		}
		return decision, shouldRevise
	}
	cases := []struct {
		decision     string
		round, max   int
		wantDecision string
		wantShould   bool
	}{
		{"revise", 1, 2, "revise", true},     // round 1 of 2 → loop
		{"revise", 2, 2, "exhausted", false}, // round 2 of 2 → give up, terminal
		{"approved", 1, 2, "approved", false},
		{"rejected", 1, 2, "rejected", false},
		{"revise", 3, 2, "exhausted", false}, // defensive: past the cap
	}
	for _, c := range cases {
		d, s := decide(c.decision, c.round, c.max)
		if d != c.wantDecision || s != c.wantShould {
			t.Fatalf("decide(%s,r%d,max%d) = (%s,%v), want (%s,%v)",
				c.decision, c.round, c.max, d, s, c.wantDecision, c.wantShould)
		}
	}
}
