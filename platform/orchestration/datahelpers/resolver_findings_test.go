// FILE: platform/orchestration/datahelpers/resolver_findings_test.go
//
// RFC_029 Phase 1, council-revision (run ae2a88a7, reuse_agent gating
// objection): the two Phase 1 WARNs are ALSO reported to a registered
// recorder so they can be persisted (agent_error_log) and the observation
// window read after the fact. These tests pin the contract at the resolver:
//
//   - with a recorder registered, one conflict occurrence → exactly ONE
//     finding (code, field, candidate paths, winner), one bypass occurrence →
//     exactly ONE finding — no dedup, because frequency is the data;
//   - the agreeing / never-resolvable controls that fire no WARN also fire NO
//     finding — the row population must equal the WARN population;
//   - with NO recorder registered (the default) behaviour is byte-identical to
//     the log-only build: same value, same WARN, no panic (default-OFF control);
//   - a recorder that panics cannot change the resolver's answer.
//
// The fixtures are the SAME ones the WARN tests use (unified_extractor_search_
// test.go, action_inputs_strict_test.go) so a finding is provably one-per-WARN
// and not a second, differently-triggered instrument. Asserted on a fake
// recorder, never a mocked INSERT — the INSERT is agenterrors' contract.
package datahelpers

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// captureResolverFindings installs a fake recorder for the test's lifetime and
// returns the slice it appends to. Package-level state, so the reset is
// load-bearing: a leaked recorder would make every later test in the package
// look "persisted".
func captureResolverFindings(t *testing.T) *[]ResolverFinding {
	t.Helper()
	got := &[]ResolverFinding{}
	SetResolverFindingRecorder(func(f ResolverFinding) { *got = append(*got, f) })
	t.Cleanup(func() { SetResolverFindingRecorder(nil) })
	return got
}

// The conflict fixture from TestWholeTreeSearchConflictWarnsAndResolvesStableWinnerPhase1.
func conflictFixture() map[string]interface{} {
	return map[string]interface{}{
		"zebra": map[string]interface{}{"purpose": "shallow-z"},
		"alpha": map[string]interface{}{"purpose": "shallow-a"},
		"mid":   map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
	}
}

