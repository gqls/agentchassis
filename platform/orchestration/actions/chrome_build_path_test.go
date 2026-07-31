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

// ── bugs_open/167 ────────────────────────────────────────────────────────────
//
// bugs_open/118 gave chrome selection ONE predicate and pinned the two
// ASSIGNMENT call sites to it. It deliberately left the three BUILD call sites —
// RenderHeader, RenderFooter, RenderHead — still resolving through
// GetComponentByFunction, which has no component_level filter. So the page-build
// path could select a component_level='section' component to serve as site
// chrome, and for both chrome functions the section component sits in the very
// same pool as the chrome one:
//
//	site-header  -> header-theme-chrome [site]     AND  site-header [section]
//	site-footer  -> footer-theme-chrome [site]     AND  site-footer [section]
//	head         -> head-seo-standard  [head]      AND  Document Head [section]
//
// Measured live 2026-07-31, the build path already returned the CHROME row for
// header and footer — but only because `ORDER BY name` happens to sort
// 'header-theme-chrome' before 'site-header' and 'footer-theme-chrome' before
// 'site-footer'. Twice lucky, on a tie-break that knows nothing about chrome.
// These tests remove the luck.
//
// This file is separate from chrome_selection_test.go on purpose: that file was
// being edited by another session (bugs_open/166) while this was written, and a
// same-file passenger is the one thing a pathspec commit cannot keep out.

// chromeBuildRowCols mirrors ResolveChromeComponent's SELECT list, including the
// computed chrome_eligible column.
func chromeBuildRowCols() []string {
	return []string{"id", "name", "function", "category", "html_template", "input_schema", "is_dark_section", "chrome_eligible"}
}

func chromeBuildRow(name, function, html string, eligible bool) *sqlmock.Rows {
	return sqlmock.NewRows(chromeBuildRowCols()).
		AddRow("11111111-1111-1111-1111-111111111111", name, function,
			"", html, []byte(`{}`), false, eligible)
}

// ── the head slot: the case that makes `eligible` a GATE and not a hint ───────
//
// ResolveChromeComponent ALWAYS answers — deliberately, so a caller can report a
// library gap instead of silently losing the slot. Live, the head pool has NO
// eligible member (both candidates are is_active=false) and the answer is
// `Document Head`, a component_level='section' component. A caller that used the
// row because a row came back would render a page section as <head> — creating
// 167 on the one slot that never had it. That is not hypothetical: it is what a
// straight swap of the lookup produces.

func TestRenderHeadFallsBackWhenTheLibraryHasNoEligibleHead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// WithArgs pins the function, so this also asserts ChromeSlotFunction("head")
	// reached the resolver rather than a slot name.
	mock.ExpectQuery("FROM content_components").
		WithArgs("head").
		WillReturnRows(chromeBuildRow("Document Head", "head", "<section>I AM A PAGE SECTION</section>", false))

	got, err := RenderHead(context.Background(), db, uuid.Nil, &RenderContext{Title: "T"}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderHead: %v", err)
	}

	if strings.Contains(got, "I AM A PAGE SECTION") {
		t.Error("an INELIGIBLE component was rendered as <head> — that is bugs_open/167 on the head slot")
	}
	if !strings.Contains(got, "HEAD SOURCE: fallback") {
		t.Errorf("an ineligible library must fall through to RenderFallbackHead, got:\n%s", got)
	}
	// RenderFallbackHead is what runs TODAY when the lookup finds no row at all,
	// so this is the assertion that the fix changes no served byte on this slot.
	if !strings.Contains(got, "<title>T</title>") {
		t.Errorf("the fallback head must still carry the page title, got:\n%s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected queries: %v", err)
	}
}

