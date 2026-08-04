package actions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// bugs_open/191. RenderSiteComponentsAction built a header's nav items and its
// CTA button in one run, from one pages table, and validated them with two
// different predicates: the nav through loadFetchablePageSet (which negates
// datahelpers.NeverDeployedPagePredicate), the CTA through loadResolverPageSet
// (which has no deployment test at all). mortgagecalculator.co.uk therefore
// shipped a header whose nav had been filtered down to its one deployed page
// and whose button, beside it, pointed at /tools/stamp-duty/index.html —
// build_status 'planned', deployed_at NULL, HTTP 404 measured on the wire
// 2026-08-04.
//
// These tests pin the policy itself; the two scans at the foot pin the door.

func chromeLinkPolicyMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func chromeLinkPolicyPageRows(urls ...string) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"url"})
	for _, u := range urls {
		r.AddRow(u)
	}
	return r
}

func chromeLinkPolicyObserved() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.WarnLevel)
	return zap.New(core), logs
}

// Matches the fetchable-page query and NOTHING that keys on build_status alone.
// Same trick as nav_visibility_test.go:57-60 — if someone "simplifies" the
// predicate, the expectation stops matching and this file goes red.
var chromeLinkPolicyFetchableQuery = regexp.QuoteMeta("NOT (deployed_at IS NULL")

