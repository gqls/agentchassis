package actions

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// bugs_open/223 — the code index holds ONE language and 5 of the 8 kinds its own
// CHECK constraint permits, and every answer it gave about the other classes
// asserted "The query was RUN; this is not an unanswered question." Measured on
// the live corpus 2026-08-10: 5,837 rows, 100% `.go`, no `var`/`const`/`type`
// row at all. A verdict built on that answered 0-row input by declaring three of
// this repo's own scripts non-existent — in a note delivered BY those scripts.
//
// Every test here holds down one half of the fix: the census is READ (so the
// wording changes itself when the corpus changes), and the classifier is
// CONSERVATIVE (it must never invent an unanswerable, because a false NOT
// ANSWERABLE would suppress a real absence — the mirror of the bug).
//
// The corpus these tests describe is the live one, so a fixture that stops
// matching production is itself a finding. Two of them are written to FAIL when
// var/const indexing lands (phase 2) — see TestMissingKindNoteDisappears.

// liveShapedScope mirrors the live index on 2026-08-10: Go only, no var/const/type.
func liveShapedScope() codeIndexScope {
	return codeIndexScope{
		total: 5837, withBody: 5837, commits: 1,
		exts:  map[string]int{".go": 5837},
		kinds: map[string]int{"func": 3653, "method": 1119, "struct": 987, "alias": 42, "interface": 36},
	}
}

// widenedScope is the same index AFTER phase 2 (var/const indexed) and after a
// hypothetical doc corpus. It exists so the disappearance of each warning is
// asserted now, in advance, rather than discovered as stale prose later.
func widenedScope() codeIndexScope {
	return codeIndexScope{
		total: 7010, withBody: 7010, commits: 1,
		exts: map[string]int{".go": 5837, ".md": 1173},
		kinds: map[string]int{"func": 3653, "method": 1119, "struct": 987, "alias": 42,
			"interface": 36, "type": 10, "var": 900, "const": 273},
	}
}

func TestPathExt(t *testing.T) {
	cases := map[string]string{
		"scripts/landmines-sync.py":                ".py",
		"platform/orchestration/actions/foo.go":    ".go",
		"docs/agent_docs/sql_for_agents/365_x.sql": ".sql",
		"scripts/":  "",
		"scripts":   "",
		"":          "",
		"doc_notes": "",
		// A dot in a DIRECTORY name is not a file extension. Reading it as one
		// would classify a whole tree as unanswerable and suppress real absences.
		"docs/v1.2/notes": "",
		// A trailing dot names no extension.
		"weird.": "",
		// A dotfile is not an extension either: ".gitignore" has no name before
		// the dot, so treating it as ext ".gitignore" would be a category error.
		".gitignore": "",
	}
	for in, want := range cases {
		if got := pathExt(in); got != want {
			t.Errorf("pathExt(%q) = %q, want %q", in, got, want)
		}
	}
}

// The classifier's positive case: the two motivating footprints from the bug.
func TestUnanswerableReasonClassifiesNonGoPaths(t *testing.T) {
	s := liveShapedScope()
	for _, q := range []string{"scripts/landmines-sync.py", "scripts/landmines-verify-dispatch.sh"} {
		reason := unanswerableReason("ls", symbolQuery{}, q, s)
		if reason == "" {
			t.Fatalf("unanswerableReason(ls, %q) = \"\" — a .py/.sh footprint is exactly what this must catch", q)
		}
		if !strings.Contains(reason, ".go (5837 rows)") {
			t.Errorf("reason must cite the census it derives from, got %q", reason)
		}
	}
	// The symbol arm classifies on the PATH half, which is why it takes a
	// symbolQuery: "scripts/x.py:main" is unanswerable for the same reason as
	// "ls scripts/x.py", and 163's parser is what separates the halves.
	sq := parseSymbolQuery("scripts/landmines_lib.py:slugify")
	if reason := unanswerableReason("symbol", sq, sq.raw, s); reason == "" {
		t.Errorf("a path-bearing symbol check on a .py file must classify as unanswerable")
	}
}

