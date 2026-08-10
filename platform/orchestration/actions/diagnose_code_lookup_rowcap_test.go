package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// bugs_open/181 — answerCodeCheck's three arms each capped at row_cap via SQL
// LIMIT and NONE of them said so, while the same function's max_checks sibling
// cap eight lines above the first arm DID report itself. A 40-of-305 answer
// rendered identically to a complete 40-row answer.
//
// Every assertion here has to be INDUCED. The cap fires on real corpora (five
// rendered blocks at exactly 40 rows were measured in llm_call_log prompts, true
// match counts 82/43/305/279) but not on demand, and the fix's whole subtlety —
// that a cap is OBSERVED from a probe row rather than INFERRED from `n ==
// rowCap` — is only visible with the cap lowered and the row count pinned.
//
// TWO facts each arm's pair of tests holds down:
//  1. rowCap+1 rows available -> rowCap rendered + the notice (the probe arrives);
//  2. exactly rowCap rows available -> byte-identical to a plainly-uncapped
//     render (the probe does not arrive, so nothing is claimed).
//
// The second is the one an inference-based fix fails, and it is the one that
// keeps every existing bundle baseline from moving.

// testCodeScope is a healthy corpus: the scope struct is only consulted for the
// zero-row branches, which the capped cases never reach.
var testCodeScope = codeIndexScope{total: 5000, withBody: 5000, commits: 1}

var (
	contentCols = []string{"path", "symbol", "body", "content", "commit_sha", "has_body", "kind"}
	lsCols      = []string{"path", "commit_sha", "has_code"}
	symbolCols  = []string{"path", "symbol", "signature", "line_start", "line_end", "commit_sha", "kind"}
)

func newCodeLookupDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return db, mock
}

// contentRowsFor renders one mocked content-arm row per (path, kind) pair. The
// body carries the query so the [body] marker branch is the one exercised.
func contentRowsFor(query string, pathsAndKinds ...[2]string) *sqlmock.Rows {
	rows := sqlmock.NewRows(contentCols)
	for _, pk := range pathsAndKinds {
		rows.AddRow(pk[0], "Sym", "func Sym() { "+query+" }", "func Sym()", "abc12345deadbeef", true, pk[1])
	}
	return rows
}

func symbolRowsFor(pathsAndKinds ...[2]string) *sqlmock.Rows {
	rows := sqlmock.NewRows(symbolCols)
	for _, pk := range pathsAndKinds {
		rows.AddRow(pk[0], "Foo", "func Foo() error", 10, 20, "abc12345deadbeef", pk[1])
	}
	return rows
}

// ── content arm ──────────────────────────────────────────────────────────────

// TestContentArmReportsCap induces the defect's exact shape on the content arm:
// three matches available, row_cap of two.
//
// WithArgs pins the LIMIT we BIND at rowCap+1, and that expectation is the guard
// against the probe being reverted to a plain rowCap: sqlmock does not implement
// LIMIT (it replays every row it was given), so the third row arriving here is
// what a real Postgres LIMIT 3 would deliver, and the loop's own break is what
// must discard it.
func TestContentArmReportsCap(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	mock.ExpectQuery("body ILIKE").
		WithArgs("needle", "", 3).
		WillReturnRows(contentRowsFor("needle",
			[2]string{"a/one.go", "func"},
			[2]string{"b/two.go", "func"},
			[2]string{"c/three.go", "func"}))

	var b strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "content", Query: "needle"}, "", 2, 400, testCodeScope, &b); err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()

	for _, want := range []string{"a/one.go", "b/two.go"} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered rows must survive the notice (%s missing); got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "c/three.go") {
		t.Errorf("the probe row exists only to be OBSERVED — it must never render; got:\n%s", got)
	}
	if !strings.Contains(got, "CAPPED (row_cap=2)") {
		t.Errorf("a capped content answer must say so, with its cap; got:\n%s", got)
	}
	if !strings.Contains(got, "UNKNOWN, not absent") {
		t.Errorf("the notice must forbid reading absence from a capped listing as absence; got:\n%s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query expectations unmet: %v", err)
	}
}

