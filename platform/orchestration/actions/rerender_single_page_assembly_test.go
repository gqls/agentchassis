// FILE: platform/orchestration/actions/rerender_single_page_assembly_test.go
//
// bugs_open/095 — "a wrong slot_name renders nothing and the run reports
// COMPLETED".
//
// An empty assembled page was ambiguous: a page with nothing built yet and a
// page whose every component failed to render both produced the same empty
// string, and both returned skipped=true. page-rerender's check_skipped
// conditional routes that to complete_skipped — a terminal step whose name
// contains "complete" and whose status is COMPLETED — so the second case was
// shaped exactly like a success, with no error recorded on the orchestration,
// the work item or the page row.
//
// These tests pin the discrimination. The blast-radius one is
// NoComponentRows_IsALegitimateSkip: measured 2026-07-27, 17 active pages
// fleet-wide have planned sections and zero component rows (5 never built), so
// treating that as a failure would convert a correct no-op into a fleet-wide
// error. It must stay a skip.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// expectAssemblyQueries wires the two reads getPageSections makes: the page's
// planned sections, then its component rows.
func expectAssemblyQueries(mock sqlmock.Sqlmock, sectionsJSON string, componentRows *sqlmock.Rows) {
	mock.ExpectQuery("SELECT COALESCE\\(sections").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(sectionsJSON)))
	mock.ExpectQuery("FROM page_components").
		WillReturnRows(componentRows)
}

func componentRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"rendered_html", "slot_name"})
}

// TestAssembly_UnrenderedComponentsAreADefect is the regression: the row shape
// bugs_open/095 opens with — a component row that looks deliberate but that
// nothing ever populated. Assembly is empty AND rows exist, so the caller must
// be able to tell this is a defect.
func TestAssembly_UnrenderedComponentsAreADefect(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectAssemblyQueries(mock, `["hero-about","about-content"]`,
		componentRows().AddRow("", "main").AddRow("", "hero-about"))

	html, diag, err := getPageSections(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("getPageSections: %v", err)
	}
	if html != "" {
		t.Fatalf("expected empty assembly, got %d bytes", len(html))
	}
	if !diag.assembledToNothingDespiteComponents() {
		t.Errorf("expected the defect shape; got rows=%d contributed=%d",
			diag.ComponentRows, diag.Contributed)
	}
	if diag.ComponentRows != 2 {
		t.Errorf("ComponentRows = %d, want 2", diag.ComponentRows)
	}
	if len(diag.UnrenderedSlots) != 2 {
		t.Errorf("UnrenderedSlots = %v, want both slots", diag.UnrenderedSlots)
	}
	// The operator needs both lists — what was wanted and what was there.
	if d := diag.describe(); d == "" ||
		!strings.Contains(d, "hero-about") || !strings.Contains(d, "main") {
		t.Errorf("describe() does not name both lists: %q", d)
	}
}

// TestAssembly_NoComponentRows_IsALegitimateSkip is the blast-radius test. A
// page that has never been built has planned sections and no rows; that is not
// a defect and must not become one.
func TestAssembly_NoComponentRows_IsALegitimateSkip(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectAssemblyQueries(mock, `["hero","generic-text-block"]`, componentRows())

	_, diag, err := getPageSections(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("getPageSections: %v", err)
	}
	if diag.assembledToNothingDespiteComponents() {
		t.Error("a page with no component rows must remain a legitimate skip, not a defect")
	}
	if len(diag.PlannedSections) != 2 {
		t.Errorf("PlannedSections = %v, want the two planned sections", diag.PlannedSections)
	}
}

// TestAssembly_BlankComponentsAreADefect: rows that carry HTML but no visible
// content are dropped by sectionHasVisibleContent. If EVERY row drops, the page
// assembles to nothing — same defect, different cause, and it must be caught
// too. This is the "9 blanked article bodies" shape.
func TestAssembly_BlankComponentsAreADefect(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectAssemblyQueries(mock, `["hero"]`,
		componentRows().AddRow("<section><h1></h1></section>", "hero"))

	html, diag, err := getPageSections(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("getPageSections: %v", err)
	}
	if html != "" {
		t.Fatalf("expected empty assembly, got %q", html)
	}
	if !diag.assembledToNothingDespiteComponents() {
		t.Error("a page whose every section is blank must be a defect, not a skip")
	}
	if len(diag.BlankSlots) != 1 || diag.BlankSlots[0] != "hero" {
		t.Errorf("BlankSlots = %v, want [hero]", diag.BlankSlots)
	}
}

// TestAssembly_HealthyPageUnaffected is the other half of the blast radius:
// removing the SQL pre-filter on rendered_html must not change what a normal
// page assembles to. A partially-blank page still contributes its good sections.
func TestAssembly_HealthyPageUnaffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	good := "<section><h1>Restructuring plans that actually complete</h1></section>"
	expectAssemblyQueries(mock, `["hero","about-content","cta"]`,
		componentRows().
			AddRow(good, "hero").
			AddRow("", "about-content").                  // never rendered
			AddRow("<section><p> </p></section>", "cta")) // blank

	html, diag, err := getPageSections(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("getPageSections: %v", err)
	}
	if diag.assembledToNothingDespiteComponents() {
		t.Error("a page with one good section is not the defect shape")
	}
	if diag.Contributed != 1 {
		t.Errorf("Contributed = %d, want 1", diag.Contributed)
	}
	if !strings.Contains(html, "Restructuring plans") {
		t.Errorf("the good section did not reach the output: %q", html)
	}
	if len(diag.UnrenderedSlots) != 1 || len(diag.BlankSlots) != 1 {
		t.Errorf("unrendered=%v blank=%v, want one of each",
			diag.UnrenderedSlots, diag.BlankSlots)
	}
}
