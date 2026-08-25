package actions

// FILE: platform/orchestration/actions/refresh_evidence_fact_history_test.go
//
// Tests for recordFactHistory (bugs_open/386) — the writer half of fact-value
// history. The reader half is claims_history_test.go in datahelpers.
//
// These assert on the RAW MAP, not a typed struct, because that is what the
// action actually mutates and because the map is the thing that can lose data.
// The register is written back with json.Marshal over map[string]interface{}
// specifically so unknown keys survive; a helper that round-tripped a fact
// through EvidenceFact to add two fields would silently drop every key the
// struct does not know about. TestHistoryPreservesUnknownFactKeys is the guard
// on that, and it is the one to keep if any of these are ever pruned.

import (
	"encoding/json"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func floatPtr(v float64) *float64 { return &v }

// armedFactMap is one armed fact carrying a key this code has never heard of,
// so every test here also checks the no-data-loss property.
func armedFactMap() map[string]interface{} {
	return map[string]interface{}{
		"id":             "F9-feed-items-collected",
		"value":          11646.0,
		"verified_at":    "2026-08-24",
		"retain_history": true,
		"writer_line":    "{value}+ feed items collected",
		"a_future_key":   map[string]interface{}{"nested": "must survive"},
	}
}

func historyOf(t *testing.T, fact map[string]interface{}) []interface{} {
	t.Helper()
	h, ok := fact["history"].([]interface{})
	if !ok {
		return nil
	}
	return h
}

func TestRecordFactHistoryAppendsTheOutgoingReading(t *testing.T) {
	fact := armedFactMap()

	if !recordFactHistory(fact, floatPtr(11646), "2026-08-24") {
		t.Fatal("recordFactHistory must report that it wrote an entry")
	}
	h := historyOf(t, fact)
	if len(h) != 1 {
		t.Fatalf("want 1 history entry, got %d", len(h))
	}
	entry, ok := h[0].(map[string]interface{})
	if !ok {
		t.Fatalf("entry is not an object: %T", h[0])
	}
	if v, _ := numericField(entry["value"]); v != 11646 {
		t.Errorf("want value 11646, got %v", entry["value"])
	}
	if entry["verified_at"] != "2026-08-24" {
		t.Errorf("want verified_at 2026-08-24, got %v", entry["verified_at"])
	}
}

// TestRecordFactHistoryIgnoresUnarmedFacts is the mutation pair for the opt-in
// gate: same call, one key removed, opposite outcome. It also asserts the KEY is
// absent rather than an empty array — an unarmed fact must be byte-identical
// after this call, so arming nothing changes nothing in the stored register.
func TestRecordFactHistoryIgnoresUnarmedFacts(t *testing.T) {
	fact := armedFactMap()
	delete(fact, "retain_history")

	if recordFactHistory(fact, floatPtr(11646), "2026-08-24") {
		t.Error("an unarmed fact must not have history recorded")
	}
	if _, present := fact["history"]; present {
		t.Error("an unarmed fact must not gain a history key at all")
	}

	// And explicitly false, not merely absent.
	fact["retain_history"] = false
	if recordFactHistory(fact, floatPtr(11646), "2026-08-24") {
		t.Error("retain_history:false must not have history recorded")
	}
}

func TestRecordFactHistoryNeedsAnOutgoingValue(t *testing.T) {
	fact := armedFactMap()

	if recordFactHistory(fact, nil, "2026-08-24") {
		t.Error("a fact with no stored value has no former reading to remember")
	}
	if _, present := fact["history"]; present {
		t.Error("a nil outgoing value must not write a zero entry")
	}
}

// TestRecordFactHistoryDoesNotDuplicateTheNewestReading — two refreshes in one
// day are real (measured: 2026-08-16 carries two), and each appending would fill
// the cap in half the calendar time it is documented to cover.
func TestRecordFactHistoryDoesNotDuplicateTheNewestReading(t *testing.T) {
	fact := armedFactMap()

	if !recordFactHistory(fact, floatPtr(11646), "2026-08-24") {
		t.Fatal("first record must write")
	}
	if recordFactHistory(fact, floatPtr(11646), "2026-08-24") {
		t.Error("recording the same reading again must be a no-op")
	}
	if got := len(historyOf(t, fact)); got != 1 {
		t.Errorf("want 1 entry after a duplicate, got %d", got)
	}

	// A DIFFERENT reading on the same day is a real second reading and appends.
	if !recordFactHistory(fact, floatPtr(11700), "2026-08-24") {
		t.Error("a different value on the same day is a new reading and must append")
	}
	if got := len(historyOf(t, fact)); got != 2 {
		t.Errorf("want 2 entries, got %d", got)
	}
}

// TestRecordFactHistoryTrimsFromTheFront — the cap must drop the OLDEST, since
// the newest readings are the ones a still-stale page is likely to be carrying.
func TestRecordFactHistoryTrimsFromTheFront(t *testing.T) {
	fact := armedFactMap()
	for i := 0; i < datahelpers.FactHistoryMaxEntries+5; i++ {
		if !recordFactHistory(fact, floatPtr(float64(1000+i)), "2026-08-24") {
			t.Fatalf("record %d must write", i)
		}
	}
	h := historyOf(t, fact)
	if len(h) != datahelpers.FactHistoryMaxEntries {
		t.Fatalf("want history capped at %d, got %d", datahelpers.FactHistoryMaxEntries, len(h))
	}
	first, _ := h[0].(map[string]interface{})
	last, _ := h[len(h)-1].(map[string]interface{})
	if v, _ := numericField(first["value"]); v != 1005 {
		t.Errorf("oldest surviving entry should be 1005 (the first 5 trimmed), got %v", first["value"])
	}
	if v, _ := numericField(last["value"]); v != float64(1000+datahelpers.FactHistoryMaxEntries+4) {
		t.Errorf("newest entry wrong: %v", last["value"])
	}
}

// TestHistoryPreservesUnknownFactKeys is the no-data-loss guard. The register
// holds human-authored keys this code does not model, and the whole action works
// on raw maps to keep them. If this ever fails, someone has started
// round-tripping a fact through a typed struct.
func TestHistoryPreservesUnknownFactKeys(t *testing.T) {
	fact := armedFactMap()

	if !recordFactHistory(fact, floatPtr(11646), "2026-08-24") {
		t.Fatal("record must write")
	}
	if fact["writer_line"] != "{value}+ feed items collected" {
		t.Error("writer_line was altered")
	}
	nested, ok := fact["a_future_key"].(map[string]interface{})
	if !ok || nested["nested"] != "must survive" {
		t.Errorf("an unmodelled key was lost or mangled: %v", fact["a_future_key"])
	}
	if fact["id"] != "F9-feed-items-collected" {
		t.Error("id was altered")
	}
}

// TestRecordFactHistoryWrittenEntriesAreReadableByTheScanner closes the loop:
// the writer's raw-map output must parse into the typed fact the scan consumes.
// Asserting the two halves separately would let a key-name mismatch pass both.
func TestRecordFactHistoryWrittenEntriesAreReadableByTheScanner(t *testing.T) {
	fact := armedFactMap()
	if !recordFactHistory(fact, floatPtr(11513), "2026-08-23") {
		t.Fatal("record must write")
	}

	eb := map[string]interface{}{"facts": []interface{}{fact}}
	raw, err := json.Marshal(eb)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := datahelpers.ParseEvidenceBase(raw)
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	if len(parsed.Facts) != 1 {
		t.Fatalf("want 1 fact, got %d", len(parsed.Facts))
	}
	f := parsed.Facts[0]
	if !f.RetainHistory {
		t.Error("retain_history did not survive into the typed fact")
	}
	if len(f.History) != 1 || f.History[0].Value != 11513 {
		t.Fatalf("history did not survive into the typed fact: %+v", f.History)
	}
	if f.History[0].VerifiedAt != "2026-08-23" {
		t.Errorf("verified_at did not survive: %q", f.History[0].VerifiedAt)
	}
}
