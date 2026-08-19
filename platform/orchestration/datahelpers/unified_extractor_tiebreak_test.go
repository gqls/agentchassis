package datahelpers

// bugs_open/306 — the whole-tree search's equal-depth tie-break is DECLARED
// (direct match ≺ ~unwrap hop ≺ sibling recursion) and pinned, and the last
// unsorted map iteration on the search path (tryUnwrapMapPatterns pattern 1)
// is sorted.
//
// Why these exist: measured 2026-08-18 on page-build-handler, 13 of 139 runs
// carried a GENUINELY DIFFERENT page at equal depth — `~unwrap.current_page`
// (the dispatched page) vs a stale page under a retry_payload sibling — and the
// unwrap-hop candidate won every time, by append order alone. No test pinned
// that, so a reordered collector would have flipped the winner silently.
//
// Mutation proof (run before committing): invert the rank comparison in
// findFieldRecursive's sort (`>` for `<`) → TestTieBreakUnwrapHopBeatsSibling
// FAILS. Remove the sort.Strings in tryUnwrapMapPatterns →
// TestUnwrapPattern1IsDeterministic FAILS under -count=50 (it flips on the
// unsorted range).

import (
	"testing"

	"go.uber.org/zap"
)

// The production shape, reduced: the page the run was dispatched for lives at
// input_data.current_page (reached via the ~unwrap hop, depth 1) and a STALE
// page lives under a retry-payload sibling that happens to sort EARLIER
// ("call_content_writer" < "input_data"), also at depth 1 once unwrapped — here
// made a direct child so the two candidates tie on depth exactly.
func tieBreakFixture() map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"current_page": map[string]interface{}{"name": "disclaimer"},
		},
		"a_sibling_sorted_earlier": map[string]interface{}{
			"current_page": map[string]interface{}{"name": "contact-index"},
		},
	}
}

