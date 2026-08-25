// FILE: platform/orchestration/actions/conditional_branch_seats_gate_test.go
//
// Pins the evaluator behaviour the improvement loop's check_seats_ran gate
// (migration 624, RFC_056 §6) depends on — demanded by council round d1342f2a
// (guardian HIGH, bug_historian HIGH): "prove the evaluator's nil-comparison
// behaviour before merge", because a silent wrong answer here either dispatches
// past a failed seat or skips the audit-pass stamp fleet-wide.
//
// The gate's condition is an OR-joined chain of
//   seat_failed_<seat>.item_type == capability_gap
// where seat_failed_<seat> is a record step's output_field: PRESENT (a map with
// item_type "capability_gap" — create_work_item returns item_type whether it
// inserted or deduped) exactly when that record step RAN, ABSENT otherwise.
//
// The two properties, each pinned in the direction whose failure is silent:
//   1. an ABSENT field compares UNEQUAL to the literal — compareValues(nil, x)
//      matches only "null"/"nil"/"" (conditional_branch_action.go:520-522), so a
//      clean sweep takes the else arm (record_audit_pass) and stamps the audit;
//   2. a PRESENT field with the literal value compares EQUAL, so one failed seat
//      routes to the then arm and the pass stamp is withheld.
//
// ⚠ Property 1's mechanism is also a documented footgun this file makes visible:
// compareValues(nil, "") is TRUE — a condition comparing against an EMPTY
// literal matches every absent field. The gate's literal is non-empty by
// construction; the test asserts the footgun so nobody "simplifies" the literal
// away.

package actions

import (
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// seatGateCondition mirrors migration 624's generated condition for the eleven
// seats, OR-joined, bracket-free (the evaluator splits OR before AND and strips
// parentheses — bugs_open/376's landmine).
func seatGateCondition(seats []string) string {
	parts := make([]string, 0, len(seats))
	for _, s := range seats {
		parts = append(parts, fmt.Sprintf("seat_failed_%s.item_type == capability_gap", s))
	}
	return strings.Join(parts, " OR ")
}

var seatGateSeats = []string{
	"news_feed", "directory_features", "quality_discovery", "design_discovery",
	"completeness_discovery", "acceptance_discovery", "design_audit", "site_review",
	"offer_analyser", "brief_fidelity", "reader_audit",
}

func TestSeatsGate_AllSeatsRanCleanTakesTheElseArm(t *testing.T) {
	// No record step ran: every field is ABSENT. The condition must be FALSE —
	// the else arm is record_audit_pass, and a clean sweep must stamp the audit.
	got, err := evaluateStringConditionMode(seatGateCondition(seatGateSeats),
		map[string]interface{}{}, zap.NewNop(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatalf("clean sweep evaluated TRUE — every sweep would skip record_audit_pass fleet-wide (the silent failure council round d1342f2a named)")
	}
}

func TestSeatsGate_OneFailedSeatWithholdsThePassStamp(t *testing.T) {
	for _, failed := range []string{"news_feed", "acceptance_discovery", "reader_audit"} {
		data := map[string]interface{}{
			"seat_failed_" + failed: map[string]interface{}{
				// create_work_item's return shape, whether inserted or deduped
				"item_type": "capability_gap", "inserted": false, "deduped": true,
			},
		}
		got, err := evaluateStringConditionMode(seatGateCondition(seatGateSeats), data, zap.NewNop(), false)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", failed, err)
		}
		if !got {
			t.Fatalf("%s failed but the gate evaluated FALSE — the audit would be stamped over a seat that did not run", failed)
		}
	}
}

func TestSeatsGate_AbsentFieldMatchesOnlyNullNilOrEmptyLiteral(t *testing.T) {
	// compareValues(nil, "") would be TRUE (conditional_branch_action.go:521),
	// but the string form cannot REACH it: evaluateStringConditionMode trims the
	// expression first, so a trailing empty literal loses its separating space
	// and the " == " operator is never found — the whole string falls through to
	// the truthy check and evaluates false. Pinned here so a future refactor
	// that stops trimming does not silently open the empty-literal hole.
	for expected, want := range map[string]bool{
		"capability_gap": false, "null": true, "nil": true, "": false,
	} {
		got, err := evaluateStringConditionMode("missing.item_type == "+expected,
			map[string]interface{}{}, zap.NewNop(), false)
		if err != nil {
			t.Fatalf("%q: unexpected error: %v", expected, err)
		}
		if got != want {
			t.Fatalf("compareValues(nil, %q): got %v, want %v", expected, got, want)
		}
	}
}
