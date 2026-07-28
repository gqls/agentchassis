// FILE: platform/orchestration/intake_repo_test.go
//
// Three SQL clauses in intake_repo.go carry the whole concurrency contract,
// and all three are inline strings a type system cannot see. In the
// state_locks_test.go style, these tests read them back out of the source and
// fail if the load-bearing part is edited away:
//   1. the claim CAS takes over ONLY an expired lease — remove the WHERE and
//      any worker can steal a live key;
//   2. the event pop is oldest-first within the key — lose the ORDER BY and
//      per-orchestration ordering silently ends;
//   3. the candidate scan skips keys with a LIVE claim — lose the NOT EXISTS
//      and every worker converges on the same busy key.

package orchestration

import (
	"regexp"
	"strings"
	"testing"
)

func intakeSQLOf(t *testing.T, funcDecl string) string {
	t.Helper()
	body := funcBodyOf(t, "intake_repo.go", funcDecl)
	m := regexp.MustCompile("(?s)`(.*?)`").FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no backtick SQL literal found inside %s", funcDecl)
	}
	return strings.ToLower(m[1])
}

func TestClaimCASTakesOverOnlyExpiredLeases(t *testing.T) {
	sql := intakeSQLOf(t, "func (r *IntakeRepository) ClaimSerialisationKey")
	if !strings.Contains(sql, "on conflict") || !strings.Contains(sql, "lease_expires_at <= now()") {
		t.Fatalf("the claim CAS lost its expired-lease guard — a live key could be stolen. SQL:\n%s", sql)
	}
}

func TestEventPopIsOldestFirstWithinKey(t *testing.T) {
	sql := intakeSQLOf(t, "func (r *IntakeRepository) NextPendingEvent")
	if !strings.Contains(sql, "order by id limit 1") {
		t.Fatalf("NextPendingEvent lost oldest-first ordering — per-orchestration order ends silently. SQL:\n%s", sql)
	}
	if !strings.Contains(sql, "status = 'pending'") {
		t.Fatalf("NextPendingEvent no longer restricts to pending rows. SQL:\n%s", sql)
	}
}

func TestCandidateScanSkipsLiveClaims(t *testing.T) {
	sql := intakeSQLOf(t, "func (r *IntakeRepository) CandidateKeys")
	if !strings.Contains(sql, "not exists") || !strings.Contains(sql, "lease_expires_at > now()") {
		t.Fatalf("CandidateKeys lost its live-claim filter — workers would converge on busy keys. SQL:\n%s", sql)
	}
}
