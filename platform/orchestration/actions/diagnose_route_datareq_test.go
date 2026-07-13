package actions

import (
	"strings"
	"testing"
)

// F0.5 — data_request answers must persist across iterations. Benchmark run
// 5120c0dc: iteration 2's bundle carried the requested rows, the tier guard
// (rightly) refused the verdict that followed, and iteration 3's bundle had
// lost them — the loop re-requested near-identical SQL and tripped
// scope-not-narrowing. withPriorRequests re-forwards the engine's accumulated
// SeenRequests keys so load_runtime re-runs them every iteration.
func TestWithPriorRequests(t *testing.T) {
	cur := func(sqls ...string) []interface{} {
		var out []interface{}
		for _, s := range sqls {
			out = append(out, map[string]interface{}{"sql": s, "why": "fresh"})
		}
		return out
	}

	t.Run("the run-3 hole: empty current verdict still re-forwards prior requests", func(t *testing.T) {
		seen := map[string]bool{
			"SELECT build_status FROM pages WHERE name='guides-index'": true,
			"SELECT 1 FROM site_work_items":                            true,
		}
		got := withPriorRequests(nil, seen, 12)
		if len(got) != 2 {
			t.Fatalf("want both prior requests forwarded, got %d: %v", len(got), got)
		}
		// deterministic order (sorted), and the why marks them as re-runs
		first := got[0].(map[string]interface{})
		if !strings.Contains(first["why"].(string), "persists across iterations") {
			t.Fatalf("re-forwarded request should say why: %v", first)
		}
	})

	t.Run("current requests keep their why and dedupe against seen", func(t *testing.T) {
		seen := map[string]bool{"SELECT a FROM b": true}
		got := withPriorRequests(cur("SELECT a FROM b"), seen, 12)
		if len(got) != 1 {
			t.Fatalf("identical current+seen must not duplicate, got %d: %v", len(got), got)
		}
		if got[0].(map[string]interface{})["why"] != "fresh" {
			t.Fatalf("current request's why must be preserved: %v", got[0])
		}
	})

	t.Run("cap honoured, current first", func(t *testing.T) {
		seen := map[string]bool{}
		for _, s := range []string{"SELECT 1", "SELECT 2", "SELECT 3"} {
			seen[s] = true
		}
		got := withPriorRequests(cur("SELECT 9"), seen, 2)
		if len(got) != 2 {
			t.Fatalf("cap 2 not honoured: %v", got)
		}
		if got[0].(map[string]interface{})["sql"] != "SELECT 9" {
			t.Fatalf("current verdict's request must come first: %v", got)
		}
	})

	t.Run("a non-read-only key in state is skipped, not forwarded", func(t *testing.T) {
		// SeenRequests round-trips through collected_data; treat keys as data.
		seen := map[string]bool{"UPDATE pages SET build_status='deployed'": true}
		if got := withPriorRequests(nil, seen, 12); len(got) != 0 {
			t.Fatalf("write SQL in state must never be re-forwarded: %v", got)
		}
	})

	t.Run("empty everything stays empty", func(t *testing.T) {
		if got := withPriorRequests(nil, nil, 12); len(got) != 0 {
			t.Fatalf("want empty, got %v", got)
		}
	})
}