// The classifier's NEGATIVE cases, which matter more: every one of these must
// keep the pre-223 wording, or the fix suppresses genuine absences.
func TestUnanswerableReasonIsConservative(t *testing.T) {
	s := liveShapedScope()
	cases := []struct {
		kind, query, why string
	}{
		{"ls", "platform/orchestration/actions/", "a directory prefix names no extension — the index may hold it"},
		{"ls", "platform/orchestration/actions/foo.go", "a .go path IS representable; 0 rows there is a real absence"},
		{"content", "stop_reason", "free text is not a path and must never be classified"},
		{"content", "doc_notes", "a table name is not a path"},
		{"symbol", "GenerateText", "a bare identifier carries no path half"},
	}
	for _, c := range cases {
		sq := parseSymbolQuery(c.query)
		if reason := unanswerableReason(c.kind, sq, c.query, s); reason != "" {
			t.Errorf("unanswerableReason(%s, %q) = %q — must be empty: %s", c.kind, c.query, reason, c.why)
		}
	}
	// And with NO census (the read failed, or an old row set): claim nothing.
	blind := codeIndexScope{total: 5837, withBody: 5837, commits: 1}
	if reason := unanswerableReason("ls", symbolQuery{}, "scripts/x.py", blind); reason != "" {
		t.Errorf("with no census the classifier must fail OPEN, got %q", reason)
	}
	// A widened corpus must stop classifying what it now holds.
	if reason := unanswerableReason("ls", symbolQuery{}, "docs/notes.md", widenedScope()); reason != "" {
		t.Errorf("a .md path must be answerable once .md rows exist, got %q", reason)
	}
}

// The wording is the guard, because the reader is a model. All three readings the
// bug recorded — removed, renamed, "does not exist" — must be blocked by name.
func TestNotAnswerableAnswerBlocksEveryWrongReading(t *testing.T) {
	got := notAnswerableAnswer("the corpus holds NO .py file at all — the indexed corpus holds only: .go (5837 rows)")
	for _, want := range []string{
		"NOT ANSWERABLE BY THIS INDEX",
		"COULD NOT have returned a row",
		"UNKNOWN",
		"NOT evidence that the target is absent, removed, renamed or inlined",
		"must not contribute to a verdict of STALE",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("notAnswerableAnswer is missing %q\ngot: %s", want, got)
		}
	}
	// It must NOT carry the sentence it replaces, or both readings ship together.
	if strings.Contains(got, "this is not an unanswered question") {
		t.Error("notAnswerableAnswer must not also assert the query was answered")
	}
}

// The third failure mode: a live package-level `var` reported as "possibly inlined
// or renamed". The census makes the absent KIND a stated fact, so the hypothesis
// has nothing to grow in.
func TestEmptyAnswerNamesMissingKinds(t *testing.T) {
	got := liveShapedScope().emptyAnswer("symbol")
	for _, want := range []string{
		"NO declarations of kind",
		"var",
		"const",
		"UNREPRESENTABLE",
		"NEITHER removal NOR a rename-or-inline hypothesis",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("emptyAnswer(symbol) is missing %q\ngot: %s", want, got)
		}
	}
	// The claim must be readable against its own evidence in the same breath.
	if !strings.Contains(got, "kinds present: alias, func, interface, method, struct") {
		t.Errorf("the note must list the kinds that ARE present, got: %s", got)
	}
}

