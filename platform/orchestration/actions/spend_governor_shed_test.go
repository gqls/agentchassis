package actions

// D4 stage B (register AGOV-013): contract tests for the spend-governor shed
// predicate and its opt-in wiring. Each assertion below is aimed at a SPECIFIC
// mutation that would ship a wrong governor while every other test stays green
// — the posture rules are load-bearing (see workItemNotGovernorShedSQL's doc
// comment) and none of them is observable from a passing happy-path query.

import (
	"strings"
	"testing"
)

// The renderer's three posture rules, each killing a named mutation:
//
//	MUTATION "fail-closed": drop the COALESCE(..., false) wrapper -> an absent
//	governor row sheds EVERYTHING (NOT NULL = NULL filters every row).
//	MUTATION "unknown-is-free": drop COALESCE(m.class,'maintenance') or
//	COALESCE(m.llm_bearing,true) -> an unmapped item_type never sheds, and the
//	map silently becomes an allow-list (the declaring-a-key-silences-your-own-
//	detector class).
//	MUTATION "order-flip": swap the CASE thresholds -> research sheds before
//	maintenance, inverting the owner's 2026-08-31 ruling.
func TestGovernorShedSQLPostureRules(t *testing.T) {
	sql := workItemNotGovernorShedSQL("wi")

	if !strings.HasPrefix(strings.TrimSpace(sql), "NOT COALESCE((") ||
		!strings.Contains(sql, "), false)") {
		t.Error("fail-open wrapper missing: the predicate must be NOT COALESCE((...), false) " +
			"so an unreadable governor NEVER sheds (an absent row must not filter every item)")
	}
	if !strings.Contains(sql, "COALESCE(m.class, 'maintenance')") {
		t.Error("unmapped item_type must default to class 'maintenance' (sheds earliest — " +
			"the safe default for an unknown spender)")
	}
	if !strings.Contains(sql, "COALESCE(m.llm_bearing, true)") {
		t.Error("unmapped item_type must default to llm_bearing=true — otherwise the class " +
			"map is an allow-list and an unlisted spender is invisible to the governor")
	}
	if !strings.Contains(sql, "gc.enabled") {
		t.Error("the master switch governor_config.enabled must gate the predicate")
	}

	// The ruled shed order: maintenance at level >= 1, build >= 2, research
	// (the ELSE arm) >= 3. Assert the pairs, not just presence.
	maint := strings.Index(sql, "WHEN 'maintenance' THEN 1")
	build := strings.Index(sql, "WHEN 'build'")
	if maint == -1 {
		t.Error("maintenance must shed at level >= 1 (owner ruling 2026-08-31: maintenance first)")
	}
	if build == -1 || !strings.Contains(sql[build:], "2") {
		t.Error("build must shed at level >= 2 (owner ruling 2026-08-31: builds second)")
	}
	if !strings.Contains(sql, "ELSE") {
		t.Error("research (the ELSE arm) must be the last class standing (level >= 3)")
	}
	if maint != -1 && build != -1 && maint > build {
		t.Error("threshold CASE lists build before maintenance — check the order was not flipped")
	}

	// The alias must reach the map join — a hardcoded alias would silently
	// misbind when a future caller uses a different one.
	if !strings.Contains(sql, "m.item_type = wi.item_type") {
		t.Error("renderer did not thread the caller's alias into the class-map join")
	}
	if other := workItemNotGovernorShedSQL("x"); !strings.Contains(other, "m.item_type = x.item_type") {
		t.Error("renderer ignores its alias argument")
	}
}

// The opt-in guarantee: no flag (or false) means the loader's statement is
// BYTE-IDENTICAL to the pre-governor one; true appends exactly the shared
// renderer (one contract, one spelling — never a second copy).
//
//	MUTATION "always-on": return the clause unconditionally -> the empty-string
//	branches fail.
//	MUTATION "private-copy": inline a second spelling of the predicate -> the
//	suffix comparison against the renderer fails.
func TestGovernorShedClauseIsOptIn(t *testing.T) {
	if got := governorShedClauseFor(map[string]interface{}{}); got != "" {
		t.Errorf("no flag must mean no clause (byte-identical statement); got %q", got)
	}
	if got := governorShedClauseFor(map[string]interface{}{"honour_spend_governor": false}); got != "" {
		t.Errorf("explicit false must mean no clause; got %q", got)
	}
	if got := governorShedClauseFor(map[string]interface{}{"honour_spend_governor": "true"}); got != "" {
		t.Errorf("a non-bool value must not enable the governor (jsonb strings do not count); got %q", got)
	}
	on := governorShedClauseFor(map[string]interface{}{"honour_spend_governor": true})
	if !strings.HasPrefix(on, "\n\t\t  AND ") || !strings.HasSuffix(on, workItemNotGovernorShedSQL("wi")) {
		t.Error("enabled clause must be exactly ' AND ' + the shared renderer output — " +
			"a drifting private copy here is the cross-media split the renderer exists to prevent")
	}
}
