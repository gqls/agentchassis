package actions

// D4 stage B (register AGOV-013): contract tests for the spend-governor wiring.
//
// The shed LOGIC is deliberately absent from Go — `governor_admits(item_type)`
// (migration 675) is the single canonical predicate, and its posture rules
// (fail-open, unmapped-sheds-earliest, disabled-is-identity, the full shed
// order) are proven by EXECUTION probes in that migration's verify, against
// synthetic governor states. What Go owes is narrower and is what these tests
// pin: the call shape (single source — no local re-spelling of the rule can
// creep back in) and the opt-in guarantee (no flag = byte-identical statement).

import (
	"strings"
	"testing"
)

// MUTATION "re-spell": someone inlines the shed logic here "to save a DB call"
// — the exact drift the function exists to prevent — and the single-call-shape
// assertions fail.
// MUTATION "alias-drop": hardcode the alias — the second-alias check fails.
func TestGovernorShedSQLIsASingleCanonicalCall(t *testing.T) {
	sql := strings.TrimSpace(workItemNotGovernorShedSQL("wi"))
	if sql != "governor_admits(wi.item_type)" {
		t.Errorf("the renderer must emit exactly the canonical one-line call — the shed LOGIC "+
			"lives in governor_admits() (migration 675) and nowhere else; got %q", sql)
	}
	if got := strings.TrimSpace(workItemNotGovernorShedSQL("x")); got != "governor_admits(x.item_type)" {
		t.Errorf("renderer ignores its alias argument; got %q", got)
	}
	for _, forbidden := range []string{"governor_work_class_map", "shed_level", "COALESCE"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("renderer output contains %q — the shed logic is being re-spelled in Go "+
				"instead of called; that is the four-copies drift the 8f4bb57d r1 round rejected", forbidden)
		}
	}
}

// The opt-in guarantee: no flag (or false, or a non-bool) means the loader's
// statement is BYTE-IDENTICAL to the pre-governor one; true appends exactly
// ' AND ' + the canonical call.
//
//	MUTATION "always-on": return the clause unconditionally -> the empty-string
//	branches fail.
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
		t.Error("enabled clause must be exactly ' AND ' + the shared renderer output")
	}
}
