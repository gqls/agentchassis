package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// These tests pin the COMPONENT-SCOPED RATCHET (bugs_open/361).
//
// The bug they exist for: the ratchet was a key-level set difference against a
// baseline that recorded only FINDINGS, so every component born after the baseline
// was cut manufactured "NEW" findings and the CronJob was red for 25 consecutive
// days (lastSuccessfulTime 2026-08-09 → 2026-09-03), with 478 NEW on the last run.
//
// ⚠ 361 §6 is explicit that BOTH arms must be proved: the fix must still fail on a
// regression in a component the baseline covered, and must not fail on growth in one
// it never saw. A fix verified only on the second arm is a fix that turned the check
// off, which is why TestRatchet_Arm1 exists and why it is named an arm.

func mkFinding(comp, field, shape string) finding {
	return finding{Component: comp, Field: field, Shape: shape, Count: 1}
}

func writeTempBaseline(t *testing.T, bf baselineFile) string {
	t.Helper()
	b, err := json.Marshal(bf)
	if err != nil {
		t.Fatalf("marshal baseline: %v", err)
	}
	p := filepath.Join(t.TempDir(), "baseline.json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	return p
}

// resetCanonical clears the package-level clone map so one test cannot leak
// identity decisions into the next.
func resetCanonical(t *testing.T) {
	t.Helper()
	prev := canonicalComponent
	canonicalComponent = map[string]string{}
	t.Cleanup(func() { canonicalComponent = prev })
}

// ARM 1 — a component the baseline COVERED gains a key. This MUST fail the run.
// Mutation proof: drop the `covered[...]` branch from classifyAgainstBaseline (so
// everything unknown becomes unbaselined) and this test fails.
func TestRatchet_Arm1_RegressionInCoveredComponentIsAFinding(t *testing.T) {
	resetCanonical(t)
	base := map[string]bool{"alpha\x00a1\x00empty_heading": true}
	covered := map[string]bool{"alpha": true}

	regressions, unbaselined, _ := classifyAgainstBaseline(
		[]finding{
			mkFinding("alpha", "a1", "empty_heading"), // already known — silent
			mkFinding("alpha", "a2", "empty_block"),   // NEW in a covered component
		}, base, covered)

	if len(regressions) != 1 {
		t.Fatalf("a covered component that gained a key must be a REGRESSION: got %d, want 1 (%+v)",
			len(regressions), regressions)
	}
	if regressions[0].Field != "a2" {
		t.Fatalf("wrong finding flagged: %+v", regressions[0])
	}
	if len(unbaselined) != 0 {
		t.Fatalf("a covered component's finding must never be filed as unbaselined: %+v", unbaselined)
	}
}

// ARM 2 — a component the baseline NEVER analysed produces a key. This must NOT
// fail the run: it is growth, not regression. Mutation proof: make covered[] always
// true and this test fails.
func TestRatchet_Arm2_UnbaselinedComponentDoesNotFail(t *testing.T) {
	resetCanonical(t)
	base := map[string]bool{"alpha\x00a1\x00empty_heading": true}
	covered := map[string]bool{"alpha": true}

	regressions, unbaselined, comps := classifyAgainstBaseline(
		[]finding{
			mkFinding("alpha", "a1", "empty_heading"), // known
			mkFinding("beta", "b1", "empty_block"),    // born after the baseline
			mkFinding("beta", "b2", "empty_cell"),
		}, base, covered)

	if len(regressions) != 0 {
		t.Fatalf("growth in an unbaselined component must NOT fail the run: %+v", regressions)
	}
	if len(unbaselined) != 2 {
		t.Fatalf("unbaselined findings must still be REPORTED: got %d, want 2", len(unbaselined))
	}
	if len(comps) != 1 || !comps["beta"] {
		t.Fatalf("unbaselined component set wrong: %+v", comps)
	}
}

// THE HOLE IN 361's OWN FIX CANDIDATE 1, pinned.
//
// 361 §4 candidate 1 proposes scoping by "does this component own zero KEYS in the
// baseline?". That is not sufficient, and the shipped artefact proves it: baseline.json's
// note says "1023 findings across 139 analysed components" while its keys span only 115
// components — so 24 components were analysed and CLEAN. Under candidate 1 those 24 read
// as unbaselined, and a regression in one would NOT fail. Recording COVERAGE separately
// from findings is what makes that state unrepresentable, and this test is the pin.
func TestRatchet_AnalysedButCleanComponentStillRatchets(t *testing.T) {
	resetCanonical(t)
	base := map[string]bool{"alpha\x00a1\x00empty_heading": true}
	// gamma was analysed at baseline time and had NO findings, so it owns no key.
	covered := map[string]bool{"alpha": true, "gamma": true}

	regressions, unbaselined, _ := classifyAgainstBaseline(
		[]finding{mkFinding("gamma", "g1", "empty_block")}, base, covered)

	if len(regressions) != 1 {
		t.Fatalf("a component that was analysed and CLEAN at baseline time and now renders a hole "+
			"is a REGRESSION — this is the case a keys-derived covered set cannot see: got %d, want 1",
			len(regressions))
	}
	if len(unbaselined) != 0 {
		t.Fatalf("must not be filed as growth: %+v", unbaselined)
	}
}

// A clone and its representative must agree about coverage, or the clone reads as
// unbaselined for ever and the ratchet quietly stops watching it.
func TestRatchet_CloneInheritsItsRepresentativesCoverage(t *testing.T) {
	resetCanonical(t)
	canonicalComponent = map[string]string{"alpha-clone": "alpha"}
	base := map[string]bool{"alpha\x00a1\x00empty_heading": true}
	covered := map[string]bool{"alpha": true}

	regressions, unbaselined, _ := classifyAgainstBaseline(
		[]finding{mkFinding("alpha-clone", "a2", "empty_block")}, base, covered)

	if len(regressions) != 1 {
		t.Fatalf("a clone of a covered component is covered: got %d regressions, want 1", len(regressions))
	}
	if len(unbaselined) != 0 {
		t.Fatalf("clone must not read as unbaselined: %+v", unbaselined)
	}
}

// A legacy baseline (no "components" list) must still WORK — refusing would turn a
// mis-scoped ratchet into a job that cannot run — but it must announce the blind spot
// it is running with, because a fallback nobody is told about becomes folklore.
func TestLoadBaseline_LegacyDerivesCoverageFromKeysAndFlagsItself(t *testing.T) {
	resetCanonical(t)
	p := writeTempBaseline(t, baselineFile{Keys: []string{"alpha\x00a1\x00empty_heading"}})

	keys, covered, legacy, err := loadBaseline(p)
	if err != nil {
		t.Fatalf("legacy baseline must load, not refuse: %v", err)
	}
	if !legacy {
		t.Fatal("a baseline with no components list MUST report itself legacy — the caller prints the blind spot")
	}
	if !keys["alpha\x00a1\x00empty_heading"] {
		t.Fatal("keys lost")
	}
	if !covered["alpha"] {
		t.Fatal("legacy coverage must be derived from the keys")
	}
	if covered["gamma"] {
		t.Fatal("legacy coverage cannot invent a component the keys never mention")
	}
}

func TestLoadBaseline_ModernReportsItselfNonLegacyAndUsesTheRecordedSet(t *testing.T) {
	resetCanonical(t)
	p := writeTempBaseline(t, baselineFile{
		Keys:       []string{"alpha\x00a1\x00empty_heading"},
		Components: []string{"alpha", "gamma"},
	})

	_, covered, legacy, err := loadBaseline(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if legacy {
		t.Fatal("a baseline carrying a components list is not legacy")
	}
	if !covered["gamma"] {
		t.Fatal("gamma was analysed and clean — the recorded set is what closes 361's blind spot")
	}
}

// The pre-existing guard must survive the change: a baseline that parses to zero keys
// would report every finding as NEW, so it is refused rather than trusted.
func TestLoadBaseline_StillRefusesAZeroKeyBaseline(t *testing.T) {
	resetCanonical(t)
	p := writeTempBaseline(t, baselineFile{Components: []string{"alpha"}})
	if _, _, _, err := loadBaseline(p); err == nil {
		t.Fatal("a zero-key baseline must be refused, not silently trusted")
	}
}

// Round-trip: what writeBaseline records must be what loadBaseline believes, in
// canonical names — including a component that was analysed and produced nothing.
func TestWriteBaseline_RecordsCanonicalCoverageIncludingCleanComponents(t *testing.T) {
	resetCanonical(t)
	canonicalComponent = map[string]string{"alpha-clone": "alpha"}
	p := filepath.Join(t.TempDir(), "b.json")

	analysed := map[string]bool{"alpha": true, "alpha-clone": true, "gamma": true}
	if err := writeBaseline(p, []finding{mkFinding("alpha", "a1", "empty_heading")}, analysed, 2); err != nil {
		t.Fatalf("write: %v", err)
	}

	// ⚠ ASSERT ON THE WRITTEN ARTEFACT, NOT ON THE ROUND-TRIP. loadBaseline
	// re-canonicalises what it reads, so a write-side mutation (recording raw names
	// instead of canonical ones) is masked by the load side — two guards in series,
	// and the round-trip assertion passed under exactly that mutation. Reading the
	// file back as JSON is what isolates the write. See WRONG_CALLS.md 2026-09-03.
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read written baseline: %v", err)
	}
	var onDisk baselineFile
	if err := json.Unmarshal(raw, &onDisk); err != nil {
		t.Fatalf("unmarshal written baseline: %v", err)
	}
	got := map[string]bool{}
	for _, c := range onDisk.Components {
		got[c] = true
	}
	// ⚠ CORRECTED from the first cut of this fix, which asserted the OPPOSITE — that
	// coverage collapses onto canonical names. That pinned a defect: a clone that is
	// later EDITED stops matching its representative's hash, so only its own raw name
	// can vouch for it, and a canonical-only set exempts it silently. See
	// TestRatchet_EditedCloneStaysAccountableUnderItsOwnName.
	if !got["alpha-clone"] {
		t.Fatalf("the FILE must record RAW names — a clone that is later edited can only be "+
			"vouched for under its own name: %v", onDisk.Components)
	}
	if !got["alpha"] || !got["gamma"] {
		t.Fatalf("written coverage lost a component: %v", onDisk.Components)
	}
	if len(onDisk.Components) != 3 {
		t.Fatalf("every component that parsed gets its own entry, clones included: %v", onDisk.Components)
	}

	_, covered, legacy, err := loadBaseline(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if legacy {
		t.Fatal("a freshly written baseline must never read as legacy")
	}
	if !covered["gamma"] {
		t.Fatal("an analysed component with NO findings must still be recorded as covered — that is the whole point")
	}
	if !covered["alpha"] {
		t.Fatal("representative missing from coverage")
	}
}

// An EDITED clone must stay accountable under its OWN name. Once its template
// diverges it is no longer mapped to a representative, so a covered set stored in
// canonical names would file its findings as unbaselined and never fail — silently
// exempting exactly the case the tool's own findingKey note calls "correct" to flag.
// This was a real defect in the first cut of this fix, found in review.
func TestRatchet_EditedCloneStaysAccountableUnderItsOwnName(t *testing.T) {
	resetCanonical(t)
	// It WAS a clone of alpha when the baseline was written, so its raw name is in
	// the covered set. It has since been edited, so nothing maps it any more.
	canonicalComponent = map[string]string{}
	base := map[string]bool{"alpha\x00a1\x00empty_heading": true}
	covered := map[string]bool{"alpha": true, "alpha-clone": true}

	regressions, unbaselined, _ := classifyAgainstBaseline(
		[]finding{mkFinding("alpha-clone", "a9", "empty_block")}, base, covered)

	if len(regressions) != 1 {
		t.Fatalf("an edited clone the baseline covered must still fail: got %d regressions, want 1 "+
			"(unbaselined=%+v)", len(regressions), unbaselined)
	}
}

// A clone BORN AFTER the baseline inherits its representative's coverage, so the
// representative and its copies are judged the same way rather than one failing and
// one passing on the same key.
func TestCovers_CloneBornAfterTheBaselineInheritsItsRepresentative(t *testing.T) {
	resetCanonical(t)
	canonicalComponent = map[string]string{"alpha-copy": "alpha"}
	covered := map[string]bool{"alpha": true}

	if !covers(covered, "alpha-copy") {
		t.Fatal("a clone of a covered component is covered, via its representative")
	}
	if covers(covered, "unrelated") {
		t.Fatal("coverage must not leak to a component with no relation to the baseline")
	}
}

// A STATIC template (no template actions) is vouched for but never probed: it is
// skipped before `checked++`. If the covered set were derived from `checked` it would
// miss every one of them, and a static component later rewritten to reference a field
// and render a hole — this check's own stated signal — would fail nothing. That was
// the second defect in the first cut of this fix.
func TestRatchet_StaticTemplateThatLaterRendersAHoleIsARegression(t *testing.T) {
	resetCanonical(t)
	base := map[string]bool{"alpha\x00a1\x00empty_heading": true}
	// "static" parsed at baseline time and had no fields at all, so it owns no key
	// and was never counted as analysed — but the baseline DID look at it.
	covered := map[string]bool{"alpha": true, "static": true}

	regressions, _, _ := classifyAgainstBaseline(
		[]finding{mkFinding("static", "headline", "empty_heading")}, base, covered)

	if len(regressions) != 1 {
		t.Fatalf("a static template that gained a field and a hole is a REGRESSION: got %d, want 1",
			len(regressions))
	}
}

// A covered set that is present but EMPTY is a ratchet switched off by hand: every
// finding would read as unbaselined and nothing could ever fail. Absent (legacy) and
// empty are different claims and must not share a code path.
func TestLoadBaseline_RefusesAnEmptyCoveredSet(t *testing.T) {
	resetCanonical(t)
	// Written as RAW JSON on purpose: the struct carries `omitempty`, so marshalling
	// an empty slice omits the field and it reads back as ABSENT (legacy), not empty.
	// The tool therefore cannot produce this artefact — only a hand edit can, which is
	// precisely the quiet, unreviewed clearing the embedded baseline exists to prevent.
	p := filepath.Join(t.TempDir(), "hand-edited.json")
	raw := "{\"keys\":[\"alpha\\u0000a1\\u0000empty_heading\"],\"components\":[]}"
	if err := os.WriteFile(p, []byte(raw), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, _, _, err := loadBaseline(p); err == nil {
		t.Fatal("an empty covered set must be refused — it would make the ratchet incapable of failing")
	}
}
