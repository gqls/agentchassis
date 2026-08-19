// FILE: platform/orchestration/actions/component_storage_identity_test.go
//
// Proves the site-aware storage-identity resolution (bugs_open/311) is both
// correct AND actually wired into StoreGeneratedComponentAction.
//
// Mutation-proof construction (the store_generated_component_source_guard
// idiom): sqlmock fails the test on any unexpected statement, so
//   - in the foreign-collision test, deleting the diversion wiring routes the
//     flow to an UPDATE of the incumbent that no expectation covers → fail;
//   - in the own-site test, a wrongly-firing diversion reaches an INSERT no
//     expectation covers → fail.
// The two tests are each other's controls: the same census with a different
// answer must flip the write path.

package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

const (
	incumbentID = "11111111-1111-1111-1111-111111111111"
	siteBID     = "22222222-2222-2222-2222-222222222222"
	siteAID     = "33333333-3333-3333-3333-333333333333"
)

// storageIdentityParams builds a generation that passes every pre-store gate,
// so the only behaviour under test is the regen-vs-create-vs-divert decision.
// The template/schema pair matches the incumbent's field set (heading), so the
// field-contract guard stays quiet on the regeneration paths.
func storageIdentityParams(db *sql.DB, siteID, domain string) ActionParams {
	inputData := map[string]interface{}{
		"section_type": "test-widget",
		"generated_template": map[string]interface{}{
			"function": "test-widget",
			"html_template": `<section class="test-widget" data-component="test-widget">` +
				`<div class="tw-inner"><h2>{{.heading}}</h2></div></section>`,
			"input_schema": map[string]interface{}{
				"fields": map[string]interface{}{
					"heading": map[string]interface{}{"type": "text", "source": "llm", "required": true},
				},
			},
		},
	}
	if siteID != "" {
		inputData["site_id"] = siteID
	}
	if domain != "" {
		inputData["domain"] = domain
	}
	return ActionParams{
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"section_type":       "input_data.section_type",
			"generated_template": "input_data.generated_template",
		}},
		CollectedData: map[string]interface{}{"input_data": inputData},
	}
}

func incumbentRow() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
		AddRow(incumbentID,
			`<section data-component="test-widget"><h2>{{.heading}}</h2></section>`,
			`{"fields":{"heading":{"type":"text","source":"llm","required":true}}}`,
			nil)
}

func aspectRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"aspect"}).AddRow("identity").AddRow("contact")
}