// TestContentArmExactlyAtCapIsSilent is the negative control in its strongest
// form: the SAME two rows rendered at row_cap=2 and at row_cap=99 must be EQUAL.
// An answer that happens to hold exactly row_cap matches is complete, and a fix
// that inferred the cap from `n == rowCap` would libel it.
func TestContentArmExactlyAtCapIsSilent(t *testing.T) {
	render := func(rowCap, wantLimit int) string {
		db, mock := newCodeLookupDB(t)
		defer db.Close()
		mock.ExpectQuery("body ILIKE").
			WithArgs("needle", "", wantLimit).
			WillReturnRows(contentRowsFor("needle",
				[2]string{"a/one.go", "func"},
				[2]string{"b/two.go", "func"}))
		var b strings.Builder
		if _, err := answerCodeCheck(context.Background(), db,
			codeCheck{Kind: "content", Query: "needle"}, "", rowCap, 400, testCodeScope, &b); err != nil {
			t.Fatalf("answerCodeCheck(rowCap=%d): %v", rowCap, err)
		}
		return b.String()
	}
	atCap, uncapped := render(2, 3), render(99, 100)
	if atCap != uncapped {
		t.Errorf("at-cap-exactly must be indistinguishable from plainly-uncapped\n at cap: %q\nuncapped: %q", atCap, uncapped)
	}
	if strings.Contains(atCap, "CAPPED") {
		t.Errorf("two rows at row_cap=2 were NOT truncated; a notice here is a false claim: %q", atCap)
	}
}

// ── ls arm ───────────────────────────────────────────────────────────────────

// TestLsArmReportsCap. This is the worst arm: ORDER BY path with nothing else to
// rank by, so the discarded tail is purely alphabetical — the identical shape
// bugs_closed/164 was filed and fixed for in the sibling file.
func TestLsArmReportsCap(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	rows := sqlmock.NewRows(lsCols).
		AddRow("platform/a.go", "abc12345deadbeef", true).
		AddRow("platform/b.go", "abc12345deadbeef", true).
		AddRow("platform/z.go", "abc12345deadbeef", true)
	// Fourth bind is codeKindsCSV and is untouched by this change; the THIRD is
	// the probe limit.
	mock.ExpectQuery("bool_or").
		WithArgs("platform/", "", 3, codeKindsCSV).
		WillReturnRows(rows)

	var b strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "ls", Query: "platform/"}, "", 2, 400, testCodeScope, &b); err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()

	for _, want := range []string{"platform/a.go", "platform/b.go"} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered listing lost %s; got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "platform/z.go") {
		t.Errorf("the probe row must never render; got:\n%s", got)
	}
	if !strings.Contains(got, "CAPPED (row_cap=2)") || !strings.Contains(got, "UNKNOWN, not absent") {
		t.Errorf("a capped ls listing must declare its cap and its alphabetical tail; got:\n%s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query expectations unmet: %v", err)
	}
}

func TestLsArmExactlyAtCapIsSilent(t *testing.T) {
	render := func(rowCap, wantLimit int) string {
		db, mock := newCodeLookupDB(t)
		defer db.Close()
		rows := sqlmock.NewRows(lsCols).
			AddRow("platform/a.go", "abc12345deadbeef", true).
			AddRow("platform/b.go", "abc12345deadbeef", true)
		mock.ExpectQuery("bool_or").
			WithArgs("platform/", "", wantLimit, codeKindsCSV).
			WillReturnRows(rows)
		var b strings.Builder
		if _, err := answerCodeCheck(context.Background(), db,
			codeCheck{Kind: "ls", Query: "platform/"}, "", rowCap, 400, testCodeScope, &b); err != nil {
			t.Fatalf("answerCodeCheck(rowCap=%d): %v", rowCap, err)
		}
		return b.String()
	}
	atCap, uncapped := render(2, 3), render(99, 100)
	if atCap != uncapped {
		t.Errorf("at-cap-exactly must be indistinguishable from plainly-uncapped\n at cap: %q\nuncapped: %q", atCap, uncapped)
	}
	if strings.Contains(atCap, "CAPPED") {
		t.Errorf("two paths at row_cap=2 were NOT truncated: %q", atCap)
	}
}

