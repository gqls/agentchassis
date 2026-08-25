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
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").WillReturnRows(incumbentRow())
	// 2. Dependent census: site A (not the requester) depends on it.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(siteAID, "sitea.uk"))
	// 3. Scoped lookup: the suffixed name is free.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget-siteb-uk").WillReturnError(sql.ErrNoRows)
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
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").WillReturnRows(incumbentRow())
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

	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").WillReturnRows(strandingRow)
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

	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").WillReturnRows(incumbentRow())
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(siteAID, "sitea.uk"))
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget-siteb-uk").
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

// ─────────────────────────────────────────────────────────────────────────────
// bugs_open/388 — the store honours the identity the ADVISORY resolved.
//
// Until this existed, which row a regeneration overwrote was decided by the
// function name the LLM wrote into its own output, while the field contract the
// writer was told to preserve came from a different resolver. The two were
// joined by a sentence in the prompt asking the model to echo a name back.
//
// These four tests are each other's controls: the same generation, with and
// without a pin, and with the pin resolvable and not, must take different write
// paths. Each names the mutation that must turn it red — a mock's own
// bookkeeping cannot assert a negative.
// ─────────────────────────────────────────────────────────────────────────────

const advisedRowID = "55555555-5555-5555-5555-555555555555"

// pinnedParams is storageIdentityParams plus the advisory's answer in
// collected_data, wired through the same OPTIONAL-EXPLICIT config key migration
// 611 writes. emittedFunction is what the MODEL chose, which is deliberately
// allowed to disagree with the pin.
func pinnedParams(db *sql.DB, componentID, advisedFunction, emittedFunction string) ActionParams {
	params := storageIdentityParams(db, siteBID, "siteb.uk")
	input := params.CollectedData["input_data"].(map[string]interface{})
	tpl := input["generated_template"].(map[string]interface{})
	tpl["function"] = emittedFunction

	params.CollectedData["existing_component"] = map[string]interface{}{
		"component_id": componentID,
		"function":     advisedFunction,
		"field_names":  "heading",
		"field_count":  1,
	}
	params.StepConfig.Config["advised_identity?"] = "existing_component"
	return params
}

func advisedRow(function string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content", "function"}).
		AddRow(advisedRowID,
			`<section data-component="test-widget"><h2>{{.heading}}</h2></section>`,
			`{"fields":{"heading":{"type":"text","source":"llm","required":true}}}`,
			nil, function)
}

