package datahelpers

// The Strategy-0 prune (RFC_029 §10.10 as corrected by §10.11): a field the
// explicit config path already resolved is removed from what Strategy 1/2 hand
// to ExtractFields, so the whole-tree search never runs for it and the conflict
// instrument stays quiet. These tests are mutation-proof against the prune
// itself: remove `withoutResolved` from either ExtractFields call and the first
// test fails on a recorded finding, because the fixture's decoys genuinely
// conflict.
//
// Scope pins, stated as tests because the diff's whole argument is "exactly
// Strategy 0 and nothing more":
//   - an UNRESOLVED declared field still searches and still records;
//   - a DEFAULTED-only field still searches (its answer is still discarded at
//     merge — pruning it too would be behaviour-equivalent but is deliberately
//     out of scope);
//   - strict `!` behaviour is untouched, including alongside a pruned field;
//   - ensureCoreFields' unconditional injections still merge in when the pruned
//     list is EMPTY — the calls must not be skipped on an empty list.

import (
	"testing"

	"go.uber.org/zap"
)

// pruneConflictFixture: "target" resolves explicitly at stored.target, AND two
// sibling decoys carry CONFLICTING values under the same key — so if the
// whole-tree search runs for "target" at all, it must record a conflict.
func pruneConflictFixture() map[string]interface{} {
	return map[string]interface{}{
		"stored": map[string]interface{}{"target": "explicit-answer"},
		"decoyA": map[string]interface{}{"target": "wrong-a"},
		"decoyB": map[string]interface{}{"target": "wrong-b"},
	}
}