// PRE-REGISTERED PHASE-2 PROOF. When var/const indexing lands, this warning must
// vanish BY ITSELF. A mock's own bookkeeping cannot assert that negative; a
// census-driven renderer can, and this is the assertion that will catch stale
// prose if someone later hardcodes the sentence.
func TestMissingKindNoteDisappears(t *testing.T) {
	if note := widenedScope().missingKindNote(); note != "" {
		t.Errorf("a corpus holding every code kind must emit no missing-kind note, got: %s", note)
	}
	if note := liveShapedScope().missingKindNote(); note == "" {
		t.Fatal("the live-shaped corpus is missing var/const/type — the note must fire, or this test proves nothing")
	}
	// And with no census at all: silence must not imply completeness, but it also
	// must not invent a claim.
	if note := (codeIndexScope{total: 1}).missingKindNote(); note != "" {
		t.Errorf("no census ⇒ no claim, got: %s", note)
	}
}

// The FALSE POSITIVE half, measured 2026-08-10: `ls scripts/` returns 110 indexed
// Go paths while every .py and .sh directly under scripts/ is invisible, so a
// generous listing reads as confirmation. This is the one caveat that rides on a
// NON-empty answer.
func TestLsReachNoteSaysWhatTheListingIs(t *testing.T) {
	got := liveShapedScope().lsReachNote()
	for _, want := range []string{"lists INDEXED paths only", "UNKNOWN, not evidence it is gone", "not a directory listing"} {
		if !strings.Contains(got, want) {
			t.Errorf("lsReachNote is missing %q\ngot: %s", want, got)
		}
	}
	// Multi-language corpus: the sentence stops being the explanation, so it goes.
	if note := widenedScope().lsReachNote(); note != "" {
		t.Errorf("a multi-extension corpus must not claim a single-language listing, got: %s", note)
	}
	if note := (codeIndexScope{total: 1}).lsReachNote(); note != "" {
		t.Errorf("no census ⇒ no claim, got: %s", note)
	}
}

func TestContentReachNoteNamesTheUnreachableClasses(t *testing.T) {
	got := liveShapedScope().emptyAnswer("content")
	// The pre-existing sentence must SURVIVE — the fix qualifies it, it does not
	// replace it, because for an in-corpus query it is true and useful.
	if !strings.Contains(got, "The query was RUN and found nothing") {
		t.Errorf("the honest-absence wording must remain for content, got: %s", got)
	}
	for _, want := range []string{"a script", "a database table", "a config value", "UNKNOWN, not absent"} {
		if !strings.Contains(got, want) {
			t.Errorf("content empty answer is missing %q\ngot: %s", want, got)
		}
	}
	if strings.Contains(widenedScope().emptyAnswer("content"), "CANNOT match here") {
		t.Error("a multi-extension corpus must not claim a footprint cannot match")
	}
}

// The mechanical census a consumer can persist beside a model's verdict.
func TestCodeEvidenceLine(t *testing.T) {
	s := liveShapedScope()

	// The motivating run: 8 checks, 3 matched code, 2 classified unanswerable,
	// 3 ran and found nothing.
	got := codeEvidenceLine(8, 3, 2, s)
	for _, want := range []string{
		"8 check(s) ran", "3 matched indexed code", "2 NOT ANSWERABLE", "3 ran and matched nothing",
		"5837 symbols", ".go (5837 rows)", "kinds with NO rows: const, type, var",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("evidence line missing %q\ngot: %s", want, got)
		}
	}
	// A round that confirmed NOTHING must say so in words, not leave it to
	// arithmetic the reader may not do.
	zero := codeEvidenceLine(5, 0, 5, s)
	if !strings.Contains(zero, "NOTHING in this round was confirmed") {
		t.Errorf("a zero-evidence round must state it, got: %s", zero)
	}
	if strings.Contains(got, "NOTHING in this round was confirmed") {
		t.Error("a round WITH confirmations must not claim it confirmed nothing")
	}
	// Scope failure must read as unknown, never as a clean census.
	if !strings.Contains(codeEvidenceLine(1, 0, 0, codeIndexScope{err: errScopeTest}), "Scope: UNKNOWN") {
		t.Error("a failed scope read must render as UNKNOWN")
	}
}

