// FILE: platform/orchestration/datahelpers/unified_extractor_search_test.go
//
// RFC_029 §9 D1/D2 (owner-delegated ruling, 2026-08-15) — the whole-tree search
// (findFieldRecursive) is collect-all / unique-or-nothing:
//
//   - unique value  → resolve, deterministically, shallowest path first;
//   - conflict      → resolve NOTHING, and emit the WARN
//     "aggressive search: conflicting candidates".
//
// PHASE 2 LANDED 2026-08-21 (RFC_029 §10.13 step 5) and these tests changed
// deliberately: the conflict tests below flipped from "resolves the stable
// winner" to "returns nil", and their WARN assertions stayed — including
// `winner_path`, which still names the candidate the ranking WOULD have picked.
// They now also assert `phase == "2-refuse"`, so a build that silently reverted
// to Phase 1 fails here rather than passing quietly. That assertion is the whole
// point: a test that cannot tell the phases apart cannot prove the flip held.
package datahelpers

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const conflictWarnMsg = "aggressive search: conflicting candidates"

// warnFields reads the single expected conflict WARN's structured fields.
// zap.Strings is an ArrayMarshaler, which ContextMap does not flatten, so the
// fields have to be routed through a MapObjectEncoder to be readable.
func warnFields(t *testing.T, logs *observer.ObservedLogs) map[string]interface{} {
	t.Helper()
	warns := logs.FilterMessage(conflictWarnMsg).All()
	if len(warns) != 1 {
		t.Fatalf("conflict WARN fired %d times, want exactly 1", len(warns))
	}
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range warns[0].Context {
		f.AddTo(enc)
	}
	if enc.Fields["phase"] != "2-refuse" {
		t.Errorf("WARN phase = %v, want \"2-refuse\" — a build that reverted to Phase 1 "+
			"would still warn, so the phase field is what proves the flip is live",
			enc.Fields["phase"])
	}
	return enc.Fields
}

// observedSearch runs findFieldRecursive under an observer core so tests can
// assert on the Phase 1 WARN instrument as well as the resolved value.
func observedSearch(t *testing.T, data map[string]interface{}, field string) (interface{}, *observer.ObservedLogs) {
	t.Helper()
	core, logs := observer.New(zapcore.WarnLevel)
	value := findFieldRecursive(data, field, 0, zap.New(core))
	return value, logs
}

