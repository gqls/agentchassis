// FILE: platform/orchestration/actions/discovery_checks/check_componentless_pages_test.go
//
// The tests that matter here are the two PREDICATE tests, not the happy path.
//
// This check exists because three sibling checks each miss the same page for a
// different reason, so the ways it can silently stop working are the ways those
// three already do:
//
//   - lose `deployed_at IS NOT NULL` and it starts filing mid-build pages
//     (measured when written: 6 of 15 candidates fleet-wide were 'planned' or
//     never deployed — a plan state, not a defect);
//   - lose the `NOT EXISTS (page_components)` guard and it becomes a
//     duplicate of check_empty_sections with none of its slot-level precision.
//
// Both are one-line deletions that leave every happy-path assertion green, which
// is exactly the shape of a rule that goes inert without anyone noticing. So they
// are asserted against the emitted SQL directly.
package discovery_checks

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func componentlessRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "page_type", "url", "sections", "build_status", "in_header",
	})
}

func newComponentlessCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
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

// TestComponentlessPagesEmitsForTheMotivatingCase is robot-hands.com/tools.html
// as it actually was: active, deployed 2026-05-10, in the header nav, three
// sections planned, zero components.
func TestComponentlessPagesEmitsForTheMotivatingCase(t *testing.T) {
	dctx, mock := newComponentlessCtx(t)
	pageID := uuid.New()

	mock.ExpectQuery(`FROM pages p`).WithArgs(dctx.SiteID).WillReturnRows(
		componentlessRows().AddRow(
			pageID.String(), "tools", "content", "/tools.html", 3, "needs_rebuild", true,
		),
	)

	res, err := (&ComponentlessPagesCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}

	wi := res.WorkItems[0]
	if wi.ItemType != "needs_content_page" {
		t.Errorf("ItemType = %q, want needs_content_page (reused, not a new type)", wi.ItemType)
	}
	if wi.HandlerAgent != "page-build-handler" {
		t.Errorf("HandlerAgent = %q, want page-build-handler", wi.HandlerAgent)
	}
	// in_header is what made this one visible to a human, so it must outrank a
	// deep-linked page.
	if wi.Severity != "high" {
		t.Errorf("Severity = %q, want high for an in-nav page", wi.Severity)
	}
	if wi.PageID == nil || *wi.PageID != pageID {
		t.Errorf("PageID not carried through; handler needs it to build the page")
	}
	// Must not collide with check_sectionless_pages' keys in idx_swi_dedup.
	if want := "componentless_page:" + pageID.String(); wi.ItemKey != want {
		t.Errorf("ItemKey = %q, want %q", wi.ItemKey, want)
	}
	// mode="recreate" would send this down load_existing_content's adoption path
	// to load content that does not exist. Its absence is load-bearing.
	if regexp.MustCompile(`"mode"`).MatchString(wi.SpecJSON) {
		t.Errorf("spec must NOT carry a mode key; got %s", wi.SpecJSON)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestComponentlessPagesSeverityDropsOffNav — a page nobody can click to is the
// same defect, lower priority. Guards against flattening severity to a constant.
func TestComponentlessPagesSeverityDropsOffNav(t *testing.T) {
	dctx, mock := newComponentlessCtx(t)
	mock.ExpectQuery(`FROM pages p`).WithArgs(dctx.SiteID).WillReturnRows(
		componentlessRows().AddRow(
			uuid.New().String(), "wholesale-pricing-explained", "content",
			"/wholesale-pricing-explained.html", 7, "deployed", false,
		),
	)

	res, err := (&ComponentlessPagesCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}
	if res.WorkItems[0].Severity != "medium" {
		t.Errorf("Severity = %q, want medium off-nav", res.WorkItems[0].Severity)
	}
}

// TestComponentlessPagesSilentWhenNothingFound — the ordinary case for a healthy
// site. No findings block either, so a clean run does not look like a finding.
func TestComponentlessPagesSilentWhenNothingFound(t *testing.T) {
	dctx, mock := newComponentlessCtx(t)
	mock.ExpectQuery(`FROM pages p`).WithArgs(dctx.SiteID).WillReturnRows(componentlessRows())

	res, err := (&ComponentlessPagesCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Findings) != 0 {
		t.Errorf("want silence, got %d items / %d findings", len(res.WorkItems), len(res.Findings))
	}
}

// TestComponentlessPagesCapIsReportedNotSilent — a cap that trims without saying
// so reads downstream as "that was all of them".
func TestComponentlessPagesCapIsReportedNotSilent(t *testing.T) {
	dctx, mock := newComponentlessCtx(t)
	rows := componentlessRows()
	total := componentlessMaxPerPass + 3
	for i := 0; i < total; i++ {
		rows.AddRow(uuid.New().String(), "p", "content", "/p.html", 2, "needs_rebuild", false)
	}
	mock.ExpectQuery(`FROM pages p`).WithArgs(dctx.SiteID).WillReturnRows(rows)

	res, err := (&ComponentlessPagesCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != componentlessMaxPerPass {
		t.Errorf("emitted %d, want cap %d", len(res.WorkItems), componentlessMaxPerPass)
	}
	var sawSkip bool
	for _, f := range res.Findings {
		if n, ok := f["skipped"].(int); ok && n == total-componentlessMaxPerPass {
			sawSkip = true
		}
	}
	if !sawSkip {
		t.Errorf("cap must report how many it dropped; findings = %v", res.Findings)
	}
}

// TestComponentlessPagesPredicatesArePresent is the anti-inertness test.
//
// It asserts on the SQL the check actually issues, because both guards can be
// deleted without breaking any assertion above: sqlmock returns whatever rows the
// test hands it regardless of the WHERE clause. Losing either one turns this
// check into something that files false positives (mid-build pages) or duplicates
// check_empty_sections. See this file's header.
func TestComponentlessPagesPredicatesArePresent(t *testing.T) {
	issued := componentlessPagesQuery

	for _, want := range []struct{ name, pattern string }{
		{"shipped-only guard", `deployed_at IS NOT NULL`},
		{"componentless guard", `NOT EXISTS`},
		{"componentless guard targets page_components", `FROM page_components pc WHERE pc\.page_id = p\.id`},
		{"active-only guard", `status = 'active'`},
		{"sections must be planned", `jsonb_array_length`},
	} {
		if !regexp.MustCompile(want.pattern).MatchString(issued) {
			t.Errorf("%s missing from query: /%s/ did not match", want.name, want.pattern)
		}
	}
}
