package actions

import (
	"reflect"
	"strings"
	"testing"
)

// bugs_open/126 candidate 2. Candidate 1 (the runner's `reload` step) stops a
// gated tool FALSE-failing. This is the other half: a GENUINE failure on a tool
// whose fence guards a consent gate or disclaimer must not be handed to
// tool-improver, because the only edits that turn such a fence green are to
// weaken, hide or delete the very markup the fence exists to protect — and a
// green run afterwards would look exactly like success.
//
// The mechanism is the escalation the convergence guard already owns: a fence
// carrying top-level `no_auto_fix: true` routes a failing verdict to the SAME
// acceptance_stuck / human-review item, regardless of how many cycles have been
// spent. Opt-in and default-OFF, so every fence written before the key existed
// behaves exactly as it did — which is the invariant these tests exist to hold.

// A fence document with the flag set, as a criteria author would write it.
const fenceWithNoAutoFix = `{
  "profiles": ["desktop", "mobile"],
  "no_auto_fix": true,
  "no_auto_fix_reason": "section B consent gate — owner-approved wording, proximate placement is load-bearing",
  "checks": [{"id": "mobile-fit", "type": "no_overflow", "selector": ".tool-container"}]
}`

// The same fence as it is written today — no mention of the key at all.
const fenceWithoutNoAutoFix = `{
  "profiles": ["desktop", "mobile"],
  "checks": [{"id": "mobile-fit", "type": "no_overflow", "selector": ".tool-container"}]
}`

// A failing verdict on a fence that forbids auto-fixing must raise the human
// escalation and NOT an improve_tool item — even on the FIRST cycle, where the
// convergence guard would happily have dispatched the fixer.
func TestNoAutoFixFence_EscalatesOnTheFirstFailure(t *testing.T) {
	run := runJudgeFailPathCriteria(t, nil, 0, nil, "comp-1", 1, fenceWithNoAutoFix)

	if got := run.raisedItemType(); got != "acceptance_stuck" {
		t.Fatalf("a no_auto_fix fence must escalate, got %q\nSQL: %s", got, run.itemInsert())
	}
	if !strings.Contains(run.itemInsert(), "'needs_human_review'") {
		t.Errorf("escalation must be raised AT needs_human_review:\n%s", run.itemInsert())
	}
	if !strings.Contains(run.itemInsert(), "'human-review'") {
		t.Errorf("escalation must be handled by human-review, never tool-improver:\n%s", run.itemInsert())
	}
	// item_type and item_key shape are deliberately the SAME as the stuck
	// producer's: one type, one dedup key, two reasons.
	if !strings.HasPrefix(run.itemKey, "acceptance_stuck:tool-loot-table-balancer:") {
		t.Errorf("escalation must reuse the existing dedup key shape, got %q", run.itemKey)
	}
	if escalated, _ := run.out["escalated"].(bool); !escalated {
		t.Errorf("result must report escalated=true, got %v", run.out["escalated"])
	}
	if created, _ := run.out["improve_tool_created"].(bool); created {
		t.Errorf("no improve_tool item may be raised for a no_auto_fix fence")
	}
	if flagged, _ := run.out["no_auto_fix"].(bool); !flagged {
		t.Errorf("result must report no_auto_fix=true, got %v", run.out["no_auto_fix"])
	}
	// The item has to say WHICH escalation this is, or a human cannot tell
	// "the fixer gave up" from "the fixer was never allowed to start".
	if !strings.Contains(run.spec, `"no_auto_fix":true`) {
		t.Errorf("spec must carry the no_auto_fix flag: %s", run.spec)
	}
	if !strings.Contains(run.spec, "section B consent gate") {
		t.Errorf("spec must carry the author's reason: %s", run.spec)
	}
	if !strings.Contains(run.spec, "why_escalated") || !strings.Contains(run.spec, "no_auto_fix:") {
		t.Errorf("why_escalated must name the no_auto_fix reason, not the cycle count: %s", run.spec)
	}
	if strings.Contains(run.spec, "improve_tool cycle(s) since the last passing") {
		t.Errorf("why_escalated must NOT claim exhausted fix cycles — none were spent: %s", run.spec)
	}
	// The summary is what a triaging human reads first in the queue.
	if !strings.Contains(run.summary, "no_auto_fix") {
		t.Errorf("summary must say the fence forbids auto-fix, got %q", run.summary)
	}
	if strings.Contains(run.summary, "not converging") {
		t.Errorf("summary must not blame non-convergence on a first failure, got %q", run.summary)
	}
	// The note is the loop's own durable record.
	if strings.Contains(run.note, "Fix: improve_tool item created") {
		t.Errorf("acceptance-fail note claims a fixer was dispatched at a protected fence:\n%s", run.note)
	}
	if !strings.Contains(run.note, "no_auto_fix") || !strings.Contains(run.note, "human review") {
		t.Errorf("acceptance-fail note does not record the no_auto_fix escalation:\n%s", run.note)
	}
}