func TestConflictingCandidatesAreRecordedOncePerOccurrence(t *testing.T) {
	got := captureResolverFindings(t)

	value, logs := observedSearch(t, conflictFixture(), "purpose")
	if value != "shallow-a" {
		t.Fatalf("purpose = %v, want \"shallow-a\": the recorder must not change the Phase 1 answer", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 1 {
		t.Fatalf("conflict WARN fired %d times, want 1 — the log line stays alongside the row", n)
	}
	if len(*got) != 1 {
		t.Fatalf("recorded %d findings for ONE conflict, want exactly 1 (no dedup, no fan-out)", len(*got))
	}
	f := (*got)[0]
	if f.Code != ResolverFindingConflictingCandidates {
		t.Errorf("Code = %q, want %q", f.Code, ResolverFindingConflictingCandidates)
	}
	if f.Field != "purpose" || f.Context["field"] != "purpose" {
		t.Errorf("Field/context field = %q/%v, want purpose", f.Field, f.Context["field"])
	}
	if f.Message != conflictWarnMsg {
		t.Errorf("Message = %q, want the WARN text verbatim %q (row and line must be joinable by eye)", f.Message, conflictWarnMsg)
	}
	if f.Context["winner_path"] != "alpha.purpose" {
		t.Errorf("winner_path = %v, want alpha.purpose", f.Context["winner_path"])
	}
	paths, _ := f.Context["candidate_paths"].([]string)
	if len(paths) != 3 {
		t.Errorf("candidate_paths = %v, want all 3 — the window needs every path to map each one before Phase 2", f.Context["candidate_paths"])
	}
	if f.Context["identity_scope"] == nil {
		t.Errorf("identity_scope missing from context — the blank orchestration_id must read as a stated limit, not a defect")
	}

	// A second occurrence is a second row: frequency is the data.
	observedSearch(t, conflictFixture(), "purpose")
	if len(*got) != 2 {
		t.Fatalf("recorded %d findings after TWO conflicts, want 2 — a deduped instrument cannot size the population", len(*got))
	}
}

func TestAgreeingCandidatesRecordNothing(t *testing.T) {
	got := captureResolverFindings(t)
	data := map[string]interface{}{
		"stored":  map[string]interface{}{"asset_id": "a-1"},
		"claimed": map[string]interface{}{"asset_id": "a-1"},
	}
	value, logs := observedSearch(t, data, "asset_id")
	if value != "a-1" {
		t.Fatalf("asset_id = %v, want a-1", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 0 {
		t.Fatalf("conflict WARN fired %d times for agreeing candidates", n)
	}
	if len(*got) != 0 {
		t.Fatalf("recorded %d findings for AGREEING candidates, want 0 — the row population must equal the WARN population or the window drowns", len(*got))
	}
}

func TestSingleSegmentMappingBypassIsRecordedOncePerOccurrence(t *testing.T) {
	got := captureResolverFindings(t)

	// The fixture from TestSingleSegmentMappingBypassWarns, verbatim.
	spec := ActionInputSpec{Optional: []string{"payload"}}
	collected := map[string]interface{}{
		"handler_output": map[string]interface{}{"ok": true},
		"sibling":        map[string]interface{}{"payload": map[string]interface{}{"foreign": true}},
	}
	config := map[string]interface{}{"payload": "handler_output"}

	inputs, err, logs := observedExtract(t, collected, config, spec)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if m := inputs.GetMap("payload"); m == nil || m["foreign"] != true {
		t.Fatalf("payload = %#v: the recorder must not change the Phase 1 answer", inputs.GetRaw("payload"))
	}
	if n := logs.FilterMessage(bypassWarnMsg).Len(); n != 1 {
		t.Fatalf("bypass WARN fired %d times, want 1", n)
	}
	if len(*got) != 1 {
		t.Fatalf("recorded %d findings for ONE bypass, want exactly 1", len(*got))
	}
	f := (*got)[0]
	if f.Code != ResolverFindingMappingBypassed {
		t.Errorf("Code = %q, want %q", f.Code, ResolverFindingMappingBypassed)
	}
	if f.Field != "payload" || f.Context["reference"] != "handler_output" {
		t.Errorf("Field/reference = %q/%v, want payload/handler_output", f.Field, f.Context["reference"])
	}
	if f.Message != bypassWarnMsg {
		t.Errorf("Message = %q, want the WARN text verbatim", f.Message)
	}

	// Control: the mapped key does not exist → no WARN (existing test) and no row.
	*got = (*got)[:0]
	agreeing := map[string]interface{}{
		"nested": map[string]interface{}{"payload": map[string]interface{}{"same": true}},
	}
	if _, err, _ := observedExtract(t, agreeing, map[string]interface{}{"payload": "missing_key"}, spec); err != nil {
		t.Fatalf("control ExtractActionInputs: %v", err)
	}
	if len(*got) != 0 {
		t.Fatalf("recorded %d findings when the mapped key does not exist, want 0", len(*got))
	}
}

// The default-OFF control: with no recorder installed the resolver's answer and
// its WARN are exactly what the log-only build produced. This is what lets a
// binary that never registers (any non-chassis consumer of datahelpers) say
// nothing false.
func TestNoRecorderIsLogOnlyAndUnchanged(t *testing.T) {
	SetResolverFindingRecorder(nil)

	value, logs := observedSearch(t, conflictFixture(), "purpose")
	if value != "shallow-a" {
		t.Fatalf("purpose = %v, want shallow-a with no recorder", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 1 {
		t.Fatalf("conflict WARN fired %d times with no recorder, want 1 — the line is not conditional on the row", n)
	}
}

// A recorder must never take the resolver down with it: a panic inside it is
// recovered, the answer stands, and the loss is logged.
func TestPanickingRecorderCannotChangeTheResolversAnswer(t *testing.T) {
	SetResolverFindingRecorder(func(ResolverFinding) { panic("sink exploded") })
	t.Cleanup(func() { SetResolverFindingRecorder(nil) })

	core, logs := observer.New(zapcore.WarnLevel)
	value := findFieldRecursive(conflictFixture(), "purpose", 0, zap.New(core))
	if value != "shallow-a" {
		t.Fatalf("purpose = %v, want shallow-a: a recorder panic must be recovered", value)
	}
	if n := logs.FilterMessage("resolver finding recorder panicked — the resolver's answer stands, the row is lost").Len(); n != 1 {
		t.Errorf("recorder panic logged %d times, want 1 — a swallowed loss reads as recorded", n)
	}
}