func TestRenderHeadUsesAnEligibleHeadComponent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM content_components").
		WithArgs("head").
		WillReturnRows(chromeBuildRow("head-seo-standard", "head", "<head><meta name=\"x\"></head>", true))

	got, err := RenderHead(context.Background(), db, uuid.Nil, &RenderContext{Title: "T"}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderHead: %v", err)
	}
	if !strings.Contains(got, `<meta name="x">`) {
		t.Errorf("an eligible head component must be rendered, got:\n%s", got)
	}
	// The marker names the COMPONENT, not the function. That is what makes a
	// future 167 visible from outside — `curl` a page and the header comment says
	// which component actually served the slot.
	if !strings.Contains(got, "HEAD SOURCE: component-db:head-seo-standard") {
		t.Errorf("the source marker must name the resolved component, got:\n%s", got)
	}
}

// ── header and footer ────────────────────────────────────────────────────────
//
// siteID=uuid.Nil with an empty Domain skips both style-collection lookups, so
// these reach the by-function branch in exactly one query — which is the branch
// this bug is about.

func TestRenderHeaderRefusesAnIneligibleComponentAndFallsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The live shape of the defect: `site-header`, component_level='section',
	// is_active=true, unforked — invisible to the old predicate, which had no
	// level filter and would have rendered this 6.6KB page section as the header.
	mock.ExpectQuery("FROM content_components").
		WithArgs("site-header").
		WillReturnRows(chromeBuildRow("site-header", "site-header", "<section>PAGE SECTION MASQUERADING AS CHROME</section>", false))

	got, err := RenderHeader(context.Background(), db, uuid.Nil, &RenderContext{}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderHeader: %v", err)
	}
	if strings.Contains(got, "PAGE SECTION MASQUERADING AS CHROME") {
		t.Error("a component_level='section' component was rendered as the site header — bugs_open/167")
	}
	if !strings.Contains(got, "HEADER SOURCE: fallback") {
		t.Errorf("an ineligible header must fall through to RenderFallbackHeader, got:\n%s", got)
	}
	// The fallback is at least a header, which the section component is not.
	if !strings.Contains(got, `class="site-header"`) {
		t.Errorf("RenderFallbackHeader must still emit a real header element, got:\n%s", got)
	}
}

func TestRenderHeaderUsesTheEligibleChromeComponent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM content_components").
		WithArgs("site-header").
		WillReturnRows(chromeBuildRow("header-theme-chrome", "site-header", `<header class="site-header">CHROME</header>`, true))

	got, err := RenderHeader(context.Background(), db, uuid.Nil, &RenderContext{}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderHeader: %v", err)
	}
	if !strings.Contains(got, "CHROME") {
		t.Errorf("the eligible chrome component must be rendered, got:\n%s", got)
	}
	if !strings.Contains(got, "HEADER SOURCE: component-db:header-theme-chrome") {
		t.Errorf("the source marker must name the resolved component, got:\n%s", got)
	}
}

func TestRenderFooterRefusesAnIneligibleComponentAndFallsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM content_components").
		WithArgs("site-footer").
		WillReturnRows(chromeBuildRow("site-footer", "site-footer", "<section>PAGE SECTION MASQUERADING AS CHROME</section>", false))

	got, err := RenderFooter(context.Background(), db, uuid.Nil, &RenderContext{}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderFooter: %v", err)
	}
	if strings.Contains(got, "PAGE SECTION MASQUERADING AS CHROME") {
		t.Error("a component_level='section' component was rendered as the site footer — bugs_open/167")
	}
	if !strings.Contains(got, "FOOTER SOURCE: fallback") {
		t.Errorf("an ineligible footer must fall through to RenderFallbackFooter, got:\n%s", got)
	}
}

func TestRenderFooterUsesTheEligibleChromeComponent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM content_components").
		WithArgs("site-footer").
		WillReturnRows(chromeBuildRow("footer-theme-chrome", "site-footer", `<footer class="site-footer">CHROME</footer>`, true))

	got, err := RenderFooter(context.Background(), db, uuid.Nil, &RenderContext{}, zap.NewNop())
	if err != nil {
		t.Fatalf("RenderFooter: %v", err)
	}
	if !strings.Contains(got, "CHROME") {
		t.Errorf("the eligible chrome component must be rendered, got:\n%s", got)
	}
	if !strings.Contains(got, "FOOTER SOURCE: component-db:footer-theme-chrome") {
		t.Errorf("the source marker must name the resolved component, got:\n%s", got)
	}
}

