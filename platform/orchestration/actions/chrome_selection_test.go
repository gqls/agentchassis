package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// ── bugs_open/118 ────────────────────────────────────────────────────────────
//
// "Which library component serves chrome function F?" had three answers, and all
// three were wrong differently (measured live 2026-07-31):
//
//	render_site_components  no predicate       -> footer-4-column   (is_active=false)
//	link_site_components    is_active only     -> header-leopardess (a client's FORK)
//	GetComponentByFunction  active + unforked  -> site-header       (component_level='section')
//
// The predicate is now one string. These tests pin each of the three clauses to
// the defect it answers, so that dropping one goes red with a reason rather than
// quietly restoring the bug that clause was added for.

func TestChromeEligiblePredicateCarriesAllThreeClauses(t *testing.T) {
	got := chromeEligibleSQL("")

	// is_active — the filed defect. Without it the winner is whatever sorts
	// first, deactivated or not.
	if !strings.Contains(got, "is_active") {
		t.Errorf("chrome predicate %q has no is_active clause — that IS bugs_open/118", got)
	}
	// forked_from IS NULL — a fork carries its parent's `function`, so an ACTIVE
	// fork of one site's header is a candidate to become every other site's.
	if !strings.Contains(got, "forked_from IS NULL") {
		t.Errorf("chrome predicate %q does not exclude forks — header-leopardess would win the site-header pool", got)
	}
	// component_level — without it the winner for site-header is `site-header`
	// itself, a component_level='section' page-section component.
	if !strings.Contains(got, "component_level IN (") {
		t.Errorf("chrome predicate %q has no level filter — a section component would serve as chrome", got)
	}

	// A whitelist, not `<> 'section'`: a body level added later must not become
	// eligible chrome by default.
	if strings.Contains(got, "<>") || strings.Contains(got, "!=") {
		t.Errorf("chrome level filter must be a whitelist, got %q", got)
	}
	for _, lvl := range []string{"'site'", "'header'", "'footer'", "'head'"} {
		if !strings.Contains(got, lvl) {
			t.Errorf("chrome level whitelist is missing %s — the estate uses all four (12/5/1/1 rows live)", lvl)
		}
	}
	for _, body := range []string{"'section'", "'tool'", "'element'"} {
		if strings.Contains(got, body) {
			t.Errorf("chrome level whitelist must not admit %s — that is a page-body level", body)
		}
	}

	// The predicate must apply the level filter as an AND on the SAME row as the
	// function match, never as an OR or a standalone clause: the council's
	// editquality seat objected that 'header'/'footer' were whitelisted
	// unverified, and the answer (measured 2026-07-31) is that they admit ZERO
	// rows because `function = $1` is applied first — all six rows at those
	// levels carry their own non-chrome functions (header-docs,
	// header-minimal-tool, footer-with-disclaimer, …). That answer only holds
	// while the two are conjunctive.
	if strings.Contains(strings.ToUpper(got), " OR ") {
		t.Errorf("chrome predicate must be a conjunction, got %q — an OR would let a level admit a row the function filter excluded", got)
	}
}

func TestChromeEligibleSQLQualifiesEveryColumn(t *testing.T) {
	// A half-qualified predicate is worse than an unqualified one: it compiles,
	// it runs, and it silently reads the wrong table in a join.
	got := chromeEligibleSQL("cc.")
	for _, col := range []string{"is_active", "forked_from", "component_level"} {
		if !strings.Contains(got, "cc."+col) {
			t.Errorf("alias form did not qualify %s: %q", col, got)
		}
	}
	if strings.Contains(strings.ReplaceAll(got, "cc.", ""), "cc.") {
		t.Errorf("alias applied more than once: %q", got)
	}
}

func TestChromeSlotFunctionIsTheOneMapping(t *testing.T) {
	for slot, want := range map[string]string{
		"header": "site-header",
		"footer": "site-footer",
		"head":   "head",
	} {
		if got := ChromeSlotFunction(slot); got != want {
			t.Errorf("ChromeSlotFunction(%q) = %q, want %q", slot, got, want)
		}
	}
	// Pass-through preserves the old inline map's behaviour for anything else,
	// which is what makes this safe to route every caller through.
	if got := ChromeSlotFunction("webdesign-couk-header"); got != "webdesign-couk-header" {
		t.Errorf("an unmapped slot must pass through unchanged, got %q", got)
	}
}

// ── the resolver ─────────────────────────────────────────────────────────────

func chromeRowCols() []string {
	return []string{"id", "name", "function", "category", "html_template", "input_schema", "is_dark_section", "chrome_eligible"}
}

