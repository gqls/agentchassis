// FILE: platform/orchestration/state_locks_test.go
//
// orchestration_states has TWO guarded-update mechanisms and they must never
// govern the same column (bugs_open/075; council objection corr 4a227ed9,
// guardian + reuse_agent). UpdateStateWithVersion's version-CAS owns every
// workflow field; TakeOverOrchestration's pod-name CAS owns processing_node
// and nothing else. If either ever writes the other's columns, two locks race
// on one field and neither is authoritative.
//
// That invariant was stated in a comment, and the guardian's objection was
// exactly right: a comment does not survive a future edit. These tests read the
// SQL out of state.go and fail if the column sets overlap — the cheapest thing
// that will actually stop it, given both statements are inline strings rather
// than anything a type system can see.

package orchestration

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// workflowColumns are the fields the version-CAS is authoritative for. Not
// exhaustive by design — it is a tripwire, not a schema.
var workflowColumns = []string{
	"status", "current_step", "awaited_steps", "awaited_requests",
	"collected_data", "execution_metadata", "version", "last_activity",
}

// setClauseOf returns the SET…WHERE region of the first UPDATE on
// orchestration_states inside the named function, so a neighbouring
// function's SQL can never be read by mistake.
func setClauseOf(t *testing.T, funcDecl string) string {
	t.Helper()

	src, err := os.ReadFile("state.go")
	if err != nil {
		t.Fatalf("read state.go: %v", err)
	}
	body := string(src)

	i := strings.Index(body, funcDecl)
	if i < 0 {
		t.Fatalf("function %q not found in state.go — renamed? update this test with it", funcDecl)
	}
	body = body[i+len(funcDecl):]
	if j := strings.Index(body, "\nfunc "); j >= 0 {
		body = body[:j] // stop at the next top-level func
	}

	m := regexp.MustCompile(`(?is)UPDATE\s+orchestration_states\s+SET(.*?)WHERE`).FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no `UPDATE orchestration_states … SET … WHERE` found inside %s", funcDecl)
	}
	return strings.ToLower(m[1])
}

// The ownership CAS must touch processing_node (and its bookkeeping stamp) and
// nothing the version-CAS is responsible for.
func TestOwnershipCASWritesOnlyOwnershipColumns(t *testing.T) {
	set := setClauseOf(t, "func (r *StateRepository) TakeOverOrchestration")

	if !strings.Contains(set, "processing_node") {
		t.Fatalf("ownership CAS does not write processing_node — the test is reading the wrong statement:\n%s", set)
	}
	for _, col := range workflowColumns {
		if regexp.MustCompile(`\b` + col + `\s*=`).MatchString(set) {
			t.Errorf("ownership CAS assigns %q, which belongs to UpdateStateWithVersion's version-CAS.\n"+
				"Two locks on one column means neither is authoritative (bugs_open/075). SET clause:\n%s", col, set)
		}
	}
}

// And the version-CAS must never write processing_node — which is also the
// original defect in a second costume: an UPDATE that omits a column silently
// discards whatever the caller assigned in memory (SetExecutingStep).
func TestVersionCASNeverWritesProcessingNode(t *testing.T) {
	set := setClauseOf(t, "func (r *StateRepository) UpdateStateWithVersion")

	if !strings.Contains(set, "status") {
		t.Fatalf("version-CAS SET clause does not mention status — the test is reading the wrong statement:\n%s", set)
	}
	if regexp.MustCompile(`\bprocessing_node\s*=`).MatchString(set) {
		t.Error("UpdateStateWithVersion assigns processing_node. Ownership is TakeOverOrchestration's column; " +
			"adding it here puts two guarded updates on one field (bugs_open/075).")
	}
}
