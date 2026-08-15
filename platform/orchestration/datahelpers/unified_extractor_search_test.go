// FILE: platform/orchestration/datahelpers/unified_extractor_search_test.go
//
// RFC_029 §9 D1/D2 (owner-delegated ruling, 2026-08-15) — the whole-tree search
// (findFieldRecursive) is collect-all / unique-or-nothing:
//
//   - unique value  → resolve, deterministically, shallowest path first;
//   - conflict      → PHASE 1 (this build): resolve to the STABLE shallowest
//     winner AND emit the WARN "aggressive search: conflicting candidates";
//     PHASE 2 (a later build, after the observation window): resolve NOTHING.
//
// WHEN PHASE 2 LANDS these tests change deliberately: the conflict tests below
// flip from "resolves the stable winner" to "returns nil", and their WARN
// assertions stay. Do not weaken them to pass both phases at once — a test that
// cannot tell the phases apart cannot prove the flip happened.
package datahelpers

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const conflictWarnMsg = "aggressive search: conflicting candidates"

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

// PHASE 1: conflicting candidates still resolve — to the stable shallowest
// winner — and the WARN names the field, every candidate path, and the winner.
// PHASE 2 flips the resolution half of this test to nil; the WARN half stays.
func TestWholeTreeSearchConflictWarnsAndResolvesStableWinnerPhase1(t *testing.T) {
	data := map[string]interface{}{
		"zebra": map[string]interface{}{"purpose": "shallow-z"},
		"alpha": map[string]interface{}{"purpose": "shallow-a"},
		"mid":   map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
	}

	value, logs := observedSearch(t, data, "purpose")

	// Depth ties break by the collector's sorted-key DFS order: "alpha" < "zebra".
	if value != "shallow-a" {
		t.Fatalf("purpose = %v, want \"shallow-a\" (Phase 1: conflicts resolve to the stable "+
			"shallowest-first winner; sorted-key order breaks the depth tie)", value)
	}

	warns := logs.FilterMessage(conflictWarnMsg).All()
	if len(warns) != 1 {
		t.Fatalf("conflict WARN fired %d times, want exactly 1", len(warns))
	}
	// zap.Strings is an ArrayMarshaler, which ContextMap does not flatten —
	// route the fields through a MapObjectEncoder to read the actual values.
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range warns[0].Context {
		f.AddTo(enc)
	}
	if enc.Fields["field"] != "purpose" {
		t.Errorf("WARN field = %v, want \"purpose\"", enc.Fields["field"])
	}
	if enc.Fields["winner_path"] != "alpha.purpose" {
		t.Errorf("WARN winner_path = %v, want \"alpha.purpose\"", enc.Fields["winner_path"])
	}
	paths, _ := enc.Fields["candidate_paths"].([]interface{})
	if len(paths) != 3 {
		t.Errorf("WARN candidate_paths = %v, want all 3 candidates named — the observation "+
			"window needs every path to hand each one an explicit mapping before Phase 2", paths)
	}
}

// Shallowest beats DFS order: a lexicographically-earlier key holding a DEEPER
// candidate must lose to a later key holding a shallower one. Depth is the
// primary sort; encounter order only breaks ties.
func TestWholeTreeSearchShallowestBeatsLexicographicallyEarlierDeep(t *testing.T) {
	data := map[string]interface{}{
		"aaa": map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
		"zzz": map[string]interface{}{"purpose": "shallow"},
	}

	value, _ := observedSearch(t, data, "purpose")
	if value != "shallow" {
		t.Fatalf("purpose = %v, want \"shallow\": depth outranks DFS encounter order", value)
	}
}

// The determinism check the old walk could never pass: the same conflicting
// fixture, many runs, one winner. Under the pre-RFC_029 walk this fixture
// resolved differently between runs of the same binary on the same data
// (measured 344/400 vs 56/400 on the real helper, LANDMINES 2026-08-08).
func TestWholeTreeSearchIsDeterministicAcrossRuns(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{"asset_id": "one"},
		"b": map[string]interface{}{"asset_id": "two"},
		"c": map[string]interface{}{"asset_id": "three"},
		"d": map[string]interface{}{"asset_id": "four"},
	}

	first, _ := observedSearch(t, data, "asset_id")
	for i := 0; i < 200; i++ {
		got := findFieldRecursive(data, "asset_id", 0, zap.NewNop())
		if got != first {
			t.Fatalf("run %d resolved %v where run 0 resolved %v — the search is still "+
				"iteration-order dependent", i, got, first)
		}
	}
	if first != "one" {
		t.Fatalf("winner = %v, want \"one\" (sorted-key DFS at equal depth)", first)
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