// errScopeTest stands in for a scope-read failure; the value is never inspected.
var errScopeTest = errors.New("scope read failed")

// ── WIRING, not just the helpers ─────────────────────────────────────────────
//
// WRITTEN BECAUSE A MUTATION SURVIVED. Replacing the `ls` arm's
// `b.WriteString(scope.lsReachNote())` with `b.WriteString("")` left every test
// above passing: they exercised the FUNCTIONS and nothing asserted the call
// sites. A helper with no caller looks exactly like a finished fix, and the
// false-positive half of bugs_open/223 — a generous `ls` listing that reads as
// confirmation — was the half left unprotected. These tests go through
// answerCodeCheck, with a census populated, so unwiring any of the three
// renderings fails a test.
func TestLsArmWiresTheReachNoteOnANonEmptyListing(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	// The measured case: a prefix that DOES resolve, to Go files, while the .py
	// and .sh files the reviewer asked about are structurally invisible.
	mock.ExpectQuery("bool_or").
		WithArgs("scripts/", "", 41, codeKindsCSV).
		WillReturnRows(sqlmock.NewRows(lsCols).
			AddRow("scripts/documentation_project/01/analyser.go", "abc12345deadbeef", true).
			AddRow("scripts/goscripts/createhashedpassword.go", "abc12345deadbeef", true))

	var b strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "ls", Query: "scripts/"}, "", 40, 400, liveShapedScope(), &b); err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()
	if !strings.Contains(got, "analyser.go") {
		t.Fatalf("the listing itself must survive; got:\n%s", got)
	}
	if !strings.Contains(got, "lists INDEXED paths only") || !strings.Contains(got, "not a directory listing") {
		t.Errorf("a non-empty ls answer must say what it is a listing OF — this is the false-positive guard; got:\n%s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The unanswerable rendering, through the arm, with the outcome flag asserted —
// the flag is what a workflow branches on, so it is as load-bearing as the prose.
func TestLsArmRendersNotAnswerableAndFlagsIt(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	mock.ExpectQuery("bool_or").
		WithArgs("scripts/landmines-sync.py", "", 41, codeKindsCSV).
		WillReturnRows(sqlmock.NewRows(lsCols))

	var b strings.Builder
	outcome, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "ls", Query: "scripts/landmines-sync.py"}, "", 40, 400, liveShapedScope(), &b)
	if err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()
	if !strings.Contains(got, "NOT ANSWERABLE BY THIS INDEX") {
		t.Errorf("a .py path with 0 rows must not render as an answered question; got:\n%s", got)
	}
	if strings.Contains(got, "The query was RUN; this is not an unanswered question") {
		t.Errorf("the pre-223 wording must be REPLACED here, not accompanied; got:\n%s", got)
	}
	if !outcome.unanswerable {
		t.Error("the outcome must flag unanswerable — the branch a consumer takes depends on this, not on the prose")
	}
	if outcome.codeRows != 0 {
		t.Errorf("no rows matched, so codeRows must be 0, got %d", outcome.codeRows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A REAL absence must keep reading as a real absence: same arm, same census, a
// .go path. This is the test that fails if the fix ever starts buying abstention
// by classifying more than it should.
func TestLsArmKeepsHonestAbsenceForAnIndexedExtension(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	mock.ExpectQuery("bool_or").
		WithArgs("platform/orchestration/actions/deleted_thing.go", "", 41, codeKindsCSV).
		WillReturnRows(sqlmock.NewRows(lsCols))

	var b strings.Builder
	outcome, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "ls", Query: "platform/orchestration/actions/deleted_thing.go"}, "", 40, 400, liveShapedScope(), &b)
	if err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()
	if strings.Contains(got, "NOT ANSWERABLE") {
		t.Errorf("a .go path IS representable — 0 rows there is a finding, not an unanswerable question; got:\n%s", got)
	}
	if !strings.Contains(got, "The query was RUN; this is not an unanswered question") {
		t.Errorf("the honest-absence wording must survive for in-corpus classes; got:\n%s", got)
	}
	if outcome.unanswerable {
		t.Error("an in-corpus miss must NOT be flagged unanswerable, or a real absence is suppressed")
	}
}

// The symbol arm's own wiring: the var/const census note must reach the answer,
// and a code-tier match must be counted as evidence.
func TestSymbolArmWiresTheMissingKindNoteAndCountsEvidence(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	// A bare name that matches nothing: the arm falls through to emptyAnswer,
	// which must carry the missing-kind census.
	mock.ExpectQuery("FROM code_symbols").WillReturnRows(sqlmock.NewRows(symbolCols))
	var b strings.Builder
	outcome, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "symbol", Query: "metaCommentaryPatterns"}, "", 40, 400, liveShapedScope(), &b)
	if err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	if !strings.Contains(b.String(), "UNREPRESENTABLE") {
		t.Errorf("the symbol arm must carry the missing-kind census — this is the 'possibly inlined or renamed' guard; got:\n%s", b.String())
	}
	if outcome.codeRows != 0 || outcome.unanswerable {
		t.Errorf("a bare-name miss is neither evidence nor unanswerable, got %+v", outcome)
	}

	// And a match: codeRows must count it, because no_code_evidence is derived
	// from it and an uncounted success reads as a round that proved nothing.
	mock.ExpectQuery("FROM code_symbols").WillReturnRows(symbolRowsFor([2]string{"a/one.go", "func"}))
	var b2 strings.Builder
	outcome2, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "symbol", Query: "Foo"}, "", 40, 400, liveShapedScope(), &b2)
	if err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	if outcome2.codeRows != 1 {
		t.Errorf("a code-tier match must be counted, got %d", outcome2.codeRows)
	}
}

