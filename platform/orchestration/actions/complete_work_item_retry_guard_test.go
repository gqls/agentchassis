package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// bugs_open/344 — a completion must not overwrite an item that is sitting out a
// retry cooldown the failure ladder scheduled seconds earlier.
//
// THE DEFECT THIS PINS, in one line: the dispatch loop calls complete_work_item on
// every returned saga; a handler that fails a step and then ends via a
// success-labelled complete_workflow returns as a SUCCESS; and `triaged` is not in
// the completion guard — so the ladder's fresh `triaged` was overwritten to
// `complete` about two seconds later, cancelling the retry and recording a failed
// build as done. Measured live: retry_after 11:02:50 against completed_at 10:32:52.
//
// WHY THESE ARE SOURCE-LEVEL ASSERTIONS AND NOT sqlmock ROUND-TRIPS. The sibling
// suite for the ladder had fifteen sqlmock tests and could not see that a bound
// parameter had been dropped from the statement text (SQLSTATE 42P18, WRONG_CALLS
// 2026-08-21) — because a mock never parses SQL and never types a placeholder. The
// property that matters here is likewise a property of the STATEMENT: does the
// predicate appear in each writer, spelled from the one shared renderer. A mock
// would happily report "0 rows affected" for a statement with no predicate at all,
// which is precisely the false pass this file exists to refuse.

func read344Source(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(".", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

// extractStatement pulls the raw-string SQL that follows a marker, so an assertion
// lands on the statement rather than anywhere in a 1,600-line file.
func extractStatement(t *testing.T, src, marker string) string {
	t.Helper()
	i := strings.Index(src, marker)
	if i < 0 {
		t.Fatalf("marker %q not found — the function was renamed or restructured; re-derive this test rather than deleting it", marker)
	}
	rest := src[i:]
	// The statement is a CONCATENATION — raw-string fragments joined by calls to
	// the shared renderers — so reading a single backtick pair would stop at the
	// first ` + "`" + ` and miss the very clause under test. Span to the closing paren of
	// the Exec/Query call instead.
	stop := strings.Index(rest, "\n\tif err")
	if stop < 0 || stop > 4000 {
		stop = 4000
		if stop > len(rest) {
			stop = len(rest)
		}
	}
	return rest[:stop]
}

func TestCompletionRefusesAnItemWithAPendingRetry(t *testing.T) {
	// Both writers of a completion-shaped outcome must carry the predicate. They
	// refuse for different reasons — CompleteWorkItemAction to preserve the
	// scheduled retry, failUnverifiedCompletion to avoid charging a second
	// attempt for one fault — and both reasons are void if the clause is absent.
	cases := []struct {
		file, marker, why string
	}{
		{
			file:   "load_work_item_actions.go",
			marker: "SET status = 'complete',",
			why:    "the dispatch loop's mark_complete — the writer measured overwriting a re-triaged item 2s after its failure",
		},
		{
			file:   "complete_work_item_verification.go",
			marker: "func failUnverifiedCompletion",
			why:    "the verification-failure path, reached after the ladder has already counted the same failure",
		},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			stmt := extractStatement(t, read344Source(t, tc.file), tc.marker)
			if !strings.Contains(stmt, "workItemRetryNotPendingSQL") {
				t.Errorf("%s does not apply the retry-cooldown predicate (%s).\nStatement:\n%s", tc.file, tc.why, stmt)
			}
		})
	}
}