// Unique-or-nothing's happy half: several paths carry the SAME value — that is
// unique, it resolves, and no conflict WARN fires. This is the "behaviour
// unchanged from today's happy path" clause of the ruling.
func TestWholeTreeSearchUniqueValueResolvesWithoutWarn(t *testing.T) {
	data := map[string]interface{}{
		"stored":  map[string]interface{}{"asset_id": "a-1"},
		"claimed": map[string]interface{}{"asset_id": "a-1"},
		"deep":    map[string]interface{}{"nested": map[string]interface{}{"asset_id": "a-1"}},
	}

	value, logs := observedSearch(t, data, "asset_id")
	if value != "a-1" {
		t.Fatalf("asset_id = %v, want \"a-1\": agreeing candidates are unique and must resolve", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 0 {
		t.Errorf("conflict WARN fired %d times for AGREEING candidates — the instrument must "+
			"only record genuine conflicts or the observation window drowns", n)
	}
}

// PHASE 2: conflicting candidates resolve to NOTHING, and the WARN still names
// the field, every candidate path, and the candidate the ranking WOULD have
// picked. Before 2026-08-21 this test asserted value == "shallow-a".
func TestWholeTreeSearchConflictWarnsAndRefusesPhase2(t *testing.T) {
	data := map[string]interface{}{
		"zebra": map[string]interface{}{"purpose": "shallow-z"},
		"alpha": map[string]interface{}{"purpose": "shallow-a"},
		"mid":   map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
	}

	value, logs := observedSearch(t, data, "purpose")

	if value != nil {
		t.Fatalf("purpose = %v, want nil: Phase 2 refuses a conflict outright. Any picking "+
			"rule here is a guess, and no field is better than a wrong field", value)
	}

	f := warnFields(t, logs)
	if f["field"] != "purpose" {
		t.Errorf("WARN field = %v, want \"purpose\"", f["field"])
	}
	// Still reported, and still the shallowest-first winner ("alpha" < "zebra" at
	// equal depth) — nothing resolves from it now, but it is the first thing a
	// reader tracing an absent field needs.
	if f["winner_path"] != "alpha.purpose" {
		t.Errorf("WARN winner_path = %v, want \"alpha.purpose\" — the ranking still decides "+
			"what gets REPORTED even though it no longer decides what is returned", f["winner_path"])
	}
	paths, _ := f["candidate_paths"].([]interface{})
	if len(paths) != 3 {
		t.Errorf("WARN candidate_paths = %v, want all 3 named — a caller whose field has "+
			"stopped arriving needs every path that competed, to choose its explicit mapping", paths)
	}
}

// Shallowest beats DFS order. Under Phase 2 this fixture CONFLICTS, so the
// resolved value is nil — but the ranking is still under test, via the WARN:
// a lexicographically-earlier key holding a DEEPER candidate must still lose to a
// later key holding a shallower one. Depth is the primary sort; encounter order
// only breaks ties. (bugs_closed/306: this preference is declared, not accidental.)
func TestWholeTreeSearchShallowestBeatsLexicographicallyEarlierDeep(t *testing.T) {
	data := map[string]interface{}{
		"aaa": map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
		"zzz": map[string]interface{}{"purpose": "shallow"},
	}

	value, logs := observedSearch(t, data, "purpose")
	if value != nil {
		t.Fatalf("purpose = %v, want nil: these candidates conflict", value)
	}
	if f := warnFields(t, logs); f["winner_path"] != "zzz.purpose" {
		t.Fatalf("WARN winner_path = %v, want \"zzz.purpose\": depth outranks DFS encounter order",
			f["winner_path"])
	}
}

// The determinism check the old walk could never pass. Under Phase 2 every run
// returns nil, so nil-equality would pass vacuously — the guarantee is now
// asserted where it still lives: the REPORTED winner must be the same every run.
// Under the pre-RFC_029 walk this fixture resolved differently between runs of the
// same binary on the same data (measured 344/400 vs 56/400, LANDMINES 2026-08-08).
func TestWholeTreeSearchIsDeterministicAcrossRuns(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{"asset_id": "one"},
		"b": map[string]interface{}{"asset_id": "two"},
		"c": map[string]interface{}{"asset_id": "three"},
		"d": map[string]interface{}{"asset_id": "four"},
	}

	value, logs := observedSearch(t, data, "asset_id")
	if value != nil {
		t.Fatalf("asset_id = %v, want nil: four differing candidates conflict", value)
	}
	first := warnFields(t, logs)["winner_path"]
	if first != "a.asset_id" {
		t.Fatalf("winner_path = %v, want \"a.asset_id\" (sorted-key DFS at equal depth)", first)
	}
	for i := 0; i < 200; i++ {
		_, runLogs := observedSearch(t, data, "asset_id")
		if got := warnFields(t, runLogs)["winner_path"]; got != first {
			t.Fatalf("run %d reported %v where run 0 reported %v — the search is still "+
				"iteration-order dependent", i, got, first)
		}
	}
}

// No candidates at all still means nil — collect-all changes what happens when
// there are many matches, not when there are none.
func TestWholeTreeSearchNoCandidatesReturnsNil(t *testing.T) {
	data := map[string]interface{}{
		"agent_config": map[string]interface{}{"asset_id": "hidden-by-skip-list"},
		"other":        map[string]interface{}{"unrelated": true},
	}

	value, logs := observedSearch(t, data, "asset_id")
	if value != nil {
		t.Fatalf("asset_id = %v, want nil: the infrastructure skip-list must still hide "+
			"agent_config from the search", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 0 {
		t.Errorf("conflict WARN fired %d times with zero candidates", n)
	}
}
