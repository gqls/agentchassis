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
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ── bugs_open/170 ────────────────────────────────────────────────────────────
//
// "Which component serves this site's chrome?" is answered by TWO stores, and
// until 2026-08-01 only one of them was guarded:
//
//	site_components.component_id            predicate (118) · detector · repair (166)
//	style_collections.header_component_id   nothing at all
//	style_collections.footer_component_id   nothing at all
//
// Both pin consumers dereferenced the pin with GetComponentByID — `WHERE id = $1`,
// no predicate — so a pin overrode every fix 118, 166 and 167 installed:
//
//   - RenderHeader/RenderFooter rendered whatever was pinned. Three deployed sites
//     were pinned to header-professional-dark (is_active=false) and four to
//     footer-4-column (is_active=false) on 2026-08-01.
//   - link_site_components WROTE the pin into site_components.component_id and
//     NULLed rendered_html — so the next run for those sites would have undone
//     166's repair, which had already moved them to the active chrome components.
//
// This file is separate from chrome_selection_test.go and chrome_build_path_test.go
// on purpose, for the reason the latter states: a same-file passenger is the one
// thing a pathspec commit cannot keep out, and this package is worked concurrently.

// ── the predicate, and the one clause that must NOT be in it ─────────────────

func TestChromePinPredicateDoesNotExcludeForks(t *testing.T) {
	got := chromePinEligibleSQL("")

	// leopardessconsulting.co.uk pins header-leopardess, a FORK, and that is
	// CORRECT — pinning a site to its own fork is what a pin is for. The pool
	// predicate excludes forks because an active fork of one client's header would
	// otherwise become every other site's default (bugs_open/118). Applying that
	// clause to a pin inverts its meaning.
	if strings.Contains(got, "forked_from") {
		t.Errorf("the PIN predicate must not exclude forks — pinning a site to its own fork is the "+
			"intended use (leopardessconsulting.co.uk, the one legitimate pin in the fleet). got %q", got)
	}
	// The two clauses that must be there.
	if !strings.Contains(got, "is_active") {
		t.Errorf("pin predicate %q has no is_active clause — that is the whole of bugs_open/170", got)
	}
	if !strings.Contains(got, "component_level IN (") {
		t.Errorf("pin predicate %q has no level filter — a page-section component could be pinned as "+
			"chrome, which is bugs_open/167 on the pin path", got)
	}
	// And it must DIFFER from the pool predicate, or the asymmetry has been lost.
	// Measured 2026-08-01 over all four live pins, this is not a stylistic
	// difference: the two disagree on exactly one row, and the pin predicate is the
	// one that gets it right.
	//
	//	                            pin    pool
	//	header-professional-dark x3 false  false
	//	header-leopardess (fork)    TRUE   false   <- collapsing them loses this site's header
	if got == chromeEligibleSQL("") {
		t.Error("pin and pool predicates are identical — the fork asymmetry has been collapsed. " +
			"That makes the guard's first live action the deletion of the single CORRECT pin in the " +
			"fleet (leopardessconsulting.co.uk's own forked header) while still catching the three real ones.")
	}
}

func TestChromePinEligibleSQLQualifiesEveryColumn(t *testing.T) {
	// A half-qualified predicate compiles, runs, and silently reads the wrong table
	// in a join — and link_site_components applies this one to TWO different aliases
	// (hc. and fc.) in a single query, so an unqualified column there would resolve
	// to whichever table Postgres liked.
	got := chromePinEligibleSQL("cc.")
	for _, col := range []string{"is_active", "component_level"} {
		if !strings.Contains(got, "cc."+col) {
			t.Errorf("alias form did not qualify %s: %q", col, got)
		}
	}
	if strings.Contains(strings.ReplaceAll(got, "cc.", ""), "cc.") {
		t.Errorf("alias applied more than once: %q", got)
	}
}

// ── the build path ───────────────────────────────────────────────────────────

func chromePinCollectionCols() []string {
	return []string{
		"id", "name", "display_name", "header_component_id", "footer_component_id",
		"css_theme_id", "color_palette", "typography", "category", "industry_tags",
	}
}

// pinnedCollection mirrors GetStyleCollectionForSite's SELECT list with one pin set.
func pinnedCollection(name, headerPin, footerPin interface{}) *sqlmock.Rows {
	return sqlmock.NewRows(chromePinCollectionCols()).
		AddRow("99999999-9999-9999-9999-999999999999", name, name, headerPin, footerPin,
			nil, []byte(`{}`), []byte(`{}`), "", []byte(`[]`))
}

func chromePinRow(name, function, html string, eligible bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "function", "category", "html_template", "input_schema",
		"is_dark_section", "pin_eligible",
	}).AddRow("88888888-8888-8888-8888-888888888888", name, function, "", html, []byte(`{}`), false, eligible)
}

