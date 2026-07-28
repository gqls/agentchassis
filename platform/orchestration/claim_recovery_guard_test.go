// FILE: platform/orchestration/claim_recovery_guard_test.go
//
// CLAIM_RECOVERY used to reset ANY claimed-but-unprocessed awaited_request back
// to 'waiting' — including one claimed milliseconds earlier by a live actor. A
// duplicate delivery of a response could therefore steal a live claim and run a
// step's side effects twice (chassis_replica_scaling CS-1; the hazard
// bugs_open/075 deliberately left unfixed as out of its scope). The fix has two
// halves: the staleness predicate lives INSIDE the UPDATE's WHERE clause so the
// decision is atomic with the reset, and the call site treats "not reset" as a
// duplicate delivery rather than retrying the claim. Like state_locks_test.go,
// these tests read the source, because both halves are inline strings and
// control flow that no type system can see — a future edit that moves the
// predicate into Go, or reverts to an unconditional reset, fails here first.

package orchestration

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// funcBodyOf returns the body of the named top-level function from the named
// file, so a neighbouring function can never be read by mistake.
func funcBodyOf(t *testing.T, file, funcDecl string) string {
	t.Helper()

	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	body := string(src)

	i := strings.Index(body, funcDecl)
	if i < 0 {
		t.Fatalf("function %q not found in %s — renamed? update this test with it", funcDecl, file)
	}
	body = body[i+len(funcDecl):]
	if j := strings.Index(body, "\nfunc "); j >= 0 {
		body = body[:j]
	}
	return body
}

// The reset's guard must be part of the UPDATE itself. A separate read followed
// by a reset reintroduces exactly the race this closes: the claim can land
// between the read and the write.
func TestStaleResetPredicateIsInsideTheUpdate(t *testing.T) {
	body := funcBodyOf(t, "coordinator.go", "func (r *StateRepository) ResetStaleAwaitedRequestForRetry")

	m := regexp.MustCompile(`(?is)UPDATE\s+awaited_requests\s+SET.*?WHERE(.*?)` + "`").FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no `UPDATE awaited_requests … WHERE …` found inside ResetStaleAwaitedRequestForRetry")
	}
	where := strings.ToLower(m[1])

	for _, predicate := range []string{
		"processed_at is null",    // never touch a processed row
		"status <> 'processing'",  // non-processing states keep resetting unconditionally (F2's late-response window)
		"processing_started_at <", // the staleness comparison itself
	} {
		if !strings.Contains(where, predicate) {
			t.Errorf("reset WHERE clause lost %q — the staleness decision must stay atomic with the reset. WHERE:\n%s", predicate, where)
		}
	}
}

// The call site must hold a fresh live claim (treat as duplicate), not loop
// back and retry the claim, and the unguarded reset must not come back.
func TestClaimRecoveryHoldsFreshClaims(t *testing.T) {
	body := funcBodyOf(t, "coordinator.go", "func processResponseClaimWithRetry")

	if !strings.Contains(body, "ResetStaleAwaitedRequestForRetry") {
		t.Fatal("processResponseClaimWithRetry no longer calls ResetStaleAwaitedRequestForRetry — " +
			"an unguarded reset lets a duplicate delivery steal a live claim (CS-1)")
	}
	if !strings.Contains(body, "CLAIM_RECOVERY_STALENESS_HELD") {
		t.Error("the held branch (fresh live claim → treat as duplicate, return) is gone from processResponseClaimWithRetry")
	}

	src, err := os.ReadFile("coordinator.go")
	if err != nil {
		t.Fatalf("read coordinator.go: %v", err)
	}
	if regexp.MustCompile(`\bfunc \(r \*StateRepository\) ResetAwaitedRequestForRetry\b`).Match(src) {
		t.Error("the unguarded ResetAwaitedRequestForRetry is back in coordinator.go — " +
			"every reset on this path must carry the staleness guard")
	}
}
