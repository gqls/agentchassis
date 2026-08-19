// FILE: platform/orchestration/actions/loop_dedup_warn_test.go
//
// bugs_open/321, the runtime net. The offline detector (config-key-audit
// --loop-sitewide-item-keys) proves at definition level that no loop-nested
// create_work_item is missing its item_key_suffix_field — but it cannot see a
// suffix that resolves to the SAME value every iteration. That shape still
// deduplicates iterations 2..N away silently, and the only place it is
// observable is the action's own !inserted result inside a loop iteration.
// shouldWarnLoopDedup is that gate; these tests pin its whole truth table,
// because a Warn that fires on every dedup fleet-wide is noise nobody reads
// and a Warn that never fires is the silence this bug is made of.
package actions

import "testing"

func TestShouldWarnLoopDedup(t *testing.T) {
	inLoop := map[string]interface{}{
		"loop_iteration":  float64(2),
		"loop_var_name":   "current_suggestion",
		"item_key_prefix": "add_tool_novel",
	}
	inLoopWithSuffix := map[string]interface{}{
		"loop_iteration":        float64(0),
		"loop_var_name":         "current_finding",
		"item_key_suffix_field": "tool_data.page_id",
	}
	topLevel := map[string]interface{}{
		"item_key_prefix": "site_plan",
	}

	cases := []struct {
		name         string
		config       map[string]interface{}
		inserted     bool
		wantFire     bool
		wantLoopVar  string
		wantSuffixed bool
	}{
		// The case the Warn exists for: a dedup inside a loop iteration.
		{"dedup in loop fires", inLoop, false, true, "current_suggestion", false},
		// The detector's blind spot: suffix configured, dedup happened anyway
		// (loop-invariant suffix, or a legitimate cross-run dedup). Fires, and
		// says the suffix was there — that flag is what tells the reader this
		// is NOT the missing-suffix shape.
		{"dedup in loop with suffix fires, flagged", inLoopWithSuffix, false, true, "current_finding", true},
		// An inserted item is not a dedup — silence, even inside a loop.
		{"insert in loop stays quiet", inLoop, true, false, "", false},
		// A top-level dedup is the INTENDED site-wide behaviour — silence.
		// This arm is what keeps the Warn from firing fleet-wide on every
		// once-per-site step doing its job.
		{"top-level dedup stays quiet", topLevel, false, false, "", false},
		{"top-level insert stays quiet", topLevel, true, false, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fire, loopVar, suffixed := shouldWarnLoopDedup(tc.config, tc.inserted)
			if fire != tc.wantFire {
				t.Fatalf("fire = %v, want %v", fire, tc.wantFire)
			}
			if loopVar != tc.wantLoopVar {
				t.Errorf("loopVar = %q, want %q", loopVar, tc.wantLoopVar)
			}
			if suffixed != tc.wantSuffixed {
				t.Errorf("suffixConfigured = %v, want %v", suffixed, tc.wantSuffixed)
			}
		})
	}
}
