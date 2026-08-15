// FILE: platform/orchestration/actions/discovery_checks/check_missing_structure_test.go
//
// bugfix_270: the predicate this check issues is the whole defect (it used to
// read pages.rendered_header/footer/head, empty on every page fleet-wide, so
// it fired on every site, every pass, forever). A happy-path test that only
// checks res.WorkItems for a hand-fed sqlmock row cannot catch a regression
// back to that column, because sqlmock returns whatever the test hands it
// regardless of the WHERE clause — so TestMissingStructureQueryReadsSiteComponents
// below asserts on the SQL actually issued, following
// check_componentless_pages_test.go's header rationale exactly.
package discovery_checks

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func newMissingStructureCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.MustParse("00ff3af5-dad8-4770-9f70-3edc267a3c92"),
		Pipeline:  "build",
		AgentType: "completeness-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

func missingStructureRow(mock sqlmock.Sqlmock, hasActivePages, headerOK, footerOK, headOK bool) {
	mock.ExpectQuery(`SELECT`).WillReturnRows(
		sqlmock.NewRows([]string{"has_active_pages", "header_ok", "footer_ok", "head_ok"}).
			AddRow(hasActivePages, headerOK, footerOK, headOK))
}

// TestMissingStructureHealthySiteSelfClears is the demand control for the
// bug this check exists to fix: the OLD predicate could never return this
// shape (rendered_header IS NULL was true for every page, unconditionally),
// so a healthy site producing zero findings plus a retraction is the proof
// the predicate can now be false at all.
func TestMissingStructureHealthySiteSelfClears(t *testing.T) {
	dctx, mock := newMissingStructureCtx(t)
	missingStructureRow(mock, true, true, true, true)

	res, err := (&MissingStructureCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Findings) != 0 || len(res.WorkItems) != 0 {
		t.Errorf("healthy site must file nothing; got %d findings, %d work items",
			len(res.Findings), len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want 1 retraction, got %d", len(res.Resolved))
	}
	r := res.Resolved[0]
	if r.ItemType != "needs_rerender" || r.ItemKey != "missing_structure:rerender" {
		t.Errorf("retraction targets ItemType=%q ItemKey=%q, want needs_rerender / missing_structure:rerender",
			r.ItemType, r.ItemKey)
	}
	if r.AllOfType {
		t.Errorf("retraction must be narrow (ItemKey), not AllOfType — this check owns exactly one key per site")
	}
	if r.Reason == "" {
		t.Errorf("retraction must carry a reason (RFC_010 requires one)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestMissingStructureFiresOnMissingSlot is the check's actual job: one
// unhealthy slot, one site with active pages, one work item under the
// UNCHANGED item_key (load-bearing: the RFC_010 retraction on a later
// healthy pass matches on this key, so a new key would orphan every
// historical open row instead of closing it).
func TestMissingStructureFiresOnMissingSlot(t *testing.T) {
	dctx, mock := newMissingStructureCtx(t)
	missingStructureRow(mock, true, true, false, true) // footer empty

	res, err := (&MissingStructureCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Resolved) != 0 {
		t.Errorf("an unhealthy site must not also retract; got %d", len(res.Resolved))
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "needs_rerender" {
		t.Errorf("ItemType = %q, want needs_rerender", wi.ItemType)
	}
	if wi.ItemKey != "missing_structure:rerender" {
		t.Errorf("ItemKey = %q, want missing_structure:rerender (unchanged — RFC_010 retraction depends on it)", wi.ItemKey)
	}
	if wi.HandlerAgent != "rerender-pages" {
		t.Errorf("HandlerAgent = %q, want rerender-pages", wi.HandlerAgent)
	}
	if !regexp.MustCompile(`footer`).MatchString(wi.Summary) {
		t.Errorf("Summary must name the unhealthy slot; got %q", wi.Summary)
	}
	if regexp.MustCompile(`likely built before`).MatchString(wi.SpecJSON) {
		t.Errorf("spec must not carry the old guessed-history reason string; got %s", wi.SpecJSON)
	}
}

// TestMissingStructureSkipsSiteWithNoActivePages — nothing to reassemble, and
// chrome-absence for a site with no live pages is not a claim we have
// grounds to make either way.
func TestMissingStructureSkipsSiteWithNoActivePages(t *testing.T) {
	dctx, mock := newMissingStructureCtx(t)
	missingStructureRow(mock, false, false, false, false)

	res, err := (&MissingStructureCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Findings) != 0 || len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Errorf("a site with no active pages must produce nothing at all; got %d/%d/%d",
			len(res.Findings), len(res.WorkItems), len(res.Resolved))
	}
}

// TestMissingStructurePropagatesQueryError — RFC_010's safety property is
// that a blinded check cannot retract real defects; the runner enforces this
// by skipping Resolved entirely whenever Run returns an error, so Run must
// actually return the error rather than swallowing it into an empty result.
func TestMissingStructurePropagatesQueryError(t *testing.T) {
	dctx, mock := newMissingStructureCtx(t)
	mock.ExpectQuery(`SELECT`).WillReturnError(context.DeadlineExceeded)

	if _, err := (&MissingStructureCheck{}).Run(dctx); err == nil {
		t.Fatal("want an error propagated, got nil")
	}
}

// TestMissingStructureQueryReadsSiteComponents is the anti-regression test:
// it asserts on the SQL actually issued, because every test above would stay
// green even if the query silently reverted to pages.rendered_header (a
// sqlmock ExpectQuery matched by `SELECT` alone does not care what follows).
// This is bug 270 itself — the query must never again read the vestigial
// columns, and must never gate on build_status (build_status='pending'
// coexists with valid, currently-serving rendered_html elsewhere in this
// codebase — see chrome_link_policy.go — so gating on it here would
// manufacture a different false-positive class).
func TestMissingStructureQueryReadsSiteComponents(t *testing.T) {
	issued := missingStructureQuery

	for _, want := range []struct{ name, pattern string }{
		{"reads site_components", `FROM site_components sc`},
		{"keys on rendered_html content", `length\(sc\.rendered_html\)`},
		{"checks all three slots", `'header'.*'footer'.*'head'`},
	} {
		re := regexp.MustCompile(`(?s)` + want.pattern)
		if !re.MatchString(issued) {
			t.Errorf("%s missing from query: /%s/ did not match:\n%s", want.name, want.pattern, issued)
		}
	}
	if regexp.MustCompile(`p\.rendered_(header|footer|head)`).MatchString(issued) {
		t.Errorf("query must not read pages.rendered_* — that is the vestigial-column defect this fix removes:\n%s", issued)
	}
	if regexp.MustCompile(`build_status`).MatchString(issued) {
		t.Errorf("query must not gate on build_status — 'pending' coexists with valid content:\n%s", issued)
	}
	if regexp.MustCompile(`'deployed'`).MatchString(issued) {
		t.Errorf("query must not filter on pages.status='deployed' — that value never occurs there:\n%s", issued)
	}
}