// THE BUG, in miniature: one deployed page, one planned page, and the policy
// must separate them. RED against the old code, where the CTA consulted a set
// built without any deployment predicate.
func TestChromeLinkPolicyDropsANeverDeployedTarget(t *testing.T) {
	db, mock := chromeLinkPolicyMockDB(t)
	logger, logs := chromeLinkPolicyObserved()

	mock.ExpectQuery(chromeLinkPolicyFetchableQuery).WillReturnRows(
		chromeLinkPolicyPageRows("/index.html"))

	policy := LoadChromeLinkPolicy(context.Background(), db, uuid.New(), logger)

	if policy.Unfiltered() {
		t.Fatal("a site WITH a deployed page must filter, not degrade")
	}
	if !policy.Allows("/index.html") {
		t.Error("the deployed page must be linkable from chrome")
	}
	if policy.Allows("/tools/stamp-duty/index.html") {
		t.Error("a never-deployed page must NOT be linkable from chrome — this is bugs_open/191 reopening")
	}
	if logs.Len() != 0 {
		t.Errorf("filtering is the normal path and must not warn, got %+v", logs.All())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// The first-build escape. Deleting it freezes a buttonless, near-empty header
// into a brand-new site's chrome, because the render is idempotence-gated.
// Keys on the row COUNT, never on set emptiness: loadFetchablePageSet always
// injects the site root, so the set is never empty.
func TestChromeLinkPolicyFirstBuildIsUnfiltered(t *testing.T) {
	db, mock := chromeLinkPolicyMockDB(t)
	logger, logs := chromeLinkPolicyObserved()

	mock.ExpectQuery(chromeLinkPolicyFetchableQuery).WillReturnRows(chromeLinkPolicyPageRows())

	policy := LoadChromeLinkPolicy(context.Background(), db, uuid.New(), logger)

	if !policy.Unfiltered() {
		t.Fatal("a site with no deployed pages must degrade to unfiltered")
	}
	if !policy.Allows("/anything-at-all.html") {
		t.Error("an unfiltered policy must allow every link")
	}
	if logs.Len() != 1 {
		t.Errorf("expected exactly 1 warning, got %d: %+v", logs.Len(), logs.All())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// An infrastructure blip during the one gated chrome render must not amputate a
// site's CTA for ever. Note this CHANGES the old CTA behaviour deliberately:
// loadResolverPageSet returned an empty set on error, so the button silently
// vanished. That was a side effect of the loader, not a decision.
func TestChromeLinkPolicyLookupErrorIsUnfiltered(t *testing.T) {
	db, mock := chromeLinkPolicyMockDB(t)
	logger, logs := chromeLinkPolicyObserved()

	mock.ExpectQuery(chromeLinkPolicyFetchableQuery).WillReturnError(errors.New("connection reset"))

	policy := LoadChromeLinkPolicy(context.Background(), db, uuid.New(), logger)

	if !policy.Unfiltered() {
		t.Fatal("a failed lookup must degrade to unfiltered, never to empty chrome")
	}
	if !policy.Allows("/index.html") {
		t.Error("an unfiltered policy must allow every link")
	}
	if logs.Len() != 1 {
		t.Errorf("expected exactly 1 warning, got %d: %+v", logs.Len(), logs.All())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// Only page links can 404 against the pages table. Dropping this guard would
// strip an external or anchor CTA that was never at risk.
func TestChromeLinkPolicyIgnoresNonPageLinks(t *testing.T) {
	db, mock := chromeLinkPolicyMockDB(t)

	mock.ExpectQuery(chromeLinkPolicyFetchableQuery).WillReturnRows(
		chromeLinkPolicyPageRows("/index.html"))

	policy := LoadChromeLinkPolicy(context.Background(), db, uuid.New(), zap.NewNop())

	for _, href := range []string{"https://example.org/", "#top", "mailto:hi@example.org"} {
		if !policy.Allows(href) {
			t.Errorf("non-page link %q must survive: it cannot 404 against the pages table", href)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// ── the door ─────────────────────────────────────────────────────────────────
//
// Comments are skipped by both scans below. They quote the old form on purpose
// — they are the record of what was wrong, not a second source of truth. The
// synthetic-line tests prove the skip works, so the doc comments added by this
// change (and every historical one naming loadResolverPageSet) stay legal.

// loosePageSetCall matches a call to the page-CONTENT link set.
var loosePageSetCall = regexp.MustCompile(`loadResolverPageSet\s*\(`)

// The files entitled to the loose set, each with the reason it is entitled.
// Anything else asking for it is a chrome link decision wearing content's
// clothes — the shape of bugs_open/191.
var loosePageSetAllowed = map[string]string{
	"resolve_internal_links_action.go": "defines it, and resolves page-CONTENT CTAs: re-resolved on every render, so a not-yet-shipped target self-corrects",
	"rerender_page_sections_action.go": "page-CONTENT CTAs on the rerender path, same repair economics",
}

func TestNoChromeCallerHandRollsALoosePageSet(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}

	var offenders []string
	hitsByAllowedFile := map[string]int{}
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
		for i, line := range strings.Split(string(src), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if !loosePageSetCall.MatchString(line) {
				continue
			}
			if _, ok := loosePageSetAllowed[name]; ok {
				hitsByAllowedFile[name]++
				continue
			}
			offenders = append(offenders, fmt.Sprintf("%s:%d", name, i+1))
		}
	}

	if scanned == 0 {
		t.Fatal("scanned no source files — the scan has gone blind, so a pass means nothing")
	}
	// An allow-list entry that matches nothing is a scan that has stopped
	// seeing, reported as a clean tree. Catch it here rather than in production.
	for name, why := range loosePageSetAllowed {
		if hitsByAllowedFile[name] == 0 {
			t.Fatalf("allow-listed %s (%s) produced NO match — either it was renamed or the matcher is dead; "+
				"in both cases this scan is no longer guarding anything", name, why)
		}
	}
	if len(offenders) > 0 {
		t.Errorf("chrome link target validated against the loose page-content set at %v\n"+
			"Use LoadChromeLinkPolicy(ctx, db, siteID, logger) and Allows(url): chrome ships on every page, "+
			"renders once behind an idempotence gate, and has no repair pass, so a never-deployed target "+
			"there is a site-wide 404 nothing later fixes (bugs_open/191).", offenders)
	}
}

// The header CTA specifically. Named separately from the scan above because
// this is the call site the bug was measured at, and a reverting edit should
// name it in the failure rather than appear as one entry in a list.
func TestHeaderCTAValidatesAgainstTheChromePolicy(t *testing.T) {
	src, err := os.ReadFile("render_site_components_action.go")
	if err != nil {
		t.Fatalf("read render_site_components_action.go: %v", err)
	}

	var sawPolicy, sawLoose bool
	for _, line := range strings.Split(string(src), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		if strings.Contains(line, "LoadChromeLinkPolicy(") {
			sawPolicy = true
		}
		if loosePageSetCall.MatchString(line) {
			sawLoose = true
		}
	}

	if !sawPolicy && !sawLoose {
		t.Fatal("found NEITHER marker in render_site_components_action.go — the CTA block has moved or been " +
			"renamed and this test has gone blind, so its pass means nothing")
	}
	if sawLoose {
		t.Error("the header CTA is validated against loadResolverPageSet again — that set has no deployment " +
			"predicate, so the button can point at a never-deployed page while the nav beside it, filtered by " +
			"the same run, cannot (bugs_open/191)")
	}
	if !sawPolicy {
		t.Error("render_site_components_action.go no longer calls LoadChromeLinkPolicy — the header CTA and the " +
			"nav beside it are answering 'which page may chrome link to?' separately again (bugs_open/191)")
	}
}

// Proof both scans can still see an offender. Without it, a matcher that
// stopped matching would report a clean tree for ever — the failure mode this
// file exists to prevent, applied to itself.
func TestLoosePageSetScanFiresOnASyntheticLine(t *testing.T) {
	lines := []string{
		`	headerPages := loadResolverPageSet(ctx, params, siteID, params.Logger)`,     // offender: the pre-fix CTA line
		`	// headerPages := loadResolverPageSet(ctx, params, siteID, logger)`,         // comment, must be ignored
		`	chromeLinks := LoadChromeLinkPolicy(ctx, params.DB, siteID, params.Logger)`, // correct form
	}
	var hits []int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		if loosePageSetCall.MatchString(line) {
			hits = append(hits, i)
		}
	}
	if len(hits) != 1 || hits[0] != 0 {
		t.Fatalf("scan matched %v; want exactly the loose call (index 0)", hits)
	}
}
