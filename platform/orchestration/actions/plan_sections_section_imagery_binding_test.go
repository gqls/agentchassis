// FILE: platform/orchestration/actions/plan_sections_section_imagery_binding_test.go
//
// A page whose sections each declare their own figure — the "one illustrated
// block per h3" shape the guides need — could not be given a different image
// per section: ensureAssets folded every section-scope site_plan_imagery row
// for the page into ONE flat map keyed by kind (first-wins), and resolve() took
// a source string with no section identity, so all N sections resolved the same
// URL. The scope_ref ordinal that says WHICH section a figure belongs to was
// read only inside the LIKE filter and then discarded.
//
// Both tests here are needed, and for different reasons:
//
//   - the binding test fails on the old code by returning the SAME url twice,
//     which is the defect stated as an assertion rather than as prose;
//   - the inertness test is the negative control the council gate asks for: a
//     page with ONE section-scope row must resolve exactly as it did before,
//     because that is every page in the estate that carries one today. A fix
//     that quietly changed those would be indistinguishable from this one
//     without it.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/gqls/agentchassis/platform/storage"
)

// illustratedBlockSchema mirrors the live `illustrated-text-block` component
// (content_components.function='illustrated-text-block', migration 644): prose
// from the writer, the figure from the resolver, and the figure optional so a
// section with no plan row renders as plain prose.
const illustratedBlockSchema = `{"fields":{
    "heading":   {"type":"text","source":"llm","required":true},
    "content":   {"type":"html","source":"llm","required":true},
    "image_url": {"type":"url","source":"site_assets.illustration","required":false,"on_missing":"skip_field"}
}}`

// expectIllustratedComponents wires the component load for a page built from
// repeated illustrated blocks. The section name IS the component name.
func expectIllustratedComponents(mock sqlmock.Sqlmock, componentID string) {
	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRowWithSchema(componentID, "illustrated-text-block", "illustrated-text-block",
			`<section>{{.content}}{{if .image_url}}<img src="{{.image_url}}">{{end}}</section>`,
			"section", illustratedBlockSchema))
}

// expectAssetLookups wires every query ensureAssets makes, with the
// section-scope arm returning the supplied rows. The four site_plan_imagery
// reads are told apart by their scope/kind literals rather than by order,
// because the field map planSection iterates is unordered.
func expectAssetLookups(mock sqlmock.Sqlmock, sectionRows *sqlmock.Rows, planOrder ...string) {
	// per-page hero, content hero, site-scope hero, logo: all empty.
	mock.ExpectQuery(`spi\.scope = 'page'`).
		WillReturnRows(sqlmock.NewRows([]string{"asset_key", "purpose"}))
	mock.ExpectQuery(`FROM assets a`).
		WillReturnRows(sqlmock.NewRows([]string{"asset_key", "purpose"}))
	mock.ExpectQuery(`spi\.scope = 'site'[\s\S]*spi\.kind = 'hero'`).
		WillReturnRows(sqlmock.NewRows([]string{"asset_key", "purpose"}))
	mock.ExpectQuery(`spi\.scope = 'section'`).WillReturnRows(sectionRows)
	mock.ExpectQuery(`spi\.scope = 'site'[\s\S]*spi\.kind = 'logo'`).
		WillReturnRows(sqlmock.NewRows([]string{"asset_key", "purpose"}))
	mock.ExpectQuery("FROM sites").WillReturnRows(sqlmock.NewRows([]string{"content_data"}))
	// The plan's section order — read only when the page HAS section figures,
	// and the only thing that can turn a scope_ref ordinal back into a section.
	order := sqlmock.NewRows([]string{"component_name"})
	for _, name := range planOrder {
		order.AddRow(name)
	}
	mock.ExpectQuery("FROM site_plan_sections").WillReturnRows(order)
	// The locked-slot ask (RFC_033 / LOCK-008): no human-pinned rows here, so
	// the plan's list IS the page's list and the binding stands.
	mock.ExpectQuery("lock_type").WillReturnRows(sqlmock.NewRows([]string{
		"row_id", "page_name", "slot", "position", "component_id",
		"component_function", "component_name", "lock_type", "locked_by"}))
}

// sectionImageryRows is the shape of the section-scope arm's result set. The
// ordinal travels in scope_ref — the column the old query never selected.
func sectionImageryRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"kind", "scope_ref", "asset_key", "purpose"})
}

// TestPlanSections_SectionScopeIllustrationBindsToItsOwnSection is the
// motivating case: two illustrated blocks, two illustrations, one per section.
func TestPlanSections_SectionScopeIllustrationBindsToItsOwnSection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectIllustratedComponents(mock, componentID)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectPostLoopReads(mock)
	expectAssetLookups(mock, sectionImageryRows().
		AddRow("illustration", "grip-styles:0", "illustration_ring_grip", "illustration").
		AddRow("illustration", "grip-styles:1", "illustration_shark_grip", "illustration"),
		"illustrated-text-block", "illustrated-text-block")
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "grip-styles",
			[]string{"illustrated-text-block", "illustrated-text-block"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 2 {
		t.Fatalf("expected two ready sections, got %d", len(items))
	}

	want := []string{
		storage.DeployedWebPath("illustration_ring_grip", "illustration"),
		storage.DeployedWebPath("illustration_shark_grip", "illustration"),
	}
	for i, item := range items {
		got, ok := item.ResolvedData["image_url"].(string)
		if !ok {
			t.Fatalf("section %d: image_url absent from resolved_data (%v)", i, item.ResolvedData)
		}
		if got != want[i] {
			t.Errorf("section %d: image_url = %q, want %q", i, got, want[i])
		}
	}
	if items[0].ResolvedData["image_url"] == items[1].ResolvedData["image_url"] {
		t.Errorf("both sections resolved the same figure (%v) — the ordinal was discarded",
			items[0].ResolvedData["image_url"])
	}
}

// TestPlanSections_SingleSectionScopeRowStillResolvesPageWide is the negative
// control. Every page in the estate carrying section-scope imagery today has at
// most one row of a given kind, and its figure is reached through the kind
// alias from whichever section declares it — NOT only the section the ordinal
// names. That behaviour must not change, or this fix silently blanks live pages
// instead of serving the ones it was written for.
func TestPlanSections_SingleSectionScopeRowStillResolvesPageWide(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectIllustratedComponents(mock, componentID)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectPostLoopReads(mock)
	// ONE row, naming ordinal 0, on a page with two illustrated sections. The
	// section it names takes it; the section it does not name must still reach
	// it through the page-wide kind alias, exactly as before.
	expectAssetLookups(mock, sectionImageryRows().
		AddRow("illustration", "about:0", "illustration_people_feature", "illustration"),
		"illustrated-text-block", "illustrated-text-block")
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "about",
			[]string{"illustrated-text-block", "illustrated-text-block"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 2 {
		t.Fatalf("expected two ready sections, got %d", len(items))
	}
	want := storage.DeployedWebPath("illustration_people_feature", "illustration")
	for i, item := range items {
		if got, _ := item.ResolvedData["image_url"].(string); got != want {
			t.Errorf("section %d: page-wide kind alias regressed: image_url = %q, want %q", i, got, want)
		}
	}
}