// ── symbol arm (both branches go through renderSymbolRows) ───────────────────

func TestSymbolArmReportsCap(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	// A bare name binds [name, repoFilter] and the LIMIT lands at $3.
	mock.ExpectQuery(`COALESCE\(signature`).
		WithArgs("Foo", "", 3).
		WillReturnRows(symbolRowsFor(
			[2]string{"a/one.go", "func"},
			[2]string{"b/two.go", "func"},
			[2]string{"c/three.go", "func"}))

	var b strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "symbol", Query: "Foo"}, "", 2, 400, testCodeScope, &b); err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()

	for _, want := range []string{"a/one.go", "b/two.go"} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered rows lost %s; got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "c/three.go") {
		t.Errorf("the probe row must never render; got:\n%s", got)
	}
	if !strings.Contains(got, "CAPPED (row_cap=2)") || !strings.Contains(got, "UNKNOWN, not absent") {
		t.Errorf("a capped symbol answer must declare its cap; got:\n%s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query expectations unmet: %v", err)
	}
}

func TestSymbolArmExactlyAtCapIsSilent(t *testing.T) {
	render := func(rowCap, wantLimit int) string {
		db, mock := newCodeLookupDB(t)
		defer db.Close()
		mock.ExpectQuery(`COALESCE\(signature`).
			WithArgs("Foo", "", wantLimit).
			WillReturnRows(symbolRowsFor(
				[2]string{"a/one.go", "func"},
				[2]string{"b/two.go", "func"}))
		var b strings.Builder
		if _, err := answerCodeCheck(context.Background(), db,
			codeCheck{Kind: "symbol", Query: "Foo"}, "", rowCap, 400, testCodeScope, &b); err != nil {
			t.Fatalf("answerCodeCheck(rowCap=%d): %v", rowCap, err)
		}
		return b.String()
	}
	atCap, uncapped := render(2, 3), render(99, 100)
	if atCap != uncapped {
		t.Errorf("at-cap-exactly must be indistinguishable from plainly-uncapped\n at cap: %q\nuncapped: %q", atCap, uncapped)
	}
	if strings.Contains(atCap, "CAPPED") {
		t.Errorf("two symbols at row_cap=2 were NOT truncated: %q", atCap)
	}
}

// TestSymbolArmElsewhereBranchReportsCap covers the branch a per-call-site fix
// would have missed. A path-qualified miss (bugs_open/163) re-runs the NAME
// alone into a separate `elsewhere` builder; because the notice lives inside
// renderSymbolRows, a capped fallback listing carries its notice INSIDE that
// builder — after the ELSEWHERE header, attached to the rows it qualifies.
func TestSymbolArmElsewhereBranchReportsCap(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	// Primary: path + name + repo, LIMIT at $4. Nothing at that path.
	mock.ExpectQuery(`COALESCE\(signature`).
		WithArgs("internal/x.go", "Foo", "", 3).
		WillReturnRows(sqlmock.NewRows(symbolCols))
	// Fallback: name + repo only, LIMIT at $3. Three matches, cap of two.
	mock.ExpectQuery(`COALESCE\(signature`).
		WithArgs("Foo", "", 3).
		WillReturnRows(symbolRowsFor(
			[2]string{"a/one.go", "func"},
			[2]string{"b/two.go", "func"},
			[2]string{"c/three.go", "func"}))

	var b strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "symbol", Query: "internal/x.go:Foo"}, "", 2, 400, testCodeScope, &b); err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()

	if !strings.Contains(got, "0 rows AT THAT PATH") {
		t.Fatalf("the path-miss disclosure did not render; got:\n%s", got)
	}
	if !strings.Contains(got, "matches 2 indexed symbol(s) ELSEWHERE") {
		t.Errorf("the ELSEWHERE count must be the RENDERED code rows, not the probe count; got:\n%s", got)
	}
	if strings.Contains(got, "c/three.go") {
		t.Errorf("the probe row must never render in the fallback either; got:\n%s", got)
	}
	notice := strings.Index(got, "CAPPED (row_cap=2)")
	elsewhereHdr := strings.Index(got, "ELSEWHERE")
	if notice < 0 {
		t.Fatalf("a capped ELSEWHERE listing must declare its cap; got:\n%s", got)
	}
	if notice < elsewhereHdr {
		t.Errorf("the notice belongs to the ELSEWHERE rows and must follow their header (notice@%d, header@%d); got:\n%s",
			notice, elsewhereHdr, got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query expectations unmet: %v", err)
	}
}

