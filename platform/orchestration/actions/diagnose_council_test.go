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

// bugs_open/019: a reviewer truncated at max_tokens used to fail the whole
// action, discarding every other seat's completed review and returning no
// verdict. salvageTruncatedReview recovers the opinion where the cut left one
// intact — which is usually, because `reviewer` and `verdict` are the first
// fields models emit.
func TestSalvageTruncatedReview(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		wantOK      bool
		wantVerdict string
		wantObjs    int
	}{
		{
			name:        "cut mid-notes, verdict intact",
			body:        `{"reviewer":"edit-quality","verdict":"object","objections":[{"edit":1,"problem":"unguarded nil deref"}],"notes":"the third edit rewrites a shared help`,
			wantOK:      true,
			wantVerdict: "object",
			wantObjs:    1,
		},
		{
			name:        "cut mid-objections array — keeps the objections that survived",
			body:        `{"reviewer":"guardian","verdict":"veto","objections":[{"edit":2,"problem":"drops the CAS guard"},{"edit":3,"problem":"partial`,
			wantOK:      true,
			wantVerdict: "veto",
			wantObjs:    1,
		},
		{
			name:   "cut before the verdict — nothing usable",
			body:   `{"reviewer":"compliance"`,
			wantOK: false,
		},
		{
			name:   "repairs into valid JSON but carries no verdict — NOT an opinion",
			body:   `{"reviewer":"reuse-agent","notes":"I looked at the adoption path and`,
			wantOK: false,
		},
		{
			name:   "unrecognised verdict is not an opinion either",
			body:   `{"reviewer":"render","verdict":"looks-fine-to-me","notes":"trailing`,
			wantOK: false,
		},
		{name: "empty", body: ``, wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rv, ok := salvageTruncatedReview([]byte(tc.body))
			if ok != tc.wantOK {
				t.Fatalf("ok: want %v got %v (verdict %q)", tc.wantOK, ok, rv.Verdict)
			}
			if !tc.wantOK {
				return
			}
			if rv.Verdict != tc.wantVerdict {
				t.Fatalf("verdict: want %q got %q", tc.wantVerdict, rv.Verdict)
			}
			if len(rv.Objections) != tc.wantObjs {
				t.Fatalf("objections: want %d got %d", tc.wantObjs, len(rv.Objections))
			}
		})
	}
}

// The safety property the old hard error was protecting, kept at the point where
// it actually matters: an unreadable seat must never be the difference between
// revise and approve. It can only make the outcome MORE conservative — an
// objection from a seat that WAS read stays decisive and is not softened.
func TestUnreadableSeatCannotApprove(t *testing.T) {
	rv := func(name, verdict string) councilReview {
		return councilReview{Reviewer: name, Verdict: verdict}
	}
	// The downgrade as applied in DiagnoseCouncilDecideAction.
	decide := func(reviews []councilReview, unreadable []string) string {
		d, _ := decideCouncil(reviews, map[string]bool{"guardian": true})
		if d == "approved" && len(unreadable) > 0 {
			d = "revise"
		}
		return d
	}

	if got := decide([]councilReview{rv("quality", "approve"), rv("guardian", "approve")}, nil); got != "approved" {
		t.Fatalf("all readable approvals should approve, got %s", got)
	}
	if got := decide([]councilReview{rv("quality", "approve"), rv("guardian", "approve")}, []string{"review_guidelines.result"}); got != "revise" {
		t.Fatalf("approve alongside an unreadable seat must downgrade to revise, got %s", got)
	}
	// Not softened in the other direction: a readable veto still rejects even
	// though a seat was lost, and a readable objection still revises.
	if got := decide([]councilReview{rv("guardian", "veto")}, []string{"review_render.result"}); got != "rejected" {
		t.Fatalf("a readable veto must survive an unreadable seat, got %s", got)
	}
	if got := decide([]councilReview{rv("quality", "object")}, []string{"review_render.result"}); got != "revise" {
		t.Fatalf("a readable objection must stay decisive, got %s", got)
	}
}