// A function name held by a component another site depends on must DIVERT to
// a fresh, site-suffixed base row — the incumbent is never touched. This is
// the bugs_open/311 deadlock: before this rule the store treated the foreign
// row as "the row to regenerate" and the field-contract guard refused for
// ever (remortgagecalculator.uk shipped without its calculator; loanzy.uk
// lost 7 of 7 tool sections).
func TestStoreGeneratedComponent_ForeignCollisionDivertsToCreation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1. Base lookup finds the incumbent.
	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").WillReturnRows(incumbentRow())
	// 2. Dependent census: site A (not the requester) depends on it.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(siteAID, "sitea.uk"))
	// 3. Scoped lookup: the suffixed name is free.
	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget-siteb-uk").WillReturnError(sql.ErrNoRows)
	// 4. The diversion is recorded durably.
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
	// 5. Source guard's aspect vocabulary read.
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	// 6. CREATION under the scoped identity — function AND name suffixed,
	//    section_type carrying the REQUESTED (unsuffixed) section name so the
	//    selector can resolve the requesting page's rebuild. An UPDATE of the
	//    incumbent here would be an unexpected statement → test failure.
	mock.ExpectQuery(`INSERT INTO content_components`).
		WithArgs("test-widget-siteb-uk", sqlmock.AnyArg(), "test-widget-siteb-uk", sqlmock.AnyArg(),
			"test-widget", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("44444444-4444-4444-4444-444444444444"))
	// 7. Initial version snapshot.
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	// 8. Quality persist (best-effort UPDATE of the NEW row).
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	// 9. Pages waiting on this section_type are re-armed.
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := StoreGeneratedComponentAction(context.Background(), storageIdentityParams(db, siteBID, "siteb.uk"))
	if err != nil {
		t.Fatalf("diverted store must succeed; got %v", err)
	}
	resp, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp["status"] != "created" {
		t.Errorf("diverted write must be a CREATION, got status %v", resp["status"])
	}
	if resp["function"] != "test-widget-siteb-uk" {
		t.Errorf("final function must be site-suffixed, got %v", resp["function"])
	}
	if resp["diverted_from_component_id"] != incumbentID {
		t.Errorf("response must name the incumbent it diverted from, got %v", resp["diverted_from_component_id"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// The same collision with NO foreign dependents is a normal same-site
// regeneration — the control for the test above. Also pins the section_type
// self-heal: the regen UPDATE must carry COALESCE(section_type, ...).
func TestStoreGeneratedComponent_OwnSiteCollisionRegenerates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1. Base lookup finds the row.
	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").WillReturnRows(incumbentRow())
	// 2. Census: no foreign dependents.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))
	// 3. Version read for the snapshot.
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\)`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(2))
	// 4. Source guard's aspect read.
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	// 5. Snapshot of the pre-regen state.
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	// 6. Regeneration UPDATE of the same row — and it must self-heal a NULL
	//    section_type (bugs_open/311's drift class shrinks on every regen).
	mock.ExpectExec(`UPDATE content_components[\s\S]*section_type    = COALESCE\(section_type`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "test-widget", incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 7. markPagesPendingRebuild: affected-site enumeration + page_components flip.
	mock.ExpectQuery(`SELECT DISTINCT p\.site_id::text`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id"}).AddRow(siteBID))
	mock.ExpectExec(`UPDATE page_components`).WithArgs(incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 8. One rerender work item for the affected site.
	mock.ExpectExec(`INSERT INTO site_work_items`).WillReturnResult(sqlmock.NewResult(1, 1))
	// 9. Quality persist.
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	// 10. Pages waiting on the section_type re-armed.
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := StoreGeneratedComponentAction(context.Background(), storageIdentityParams(db, siteBID, "siteb.uk"))
	if err != nil {
		t.Fatalf("own-site regeneration must succeed; got %v", err)
	}
	resp := result.(map[string]interface{})
	if resp["status"] != "regenerated" {
		t.Errorf("own-site collision must REGENERATE, got status %v", resp["status"])
	}
	if resp["function"] != "test-widget" {
		t.Errorf("own-site regeneration must keep the function name, got %v", resp["function"])
	}
	if _, present := resp["diverted_from_component_id"]; present {
		t.Error("own-site regeneration must not report a diversion")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// With no requester site_id (direct programmatic invocation) the behaviour is
// the legacy one: the function lookup's row is the regeneration target and
// the field-contract guard still refuses a stranding regen. No census query
// may run — the diversion needs a requester to compare against.
func TestStoreGeneratedComponent_UnknownRequesterKeepsLegacyRefusal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Incumbent whose schema carries a field the new generation drops.
	strandingRow := sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
		AddRow(incumbentID,
			`<section data-component="test-widget"><h2>{{.heading}}</h2><p>{{.legacy_note}}</p></section>`,
			`{"fields":{"heading":{"type":"text"},"legacy_note":{"type":"text"}}}`,
			nil)

	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").WillReturnRows(strandingRow)
	// NO census expectation: a census here is an unexpected statement → fail.
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\)`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1))
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	// The field-contract refusal is recorded; no content_components write.
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = StoreGeneratedComponentAction(context.Background(), storageIdentityParams(db, "", ""))
	if err == nil {
		t.Fatal("stranding regeneration with unknown requester must still be REFUSED; got nil error")
	}
	if !strings.Contains(err.Error(), "regeneration removes/renames") {
		t.Errorf("refusal must come from the field-contract guard; got %q", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// The pathological tail: the base name AND the site-scoped name are both
// depended on by other sites. Refuse loudly, never loop through suffixes.
func TestResolveStorageIdentity_DoubleCollisionRefuses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	scopedID := "55555555-5555-5555-5555-555555555555"

	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").WillReturnRows(incumbentRow())
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(siteAID, "sitea.uk"))
	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget-siteb-uk").
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
			AddRow(scopedID, "<section></section>", "{}", nil))
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(scopedID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow("66666666-6666-6666-6666-666666666666", "sitec.uk"))

	_, err = resolveStorageIdentity(context.Background(), db, "test-widget", siteBID, "siteb.uk", zap.NewNop())
	if err == nil {
		t.Fatal("double collision must refuse; got nil error")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite either") {
		t.Errorf("refusal must name the double collision; got %q", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}