// ── the door: a Go-level lookup, which the 118 scan cannot see ───────────────
//
// TestNoChromeSelectionHandTypesItsOwnLookup (118) scans for hand-typed SQL
// (`function = 'site-header'`) and SKIPS component_library.go outright. Both
// exemptions were right for what it guards and both are why it was blind here:
// these three call sites live in component_library.go and contained no SQL at
// all. They were Go calls handing a chrome function name to a section-shaped
// lookup.
//
// So this scan is the complement, not a duplicate: it matches the CALL, it
// covers component_library.go, and it fails on the exact form the bug had.

var chromeFunctionGoLookup = regexp.MustCompile(
	`Get(?:ComponentByFunction|ComponentWithFallback)\s*\([^)]*"(?:site-header|site-footer|head)"`)

func TestNoBuildPathResolvesChromeByPlainFunctionLookup(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}

	var offenders []string
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		// NOTE: component_library.go is deliberately NOT skipped here — it is the
		// file all three defects lived in.
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		scanned++
		for i, line := range strings.Split(string(src), "\n") {
			// The old calls are quoted in comments as the record of what was wrong.
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if chromeFunctionGoLookup.MatchString(line) {
				offenders = append(offenders, fmt.Sprintf("%s:%d", name, i+1))
			}
		}
	}

	// A scan that stopped matching would report a clean tree for ever — the exact
	// failure mode this file exists to prevent, turned on itself.
	if scanned == 0 {
		t.Fatal("scanned no source files — the scan has gone blind, so a pass means nothing")
	}
	if len(offenders) > 0 {
		t.Errorf("chrome function resolved through a section-shaped lookup at %v\n"+
			"GetComponentByFunction has NO component_level filter, so it can return a "+
			"component_level='section' component to serve as chrome (bugs_open/167). Use:\n"+
			"    comp, eligible, err := ResolveChromeComponent(ctx, db, ChromeSlotFunction(slot), logger)\n"+
			"and treat !eligible as a reason to fall back, never as a component to render.", offenders)
	}
}

// Proof the scan can see the defect it was written for, and that it does NOT
// fire on the corrected form. Without both halves a green run means nothing:
// a scan that matches everything and one that matches nothing both stay quiet
// once the offending line is gone.
func TestChromeGoLookupScanFiresOnTheOriginalDefect(t *testing.T) {
	cases := []struct {
		line string
		want bool
		why  string
	}{
		{`		comp, err = GetComponentByFunction(ctx, db, "site-header", logger)`, true, "RenderHeader as filed"},
		{`		comp, err = GetComponentByFunction(ctx, db, "site-footer", logger)`, true, "RenderFooter as filed"},
		{`	comp, err := GetComponentByFunction(ctx, db, "head", logger)`, true, "RenderHead as filed"},
		{`	comp, err := GetComponentWithFallback(ctx, db, "site-header", logger)`, true, "the sibling lookup routes to the same query"},
		{`		// comp, err = GetComponentByFunction(ctx, db, "site-header", logger)`, false, "a comment is the record, not a call"},
		{`	comp, eligible, err := ResolveChromeComponent(ctx, db, ChromeSlotFunction("header"), logger)`, false, "the corrected form"},
		{`	return GetComponentByFunction(ctx, db, "generic-text-block", logger)`, false, "the SECTION path must stay untouched"},
		{`	comp, err := GetComponentByFunction(ctx, db, function, logger)`, false, "a variable function is the generic section lookup"},
	}

	for _, c := range cases {
		line := c.line
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			if c.want {
				t.Fatalf("test case is self-contradictory: %q", c.why)
			}
			continue
		}
		if got := chromeFunctionGoLookup.MatchString(line); got != c.want {
			t.Errorf("scan matched=%v, want %v for %s\n  line: %s", got, c.want, c.why, line)
		}
	}
}
