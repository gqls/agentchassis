package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/livespec"
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
	//
	// THE CENSUS, settled and stated once (three seats objected to a submission
	// that gave three different numbers — 3, 4 and 5 — for the same claim, and
	// they were right that an unclosed census is where a missed duplicate hides):
	// FOUR call sites across THREE consumer files, plus the definition. Two more
	// copies of the same contract live in SQL (migration 506's two dispatch reads),
	// which is what TestGoAndSQLAgreeOnTheCooldownBoundary exists for.
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

// TestClaimPredicateIsByteIdenticalToWhatItReplaced answers the council's GATING
// objection (guardian, high, corr 2c21e214): ClaimWorkItemAction is the fleet's
// central claim path — every dispatch loop claiming any item type — and the
// submission asserted the rendered SQL was "byte-identical" to the hand-typed
// clause it replaced WITHOUT proving it. A stray paren or space there is not a
// contained regression, it is a fleet-wide claim outage.
//
// The assertion is cheap and it is the one that was missing.
func TestClaimPredicateIsByteIdenticalToWhatItReplaced(t *testing.T) {
	// The exact text that stood in claim_work_item_action.go and
	// load_work_item_actions.go before the renderer existed, transcribed from
	// the 307 commit rather than retyped from memory.
	const wasBare = "(retry_after IS NULL OR retry_after <= NOW())"
	const wasAliased = "(wi.retry_after IS NULL OR wi.retry_after <= NOW())"

	if got := workItemRetryNotPendingSQL(""); got != wasBare {
		t.Errorf("claim path predicate CHANGED: got %q, want %q — this is the fleet's claim gate", got, wasBare)
	}
	if got := workItemRetryNotPendingSQL("wi"); got != wasAliased {
		t.Errorf("dispatch selection predicate CHANGED: got %q, want %q", got, wasAliased)
	}

	// And it must still be syntactically embeddable: balanced parens, no stray
	// quotes, no trailing comma. A renderer is a string function; nothing else
	// checks its output until Postgres does, at which point the claim is already
	// failing in production.
	for _, alias := range []string{"", "wi", "active"} {
		p := workItemRetryNotPendingSQL(alias)
		if strings.Count(p, "(") != strings.Count(p, ")") {
			t.Errorf("alias %q: unbalanced parens in %q", alias, p)
		}
		if strings.Contains(p, "'") || strings.HasSuffix(p, ",") {
			t.Errorf("alias %q: %q is not safely embeddable in a WHERE clause", alias, p)
		}
	}
}

// TestGoAndTheDeclarationAgreeOnTheCooldownBoundary answers guardian's low objection: the
// same contract lives in Go AND in DB-resident SQL (migration 506's two dispatch
// reads; 341's held 524 adds a third). The Go side is now one renderer, but that
// closes only half the drift — if either side later moved to `<` while the other
// kept `<=`, an item would be claimable a moment before or after it is
// completable, and the disagreement would be invisible until a race surfaced it.
//
// It reads platform/livespec, NOT migration 506 (bugs_open/363). The old form had
// two defects that made its green uninformative: it asserted the TEXT of an applied
// migration, which the checksum rule in schema_migrations forbids anyone changing,
// so it could not fail in the direction it existed to detect; and it called t.Skipf
// when the file was unreadable, so a RENAME was a silent pass. ⚠ Nothing here
// compares the declaration to the LIVE row — that is the phase-2 auditor.
func TestGoAndTheDeclarationAgreeOnTheCooldownBoundary(t *testing.T) {
	d := livespec.MustGet("scheduled_task.build-pipeline-trigger.retry_cooldown")

	// The rendered Go predicate must be what the declaration requires of the live row.
	want := workItemRetryNotPendingSQL("wi")
	if !d.HasFragment(want) {
		t.Errorf("the declaration does not require the Go renderer's predicate %q.\n"+
			"One contract in two media has drifted — check the boundary (<= vs <) and the NULL arm.", want)
	}
	// The specific drift that would otherwise be silent: a strict inequality. The
	// declaration must FORBID it outright, not merely omit it.
	if !d.Forbids("retry_after < NOW()") {
		t.Error("the declaration no longer forbids the STRICT boundary — with '<' an item would be " +
			"claimable a moment before it is completable, and nothing else would report the disagreement")
	}
}

// TestCompletionSkipIsRecordedDurablyOnTheRow answers the council's bug_historian
// objection (corr 2c21e214, medium): a refusal recorded only in a pod log is the
// shape of bugs_closed/034 — logs rotate in ~90s, and the operator who later asks
// "why did this item never complete?" has nothing to read.
//
// The mutation that exposed the need for this test is worth recording: removing
// the durable write entirely produced "[no tests to run]", i.e. the change I made
// to answer the objection had no assertion behind it at all. That is the same
// failure as the 42P18 defect one file over — a fix believed rather than pinned.
func TestCompletionSkipIsRecordedDurablyOnTheRow(t *testing.T) {
	src := read344Source(t, "load_work_item_actions.go")
	i := strings.Index(src, "rows == 0")
	if i < 0 {
		if i = strings.Index(src, "if rows == 0 {"); i < 0 {
			t.Fatal("could not locate CompleteWorkItemAction's skip branch")
		}
	}
	branch := src[i:]
	if j := strings.Index(branch, "logger.Info(\"CompleteWorkItemAction: Done\""); j > 0 {
		branch = branch[:j]
	}

	if !strings.Contains(branch, "completion_skipped") {
		t.Error("the skip branch writes no durable record to the row — only a log line, which rotates " +
			"in ~90s (bugs_closed/034's shape, raised by the council's bug_historian seat)")
	}
	if !strings.Contains(branch, "UPDATE site_work_items") {
		t.Error("the skip branch does not persist anything at all")
	}
	// The classification must happen INSIDE that write, not in a second read —
	// editquality's objection: an unsynchronised follow-up SELECT can mislabel a
	// refusal if the row moved in between.
	if !strings.Contains(branch, "CASE WHEN retry_after") {
		t.Error("the skip reason is not derived inside the recording UPDATE — a separate read " +
			"re-opens the race the council flagged")
	}
	if strings.Contains(branch, "SELECT retry_after IS NOT NULL AND retry_after > NOW()") {
		t.Error("the superseded second-read classification is still present")
	}
}

func TestASuccessfulCompletionWIPESTheSkipMarker(t *testing.T) {
	// guardian's stale-state objection: this fix deliberately increases the rate of
	// refused completions, and a landmine on record says a completed row can keep
	// error/result text from an EARLIER refused completion. It cannot here, and the
	// reason is structural rather than careful: the success path REPLACES result
	// (`result = $2::jsonb`) rather than merging it, so the marker cannot survive
	// into the record of a genuine completion.
	//
	// Pinned because it is load-bearing for the objection's answer: if someone
	// later "improves" that write into a merge, the marker starts leaking into
	// successful records and the answer silently stops being true.
	src := read344Source(t, "load_work_item_actions.go")
	stmt := extractStatement(t, src, "SET status = 'complete',")
	if !strings.Contains(stmt, "result = $2::jsonb") {
		t.Error("the success path no longer REPLACES result — a completion_skipped marker from an " +
			"earlier refusal could now survive into a successful record (guardian's stale-state objection)")
	}
	if strings.Contains(stmt, "COALESCE(result") && strings.Contains(stmt, "|| $2::jsonb") {
		t.Error("the success path now MERGES result — see above; the skip marker would leak")
	}
}