// A [doc] row is NOT evidence: the D12 guard's whole point is that a document
// says where only code shows, and a field named "did this find evidence" must
// honour the same line or it launders prose into proof one layer up.
func TestDocRowsAreNotCountedAsEvidence(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	mock.ExpectQuery("FROM code_symbols").
		WillReturnRows(symbolRowsFor([2]string{"docs/guide.md", kindDoc}))
	var b strings.Builder
	outcome, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "symbol", Query: "Foo"}, "", 40, 400, liveShapedScope(), &b)
	if err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	if outcome.codeRows != 0 {
		t.Errorf("a [doc] row must not count as code evidence, got codeRows=%d", outcome.codeRows)
	}
}

// The content arm's half of the same guard, kept separate because the two arms
// count evidence through different code (renderSymbolRows splits code from prose
// for the symbol arm; the content arm does it inline).
func TestContentArmDoesNotCountDocRowsAsEvidence(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	mock.ExpectQuery("body ILIKE").
		WithArgs("needle", "", 41).
		WillReturnRows(contentRowsFor("needle", [2]string{"docs/guide.md", kindDoc}))

	var b strings.Builder
	outcome, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "content", Query: "needle"}, "", 40, 400, liveShapedScope(), &b)
	if err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	if outcome.codeRows != 0 {
		t.Errorf("a [doc] content match must not count as code evidence, got codeRows=%d", outcome.codeRows)
	}
}

// ── objections raised by council 495df717, closed here ───────────────────────