// A fence with no_auto_fix that is ALSO stuck escalates once, under the same
// key, and the protected-markup reason is the one a human is shown: it decides
// what may be changed, where the cycle count is only context.
func TestNoAutoFixFence_TakesPrecedenceOverTheCycleCount(t *testing.T) {
	run := runJudgeFailPathCriteria(t, nil, 2, nil, "comp-1", 1, fenceWithNoAutoFix)

	if got := run.raisedItemType(); got != "acceptance_stuck" {
		t.Fatalf("expected an escalation item, got %q", got)
	}
	if !strings.Contains(run.spec, `"no_auto_fix":true`) {
		t.Errorf("spec must still carry the flag when both reasons hold: %s", run.spec)
	}
	if !strings.Contains(run.note, "no_auto_fix") {
		t.Errorf("note must lead with the no_auto_fix reason when both hold:\n%s", run.note)
	}
	// The cycles really were spent, so the count must still be recorded honestly.
	if spent, _ := run.out["fix_cycles_spent"].(int); spent != 2 {
		t.Errorf("fix_cycles_spent = %v, want 2", run.out["fix_cycles_spent"])
	}
}

// The invariant that matters most: a fence that does not mention the key is
// untouched by this change. Asserted as an EQUALITY against the shapes that must
// all mean "false" — absent key, explicit false, and criteria that do not parse
// at all — so a future edit cannot make one of them diverge quietly.
func TestNoAutoFixAbsent_IsExactlyTodaysBehaviour(t *testing.T) {
	baseline := runJudgeFailPathCriteria(t, nil, 0, nil, "comp-1", 1, fenceWithoutNoAutoFix)

	if got := baseline.raisedItemType(); got != "improve_tool" {
		t.Fatalf("a fence without the flag must still auto-fix, got %q\nSQL: %s", got, baseline.itemInsert())
	}
	if created, _ := baseline.out["improve_tool_created"].(bool); !created {
		t.Errorf("improve_tool_created should be true; out=%v", baseline.out)
	}
	if escalated, _ := baseline.out["escalated"].(bool); escalated {
		t.Errorf("an unflagged fence must not escalate")
	}
	// Absent means ABSENT: a workflow branching on key presence must not start
	// seeing a new key on every ordinary verdict.
	if _, present := baseline.out["no_auto_fix"]; present {
		t.Errorf("result map gained a no_auto_fix key on an unflagged verdict: %v", baseline.out)
	}
	if !strings.Contains(baseline.note, "Fix: improve_tool item created") {
		t.Errorf("note should record the queued fix:\n%s", baseline.note)
	}
	if !strings.Contains(baseline.spec, "acceptance_test") {
		t.Errorf("improve_tool spec lost its criteria: %s", baseline.spec)
	}
	if !strings.HasPrefix(baseline.itemKey, "acceptance_fail:tool-loot-table-balancer:") {
		t.Errorf("normal path must keep its own dedup key, got %q", baseline.itemKey)
	}

	// Every other "the flag is false" shape must land in exactly the same place.
	for _, tc := range []struct{ name, criteria string }{
		{"explicit false", `{"no_auto_fix": false, "checks": [{"id": "mobile-fit"}]}`},
		{"malformed — truncated mid-document", `{"profiles": ["desktop"], "checks": [{"id": "mobi`},
		{"malformed — not JSON at all", "no_auto_fix: true"},
		{"wrong type — a string, not a bool", `{"no_auto_fix": "true", "checks": []}`},
		{"empty criteria", ""},
		{"JSON null", "null"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := runJudgeFailPathCriteria(t, nil, 0, nil, "comp-1", 1, tc.criteria)
			if got := run.raisedItemType(); got != "improve_tool" {
				t.Fatalf("must fall back to the normal path, got %q", got)
			}
			if _, present := run.out["no_auto_fix"]; present {
				t.Errorf("result map gained a no_auto_fix key: %v", run.out)
			}
			if !reflect.DeepEqual(run.out, baseline.out) {
				t.Errorf("result differs from the unflagged baseline:\n got %v\nwant %v", run.out, baseline.out)
			}
			// The note carries no criteria, so it must be byte-identical.
			if run.note != baseline.note {
				t.Errorf("acceptance-fail note differs from the unflagged baseline:\n got %s\nwant %s", run.note, baseline.note)
			}
			if run.summary != baseline.summary {
				t.Errorf("summary differs from the unflagged baseline:\n got %q\nwant %q", run.summary, baseline.summary)
			}
			if run.itemKey != baseline.itemKey {
				t.Errorf("item_key differs from the unflagged baseline: got %q want %q", run.itemKey, baseline.itemKey)
			}
		})
	}
}