func TestRetryPredicateIsRenderedOnceNotCopied(t *testing.T) {
	// Two media already hold this contract (Go here, the claimed-item-timeout
	// pre_query in SQL). Within Go it must come from ONE renderer, or the third
	// copy is where the drift starts — the failure this estate has filed under
	// 034/195/197 and again under 213/285.
	// EVERY Go consumer of the predicate, not just the two 344 touches. The claim
	// path was omitted from this list at first and a mutation re-inlining a copy
	// there passed — the test was narrower than its own name. work_items_common.go
	// is excluded on purpose: it is where the renderer and its doc comment live.
	for _, f := range []string{
		"load_work_item_actions.go",          // dispatch selection + the completion writer
		"complete_work_item_verification.go", // the verification-failure writer
		"claim_work_item_action.go",          // the atomic claim
	} {
		src := read344Source(t, f)
		// A LITERAL copy of the whole predicate is the drift; a reference to the
		// renderer is not. (This caught two real inline copies when it was first
		// run — the dispatch read path and the claim path, both written for
		// bugs_open/307 before this renderer existed. Both now call it.)
		if regexp.MustCompile(`retry_after IS NULL OR .*retry_after <= NOW\(\)`).MatchString(src) {
			t.Errorf("%s inlines the predicate instead of calling workItemRetryNotPendingSQL — one rule, one renderer", f)
		}
	}
	common := read344Source(t, "work_items_common.go")
	if !strings.Contains(common, "func workItemRetryNotPendingSQL(") {
		t.Fatal("the shared renderer is missing from work_items_common.go, where the other shared work-item SQL lives")
	}
}

func TestRetryPredicateRendersForBothAliasedAndBareStatements(t *testing.T) {
	// An UPDATE has no alias; a SELECT joining sites does. Getting the bare case
	// wrong yields `.retry_after`, which is a syntax error the caller only meets
	// at runtime — the 42P18 lesson one file over.
	if got, want := workItemRetryNotPendingSQL(""), "(retry_after IS NULL OR retry_after <= NOW())"; got != want {
		t.Errorf("bare: got %q, want %q", got, want)
	}
	if got, want := workItemRetryNotPendingSQL("wi"), "(wi.retry_after IS NULL OR wi.retry_after <= NOW())"; got != want {
		t.Errorf("aliased: got %q, want %q", got, want)
	}
}

func TestCompletionGuardStillAllowsTheLEGITIMATERetryThenSuccess(t *testing.T) {
	// THE DISCONFIRMING CONTROL, and the reason this fix is a predicate rather
	// than a status word. A fix that simply refused completions would pass every
	// other test in this file and would be catastrophic: it would strand every
	// item that legitimately failed, waited, and then succeeded.
	//
	// The recorded live shape (2026-08-20, two natural transient rows): stamp
	// 18:34:00Z → re-claim 18:34:25Z → completion 18:34:51Z. At completion time
	// the stamp is in the PAST, so `retry_after <= NOW()` is TRUE and the row
	// completes. That is guaranteed by the claim path, which refuses to re-claim
	// before the stamp expires — the two halves are one mechanism.
	//
	// Asserted on the predicate's own semantics, since it is a pure renderer:
	// the clause must admit a NULL stamp (never failed) and a past stamp
	// (cooldown served), and exclude only a future one.
	pred := workItemRetryNotPendingSQL("")
	if !strings.Contains(pred, "IS NULL OR") {
		t.Error("the predicate does not admit a NULL retry_after — every item that has never failed would become uncompletable")
	}
	if !strings.Contains(pred, "<= NOW()") {
		t.Error("the predicate does not admit an EXPIRED cooldown — an item that failed, waited and then succeeded could never be completed")
	}
	if strings.Contains(pred, "> NOW()") {
		t.Error("the predicate is inverted: it would admit exactly the rows it must refuse")
	}
}

func TestTriagedIsNotAddedToTheCompletionGuardList(t *testing.T) {
	// 344 candidate 3, explicitly rejected in the bug file and pinned here so a
	// later reader does not "simplify" the predicate into the list. `triaged` is
	// an in-progress status; guarding it would refuse every legitimate completion
	// of an item re-triaged mid-run for unrelated reasons, and would protect the
	// status WORD rather than the decision ("a retry is scheduled") that the word
	// only sometimes carries.
	for _, s := range workItemCompletionGuardStatuses {
		if s == "triaged" {
			t.Error("'triaged' is in the completion guard list — that is 344 candidate 3, which the bug file rejects: " +
				"it protects a word rather than the decision, and refuses legitimate completions")
		}
	}
}