// editquality (MEDIUM, edit 2): the diagnosis names THREE false-positive modes and
// the first version guarded two. The third is a `content` check aimed at a non-Go
// file being answered by a same-named GO symbol — measured: `content: slugify`,
// stated purpose "confirms the slugify function exists in landmines_lib.py",
// returned six confident hits on slugifyPathSegments/slugifyForCompositionName. A
// false positive WITH CITATIONS. The query is free text so nothing can know which
// file was meant; what is knowable is where every match came FROM.
func TestContentArmSaysWhereItsMatchesCameFrom(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()

	mock.ExpectQuery("body ILIKE").
		WithArgs("slugify", "", 41).
		WillReturnRows(contentRowsFor("slugify",
			[2]string{"platform/orchestration/actions/adopt_verbatim.go", "func"},
			[2]string{"platform/orchestration/actions/resolve_composition_helpers.go", "func"}))

	var b strings.Builder
	outcome, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "content", Query: "slugify", Why: "confirms the slugify function exists in landmines_lib.py"},
		"", 40, 400, liveShapedScope(), &b)
	if err != nil {
		t.Fatalf("answerCodeCheck: %v", err)
	}
	got := b.String()
	if !strings.Contains(got, "adopt_verbatim.go") {
		t.Fatalf("the matches themselves must survive; got:\n%s", got)
	}
	for _, want := range []string{"every match above comes from a .go file", "SHARES A NAME", "UNSEARCHABLE here"} {
		if !strings.Contains(got, want) {
			t.Errorf("a non-empty content answer must say where its matches came from — missing %q; got:\n%s", want, got)
		}
	}
	// It is still evidence OF SOMETHING (two real Go rows), so the count must not
	// be suppressed: this note qualifies a match, it does not retract it.
	if outcome.codeRows != 2 {
		t.Errorf("the caveat must not reduce the evidence count, got %d", outcome.codeRows)
	}
	// Multi-language corpus: the sentence stops being the explanation.
	if strings.Contains(widenedScope().contentMatchReachNote(), "every match above comes from") {
		t.Error("a multi-extension corpus must not claim every match is one file type")
	}
}

// bug_historian (MEDIUM, edit 6): "gate_evidence depends on runtime resolution of
// `lookup.no_code_evidence` across a step boundary — the fix could ship, look
// wired, and never actually gate a single verdict, with no error surfaced
// anywhere." That is the strongest objection in the round and it is right: a
// resolution failure degrades to else_step, which is today's behaviour, silently
// and for ever.
//
// It is also closable BEFORE the roll, because the condition evaluator lives in
// this package. This test runs the seed's EXACT condition string against the
// EXACT shape DiagnoseCodeLookupAction returns — verified against a live run's
// stored state, which nests the action's map directly under the step's
// output_field with no "result" wrapper (unlike execute_llm_prompt, whose results
// ARE wrapped, which is why the sibling steps read `verdict.result.body`).
//
// If someone renames the key, or the evaluator stops resolving a two-segment
// dotted path, or a future action wraps its return, this fails here rather than
// silently never gating in production.
func TestSeedConditionResolvesAgainstTheActionsReturnShape(t *testing.T) {
	const seedCondition = "lookup.no_code_evidence == true" // byte-identical to seed 365

	// LOCKSTEP, and this line is why the test can be trusted. A first version built
	// its own literal map and therefore proved only that the EVALUATOR resolves a
	// dotted path — renaming the key in the action left it green, which is the same
	// helper-versus-wiring hole that let a mutation survive earlier in this lane.
	// Pinning the seed's string to the action's own constant closes it: rename the
	// key and this fails here, instead of the gate silently never firing in
	// production.
	if want := "lookup." + codeEvidenceGateField + " == true"; seedCondition != want {
		t.Fatalf("seed 365's condition %q no longer matches the action's key %q — the config and the code have drifted, and the symptom would be a gate that never fires",
			seedCondition, want)
	}

	// The action's own return map, keys spelled as the action spells them.
	lookupReturn := func(noCodeEvidence bool) map[string]interface{} {
		return map[string]interface{}{
			"lookup": map[string]interface{}{
				"results_text":        "…",
				"checks_run":          8,
				"checks_dropped":      0,
				"checks_with_rows":    0,
				"checks_unanswerable": 5,
				codeEvidenceGateField: noCodeEvidence,
				"evidence_line":       "[code-lookup evidence: …]",
			},
		}
	}

	met, err := evaluateStringCondition(seedCondition, lookupReturn(true), zap.NewNop())
	if err != nil {
		t.Fatalf("the seed's condition must evaluate without error: %v", err)
	}
	if !met {
		t.Fatal("no_code_evidence=true must take the then_step (verify_unverifiable) — otherwise the gate ships wired and never fires")
	}

	met, err = evaluateStringCondition(seedCondition, lookupReturn(false), zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error on the false branch: %v", err)
	}
	if met {
		t.Fatal("no_code_evidence=false must take the else_step (verify) — a gate that always fires is as broken as one that never does")
	}

	// The pre-roll degradation, asserted rather than assumed: on a binary that does
	// not yet return the key, the condition must resolve to false with NO error, so
	// the pipeline behaves exactly as it does today. This is the property that made
	// applying the seed ahead of the roll safe, and it was observed live
	// (evidence_gate recorded {"condition_met": false, "next_step_override":
	// "verify"} on v1.0.1277) — this test is what keeps it true.
	old := map[string]interface{}{"lookup": map[string]interface{}{
		"results_text": "…", "checks_run": 8, "checks_dropped": 0,
	}}
	met, err = evaluateStringCondition(seedCondition, old, zap.NewNop())
	if err != nil {
		t.Fatalf("an absent field must not error — a hard failure here would break every verification on an older binary: %v", err)
	}
	if met {
		t.Fatal("an absent field must resolve FALSE, so an old binary keeps today's route")
	}
}

