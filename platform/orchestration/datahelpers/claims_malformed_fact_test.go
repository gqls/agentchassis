// FILE: platform/orchestration/datahelpers/claims_malformed_fact_test.go
//
// One undecodable fact must cost that fact and nothing else.
//
// THE LIVE CASE THIS WAS WRITTEN FROM (bugs_closed/161's residual, measured
// 2026-09-03 over all 27 live registers): two sites' registers did not parse
// at all, so every claims gate skipped them and 10 banned_claims were inert —
// finetuning.uk (3 bans, since 2026-08-24) and noted.co.uk (7 bans, since
// 2026-08-25). Both failed on the same shape: a fact whose `value` is text
// ("MIT", "Apache 2.0", "30 days"), which EvidenceFact.Value (*float64)
// cannot hold. The old all-or-nothing decode turned one unsupported field
// into total, silent disarmament of the site's guard list.
//
// The fixtures below are the real shapes, reduced.
package datahelpers

import (
	"encoding/json"
	"strings"
	"testing"
)

// notedShapedRegister is noted.co.uk's live shape: one text-valued fact whose
// `source` is also a bare string (a second, independent defect the old decode
// never even reached), alongside the security bans that were inert because of
// it.
const notedShapedRegister = `{
  "audit_doc": "docs/.../noted",
  "facts": [
    {"claim": "Deleted text persists in encrypted backups for at most 30 days",
     "value": "30 days",
     "source": "B2 offsite backup object lock = 30 days",
     "registered": "2026-08-25"}
  ],
  "banned_claims": [
    {"pattern": "end[- ]?to[- ]?end encrypt|zero[- ]?knowledge", "reason": "Not built."},
    {"pattern": "gdpr[- ]compliant|iso ?27001|soc ?2", "reason": "No auditor, no date."}
  ]
}`

// TestMalformedFactDoesNotVoidTheBannedClaims is the whole bug in one test.
//
// MUTATION THAT MUST GO RED: restore the all-or-nothing decode in
// ParseEvidenceBase —
//
//	var eb EvidenceBase; json.Unmarshal(data, &eb)
//
// — and this returns (nil, error), so every assertion below fails and the two
// bans are gone.
func TestMalformedFactDoesNotVoidTheBannedClaims(t *testing.T) {
	eb, err := ParseEvidenceBase([]byte(notedShapedRegister))
	if err != nil {
		t.Fatalf("a register with one bad fact must still parse: %v", err)
	}
	if eb == nil {
		t.Fatal("register parsed to nil — the site would read as 'not opted in'")
	}

	// The point of the change: the guard list survives its neighbour.
	if len(eb.BannedClaims) != 2 {
		t.Fatalf("banned_claims must survive a malformed fact: got %d, want 2", len(eb.BannedClaims))
	}
	// And they are COMPILED, not merely present — an uncompiled pattern
	// matches nothing and would pass this test's count while failing live.
	for i, bc := range eb.BannedClaims {
		if bc.re == nil {
			t.Fatalf("banned_claims[%d] was not compiled — it would match nothing", i)
		}
	}
	if !eb.BannedClaims[0].re.MatchString("we are end-to-end encrypted") {
		t.Fatal("the compiled ban does not match the sentence it exists to refuse")
	}

	// The bad fact is dropped, not silently kept as a zero value.
	if len(eb.Facts) != 0 {
		t.Fatalf("the undecodable fact must not appear as a fact: got %d", len(eb.Facts))
	}
	// ...and it is REPORTED, which is what makes it fixable.
	if len(eb.MalformedFacts) != 1 {
		t.Fatalf("the undecodable fact must be reported: got %d entries, want 1", len(eb.MalformedFacts))
	}
	mf := eb.MalformedFacts[0]
	if mf.Index != 0 {
		t.Fatalf("MalformedFact.Index = %d, want 0 — a human needs the row", mf.Index)
	}
	if !strings.Contains(mf.Err, "value") {
		t.Fatalf("the error must name the offending field, got %q", mf.Err)
	}
}

// TestMalformedFactIsNamedWhenItHasAnID: the id is read separately from the
// failed decode, so one bad field does not also cost us the fact's name.
//
// MUTATION: have rawFactID decode into a struct with `ID string` instead of a
// map — it then fails on the SAME type mismatch and every name comes back "".
func TestMalformedFactIsNamedWhenItHasAnID(t *testing.T) {
	const reg = `{"facts":[
	  {"id":"ft-licence-mistral7b","claim":"licence","value":"Apache 2.0","source":{"attested_by":"owner"}},
	  {"id":"ft-good","claim":"models","value":3,"kind":"count","source":{"attested_by":"owner"},"verified_at":"2026-08-24"}
	],"banned_claims":[]}`
	eb, err := ParseEvidenceBase([]byte(reg))
	if err != nil || eb == nil {
		t.Fatalf("parse: %v (eb nil: %v)", err, eb == nil)
	}
	if len(eb.MalformedFacts) != 1 || eb.MalformedFacts[0].ID != "ft-licence-mistral7b" {
		t.Fatalf("the malformed fact must be named: %+v", eb.MalformedFacts)
	}
	// The GOOD fact beside it is untouched — a bad neighbour costs it nothing.
	if len(eb.Facts) != 1 || eb.Facts[0].ID != "ft-good" {
		t.Fatalf("the well-formed fact must survive: %+v", eb.Facts)
	}
}

