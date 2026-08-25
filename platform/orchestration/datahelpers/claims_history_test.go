package datahelpers

// FILE: platform/orchestration/datahelpers/claims_history_test.go
//
// Tests for fact-value history (bugs_open/386): a page that rendered a number
// while the register held it is not inventing that number when the register
// moves on.
//
// EVERY fixture here is MINIMAL — one fact, and where a negative is asserted,
// nothing else in the register that could support the value for an unrelated
// reason. That is deliberate and it is the whole reason these tests are worth
// anything: the bugs_open/364 lane's first draft used a populated register and
// PASSED with its fix reverted, because some other fact was quietly supporting
// the number. So each positive below is paired with the same fixture mutated at
// exactly one point, asserting the opposite outcome. A test that only ever sees
// the armed case cannot tell "the mechanism fired" from "the value was supported
// anyway".

import (
	"encoding/json"
	"testing"
)

// armedFact is one number-bearing fact with retained history, scoped by a
// context term. Callers mutate the single field under test.
func armedFact() EvidenceFact {
	v := 11828.0
	return EvidenceFact{
		ID:            "F9-feed-items-collected",
		Claim:         "feed items collected",
		Value:         &v,
		Kind:          FactKindMetric,
		VerifiedAt:    "2026-08-25",
		Tolerance:     "exact",
		ContextTerms:  []string{"feed items collected"},
		RetainHistory: true,
		History: []FactHistoryEntry{
			{Value: 11513, VerifiedAt: "2026-08-23"},
			{Value: 11646, VerifiedAt: "2026-08-24"},
		},
	}
}

const windowMatching = "…feed items collected 11513 verified 2026-08-23…"

func TestHistorySupportsAFormerReading(t *testing.T) {
	eb := &EvidenceBase{Facts: []EvidenceFact{armedFact()}}

	// The current value still passes — history must not have displaced it.
	if !eb.numberSupported(11828, windowMatching) {
		t.Fatal("current value 11828 must remain supported")
	}
	// The reading the page actually rendered.
	if !eb.numberSupported(11513, windowMatching) {
		t.Error("11513 was the register's value on 2026-08-23 and must be supported")
	}
	if !eb.numberSupported(11646, windowMatching) {
		t.Error("11646 was the register's value on 2026-08-24 and must be supported")
	}
}

// TestHistoryIsInertWhenNotArmed is the mutation proof for the opt-in gate: the
// SAME fixture with one bool flipped must reach the opposite verdict. Without
// this, TestHistorySupportsAFormerReading passes on a build where RetainHistory
// is ignored entirely.
func TestHistoryIsInertWhenNotArmed(t *testing.T) {
	f := armedFact()
	f.RetainHistory = false
	eb := &EvidenceBase{Facts: []EvidenceFact{f}}

	if eb.numberSupported(11513, windowMatching) {
		t.Error("an unarmed fact must not support a former reading — the unsafe side is off by default")
	}
	if !eb.numberSupported(11828, windowMatching) {
		t.Error("the current value must still be supported when history is off")
	}
}

// TestHistoryDoesNotBecomeBlanketSupport is the other direction: history accepts
// exactly the values the register held, never a range. A gte tolerance would
// support 11512; exact history must not.
func TestHistoryDoesNotBecomeBlanketSupport(t *testing.T) {
	eb := &EvidenceBase{Facts: []EvidenceFact{armedFact()}}

	for _, val := range []float64{11512, 11514, 11600, 11647, 1, 0} {
		if eb.numberSupported(val, windowMatching) {
			t.Errorf("%v was never a register value and must not be supported", val)
		}
	}
}

// TestHistoryStaysInsideTheContextTermGate — a former reading is only evidence
// for the claim the fact scopes. In a window about something else it must not
// support the number, or one armed counter starts vouching for unrelated figures
// across the page (the accidental-support failure mode of bugs_open/364 §2).
func TestHistoryStaysInsideTheContextTermGate(t *testing.T) {
	eb := &EvidenceBase{Facts: []EvidenceFact{armedFact()}}

	const unrelated = "…we have advised 11513 clients since founding…"
	if eb.numberSupported(11513, unrelated) {
		t.Error("a former reading must not support a number in a window the fact does not scope")
	}
	// Control: the same value in the matching window IS supported, so the
	// negative above is the context gate and not history being broken.
	if !eb.numberSupported(11513, windowMatching) {
		t.Fatal("control failed: 11513 must be supported in the matching window")
	}
}

// TestHistoryDoesNotMakeAFactASeries guards the field-collision this design
// exists to avoid: IsSeries() keys on len(Observations), so had history reused
// that slot every armed fact would silently become a series and take the
// series branch in numberSupported instead.
func TestHistoryDoesNotMakeAFactASeries(t *testing.T) {
	f := armedFact()
	if f.IsSeries() {
		t.Error("a fact with history and no observations must not report as a series")
	}
	if len(f.Observations) != 0 {
		t.Error("history must not populate Observations")
	}
}

// TestHistoryCapIsEnforcedAtTheReader — the cap bounds what the scan ACCEPTS, so
// it cannot live only in the writer that trims. A hand seed or a migration
// backfill can produce an over-cap array, and an armed fact must still accept no
// more than the cap allows.
func TestHistoryCapIsEnforcedAtTheReader(t *testing.T) {
	f := armedFact()
	f.History = nil
	// Oldest first, newest last, one more than the cap.
	for i := 0; i <= FactHistoryMaxEntries; i++ {
		f.History = append(f.History, FactHistoryEntry{Value: float64(1000 + i)})
	}
	eb := &EvidenceBase{Facts: []EvidenceFact{f}}

	oldest := 1000.0
	newest := float64(1000 + FactHistoryMaxEntries)
	if eb.numberSupported(oldest, windowMatching) {
		t.Errorf("the oldest entry (%v) is past the cap and must not be accepted", oldest)
	}
	if !eb.numberSupported(newest, windowMatching) {
		t.Errorf("the newest entry (%v) must be accepted", newest)
	}
	// And the entry exactly at the cap boundary survives, so the trim is
	// off-by-none rather than silently dropping a valid reading.
	if !eb.numberSupported(1001, windowMatching) {
		t.Error("the entry at the cap boundary must be accepted")
	}
}

// TestHistoryRoundTripsThroughJSON — the register is stored as jsonb and read
// back by ParseEvidenceBase, so the wire form is part of the contract. omitempty
// on both fields matters: an unarmed fact must serialise with neither key, or
// every one of the ~295 facts in the fleet grows two.
func TestHistoryRoundTripsThroughJSON(t *testing.T) {
	raw, err := json.Marshal(armedFact())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back EvidenceFact
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !back.RetainHistory || len(back.History) != 2 {
		t.Fatalf("history did not survive the round trip: %+v", back)
	}
	if back.History[0].Value != 11513 || back.History[0].VerifiedAt != "2026-08-23" {
		t.Errorf("first entry corrupted: %+v", back.History[0])
	}

	bare, err := json.Marshal(EvidenceFact{ID: "x", Kind: FactKindMetric})
	if err != nil {
		t.Fatalf("marshal bare: %v", err)
	}
	for _, key := range []string{"retain_history", "history"} {
		var m map[string]interface{}
		if err := json.Unmarshal(bare, &m); err != nil {
			t.Fatalf("unmarshal bare: %v", err)
		}
		if _, present := m[key]; present {
			t.Errorf("an unarmed fact must not serialise %q", key)
		}
	}
}