func TestStrategy0ResolvedFieldIsNotSearchedAndRecordsNoConflict(t *testing.T) {
	got := captureResolverFindings(t)

	spec := ActionInputSpec{Optional: []string{"target"}}
	config := map[string]interface{}{"target": "stored.target"}

	inputs, err, logs := observedExtract(t, pruneConflictFixture(), config, spec)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if v := inputs.Get("target"); v != "explicit-answer" {
		t.Fatalf("target = %q, want the Strategy 0 value \"explicit-answer\"", v)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 0 {
		t.Fatalf("conflict WARN fired %d times for a Strategy-0-resolved field — the search ran when the prune should have removed the field", n)
	}
	if len(*got) != 0 {
		t.Fatalf("recorded %d findings, want 0: the decoys conflict, so any finding means the whole-tree search ran for a field Strategy 0 already resolved", len(*got))
	}
}

func TestUnresolvedFieldStillSearchesAndStillRecords(t *testing.T) {
	got := captureResolverFindings(t)

	// No config mapping at all: the search is the only resolver, and the decoys
	// conflict — the instrument must still see it. This is the prune's
	// overreach guard: the surviving rows are the population Phase 2 is about.
	spec := ActionInputSpec{Optional: []string{"target"}}
	data := map[string]interface{}{
		"decoyA": map[string]interface{}{"target": "wrong-a"},
		"decoyB": map[string]interface{}{"target": "wrong-b"},
	}

	inputs, err, logs := observedExtract(t, data, map[string]interface{}{}, spec)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if v := inputs.Get("target"); v != "wrong-a" {
		t.Fatalf("target = %q, want \"wrong-a\" (stable shallowest-first winner unchanged)", v)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 1 {
		t.Fatalf("conflict WARN fired %d times, want 1 — an unresolved field must keep searching and recording", n)
	}
	if len(*got) != 1 {
		t.Fatalf("recorded %d findings, want 1", len(*got))
	}
}

func TestDefaultedOnlyFieldIsStillSearched(t *testing.T) {
	got := captureResolverFindings(t)

	// "mode" carries a spec Default and no config mapping. Its search answer is
	// discarded at merge (a default IS a value), but the search itself still
	// runs — the prune is scoped to Strategy-0-resolved fields ONLY, and this
	// test is what holds that line.
	spec := ActionInputSpec{
		Optional: []string{"mode"},
		Defaults: map[string]interface{}{"mode": "default-mode"},
	}
	data := map[string]interface{}{
		"decoyA": map[string]interface{}{"mode": "wrong-a"},
		"decoyB": map[string]interface{}{"mode": "wrong-b"},
	}

	inputs, err, logs := observedExtract(t, data, map[string]interface{}{}, spec)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if v := inputs.Get("mode"); v != "default-mode" {
		t.Fatalf("mode = %q, want the default (merge discards the search answer, unchanged)", v)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 1 {
		t.Fatalf("conflict WARN fired %d times, want 1 — a defaulted-only field is NOT pruned", n)
	}
	if len(*got) != 1 {
		t.Fatalf("recorded %d findings, want 1", len(*got))
	}
}

func TestPruneComposesWithStrictAndLeavesTheSearchedFieldAlone(t *testing.T) {
	got := captureResolverFindings(t)

	// Three declared fields, three fates in one call:
	//   asset_id!  — strict, resolves via its reference (never meets the search)
	//   target     — Strategy 0 resolves it (pruned from the search)
	//   loose      — unmapped, genuinely searched, single agreeing source
	spec := ActionInputSpec{Optional: []string{"asset_id", "target", "loose"}}
	data := map[string]interface{}{
		"stored": map[string]interface{}{
			"asset_id": "a-real",
			"target":   "explicit-answer",
		},
		"decoyA": map[string]interface{}{"target": "wrong-a"},
		"decoyB": map[string]interface{}{"target": "wrong-b"},
		"other":  map[string]interface{}{"loose": "found-by-search"},
	}
	config := map[string]interface{}{
		"asset_id!": "stored.asset_id",
		"target":    "stored.target",
	}

	inputs, err := ExtractActionInputs(data, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if v := inputs.Get("asset_id"); v != "a-real" {
		t.Fatalf("asset_id = %q, want a-real (strict unchanged)", v)
	}
	if v := inputs.Get("target"); v != "explicit-answer" {
		t.Fatalf("target = %q, want explicit-answer", v)
	}
	if v := inputs.Get("loose"); v != "found-by-search" {
		t.Fatalf("loose = %q, want found-by-search — the prune must not remove fields it did not resolve", v)
	}
	if len(*got) != 0 {
		t.Fatalf("recorded %d findings, want 0 (target pruned; loose has one agreeing source)", len(*got))
	}
}

func TestStrictFailureIsUnchangedByThePrune(t *testing.T) {
	// A strict field whose reference does not resolve must still hard-fail,
	// even when every OTHER declared field was pruned (empty search list).
	spec := ActionInputSpec{Optional: []string{"asset_id", "target"}}
	data := map[string]interface{}{
		"stored": map[string]interface{}{"target": "explicit-answer"},
	}
	config := map[string]interface{}{
		"asset_id!": "nowhere.asset_id",
		"target":    "stored.target",
	}

	_, err := ExtractActionInputs(data, config, spec, zap.NewNop())
	if err == nil {
		t.Fatalf("want the strict hard error, got nil — the prune must not soften '!'")
	}
}

func TestCoreFieldInjectionsSurviveAnEmptyPrunedList(t *testing.T) {
	// Every declared field is Strategy-0-resolved, so the pruned list handed to
	// ExtractFields is EMPTY — and the call must still happen: ensureCoreFields'
	// unconditional injections (here: domain) merge into Values today, and
	// skipping the call on an empty list would silently remove them.
	spec := ActionInputSpec{Optional: []string{"target"}}
	data := map[string]interface{}{
		"stored":      map[string]interface{}{"target": "explicit-answer"},
		"site_record": map[string]interface{}{"domain": "example.co.uk"},
	}
	config := map[string]interface{}{"target": "stored.target"}

	inputs, err := ExtractActionInputs(data, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if v := inputs.Get("target"); v != "explicit-answer" {
		t.Fatalf("target = %q, want explicit-answer", v)
	}
	if v := inputs.Get("domain"); v != "example.co.uk" {
		t.Fatalf("domain = %q, want the ensureCoreFields injection to survive an empty pruned list", v)
	}
}
