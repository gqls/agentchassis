// FILE: platform/orchestration/actions/discovery_checks/check_directory_test.go
//
// The property that matters: these checks must stay silent until BOTH a
// site has opted in AND the global register actually has publishable data
// OF THE RIGHT KIND — raising a work item for an empty section would push
// content-gap-planner to build a section with nothing to show. Each gate is
// tested in isolation so a future change to one can't silently widen or
// narrow the other.

package discovery_checks

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// testProfile resolves a register kind to its check profile, failing loudly
// rather than returning a zero profile — a zero profile would make every
// query match nothing and the tests would pass for the wrong reason.
func testProfile(kind string) directoryCheckProfile {
	for _, p := range directoryCheckProfiles {
		if p.Kind == kind {
			return p
		}
	}
	panic("no directory check profile for kind " + kind)
}

func TestMissingModelDirectorySectionCheck_NotOptedIn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).WillReturnError(sql.ErrNoRows)

	res, err := (&MissingDirectorySectionCheck{profile: testProfile("model")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items when the site hasn't opted in, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectorySectionCheck_OptedInButRegistryEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`FROM directory_claims`).WithArgs("model").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	res, err := (&MissingDirectorySectionCheck{profile: testProfile("model")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Opted in, but discovery hasn't produced anything yet — must wait, not
	// build an empty section.
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items with an empty registry, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectorySectionCheck_RaisesFindingWhenDataExistsButNoSection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`FROM directory_claims`).WithArgs("model").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM page_components`).WithArgs(siteID, "model-directory").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	res, err := (&MissingDirectorySectionCheck{profile: testProfile("model")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("expected exactly one work item, got %d", len(res.WorkItems))
	}
	item := res.WorkItems[0]
	if item.ItemType != "missing_model_directory_section" {
		t.Errorf("item_type = %q, want missing_model_directory_section", item.ItemType)
	}
	if item.HandlerAgent != "content-gap-planner" {
		t.Errorf("handler = %q, want content-gap-planner", item.HandlerAgent)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectorySectionCheck_SilentWhenSectionAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`FROM directory_claims`).WithArgs("model").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM page_components`).WithArgs(siteID, "model-directory").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	res, err := (&MissingDirectorySectionCheck{profile: testProfile("model")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items when the section already exists, got %d", len(res.WorkItems))
	}
}

func TestMissingModelDirectoryPageCheck_RequiresSeparatePageFlag(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	// recommended=true but separate_page is absent (defaults false) — must
	// stay silent, same as missing_news_page's identical gate.
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))

	res, err := (&MissingDirectoryPageCheck{profile: testProfile("model")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items without separate_page=true, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectoryPageCheck_RaisesFinding(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true,"separate_page":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`FROM directory_claims`).WithArgs("model").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM pages`).WithArgs(siteID, "model-directory").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT domain FROM sites`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("ai-agent-orchestration.com"))

	res, err := (&MissingDirectoryPageCheck{profile: testProfile("model")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("expected exactly one work item, got %d", len(res.WorkItems))
	}
	if res.WorkItems[0].ItemType != "missing_model_directory_page" {
		t.Errorf("item_type = %q, want missing_model_directory_page", res.WorkItems[0].ItemType)
	}
}

// ===========================================================================
// Phase E — the kind gate
// ===========================================================================
// The failure this prevents: the register is shared, so a site that opts
// into the adoption tracker while the register holds ONLY model claims must
// stay silent. An earlier draft counted claims without joining
// directory_entities, which would have raised a "build an adoption tracker
// section" item whose section would render empty — the exact
// build-something-with-nothing-in-it outcome the claim-count gate exists to
// prevent, reintroduced through the back door.

func TestMissingDirectorySectionCheck_KindGate_SilentWhenOnlyOtherKindsHaveData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"adoption_tracker":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	// The count query is kind-scoped; with only model claims in the register
	// the 'company' count is 0.
	mock.ExpectQuery(`FROM directory_claims`).WithArgs("company").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	res, err := (&MissingDirectorySectionCheck{profile: testProfile("company")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items when no company claims exist, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingDirectorySectionCheck_KindGate_ReadsItsOwnSpecKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	// Opted into the MODEL directory only. The adoption-tracker check must
	// not read model_directory's flag as its own.
	specJSON := `{"content_features":{"model_directory":{"recommended":true,"separate_page":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))

	res, err := (&MissingDirectorySectionCheck{profile: testProfile("company")}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("adoption-tracker check fired on the model_directory opt-in flag: %d items", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Every profile must be distinct in the fields that key a work item or a
// query. A copy-paste slip (two profiles sharing an ItemKey prefix, a
// component function, or a page_type) would make one register's checks
// silently satisfy the other's — and because the checks only ever fire on
// opted-in sites, that would surface as "the tracker never appears" on a
// site nobody is watching.
func TestDirectoryCheckProfilesAreDistinct(t *testing.T) {
	seen := map[string]string{}
	claim := func(field, value, kind string) {
		if value == "" {
			t.Errorf("profile %q has an empty %s", kind, field)
			return
		}
		key := field + "=" + value
		if prev, dup := seen[key]; dup {
			t.Errorf("profiles %q and %q share %s %q", prev, kind, field, value)
			return
		}
		seen[key] = kind
	}
	for _, p := range directoryCheckProfiles {
		claim("kind", p.Kind, p.Kind)
		claim("spec_key", p.SpecKey, p.Kind)
		claim("snippet_component", p.SnippetComponent, p.Kind)
		claim("listing_component", p.ListingComponent, p.Kind)
		claim("page_type", p.PageType, p.Kind)
		claim("section_item_type", p.SectionItemType, p.Kind)
		claim("page_item_type", p.PageItemType, p.Kind)
	}
}