// T-A — THE PIN DECIDES THE WRITE TARGET, NOT THE MODEL.
//
// The writer emits "test-widget-renamed" after being pinned to "test-widget"
// (component advisedRowID). The write must land on the PINNED row. There is
// deliberately NO expectation for a by-function lookup anywhere in this test:
// if the store falls back to resolving the emitted name, it reaches a statement
// nothing covers and sqlmock fails.
//
// MUTATION: replace the pinned branch with the legacy
// resolveStorageIdentity(functionName, ...) call -> the by-function query is
// unexpected and this fails. Second mutation: delete the divergence
// LogActionFindings call -> the agent_error_log INSERT goes unmatched.
func TestStoreGeneratedComponent_AdvisedIdentityPinDecidesTheWriteTarget(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1. The pinned row, BY ID — never by the name the model chose.
	mock.ExpectQuery(`WHERE id = \$1::uuid AND forked_from IS NULL`).WithArgs(advisedRowID).
		WillReturnRows(advisedRow("test-widget"))
	// 2. bugs_open/311's census still runs over the pinned row.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(advisedRowID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))
	// 3. The disagreement is recorded — harmless, and therefore the honest meter.
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
	// 4. Regeneration of the PINNED row.
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\)`).WithArgs(advisedRowID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1))
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE content_components`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "test-widget", advisedRowID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p\.site_id::text`).WithArgs(advisedRowID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id"}).AddRow(siteBID))
	mock.ExpectExec(`UPDATE page_components`).WithArgs(advisedRowID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO site_work_items`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := StoreGeneratedComponentAction(context.Background(),
		pinnedParams(db, advisedRowID, "test-widget", "test-widget-renamed"))
	if err != nil {
		t.Fatalf("pinned store must succeed; got %v", err)
	}
	resp := result.(map[string]interface{})
	if resp["status"] != "regenerated" {
		t.Errorf("the pinned row must be REGENERATED, got status %v", resp["status"])
	}
	if resp["component_id"] != advisedRowID {
		t.Errorf("the write must land on the ADVISED row; got %v", resp["component_id"])
	}
	if resp["function"] != "test-widget" {
		t.Errorf("the pinned row's function must win over the model's; got %v", resp["function"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T-B — A PIN THAT NO LONGER RESOLVES FALLS BACK, AND SAYS SO.
//
// The advised row was deleted or forked between advice and store. Falling back
// is right; falling back SILENTLY is not, because that race is otherwise
// invisible and indistinguishable from the pin never having been wired.
//
// MUTATION: remove the fallback -> no write path has expectations. Remove the
// COMPONENT_ADVISED_ROW_VANISHED finding -> the first agent_error_log INSERT is
// unmatched.
func TestStoreGeneratedComponent_VanishedPinFallsBackAndSaysSo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`WHERE id = \$1::uuid AND forked_from IS NULL`).WithArgs(advisedRowID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
	// Legacy resolution resumes, keyed on the name the model emitted.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").
		WillReturnRows(incumbentRow())
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\)`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(0))
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE content_components`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "test-widget", incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p\.site_id::text`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id"}))
	mock.ExpectExec(`UPDATE page_components`).WithArgs(incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := StoreGeneratedComponentAction(context.Background(),
		pinnedParams(db, advisedRowID, "test-widget", "test-widget")); err != nil {
		t.Fatalf("a vanished pin must degrade, not fail; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T-C — THE 311 COMPOSITION PROOF, and the reason the diversion decision was
// SPLIT rather than copied.
//
// A pinned row that other sites depend on must STILL divert. If the pin were
// allowed to skip the census, bugs_open/311 would silently reopen: one site's
// build would overwrite a component another site is serving. There is
// deliberately no expectation for an UPDATE of the pinned row.
//
// MUTATION: set IsRegeneration = true on the pinned path without calling
// decideStorageIdentity -> the incumbent UPDATE hits no expectation and this
// fails. That is a real negative proven by mutation, not by counting calls.
func TestStoreGeneratedComponent_PinnedRowForeignDependentsStillDivert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`WHERE id = \$1::uuid AND forked_from IS NULL`).WithArgs(advisedRowID).
		WillReturnRows(advisedRow("test-widget"))
	// A foreign site depends on the pinned row.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(advisedRowID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(siteAID, "sitea.uk"))
	// The scoped name is free -> creation under it.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget-siteb-uk").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	mock.ExpectQuery(`INSERT INTO content_components`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("66666666-6666-6666-6666-666666666666"))
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := StoreGeneratedComponentAction(context.Background(),
		pinnedParams(db, advisedRowID, "test-widget", "test-widget"))
	if err != nil {
		t.Fatalf("pinned diverted store must succeed; got %v", err)
	}
	resp := result.(map[string]interface{})
	if resp["status"] != "created" {
		t.Errorf("a foreign-depended pinned row must DIVERT to a creation, got %v", resp["status"])
	}
	if resp["function"] != "test-widget-siteb-uk" {
		t.Errorf("the diverted identity must be site-scoped, got %v", resp["function"])
	}
	if resp["diverted_from_component_id"] != advisedRowID {
		t.Errorf("the diversion must name the PINNED row it steered away from, got %v",
			resp["diverted_from_component_id"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T-D — THE SILENT HALF, MADE LOUD.
//
// An UNPINNED creation on a section_type that already has an active component
// is the duplicate the field-contract guard cannot see: isRegeneration is false,
// so the guard is vacuous and the parallel row is born with no error and no work
// item. Two such pairs exist in the live library from the generated route.
//
// This is unrepresentable on the pinned path, so it can only fire on the
// fail-open residual — which is exactly what makes it worth counting.
//
// MUTATION: delete the census block -> the census query and the finding INSERT
// both go unmatched.
func TestStoreGeneratedComponent_SilentParallelBirthIsRecorded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// No pin at all, and no row under the emitted function -> creation.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	mock.ExpectQuery(`INSERT INTO content_components`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("77777777-7777-7777-7777-777777777777"))
	// The census that the resolver did NOT ask: by section_type, not by name.
	mock.ExpectQuery(`SELECT count\(\*\), string_agg\(function`).
		WithArgs("test-widget", "77777777-7777-7777-7777-777777777777").
		WillReturnRows(sqlmock.NewRows([]string{"count", "string_agg"}).AddRow(1, "test-widget-incumbent"))
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := StoreGeneratedComponentAction(context.Background(),
		storageIdentityParams(db, siteBID, "siteb.uk")); err != nil {
		t.Fatalf("the census is best-effort and must not fail the store; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T-E — THE UN-PINNED REGENERATION SAYS WHEN IT WAS A GUESS.
//
// Added in council round 2 on bug_historian's advisory objection: round 1
// instrumented the un-pinned CREATE path and left the un-pinned REGENERATE path
// silent. Regeneration is the worse of the two — an existing row is OVERWRITTEN,
// and which one was decided by `ORDER BY is_active DESC, updated_at DESC LIMIT 1`
// over a name that 25 of 330 non-forked rows do not hold uniquely.
//
// ⚠ THIS TEST EXISTS BECAUSE THE OTHER TESTS COULD NOT CATCH IT. The census is
// best-effort by design — a census failure must never fail a store that would
// otherwise succeed — so under sqlmock an UNEXPECTED census query returns an
// error, the code logs and returns, and ExpectationsWereMet() still passes. The
// swallow that is correct in production makes every neighbouring test blind to
// this code path. Only a POSITIVE expectation proves it runs.
//
// MUTATION: delete the reportAmbiguousUnpinnedRegeneration call -> the census
// query and the agent_error_log INSERT both go unmatched and this fails.
func TestStoreGeneratedComponent_UnpinnedRegenerationReportsAmbiguity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// No pin. The legacy by-function lookup finds the incumbent.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").
		WillReturnRows(incumbentRow())
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))
	// THE QUESTION THE RESOLVER DISCARDED: how many rows shared that name?
	mock.ExpectQuery(`SELECT count\(\*\)\s+FROM content_components\s+WHERE function = \$1 AND forked_from IS NULL`).
		WithArgs("test-widget").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\)`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(0))
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE content_components`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "test-widget", incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p\.site_id::text`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id"}))
	mock.ExpectExec(`UPDATE page_components`).WithArgs(incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := StoreGeneratedComponentAction(context.Background(),
		storageIdentityParams(db, siteBID, "siteb.uk")); err != nil {
		t.Fatalf("an ambiguous regeneration must still succeed; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T-F — the DECISION, tested directly, because through the mock it cannot be.
//
// ⚠ THE FIRST VERSION OF THIS TEST WAS VACUOUS AND THE MUTATION PROVED IT.
// It drove the whole action with the census returning 1 and declared no
// agent_error_log expectation, on the assumption that a spurious finding would
// surface as an unexpected statement. It does not: LogActionFindings is
// best-effort and swallows the write error, so ExpectationsWereMet() passes
// either way. Removing the `siblings <= 1` gate left that test GREEN. A positive
// expectation can prove a finding fires; no arrangement of expectations proves
// one does not — so the predicate is a pure function and is asserted here.
//
// MUTATION: drop the `siblings <= 1` early return in
// ambiguousRegenerationFinding -> the first sub-case fails.
func TestAmbiguousRegenerationFinding_OnlyFiresOnARealContest(t *testing.T) {
	ident := storageIdentity{FunctionName: "site-footer", ExistingID: incumbentID, IsRegeneration: true}

	for _, tc := range []struct {
		name     string
		siblings int
		want     bool
	}{
		{"one row holds the name — no contest, stay silent", 1, false},
		{"zero (census raced the write) — nothing to report", 0, false},
		{"two rows share it — the winner was picked by recency", 2, true},
		{"five, the live worst case for site-footer/site-header", 5, true},
	} {
		finding, got := ambiguousRegenerationFinding(ident, tc.siblings, "site-footer", "footer")
		if got != tc.want {
			t.Errorf("%s: siblings=%d reported=%v, want %v", tc.name, tc.siblings, got, tc.want)
		}
		if !got {
			continue
		}
		if finding.ErrorCode != "COMPONENT_UNPINNED_REGENERATION_AMBIGUOUS" {
			t.Errorf("wrong code: %s", finding.ErrorCode)
		}
		// The count is the whole point — a finding that does not carry it cannot
		// be triaged, because the reader cannot tell a 2-way tie from a 5-way one.
		if finding.Context["rows_sharing_function"] != tc.siblings {
			t.Errorf("finding must carry the sibling count; got %v", finding.Context["rows_sharing_function"])
		}
		if finding.Context["chosen_component_id"] != incumbentID {
			t.Errorf("finding must name the row that was overwritten; got %v", finding.Context["chosen_component_id"])
		}
	}
}

// T-G — the wired quiet case. Retained because it still pins the CENSUS running
// on the un-pinned regeneration path (a positive expectation, which does work);
// it deliberately no longer claims to prove the absence of a finding — T-F does
// that. Keeping the two apart is the point: this one would pass either way.
func TestStoreGeneratedComponent_UnambiguousRegenerationStillCensuses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("test-widget").
		WillReturnRows(incumbentRow())
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(incumbentID, siteBID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))
	// Exactly one row holds the name -> nothing to report.
	mock.ExpectQuery(`SELECT count\(\*\)\s+FROM content_components\s+WHERE function = \$1 AND forked_from IS NULL`).
		WithArgs("test-widget").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// NO agent_error_log expectation: a finding here would be an unexpected statement.

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\)`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(0))
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).WillReturnRows(aspectRows())
	mock.ExpectExec(`INSERT INTO component_versions`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE content_components`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "test-widget", incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p\.site_id::text`).WithArgs(incumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id"}))
	mock.ExpectExec(`UPDATE page_components`).WithArgs(incumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE pages SET build_status`).WithArgs("test-widget").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := StoreGeneratedComponentAction(context.Background(),
		storageIdentityParams(db, siteBID, "siteb.uk")); err != nil {
		t.Fatalf("store must succeed; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}