// ── the staleness half: the indexed commit travels WITH the empty answer ─────
//
// 090 round 520b2f7e (2026-08-11): the freshness BANNER named the indexed commit
// and said "local unpushed work is never visible" in the header of every verify
// prompt of the motivating incident — measured in llm_call_log.prompt_rendered —
// and the verdict still explained an unpushed symbol's absence as "of kinds not
// indexed", quoting the kind census rendered beside the empty answer while
// talking past the header. These tests hold down the remedy: the same fact,
// restated where the explanation is formed, and carried into the persisted
// evidence line so a doc_notes verdict can be dated against the code.

// testFreshness mirrors the motivating incident: index at the pushed tip of
// 2026-08-10 16:27 UTC while the missing symbol was committed 23:13 the same
// day — which is why the note carries minute precision, not a date.
func testFreshness() indexFreshness {
	return indexFreshness{
		sha: "5a68d6caf00d", ref: "087_towards_multiple_domains",
		commitTime: time.Date(2026, 8, 10, 16, 27, 0, 0, time.UTC),
		updatedAt:  time.Date(2026, 8, 11, 9, 0, 0, 0, time.UTC),
	}
}

func datedScope() codeIndexScope {
	s := liveShapedScope()
	s.indexed = testFreshness()
	return s
}

func TestIndexedAsOfNote(t *testing.T) {
	got := datedScope().indexedAsOfNote()
	for _, want := range []string{
		"5a68d6ca", // shortSHA
		"087_towards_multiple_domains",
		"2026-08-10 16:27 UTC", // absolute, because an age is only true at render time
		"PUSHED tip",
		"INDEX STALENESS",
		"not absence, not removal, not a rename", // the three wrong readings, blocked by name
	} {
		if !strings.Contains(got, want) {
			t.Errorf("as-of note missing %q\ngot: %s", want, got)
		}
	}
	// Unreadable freshness or an empty index: the banner branches already shout
	// those states, and an as-of note naming a commit nobody read would be an
	// invented fact — the exact class this file exists to remove.
	if got := (codeIndexScope{indexed: indexFreshness{err: errScopeTest}}).indexedAsOfNote(); got != "" {
		t.Errorf("a failed freshness read must yield no as-of note, got %q", got)
	}
	if got := (codeIndexScope{}).indexedAsOfNote(); got != "" {
		t.Errorf("an empty index must yield no as-of note, got %q", got)
	}
	// A pre-migration row set (rows written, commit unrecorded) must be dated as
	// UNDATABLE — silence there would read as "nothing to say".
	undated := codeIndexScope{indexed: indexFreshness{updatedAt: time.Date(2026, 8, 11, 9, 0, 0, 0, time.UTC)}}
	if got := undated.indexedAsOfNote(); !strings.Contains(got, "UNRECORDED") {
		t.Errorf("rows with no commit_sha must render as undatable, got %q", got)
	}
}

