package orchestration

import (
	"fmt"
	"testing"
	"time"
)

// bugs_open/243. These pin the WRITER half of __step_errors. The reader half
// (reviewStepFailed, in the actions package) had four tests from the day it
// shipped; the writer had none, because it was inline in routeToErrorStep, which
// needs a DB. The two sides agree on a contract nothing asserted — the map is
// keyed by PLAIN STEP NAME — so a refactor of either could silently return the
// council to counting a lost seat as an abstention, with every test still green.
//
// That is the failure this file exists to make impossible.

var at = time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

// THE CONTRACT. The reader derives the step name from a collected-data path by
// splitting on the FIRST "." ("review_editquality.result" -> "review_editquality")
// and looks that up as a key. So the writer must key by the bare step name — no
// prefix, no suffix, no path. If this test and the reader's test ever disagree,
// one of them is wrong and the council miscounts silently.
func TestRecordStepErrorKeysByBareStepName(t *testing.T) {
	collected := map[string]interface{}{}
	recordStepError(collected, "review_editquality", "step review_editquality failed: 400 usage limit", at)

	m, ok := collected["__step_errors"].(map[string]interface{})
	if !ok {
		t.Fatal("__step_errors must be a map[string]interface{} — the reader type-asserts exactly this")
	}
	entry, ok := m["review_editquality"]
	if !ok {
		t.Fatalf("must be keyed by the BARE step name; got keys %v", keysOf(m))
	}
	e, ok := entry.(map[string]interface{})
	if !ok {
		t.Fatal("each entry must be a map — the reader reads entry[\"message\"] as a string")
	}
	if msg, _ := e["message"].(string); msg == "" {
		t.Error(`entry["message"] must carry the error text; the reader returns it so a loss is attributable`)
	}
	if _, hasAt := e["at"]; !hasAt {
		t.Error(`entry["at"] must be present`)
	}
}

// __step_error is read by 33 Go sites and 6 live agent configs (measured
// 2026-08-24). The whole design rests on leaving it alone.
func TestRecordStepErrorDoesNotTouchTheSingularKey(t *testing.T) {
	original := map[string]interface{}{"failed_step": "load_page", "message": "boom"}
	collected := map[string]interface{}{"__step_error": original}

	recordStepError(collected, "load_page", "boom", at)

	got, ok := collected["__step_error"].(map[string]interface{})
	if !ok {
		t.Fatal("__step_error must survive unchanged in shape")
	}
	if got["failed_step"] != "load_page" || got["message"] != "boom" {
		t.Errorf("__step_error was mutated: %v", got)
	}
}

// Accumulation is the entire point — the singular key already holds "the last one".
func TestRecordStepErrorAccumulatesDistinctSteps(t *testing.T) {
	collected := map[string]interface{}{}
	recordStepError(collected, "review_editquality", "first", at)
	recordStepError(collected, "review_guardian", "second", at)

	m := collected["__step_errors"].(map[string]interface{})
	if len(m) != 2 {
		t.Fatalf("both failures must be retained, got %d: %v", len(m), keysOf(m))
	}
}

// A retrying step should report its LATEST error, not its first, and must not
// consume a second slot against the cap.
func TestRecordStepErrorUpdatesARepeatInPlace(t *testing.T) {
	collected := map[string]interface{}{}
	recordStepError(collected, "flaky", "attempt one", at)
	recordStepError(collected, "flaky", "attempt two", at)

	m := collected["__step_errors"].(map[string]interface{})
	if len(m) != 1 {
		t.Fatalf("a repeat must not add a key, got %d", len(m))
	}
	if msg := m["flaky"].(map[string]interface{})["message"]; msg != "attempt two" {
		t.Errorf("expected the latest message, got %q", msg)
	}
}

// The guardian seat's objection made concrete: this path is fleet-wide, and a loop
// expanding into many failing iterations makes a DISTINCT step name each time.
func TestRecordStepErrorIsBoundedAndSaysSoWhenItBinds(t *testing.T) {
	collected := map[string]interface{}{}
	for i := 0; i < maxStepErrors+25; i++ {
		recordStepError(collected, fmt.Sprintf("loop_step_%d", i), "boom", at)
	}
	m := collected["__step_errors"].(map[string]interface{})

	// cap + the marker, and nothing more.
	if len(m) != maxStepErrors+1 {
		t.Fatalf("expected %d entries (cap + marker), got %d", maxStepErrors+1, len(m))
	}
	if _, marked := m[stepErrorTruncatedKey]; !marked {
		t.Fatal("a capped record MUST say so — without the marker a seat whose failure fell off the end reads as an abstention, which is the exact conflation this record prevents")
	}
	// The marker must not be mistakable for a step, or the reader would report a
	// step named __truncated as having failed.
	if _, looksLikeAStep := m["loop_step_0"].(map[string]interface{})["message"]; !looksLikeAStep {
		t.Error("early entries must be retained")
	}

	// Re-failing a step ALREADY present must still work past the cap — otherwise a
	// retrying step silently stops reporting once an unrelated loop filled the map.
	recordStepError(collected, "loop_step_0", "later error", at)
	if msg := m["loop_step_0"].(map[string]interface{})["message"]; msg != "later error" {
		t.Errorf("a step already present must keep updating past the cap, got %q", msg)
	}
}

// Defensive: routeToErrorStep is on the failure path of every workflow, so this
// must never be the thing that panics.
func TestRecordStepErrorIsInertOnJunk(t *testing.T) {
	recordStepError(nil, "step", "msg", at) // must not panic

	collected := map[string]interface{}{}
	recordStepError(collected, "", "msg", at)
	if _, present := collected["__step_errors"]; present {
		t.Error("an empty step name must record nothing rather than an unkeyed entry")
	}

	// A pre-existing value of the wrong type must be replaced, not panicked on.
	collected2 := map[string]interface{}{"__step_errors": "not-a-map"}
	recordStepError(collected2, "step", "msg", at)
	if m, ok := collected2["__step_errors"].(map[string]interface{}); !ok || len(m) != 1 {
		t.Error("a malformed pre-existing value must be replaced with a usable map")
	}
}

func keysOf(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