func TestResolveChromeComponentReportsEligible(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// WithArgs pins the FUNCTION, not merely that a query happened: a resolver
	// that asked about the slot name instead would find nothing for every site.
	mock.ExpectQuery("FROM content_components").
		WithArgs("site-footer").
		WillReturnRows(sqlmock.NewRows(chromeRowCols()).
			AddRow("11111111-1111-1111-1111-111111111111", "footer-theme-chrome", "site-footer",
				"", "<footer></footer>", []byte(`{"logo_text":{}}`), false, true))

	comp, eligible, err := ResolveChromeComponent(context.Background(), db, "site-footer", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if !eligible {
		t.Error("a row the query reports as chrome_eligible must come back eligible")
	}
	if comp.Name != "footer-theme-chrome" {
		t.Errorf("component = %q", comp.Name)
	}
	if _, ok := comp.InputSchema["logo_text"]; !ok {
		t.Error("input_schema must be decoded — render_site_components fills a component's own field vocabulary from it (bugs_open/018)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected queries: %v", err)
	}
}

// The `head` case as it stands live: both candidates are is_active=false, so the
// library has nothing eligible. The resolver must still answer — refusing would
// take the head slot away from new sites and injectBrandHeadTags (favicon,
// og-card) with it — but it must say so.
func TestResolveChromeComponentReportsAnIneligibleFallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM content_components").
		WithArgs("head").
		WillReturnRows(sqlmock.NewRows(chromeRowCols()).
			AddRow("22222222-2222-2222-2222-222222222222", "Document Head", "head",
				"", "<head></head>", []byte(`{}`), false, false))

	comp, eligible, err := ResolveChromeComponent(context.Background(), db, "head", zap.NewNop())
	if err != nil {
		t.Fatalf("an ineligible library must not be an error — the caller needs the row: %v", err)
	}
	if eligible {
		t.Fatal("chrome_eligible=false must surface as eligible=false; reporting it as fine is the whole of bugs_open/118")
	}
	if comp == nil || comp.Name != "Document Head" {
		t.Fatalf("the last-resort row must still be returned, got %#v", comp)
	}
}

func TestResolveChromeComponentNoRowIsAnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM content_components").
		WithArgs("site-header").
		WillReturnRows(sqlmock.NewRows(chromeRowCols()))

	if _, _, err := ResolveChromeComponent(context.Background(), db, "site-header", zap.NewNop()); err == nil {
		t.Fatal("no row at all is a different thing from an ineligible row and must be an error")
	}
}

// The ordering is what makes the answer reproducible. Two sites resolving the
// same slot on the same library must get the same component, and the eligible
// rows must sort ahead of the ineligible ones or the flag would be decorative.
func TestResolveChromeComponentOrdersEligibleFirstThenByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// sqlmock's default matcher is a regexp over the SQL actually sent, so this
	// expectation asserts the shipped query text, not a copy of it. The eligible
	// flag sorts first; `name` breaks the tie. Without the second, `LIMIT 1` over
	// a two-member pool returns whatever the plan felt like — which is what
	// GetComponentByFunction did for both chrome functions until this bug.
	mock.ExpectQuery(`ORDER BY \(.+\) DESC, name`).
		WithArgs("site-header").
		WillReturnRows(sqlmock.NewRows(chromeRowCols()).
			AddRow("33333333-3333-3333-3333-333333333333", "header-theme-chrome", "site-header",
				"", "<header></header>", []byte(`{}`), false, true))

	if _, _, err := ResolveChromeComponent(context.Background(), db, "site-header", zap.NewNop()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the shipped query does not carry `ORDER BY (<eligible>) DESC, name`: %v", err)
	}
}

// The sibling lookup must be deterministic for the same reason, and it is the one
// with a live two-member pool for BOTH chrome functions.
func TestGetComponentByFunctionIsOrdered(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`is_active = true AND forked_from IS NULL\s+ORDER BY name`).
		WithArgs("site-footer").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "function", "category", "html_template", "input_schema", "is_dark_section"}).
			AddRow("44444444-4444-4444-4444-444444444444", "site-footer", "site-footer", "", "<footer></footer>", []byte(`{}`), false))

	if _, err := GetComponentByFunction(context.Background(), db, "site-footer", zap.NewNop()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("GetComponentByFunction lost its ORDER BY: %v", err)
	}
}

// ── the lockstep door: no fourth copy of the predicate ───────────────────────
//
// This is the test that stops 118 recurring. The three call sites did not
// disagree because anyone chose differently; they disagreed because each was
// written separately and nothing connected them. Same construction as
// TestNoWriterHandTypesTheAssetLockPredicate (LOCK-007).

// chromeFunctionLiteral matches a SQL comparison against a chrome function name,
// which is how all three of the old hand-typed lookups began.
var chromeFunctionLiteral = regexp.MustCompile(`function\s*=\s*'(site-header|site-footer|head)'`)

func TestNoChromeSelectionHandTypesItsOwnLookup(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}

	var offenders []string
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") || name == "component_library.go" {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		scanned++
		for i, line := range strings.Split(string(src), "\n") {
			// Comments quote the old queries on purpose — they are the record of
			// what was wrong, not a second source of truth.
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if chromeFunctionLiteral.MatchString(line) {
				offenders = append(offenders, fmt.Sprintf("%s:%d", name, i+1))
			}
		}
	}

	if scanned == 0 {
		t.Fatal("scanned no source files — the scan has gone blind, so a pass means nothing")
	}
	if len(offenders) > 0 {
		t.Errorf("hand-typed chrome component lookup at %v\n"+
			"Use ResolveChromeComponent(ctx, db, ChromeSlotFunction(slot), logger) so the "+
			"selection cannot drift back into three predicates (bugs_open/118).", offenders)
	}
}

// Proof the scan can see an offender. Without it, a scan that stopped matching
// would report a clean tree for ever — the failure mode this whole file exists
// to prevent, applied to itself.
func TestChromeLookupScanFiresOnASyntheticLine(t *testing.T) {
	lines := []string{
		`			WHERE function = 'site-footer' AND is_active = true`,                                     // offender: the old link_site_components form
		`		// WHERE function = 'site-header' ORDER BY name LIMIT 1`,                                  // comment, must be ignored
		"		comp, eligible, err := ResolveChromeComponent(ctx, db, ChromeSlotFunction(slot), logger)", // correct form
	}
	var hits []int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		if chromeFunctionLiteral.MatchString(line) {
			hits = append(hits, i)
		}
	}
	if len(hits) != 1 || hits[0] != 0 {
		t.Fatalf("scan matched %v; want exactly the hand-typed lookup (index 0)", hits)
	}
}