// ── placement against the deferred doc block ─────────────────────────────────

// TestCapNoticeLandsBeforeTheDocBlock pins the interaction that makes the
// wording load-bearing. A check's rows are SPLIT: code-kind rows go to the
// caller's builder, doc-kind rows are buffered and flushed by a defer AFTER the
// arm returns. So the notice sits between the code rows and the doc block, and
// says "this answer" rather than "the rows above" — which stays true whatever
// the split. The cap counts RENDERED rows across both destinations, so a doc row
// spends cap budget exactly as a code row does.
func TestCapNoticeLandsBeforeTheDocBlock(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	mock.ExpectQuery("body ILIKE").
		WithArgs("needle", "", 3).
		WillReturnRows(contentRowsFor("needle",
			[2]string{"a/one.go", "func"},
			[2]string{"docs/guide.md", kindDoc},
			[2]string{"z/three.go", "func"}))

	var b strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "content", Query: "needle"}, "", 2, 400, testCodeScope, &b); err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()

	docHdr := strings.Index(got, docBlockHeader)
	if docHdr < 0 {
		t.Fatalf("the doc-kind row must render under its own header; got:\n%s", got)
	}
	docRow := strings.Index(got, "docs/guide.md")
	if docRow < docHdr {
		t.Errorf("the doc row must sit UNDER the doc header (row@%d, header@%d); got:\n%s", docRow, docHdr, got)
	}
	if strings.Contains(got, "z/three.go") {
		t.Errorf("a doc row spends cap budget too — the third row must be discarded; got:\n%s", got)
	}
	notice := strings.Index(got, "CAPPED (row_cap=2)")
	if notice < 0 {
		t.Fatalf("a capped mixed answer must declare its cap; got:\n%s", got)
	}
	if notice > docHdr {
		t.Errorf("the notice must land with the code rows, BEFORE the deferred doc block (notice@%d, header@%d); got:\n%s",
			notice, docHdr, got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query expectations unmet: %v", err)
	}
}

// ── the pure halves ──────────────────────────────────────────────────────────

// TestRowCapNotice — "" when nothing was capped is what keeps every uncapped
// answer byte-identical, so the silent case is asserted as hard as the loud one.
func TestRowCapNotice(t *testing.T) {
	if got := rowCapNotice(false, 40); got != "" {
		t.Errorf("an uncapped answer must gain nothing, got %q", got)
	}
	got := rowCapNotice(true, 40)
	for _, want := range []string{"CAPPED (row_cap=40)", "UNKNOWN, not absent", "ordered by path", "raise row_cap"} {
		if !strings.Contains(got, want) {
			t.Errorf("notice missing %q; got %q", want, got)
		}
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("the notice is one line and must terminate it: %q", got)
	}
}

// TestProbeLimit — rowCap+1 is the whole mechanism; the <= 0 passthrough keeps
// the pre-existing LIMIT 0 -> emptyAnswer edge exactly as it was.
func TestProbeLimit(t *testing.T) {
	cases := map[int]int{40: 41, 2: 3, 1: 2, 0: 0, -1: -1}
	for in, want := range cases {
		if got := probeLimit(in); got != want {
			t.Errorf("probeLimit(%d) = %d, want %d", in, got, want)
		}
	}
}
