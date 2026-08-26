// Tests for --ungraded-completions (bugs_open/393 part B).
//
// MUTATION PROOFS the acceptance demands ("mutation-proved both ways"), each
// stated with the line that breaks it:
//   - markAcknowledged stamping `Acknowledged: true` unconditionally →
//     TestUngradedTypeIsUnacknowledgedByDefault fails (the novel arm).
//   - markAcknowledged stamping `false` unconditionally (ack lookup dropped) →
//     the same test's acked arm fails, and TestHollowUngradedAckDoesNotCount's
//     CONTROL fails.
//   - loadUngradedCompletionsAcks accepting an empty reason →
//     TestHollowUngradedAckDoesNotCount fails.
//   - the aliveness refusal removed → TestUngradedSummaryStatesScope still pins
//     the total into the body, and the emit-path refusal is pinned by
//     TestUngradedStdinShapeDemandsTheAlivenessTotal (ErrorLogRows required).
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ungradedFixture() []ungradedGroup {
	return []ungradedGroup{
		{ItemType: "dark_section_audit", Rows: 11,
			FirstSeen: "2026-08-14 14:24:46", LastSeen: "2026-08-17 12:44:02"},
		{ItemType: "brand_new_drifter", Rows: 2,
			FirstSeen: "2026-08-26 06:00:00", LastSeen: "2026-08-26 06:05:00"},
	}
}

// TestUngradedTypeIsUnacknowledgedByDefault is the canonical both-directions
// shape (optionalexplicitwires_test.go): one fixture, run twice.
func TestUngradedTypeIsUnacknowledgedByDefault(t *testing.T) {
	// Novel: no acks at all — both types are NEW and the run must fail.
	novel := markAcknowledged(ungradedFixture(), map[string]bool{})
	if n := unackedUngradedCount(novel); n != 2 {
		t.Fatalf("no acks: unacked = %d, want 2 — a novel drifting type must be a finding", n)
	}
	// Acked: the diagnosed type is quiet, the new one still fails.
	acked := markAcknowledged(ungradedFixture(), map[string]bool{"dark_section_audit": true})
	if n := unackedUngradedCount(acked); n != 1 {
		t.Fatalf("with dark_section_audit acked: unacked = %d, want 1", n)
	}
	for _, g := range acked {
		if g.ItemType == "dark_section_audit" && !g.Acknowledged {
			t.Error("the acknowledged type is not stamped Acknowledged")
		}
		if g.ItemType == "brand_new_drifter" && g.Acknowledged {
			t.Error("an unacked type is stamped Acknowledged — the ratchet is broken open")
		}
	}
}

// TestHollowUngradedAckDoesNotCount: an entry with an empty reason is ignored —
// with the CONTROL that a real ack DOES count, so the assertion can distinguish
// hollow from real (the sibling tests' pinned pattern).
func TestHollowUngradedAckDoesNotCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "acks.json")
	if err := os.WriteFile(path, []byte(`{
		"_doc": "test",
		"hollow_type": {"reason": "   ", "date": "2026-08-26"},
		"real_type":   {"reason": "diagnosed: see evidence", "date": "2026-08-26"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	acked, err := loadUngradedCompletionsAcks(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if acked["hollow_type"] {
		t.Error("an empty-reason ack counted — a hollow ack is not an ack")
	}
	if !acked["real_type"] {
		t.Fatal("CONTROL FAILED: a real ack did not count, so the assertion above " +
			"cannot distinguish hollow from real")
	}
	if acked["_doc"] {
		t.Error("a documentation key counted as an ack")
	}
}

// TestMissingUngradedAcksFileIsAnError: "could not read the exceptions" and
// "there are no exceptions" have opposite meanings.
func TestMissingUngradedAcksFileIsAnError(t *testing.T) {
	if _, err := loadUngradedCompletionsAcks(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Fatal("a missing acks file loaded as an empty ack set — the run would grade " +
			"every diagnosed type as NEW, or worse, a typo'd path would never be noticed")
	}
}

// TestTheShippedUngradedAcksFileLoads: the file the image will COPY must parse
// and must acknowledge dark_section_audit with a real reason — the one
// population bugs_open/393 diagnosed.
func TestTheShippedUngradedAcksFileLoads(t *testing.T) {
	acked, err := loadUngradedCompletionsAcks(
		"../../docs/agent_docs/docs024_key_docs_latest/architecture_review/no_change_unreadable_acks.json")
	if err != nil {
		t.Fatalf("the shipped acks file does not load: %v", err)
	}
	if !acked["dark_section_audit"] {
		t.Error("dark_section_audit is not acknowledged in the shipped file — the first " +
			"scheduled run would page on the population 393 already diagnosed and closed")
	}
}

// TestUngradedSummaryStatesScope: the doc_notes body must carry the SCOPE (the
// aliveness total and the acks path), not just the verdict — "0 drifting types
// over 41,000 error rows" and "0 over 0" have opposite meanings.
func TestUngradedSummaryStatesScope(t *testing.T) {
	groups := markAcknowledged(ungradedFixture(), map[string]bool{"dark_section_audit": true})
	body := ungradedCompletionsRunSummary(41234, "/app/no_change_unreadable_acks.json", groups)
	for _, want := range []string{
		"41234",                            // the aliveness total
		"no_change_unreadable_acks.json",   // where the exceptions live
		"NO_CHANGE_GATE_UNREADABLE_RESULT", // the code, so the row is self-describing
		"[ack] dark_section_audit",         // the diagnosed population, visible not hidden
		"[NEW] brand_new_drifter",          // the finding
		"complete_work_item_no_change.go",  // where the remedy starts
	} {
		if !strings.Contains(body, want) {
			t.Errorf("summary is missing %q\n---\n%s", want, body)
		}
	}
	// And a clean run must still state its scope.
	clean := ungradedCompletionsRunSummary(41234, "acks.json", nil)
	if !strings.Contains(clean, "41234") || !strings.Contains(clean, "0 item_type(s)") {
		t.Errorf("a clean run's summary does not state its scope:\n%s", clean)
	}
}

// TestUngradedStdinShapeDemandsTheAlivenessTotal pins the stdin contract: a
// payload without error_log_rows must be refusable, which the emit path does by
// checking the pointer. Here we pin the decode-side property the refusal keys
// on, so a "helpful" future change to a plain int (zero-defaulting) fails.
func TestUngradedStdinShapeDemandsTheAlivenessTotal(t *testing.T) {
	var in ungradedCompletionsStdin
	if err := json.Unmarshal([]byte(`{"groups": []}`), &in); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if in.ErrorLogRows != nil {
		t.Fatal("error_log_rows decoded non-nil from a payload that omits it — the emit " +
			"path can no longer tell a missing total from a zero one, and a blind read " +
			"would print a clean report")
	}
}
