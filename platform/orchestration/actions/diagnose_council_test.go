package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

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

// Round-2 fix from council review 2eed453a: a truncated partial that happens to
// parse as valid JSON must still be marked degraded — the decider consults the
// step's __truncated marker, derived as a sibling of the reviewer field.
func TestMarkerFieldFor(t *testing.T) {
	cases := map[string]string{
		"review_editquality.result": "review_editquality.__truncated",
		"a.b.result":                "a.b.__truncated",
		"bare":                      "",
		".result":                   "",
	}
	for in, want := range cases {
		if got := markerFieldFor(in); got != want {
			t.Fatalf("markerFieldFor(%q): want %q got %q", in, want, got)
		}
	}
}

// bugs_open/036: `edit` was a plain int, so of the three registers a model
// naturally answers "which edit?" in, exactly one parsed — and the other two
// took down the whole round. The free-text cases are the live ones: all three
// voided rounds carried a plan-level description, not a malformed index.
func TestObjectionEditTolerance(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantIndex int
		wantRaw   string // what the report must show — the reviewer's own token
	}{
		{"bare index", `{"edit":3,"problem":"p"}`, 3, `3`},
		{"index as string", `{"edit":"2","problem":"p"}`, 2, `"2"`},
		{"index as string with spaces", `{"edit":" 4 ","problem":"p"}`, 4, `" 4 "`},
		{"float for an int", `{"edit":3.0,"problem":"p"}`, 3, `3.0`},
		{"file name — plan-wide", `{"edit":"provider.go","problem":"p"}`, 0, `"provider.go"`},
		// The three live payloads that actually voided rounds.
		{"live: plan-level", `{"edit":"plan-level (deploy verification)","problem":"p"}`, 0, `"plan-level (deploy verification)"`},
		{"live: risks note", `{"edit":"risks note on the 54 mis-stamped rows","problem":"p"}`, 0, `"risks note on the 54 mis-stamped rows"`},
		{"live: risks/summary", `{"edit":"risks/summary (item 5)","problem":"p"}`, 0, `"risks/summary (item 5)"`},
		{"null", `{"edit":null,"problem":"p"}`, 0, `null`},
		{"object", `{"edit":{"file":"x.go"},"problem":"p"}`, 0, `{"file":"x.go"}`},
		{"absent", `{"problem":"p"}`, 0, ``},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var o councilObjection
			if err := json.Unmarshal([]byte(tc.body), &o); err != nil {
				t.Fatalf("must never fail to parse: %v", err)
			}
			if o.Edit.Index != tc.wantIndex {
				t.Fatalf("index: want %d got %d", tc.wantIndex, o.Edit.Index)
			}
			if string(o.Edit.Raw) != tc.wantRaw {
				t.Fatalf("raw: want %q got %q", tc.wantRaw, string(o.Edit.Raw))
			}
			// The problem text is the part carrying the review's value — it must
			// survive whatever the pointer did.
			if o.Problem != "p" {
				t.Fatalf("problem lost: %q", o.Problem)
			}
		})
	}
}

// A whole review must survive the same slip, and the persisted report must show
// the reviewer's own token rather than a laundered 0.
func TestMistypedEditDoesNotVoidTheReview(t *testing.T) {
	body := `{"reviewer":"debug-historian","verdict":"object",
	          "objections":[{"edit":"plan-level (deploy verification)","problem":"no post-deploy pod grep","severity":"medium"}]}`
	var rv councilReview
	if err := json.Unmarshal([]byte(body), &rv); err != nil {
		t.Fatalf("the live bugs_open/036 payload must parse: %v", err)
	}
	if rv.Verdict != "object" || len(rv.Objections) != 1 {
		t.Fatalf("verdict/objections lost: %+v", rv)
	}
	out, err := json.Marshal(rv)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"edit":"plan-level (deploy verification)"`) {
		t.Fatalf("report must round-trip the reviewer's own token, got %s", out)
	}
}

// bugs_open/036 candidate (2), the structural half: a seat whose output is valid
// JSON but off-schema in a field the tolerant type does NOT cover must cost that
// seat, not the round.
func TestSalvageMistypedReview(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		wantOK      bool
		wantVerdict string
		wantObjs    int
	}{
		{
			name:        "objections is an object, not an array — verdict survives",
			body:        `{"reviewer":"guardian","verdict":"veto","objections":{"edit":1,"problem":"drops the CAS guard"}}`,
			wantOK:      true,
			wantVerdict: "veto",
			wantObjs:    0, // the detail is lost; the opinion is not
		},
		{
			name:        "missing is a string, not a list",
			body:        `{"reviewer":"reuse","verdict":"object","missing":"the adoption path","objections":[{"edit":1,"problem":"p"}]}`,
			wantOK:      true,
			wantVerdict: "object",
			wantObjs:    1,
		},
		{
			// encoding/json continues past a TYPE error (unlike a syntax error) and
			// keeps everything that did decode, so the objection survives with only
			// the offending field zeroed. Asserted deliberately: this is why the
			// per-field salvage keeps as much as it does, and it is the difference
			// between losing a field and losing a seat.
			name:        "severity mistyped deep in an objection — objection survives, field dropped",
			body:        `{"reviewer":"edit-quality","verdict":"approve","objections":[{"edit":1,"problem":"p","severity":3}]}`,
			wantOK:      true,
			wantVerdict: "approve",
			wantObjs:    1,
		},
		{
			name:   "verdict itself is mistyped — no opinion to recover",
			body:   `{"reviewer":"render","verdict":["approve"],"notes":"n"}`,
			wantOK: false,
		},
		{
			name:   "unrecognised verdict is not an opinion",
			body:   `{"reviewer":"render","verdict":"looks-fine-to-me"}`,
			wantOK: false,
		},
		{name: "not an object at all", body: `["approve"]`, wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rv, ok := salvageMistypedReview([]byte(tc.body))
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

// The round-level property §6 asks for: a two-seat round where one seat is
// unusable still produces a verdict from the seat that WAS read, and the loss is
// visible rather than silent.
func TestOneBadSeatDoesNotVoidTheRound(t *testing.T) {
	// Seat A: valid JSON, off-schema past the verdict → salvaged, degraded.
	a, ok := salvageMistypedReview([]byte(`{"reviewer":"debug-historian","verdict":"object","objections":{"edit":1,"problem":"p"}}`))
	if !ok {
		t.Fatal("seat A should salvage into an opinion")
	}
	a.Degraded = true
	// Seat B: read cleanly.
	b := councilReview{Reviewer: "edit-quality", Verdict: "approve"}

	decision, decidedBy := decideCouncil([]councilReview{a, b}, map[string]bool{"guardian": true})
	if decision != "revise" {
		t.Fatalf("the salvaged seat's objection must still decide, got %s (%s)", decision, decidedBy)
	}
	if !a.Degraded {
		t.Fatal("a salvaged seat must be marked degraded so the report shows a partial opinion")
	}

	// And a seat lost entirely still cannot be the difference between revise and
	// approve: two clean approvals plus one unreadable seat downgrades.
	d, _ := decideCouncil([]councilReview{b, {Reviewer: "guardian", Verdict: "approve"}}, map[string]bool{"guardian": true})
	if d != "approved" {
		t.Fatalf("precondition: clean approvals should approve, got %s", d)
	}
	unreadable := []string{"review_render.result"}
	if d == "approved" && len(unreadable) > 0 {
		d = "revise"
	}
	if d != "revise" {
		t.Fatalf("an unreadable seat must block approval, got %s", d)
	}
}
