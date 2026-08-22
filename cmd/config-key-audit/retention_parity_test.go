package main

import "testing"

// The sweep arm as migration 567 leaves it. Trimmed to the shape the parity
// check reads — the DELETE, the list, and the RETURNING that bounds it.
const testSweepArm = `
    WITH deleted_errors AS (
        DELETE FROM agent_error_log
        WHERE (occurred_at < NOW() - INTERVAL '30 days'
                AND split_part(error_code, ':', 1) = ANY (ARRAY[
                      'UNKNOWN', 'PROCESSING_FAILED', 'TIMEOUT',
                      'RESOLVER_CONFLICTING_CANDIDATES'
                    ]))
           OR occurred_at < NOW() - INTERVAL '365 days'
        RETURNING id
    ),
    deleted_audit AS (
        DELETE FROM orchestration_state_audit WHERE id < 5 RETURNING id
    )`

func armFor(t *testing.T, pre string) string {
	t.Helper()
	arm, ok := extractErrorLogArm(pre)
	if !ok {
		t.Fatalf("extractErrorLogArm found no agent_error_log arm")
	}
	return arm
}

// The arm must stop at its own RETURNING. Without that bound a code named by a
// LATER arm would count as short-retention here and the check would report
// parity that is not there.
func TestErrorLogArmStopsAtItsOwnReturning(t *testing.T) {
	arm := armFor(t, testSweepArm)
	if !namedInShortRetention(arm, "UNKNOWN") {
		t.Errorf("arm should contain its own list entry UNKNOWN")
	}
	if namedInShortRetention(arm, "orchestration_state_audit") {
		t.Errorf("arm leaked into the next CTE — the RETURNING bound is not holding")
	}
}

func TestErrorLogArmAbsentIsReportedNotAssumed(t *testing.T) {
	if _, ok := extractErrorLogArm("SELECT 1"); ok {
		t.Errorf("a pre_query with no agent_error_log arm must not report one")
	}
	// A DELETE with no RETURNING is also unreadable rather than empty: returning
	// ok=true with a runaway string would make every code look short-retention.
	if _, ok := extractErrorLogArm("DELETE FROM agent_error_log WHERE true"); ok {
		t.Errorf("an unbounded arm must be reported unreadable, not parsed")
	}
}

// The load-bearing direction: a deliberate finding that ends up in the
// short-retention list is bugs_open/358 happening again, and must be a finding.
func TestFindingCodeInShortRetentionListIsAFinding(t *testing.T) {
	arm := armFor(t, testSweepArm)
	reg := map[string]findingCodeEntry{
		"UNKNOWN":                         {Disposition: "operational"},
		"PROCESSING_FAILED":               {Disposition: "operational"},
		"TIMEOUT":                         {Disposition: "unruled"}, // <- mis-listed on purpose
		"RESOLVER_CONFLICTING_CANDIDATES": {Disposition: "instrumented"},
	}
	got := auditRetentionParity(arm, reg)
	if len(got) != 1 {
		t.Fatalf("want exactly 1 finding, got %d: %+v", len(got), got)
	}
	if got[0].Code != "TIMEOUT" || got[0].Kind != "retention_parity_finding_expires_early" {
		t.Errorf("wrong finding: %+v", got[0])
	}
}

// The other direction: plumbing that is NOT in the list silently gains a
// 365-day clock, and nothing else in the estate would say so.
func TestOperationalCodeMissingFromListIsAFinding(t *testing.T) {
	arm := armFor(t, testSweepArm)
	reg := map[string]findingCodeEntry{
		"UNKNOWN":       {Disposition: "operational"},
		"LLM_API_ERROR": {Disposition: "operational"}, // absent from the list
	}
	got := auditRetentionParity(arm, reg)
	if len(got) != 1 {
		t.Fatalf("want exactly 1 finding, got %d: %+v", len(got), got)
	}
	if got[0].Code != "LLM_API_ERROR" || got[0].Kind != "retention_parity_missing" {
		t.Errorf("wrong finding: %+v", got[0])
	}
}

// `instrumented` is deliberately unasserted on BOTH sides — whether a
// time-boxed measurement wants history or only frequency is its review_by's
// call, not this check's. A test pins that, so a later "tidy-up" that starts
// asserting it has to delete a test that says why.
func TestInstrumentedIsUnassertedOnBothSides(t *testing.T) {
	arm := armFor(t, testSweepArm)
	for _, reg := range []map[string]findingCodeEntry{
		{"RESOLVER_CONFLICTING_CANDIDATES": {Disposition: "instrumented"}}, // in the list
		{"RESOLVER_MAPPING_BYPASSED": {Disposition: "instrumented"}},       // not in the list
	} {
		if got := auditRetentionParity(arm, reg); len(got) != 0 {
			t.Errorf("instrumented must not be asserted either way, got %+v", got)
		}
	}
}

// A registry that agrees with the sweep produces nothing. Stated as its own
// test because every assertion above is a positive: without this one, a check
// that always fired would still pass them all.
func TestAgreeingRegistryIsSilent(t *testing.T) {
	arm := armFor(t, testSweepArm)
	reg := map[string]findingCodeEntry{
		"UNKNOWN":                         {Disposition: "operational"},
		"PROCESSING_FAILED":               {Disposition: "operational"},
		"TIMEOUT":                         {Disposition: "operational"},
		"RESOLVER_CONFLICTING_CANDIDATES": {Disposition: "instrumented"},
		"CONTENT_LINK_REPAIR_DETAIL":      {Disposition: "unruled"},
		"CONTENT_KEY_LOSS":                {Disposition: "consumed"},
	}
	if got := auditRetentionParity(arm, reg); len(got) != 0 {
		t.Errorf("an agreeing registry must be silent, got %+v", got)
	}
}