// The live defect: a collection pinning a DEACTIVATED header. The pin must be
// abandoned and the eligible library component served instead.
//
// This test is the one that asserts the FLEET-VISIBLE half of the fix. Before it,
// three deployed sites rendered header-professional-dark; after it they render
// header-theme-chrome — which is the component bugs_open/166 had already repointed
// their site_components row to, so the change makes the two paths AGREE rather
// than choosing something new.
func TestRenderHeaderAbandonsAnIneligiblePinForTheLibrary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pinID := "11111111-1111-1111-1111-111111111111"

	mock.ExpectQuery("FROM sites s").
		WillReturnRows(pinnedCollection("professional-dark", pinID, nil))
	// The pin dereference, reporting pin_eligible=false (is_active=false live).
	mock.ExpectQuery("FROM content_components").
		WillReturnRows(chromePinRow("header-professional-dark", "site-header",
			`<header class="site-header">DEACTIVATED PINNED HEADER</header>`, false))
	// …and the fall-through to the eligible-only pool lookup.
	mock.ExpectQuery("FROM content_components").
		WithArgs("site-header").
		WillReturnRows(chromeBuildRow("header-theme-chrome", "site-header",
			`<header class="site-header">LIBRARY CHROME</header>`, true))

	got, err := RenderHeader(context.Background(), db, siteID, &RenderContext{}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderHeader: %v", err)
	}
	if strings.Contains(got, "DEACTIVATED PINNED HEADER") {
		t.Error("a deactivated component pinned by the style collection was rendered as site chrome — bugs_open/170")
	}
	if !strings.Contains(got, "LIBRARY CHROME") {
		t.Errorf("an ineligible pin must fall through to ResolveChromeComponent, got:\n%s", got)
	}
	// The source marker must name the COMPONENT, not the collection: the pin branch
	// reports `component-db:<collection>` and the pool branch `component-db:<component>`,
	// so this asserts which branch actually produced the output rather than merely
	// that the right bytes appeared.
	if !strings.Contains(got, "HEADER SOURCE: component-db:header-theme-chrome") {
		t.Errorf("the source marker must show the POOL branch answered, got:\n%s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected collection + pin dereference + pool resolve: %v", err)
	}
}

// The positive control, and the row that makes the fork asymmetry real rather
// than theoretical: an ELIGIBLE pin — including one naming the site's own active
// fork — must still be honoured, with no library lookup at all.
//
// Without this test the whole fix is satisfied by ignoring pins entirely.
func TestRenderHeaderHonoursAnEligiblePinIncludingAFork(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pinID := "22222222-2222-2222-2222-222222222222"

	mock.ExpectQuery("FROM sites s").
		WillReturnRows(pinnedCollection("leopardess-dark-gold", pinID, nil))
	// header-leopardess: is_active=true, forked_from NOT NULL, component_level='site'.
	// The POOL predicate rejects it; the PIN predicate accepts it, and that is correct.
	mock.ExpectQuery("FROM content_components").
		WillReturnRows(chromePinRow("header-leopardess", "site-header",
			`<header class="site-header">BESPOKE CLIENT HEADER</header>`, true))

	got, err := RenderHeader(context.Background(), db, siteID, &RenderContext{}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderHeader: %v", err)
	}
	if !strings.Contains(got, "BESPOKE CLIENT HEADER") {
		t.Errorf("an eligible pin — a site's own active fork — must be rendered, got:\n%s", got)
	}
	if !strings.Contains(got, "HEADER SOURCE: component-db:leopardess-dark-gold") {
		t.Errorf("an honoured pin must report the COLLECTION as its source, got:\n%s", got)
	}
	// sqlmock fails on any unexpected query, so this also asserts NO pool lookup
	// happened — an honoured pin must cost exactly two queries.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("an honoured pin must short-circuit the library lookup: %v", err)
	}
}

// The footer half, which is WIDER than the header half: four collections pin
// footer-4-column (is_active=false), including leopardess, whose header pin is the
// one legitimate row. A fix tested only on the header would leave the larger half.
func TestRenderFooterAbandonsAnIneligiblePinForTheLibrary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pinID := "33333333-3333-3333-3333-333333333333"

	mock.ExpectQuery("FROM sites s").
		WillReturnRows(pinnedCollection("professional-dark", nil, pinID))
	mock.ExpectQuery("FROM content_components").
		WillReturnRows(chromePinRow("footer-4-column", "site-footer",
			`<footer class="site-footer">DEACTIVATED PINNED FOOTER</footer>`, false))
	mock.ExpectQuery("FROM content_components").
		WithArgs("site-footer").
		WillReturnRows(chromeBuildRow("footer-theme-chrome", "site-footer",
			`<footer class="site-footer">LIBRARY CHROME</footer>`, true))

	got, err := RenderFooter(context.Background(), db, siteID, &RenderContext{}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderFooter: %v", err)
	}
	if strings.Contains(got, "DEACTIVATED PINNED FOOTER") {
		t.Error("a deactivated component pinned by the style collection was rendered as the site footer — bugs_open/170")
	}
	if !strings.Contains(got, "FOOTER SOURCE: component-db:footer-theme-chrome") {
		t.Errorf("an ineligible footer pin must fall through to the library, got:\n%s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected collection + pin dereference + pool resolve: %v", err)
	}
}