// TestMalformedOnlyRegisterStillParsesNonNil guards the clause that keeps the
// signal alive on a register with no bans to save it: nil means "not opted
// in", and a site whose every fact failed to decode HAS opted in and got
// nothing.
//
// MUTATION: drop `&& len(eb.MalformedFacts) == 0` from the empty-base test —
// this returns nil and the only evidence of the defect is discarded.
func TestMalformedOnlyRegisterStillParsesNonNil(t *testing.T) {
	const reg = `{"facts":[{"id":"only","claim":"x","value":"text"}],"banned_claims":[]}`
	eb, err := ParseEvidenceBase([]byte(reg))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if eb == nil {
		t.Fatal("a register whose only fact is malformed must not read as 'no register'")
	}
	if len(eb.MalformedFacts) != 1 {
		t.Fatalf("want 1 malformed fact, got %d", len(eb.MalformedFacts))
	}
	// It is still not SCANNABLE — no facts, no bans — and that must not change.
	if eb.HasScannableRegister() {
		t.Fatal("a base with no facts and no bans must not be scannable (bugs_open/364)")
	}
}

// TestTopLevelGarbageStillErrors: tolerance is per-FACT only. A register whose
// JSON is broken above the facts array is still an error, exactly as before.
func TestTopLevelGarbageStillErrors(t *testing.T) {
	for _, bad := range []string{`{"facts": [}`, `not json at all`, `{"banned_claims": 7}`} {
		if _, err := ParseEvidenceBase([]byte(bad)); err == nil {
			t.Fatalf("input %q must still be an error", bad)
		}
	}
}

// TestWellFormedRegisterIsUnaffected is the regression control: the 25 of 27
// live registers that already parsed must be unchanged by this, including the
// empty-base nil contract.
func TestWellFormedRegisterIsUnaffected(t *testing.T) {
	const reg = `{"audit_doc":"d","facts":[
	  {"id":"a","claim":"tools live","value":11,"kind":"count","source":{"sql":"SELECT 1"},"verified_at":"2026-07-24"},
	  {"id":"b","claim":"guides live","value":10,"kind":"count","source":{"attested_by":"owner"},"verified_at":"2026-07-24"}
	],"banned_claims":[{"pattern":"unhackable","reason":"absolute"}]}`
	eb, err := ParseEvidenceBase([]byte(reg))
	if err != nil || eb == nil {
		t.Fatalf("parse: %v", err)
	}
	if len(eb.Facts) != 2 || len(eb.BannedClaims) != 1 || len(eb.MalformedFacts) != 0 {
		t.Fatalf("a clean register must be unchanged: %d facts, %d bans, %d malformed",
			len(eb.Facts), len(eb.BannedClaims), len(eb.MalformedFacts))
	}
	if eb.Facts[0].Value == nil || *eb.Facts[0].Value != 11 {
		t.Fatal("numeric values must still decode")
	}
	// The nil contract for a genuinely empty base is untouched.
	empty, err := ParseEvidenceBase([]byte(`{"facts":[],"banned_claims":[]}`))
	if err != nil || empty != nil {
		t.Fatalf("an empty base must still parse to (nil, nil), got %v %v", empty, err)
	}
}

// TestRawFactIDNeverGuesses: an unreadable fact yields "", never a wrong name.
func TestRawFactIDNeverGuesses(t *testing.T) {
	if got := rawFactID([]byte(`{"id": 7}`)); got != "" {
		t.Fatalf("a non-string id must yield \"\", got %q", got)
	}
	if got := rawFactID([]byte(`{{{`)); got != "" {
		t.Fatalf("unparseable raw must yield \"\", got %q", got)
	}
	if got := rawFactID([]byte(`{"id":"ok"}`)); got != "ok" {
		t.Fatalf("a readable id must be returned, got %q", got)
	}
}

// TestMalformedFactsDoNotRoundTripIntoAConsumersRegister: MalformedFacts is
// produced by the parse, not read from the register, and must never be
// written back into one.
func TestMalformedFactsDoNotRoundTripIntoAConsumersRegister(t *testing.T) {
	eb, _ := ParseEvidenceBase([]byte(notedShapedRegister))
	out, err := json.Marshal(eb)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "MalformedFacts") || strings.Contains(string(out), "malformed") {
		t.Fatalf("MalformedFacts must not serialise into a register: %s", out)
	}
}
