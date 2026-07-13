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

// F2.2 revise loop + F2.3 reframe: the round/cap mapping, exercised on the
// REAL function (applyCouncilCaps), not a mirror. shouldRevise = revise AND
// round<max; a revise out of rounds becomes 'exhausted' (terminal), not a
// silent approve. shouldReframe = FIRST rejection with rounds left — one
// narrower replan before escalation; a second rejection or a spent cap
// escalates.
func TestApplyCouncilCaps(t *testing.T) {
	cases := []struct {
		name          string
		decision      string
		round, max    int
		rejectedCount int
		wantDecision  string
		wantRevise    bool
		wantReframe   bool
	}{
		{"revise round 1 of 2 → loop", "revise", 1, 2, 0, "revise", true, false},
		{"revise round 2 of 2 → exhausted", "revise", 2, 2, 0, "exhausted", false, false},
		{"approved is terminal", "approved", 1, 2, 0, "approved", false, false},
		{"revise past the cap (defensive)", "revise", 3, 2, 0, "exhausted", false, false},
		{"first rejection with rounds left → reframe", "rejected", 1, 3, 1, "rejected", false, true},
		{"second rejection → escalate", "rejected", 2, 3, 2, "rejected", false, false},
		{"rejection at the cap → escalate", "rejected", 3, 3, 1, "rejected", false, false},
		{"rejection, count failed closed (=2) → escalate", "rejected", 1, 3, 2, "rejected", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, _, revise, reframe := applyCouncilCaps(c.decision, "by", c.round, c.max, c.rejectedCount)
			if d != c.wantDecision || revise != c.wantRevise || reframe != c.wantReframe {
				t.Fatalf("applyCouncilCaps(%s,r%d,max%d,rej%d) = (%s,revise=%v,reframe=%v), want (%s,%v,%v)",
					c.decision, c.round, c.max, c.rejectedCount, d, revise, reframe,
					c.wantDecision, c.wantRevise, c.wantReframe)
			}
		})
	}
}