// No content_components row: the known gap the `stuck` producer already had.
// Nothing can be escalated cleanly — but nothing is dispatched at the protected
// markup either, which is the direction that matters. The note must say so
// rather than read as a broken insert.
func TestNoAutoFixFence_WithNoComponentRaisesNothingAndSaysSo(t *testing.T) {
	// The driver sets NO exec expectation when the component id is "", so
	// sqlmock fails the test if the judge writes a work item of either kind.
	run := runJudgeFailPathCriteria(t, nil, 0, nil, "", 1, fenceWithNoAutoFix)

	if got := run.raisedItemType(); got != "" {
		t.Fatalf("no component means no work item of any kind, got %q", got)
	}
	if created, _ := run.out["improve_tool_created"].(bool); created {
		t.Errorf("nothing was inserted, so improve_tool_created must be false; out=%v", run.out)
	}
	if escalated, _ := run.out["escalated"].(bool); escalated {
		t.Errorf("nothing was inserted, so escalated must be false; out=%v", run.out)
	}
	if !strings.Contains(run.note, "no_auto_fix") {
		t.Errorf("note must still record that the fence forbids auto-fixing:\n%s", run.note)
	}
	if !strings.Contains(run.note, "route this manually") {
		t.Errorf("note must tell the reader this needs manual routing:\n%s", run.note)
	}
}

// A fence that sets the flag but writes no reason must still produce readable
// prose — "no_auto_fix ()" reads as a defect in the message rather than as a
// deliberate protection.
func TestNoAutoFixFence_WithoutAReasonStillReads(t *testing.T) {
	run := runJudgeFailPathCriteria(t, nil, 0, nil, "comp-1", 1,
		`{"no_auto_fix": true, "checks": [{"id": "mobile-fit"}]}`)

	if got := run.raisedItemType(); got != "acceptance_stuck" {
		t.Fatalf("expected an escalation item, got %q", got)
	}
	if !strings.Contains(run.spec, `"no_auto_fix":true`) {
		t.Errorf("spec must carry the flag: %s", run.spec)
	}
	if strings.Contains(run.spec, `"no_auto_fix_reason"`) {
		t.Errorf("an unstated reason must not be written as an empty spec key: %s", run.spec)
	}
	if !strings.Contains(run.note, "no reason stated on the fence") {
		t.Errorf("note must say the reason was not stated:\n%s", run.note)
	}
}

// The parser in isolation: the fallback is what keeps the flag safe to add to a
// shared schema, so it is asserted directly as well as through the judge.
func TestParseNoAutoFix(t *testing.T) {
	for _, tc := range []struct {
		name       string
		criteria   string
		wantFlag   bool
		wantReason string
	}{
		{"flag with a reason", fenceWithNoAutoFix, true,
			"section B consent gate — owner-approved wording, proximate placement is load-bearing"},
		{"flag alone", `{"no_auto_fix":true}`, true, ""},
		{"reason is trimmed", `{"no_auto_fix":true,"no_auto_fix_reason":"  consent gate  "}`, true, "consent gate"},
		{"a reason without the flag protects nothing", `{"no_auto_fix_reason":"consent gate"}`, false, "consent gate"},
		{"absent", fenceWithoutNoAutoFix, false, ""},
		{"explicit false", `{"no_auto_fix":false}`, false, ""},
		{"empty", "", false, ""},
		{"whitespace only", "   \n\t", false, ""},
		{"truncated", `{"no_auto_fix":tr`, false, ""},
		{"not JSON", "no_auto_fix: true", false, ""},
		{"wrong type", `{"no_auto_fix":"true"}`, false, ""},
		{"JSON null", "null", false, ""},
		{"a JSON array, not an object", `[{"no_auto_fix":true}]`, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flag, reason := parseNoAutoFix(tc.criteria)
			if flag != tc.wantFlag {
				t.Errorf("flag = %v, want %v", flag, tc.wantFlag)
			}
			if reason != tc.wantReason {
				t.Errorf("reason = %q, want %q", reason, tc.wantReason)
			}
		})
	}
}