func TestTieBreakUnwrapHopBeatsSibling(t *testing.T) {
	value, logs := observedSearch(t, tieBreakFixture(), "current_page")
	page, _ := value.(map[string]interface{})
	if page["name"] != "disclaimer" {
		t.Fatalf("current_page = %v, want the ~unwrap (input_data) candidate \"disclaimer\" — "+
			"the declared tie-break (direct ≺ unwrap ≺ sibling) must beat sorted-key order at equal depth", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 1 {
		t.Fatalf("conflict WARN fired %d times, want 1 — the fixture IS a conflict; the tie-break resolves it, it does not hide it", n)
	}
}

func TestTieBreakDirectMatchBeatsUnwrapHop(t *testing.T) {
	// A direct key at the root beats the same key reached through the hop, even
	// though the hop's value is "closer to the step's own inputs" — rank order is
	// direct first, and this pins that the declaration matches the old append.
	data := map[string]interface{}{
		"current_page": map[string]interface{}{"name": "root-direct"},
		"input_data": map[string]interface{}{
			"current_page": map[string]interface{}{"name": "via-unwrap"},
		},
	}
	value, _ := observedSearch(t, data, "current_page")
	page, _ := value.(map[string]interface{})
	if page["name"] != "root-direct" {
		t.Fatalf("current_page = %v, want the root direct match (depth 0 beats depth 1 regardless of rank)", value)
	}
}

func TestTieBreakDepthStillOutranksRank(t *testing.T) {
	// Rank only breaks DEPTH ties: a shallower sibling beats a deeper unwrap-hop
	// candidate. Pins that declaring the rank did not promote it above depth.
	data := map[string]interface{}{
		"input_data": map[string]interface{}{
			"nested": map[string]interface{}{
				"current_page": map[string]interface{}{"name": "deep-via-unwrap"},
			},
		},
		"zz_sibling": map[string]interface{}{
			"current_page": map[string]interface{}{"name": "shallow-sibling"},
		},
	}
	value, _ := observedSearch(t, data, "current_page")
	page, _ := value.(map[string]interface{})
	if page["name"] != "shallow-sibling" {
		t.Fatalf("current_page = %v, want the shallower sibling — depth is primary, rank is the tie-break only", value)
	}
}

func TestTieBreakRankIsInheritedBelowTheFirstHop(t *testing.T) {
	// Two candidates at depth 2: one under the unwrap hop, one under a sibling
	// whose key sorts earlier. The hop's subtree must still win — the rank is
	// set by the FIRST hop from the root and inherited, not re-derived per level.
	data := map[string]interface{}{
		"input_data": map[string]interface{}{
			"wrapper": map[string]interface{}{
				"current_page": map[string]interface{}{"name": "under-hop"},
			},
		},
		"a_earlier": map[string]interface{}{
			"wrapper": map[string]interface{}{
				"current_page": map[string]interface{}{"name": "under-sibling"},
			},
		},
	}
	value, _ := observedSearch(t, data, "current_page")
	page, _ := value.(map[string]interface{})
	if page["name"] != "under-hop" {
		t.Fatalf("current_page = %v, want the unwrap-hop subtree's candidate at equal depth 2", value)
	}
}

func TestUnwrapPattern1IsDeterministic(t *testing.T) {
	// Two *_result keys, each carrying a `result` child with a DIFFERENT value
	// for the searched field. Before the sort, pattern 1 returned whichever the
	// map iterator met first; `go test -count=50` flipped it. Now it is the
	// lexicographically first key, every run.
	data := map[string]interface{}{
		"zeta_result":  map[string]interface{}{"result": map[string]interface{}{"asset_id": "from-zeta"}},
		"alpha_result": map[string]interface{}{"result": map[string]interface{}{"asset_id": "from-alpha"}},
	}
	first := tryUnwrapMapPatterns(data, zap.NewNop())
	for i := 0; i < 200; i++ {
		got := tryUnwrapMapPatterns(data, zap.NewNop())
		gm, _ := got.(map[string]interface{})
		fm, _ := first.(map[string]interface{})
		if gm["asset_id"] != fm["asset_id"] {
			t.Fatalf("run %d unwrapped %v where run 0 unwrapped %v — pattern 1 is still iteration-order dependent", i, gm["asset_id"], fm["asset_id"])
		}
	}
	fm, _ := first.(map[string]interface{})
	if fm["asset_id"] != "from-alpha" {
		t.Fatalf("pattern 1 picked %v, want the sorted-first key alpha_result", fm["asset_id"])
	}
}

func TestRankDeclarationChangesNoExistingWinner(t *testing.T) {
	// The three fixtures from unified_extractor_search_test.go, re-asserted here
	// so the rank declaration is provably a no-op on every previously pinned
	// winner (all of them are sibling-vs-sibling ties, rank equal, sorted-key
	// order still decides).
	cases := []struct {
		name  string
		data  map[string]interface{}
		field string
		want  interface{}
	}{
		{"conflict-shallow-a", map[string]interface{}{
			"zebra": map[string]interface{}{"purpose": "shallow-z"},
			"alpha": map[string]interface{}{"purpose": "shallow-a"},
			"mid":   map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
		}, "purpose", "shallow-a"},
		{"shallowest-beats-earlier-deep", map[string]interface{}{
			"aaa": map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
			"zzz": map[string]interface{}{"purpose": "shallow"},
		}, "purpose", "shallow"},
		{"four-way-sorted", map[string]interface{}{
			"a": map[string]interface{}{"asset_id": "one"},
			"b": map[string]interface{}{"asset_id": "two"},
			"c": map[string]interface{}{"asset_id": "three"},
			"d": map[string]interface{}{"asset_id": "four"},
		}, "asset_id", "one"},
	}
	for _, c := range cases {
		if got := findFieldRecursive(c.data, c.field, 0, zap.NewNop()); got != c.want {
			t.Errorf("%s: got %v, want %v — the rank declaration must not move any previously pinned winner", c.name, got, c.want)
		}
	}
}