// Through answerCodeCheck, one arm per empty-answer branch, because a mutation
// already survived the helper-only version of this file once (the lsReachNote
// unwiring, WRONG_CALLS 2026-08-10): unwiring indexedAsOfNote from any arm's
// empty answer must fail HERE, not in production prose.
func TestEmptyAnswersCarryTheIndexedCommit(t *testing.T) {
	db, mock := newCodeLookupDB(t)
	defer db.Close()
	scope := datedScope()

	mock.ExpectQuery("FROM code_symbols").WillReturnRows(sqlmock.NewRows(symbolCols))
	var symbolOut strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "symbol", Query: "metaCommentaryPatterns"}, "", 40, 400, scope, &symbolOut); err != nil {
		t.Fatalf("symbol arm: %v", err)
	}

	// The ls miss on an INDEXED extension — the honest-absence branch, which is
	// exactly where the misverdict formed: representable, searched, 0 rows.
	mock.ExpectQuery("bool_or").
		WithArgs("platform/orchestration/actions/gone.go", "", 41, codeKindsCSV).
		WillReturnRows(sqlmock.NewRows(lsCols))
	var lsOut strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "ls", Query: "platform/orchestration/actions/gone.go"}, "", 40, 400, scope, &lsOut); err != nil {
		t.Fatalf("ls arm: %v", err)
	}

	mock.ExpectQuery("body ILIKE").
		WithArgs("no-such-needle", "", 41).
		WillReturnRows(sqlmock.NewRows(contentCols))
	var contentOut strings.Builder
	if _, err := answerCodeCheck(context.Background(), db,
		codeCheck{Kind: "content", Query: "no-such-needle"}, "", 40, 400, scope, &contentOut); err != nil {
		t.Fatalf("content arm: %v", err)
	}

	for arm, got := range map[string]string{
		"symbol": symbolOut.String(), "ls": lsOut.String(), "content": contentOut.String(),
	} {
		if !strings.Contains(got, "as-of: this answer describes commit 5a68d6ca") ||
			!strings.Contains(got, "INDEX STALENESS") {
			t.Errorf("the %s arm's empty answer must carry the indexed commit; got:\n%s", arm, got)
		}
	}
}

// The persisted half: codeEvidenceLine is what append_doc_note suffixes onto the
// verdict row, and before this change nothing the run persisted recorded which
// commit the answers described — a verdict read months later could not be dated
// against the code at all.
func TestCodeEvidenceLineNamesTheIndexedCommit(t *testing.T) {
	got := codeEvidenceLine(8, 3, 2, datedScope())
	for _, want := range []string{
		"Answers describe indexed commit 5a68d6ca",
		"087_towards_multiple_domains",
		"2026-08-10 16:27 UTC",
		"not the present tree",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("evidence line missing %q\ngot: %s", want, got)
		}
	}
	// Without a readable freshness row the clause must vanish, not invent: the
	// line keeps its census and loses only the dating.
	bare := codeEvidenceLine(8, 3, 2, liveShapedScope())
	if strings.Contains(bare, "indexed commit") {
		t.Errorf("no freshness read ⇒ no commit claim, got: %s", bare)
	}
	if !strings.Contains(bare, "5837 symbols") {
		t.Errorf("the census must survive the missing clause, got: %s", bare)
	}
}