// A pin whose eligibility is decided by a SEPARATE round trip could be told the
// component is fine by one query and the truth by another. Same construction, and
// same reason, as ResolveChromeComponent: the flag rides on the fetch.
func TestChromePinEligibilityRidesOnTheFetch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	id := uuid.New()
	mock.ExpectQuery(`\(.*is_active.*component_level.*\) as pin_eligible[\s\S]*WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(chromePinRow("header-professional-dark", "site-header", "<header></header>", false))

	comp, eligible, err := GetChromePinComponent(context.Background(), db, id, zap.NewNop())
	if err != nil {
		t.Fatalf("GetChromePinComponent: %v", err)
	}
	if eligible {
		t.Error("pin_eligible=false must surface as eligible=false")
	}
	if comp == nil || comp.Name != "header-professional-dark" {
		t.Fatalf("the row must still be returned so the caller can name it in a log, got %#v", comp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the shipped query must carry the predicate as a computed column on the SAME fetch: %v", err)
	}
}

// ── the door: no consumer may dereference a pin with the unguarded fetch ─────
//
// This is the test that stops 170 recurring, and it is the framework-level half
// of the fix. The two pin consumers did not disagree with the eligibility rule on
// purpose; they were each written separately against a general-purpose by-id
// fetch, and nothing connected them to the predicate. A third consumer written
// tomorrow would do the same.
//
// Same construction as TestNoChromeSelectionHandTypesItsOwnLookup (118) and
// TestNoWriterHandTypesTheAssetLockPredicate (LOCK-007).

// chromePinDeref matches a by-id component fetch whose argument is a chrome pin
// field — the exact form both consumers had.
var chromePinDeref = regexp.MustCompile(
	`GetComponentByID\([^)]*(HeaderComponentID|FooterComponentID|header_component_id|footer_component_id)`)

// chromePinSQLRead matches SQL that selects a pin column without the predicate
// beside it. Deliberately narrow: it fires on a raw read of the pin column in a
// query that never mentions component_level, which is what link_site_components
// looked like before this fix.
var chromePinSQLRead = regexp.MustCompile(`(?i)(header_component_id|footer_component_id)\s*::\s*text`)

func TestNoConsumerDereferencesAChromePinUnguarded(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}

	var offenders []string
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		scanned++
		s := string(src)

		for i, line := range strings.Split(s, "\n") {
			// Comments quote the old form on purpose — they are the record of what
			// was wrong, not a second source of truth.
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if chromePinDeref.MatchString(line) {
				offenders = append(offenders, fmt.Sprintf("%s:%d (Go dereference)", name, i+1))
			}
		}

		// The SQL form: a file that reads a pin column must also carry the pin
		// predicate. component_library.go is exempt because it DEFINES both the
		// predicate and the StyleCollection scan that legitimately reads the raw
		// columns into the struct.
		if name == "component_library.go" {
			continue
		}
		for _, q := range chromePinQueries(s) {
			if chromePinSQLRead.MatchString(q) && !chromePinQueryIsGuarded(q) {
				offenders = append(offenders, name+" (SQL read with no pin predicate)")
				break
			}
		}
	}

	if scanned == 0 {
		t.Fatal("scanned no source files — the scan has gone blind, so a pass means nothing")
	}
	if len(offenders) > 0 {
		t.Errorf("unguarded chrome-pin dereference at %v\n"+
			"Read a style_collections chrome pin through GetChromePinComponent (or apply "+
			"chromePinEligibleSQL in the query), so a pin cannot bypass the predicate the rest of "+
			"chrome selection shares (bugs_open/170). An unguarded pin does not only render a "+
			"deactivated component — via link_site_components it WRITES one into "+
			"site_components.component_id, undoing bugs_open/166's repair.", offenders)
	}
}

// chromePinQueryIsGuarded says whether a pin read carries the predicate. Calling
// the shared helper counts, and is in fact the PREFERRED form — the point of the
// rule is that there is one predicate, not that the words appear inline. An
// inline `component_level` also counts, for a query built without concatenation.
func chromePinQueryIsGuarded(q string) bool {
	return strings.Contains(q, "chromePinEligibleSQL") || strings.Contains(q, "component_level")
}

// chromePinQueries returns the SQL that mentions a pin column, as the CALLER
// assembles it rather than as the file stores it.
//
// > **CORRECTED — the first version FAILED ON ITS OWN FIX.** It returned the raw
// > backtick literals, and the sanctioned guarded form injects the predicate by
// > concatenation:
// >
// >	`… scol.header_component_id::text, hc.name, (` + chromePinEligibleSQL("hc.") + `),
// >
// > so `component_level` lives in the Go glue BETWEEN two literals and never
// > appears inside either. The scan therefore reported link_site_components — the
// > file this bug's fix had just corrected — as an offender, while a genuinely
// > unguarded read would have looked identical to it.
// >
// > The fix is to stitch each literal back together with the glue chunks either
// > side of it, so the population is the query as EXECUTED. It is deliberately
// > statement-local rather than file-wide: a file that guards one pin read and
// > leaves another bare must still fail, which a "does this file mention the
// > predicate anywhere?" test would wave through.
func chromePinQueries(src string) []string {
	var out []string
	parts := strings.Split(src, "`")
	for i := 1; i < len(parts); i += 2 { // odd indices are inside backticks
		if !strings.Contains(parts[i], "header_component_id") && !strings.Contains(parts[i], "footer_component_id") {
			continue
		}
		// The literal plus the Go glue on each side. `+ chromePinEligibleSQL("hc.") +`
		// sits in exactly those chunks.
		effective := parts[i]
		if i > 0 {
			effective = parts[i-1] + effective
		}
		if i+1 < len(parts) {
			effective += parts[i+1]
		}
		out = append(out, effective)
	}
	return out
}

// Proof the scan can see an offender. Without it, a scan that stopped matching
// would report a clean tree for ever — the failure mode this file exists to
// prevent, applied to itself. Both forms are exercised, because the two consumers
// had one each.
func TestChromePinScanFiresOnBothOriginalForms(t *testing.T) {
	// The RenderHeader form, as it stood before this fix.
	goOffender := "		comp, err = GetComponentByID(ctx, db, *coll.HeaderComponentID, logger)"
	if !chromePinDeref.MatchString(goOffender) {
		t.Error("the Go scan no longer sees RenderHeader's original dereference — it has gone inert")
	}
	// A by-id fetch for an ordinary page section must NOT match: RenderComponentAction
	// legitimately calls GetComponentByID for arbitrary components, and a chrome
	// predicate there would be wrong.
	quiet := "		comp, err = GetComponentByID(ctx, params.DB, compUUID, params.Logger)"
	if chromePinDeref.MatchString(quiet) {
		t.Error("the scan must not fire on a general by-id component fetch — RenderComponentAction " +
			"renders arbitrary page sections and has no business with a chrome predicate")
	}

	// The link_site_components form, as it stood before this fix: one plain literal,
	// no predicate anywhere.
	sqlOffender := "q := `SELECT scol.header_component_id::text, scol.footer_component_id::text FROM sites s`"
	qs := chromePinQueries(sqlOffender)
	if len(qs) != 1 || !chromePinSQLRead.MatchString(qs[0]) || chromePinQueryIsGuarded(qs[0]) {
		t.Fatalf("the SQL scan no longer sees link_site_components' original read — it has gone inert. got %q", qs)
	}

	// The CONCATENATED guarded form — the one the scan's first version failed on,
	// and the reason chromePinQueries stitches the Go glue back in. `component_level`
	// appears in NEITHER literal; it arrives through the helper call between them.
	// If this ever reports unguarded again, the stitching has been lost.
	guardedConcat := "q := `SELECT scol.header_component_id::text, hc.name, (` + chromePinEligibleSQL(\"hc.\") + `) FROM sites s`"
	gq := chromePinQueries(guardedConcat)
	if len(gq) != 1 {
		t.Fatalf("the concatenated read must yield exactly one effective query, got %d: %q", len(gq), gq)
	}
	if strings.Contains(gq[0], "component_level") {
		t.Fatal("this fixture is meant to prove the STITCHING works — if component_level appears " +
			"inline it is testing the easy case and would pass even with the stitching removed")
	}
	if !chromePinQueryIsGuarded(gq[0]) {
		t.Fatalf("a pin read guarded by the shared helper must not be reported — that is the false "+
			"positive this scan's first version raised against its own fix. got %q", gq[0])
	}

	// And the inline guarded form, for a query built without concatenation.
	guardedInline := "q := `SELECT scol.header_component_id::text, (hc.is_active AND hc.component_level IN ('site')) FROM sites s`"
	iq := chromePinQueries(guardedInline)
	if len(iq) != 1 || !chromePinQueryIsGuarded(iq[0]) {
		t.Fatalf("an inline-guarded read must also be accepted, got %q", iq)
	}
}
