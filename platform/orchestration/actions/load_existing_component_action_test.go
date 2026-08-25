// FILE: platform/orchestration/actions/load_existing_component_action_test.go
//
// Proves the pre-generation advisory tells the writer what the birth gate will
// actually enforce (bugs_open/337). Until this file existed the action had NO
// tests at all, which is part of how it went three predicates out of step with
// the guard without anyone noticing.
//
// Mutation-proof construction (the component_storage_identity idiom): sqlmock
// fails on any unexpected statement, so a fallback that fires when it should
// not — or does not fire when it should — reaches a query no expectation covers
// and fails. The vocabulary tests are each other's controls: the same run with
// the aspect read succeeding and failing must differ in exactly one key.

package actions

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

const (
	advisoryBaseID  = "aaaaaaaa-1111-1111-1111-111111111111"
	advisorySiteID  = "bbbbbbbb-2222-2222-2222-222222222222"
	advisoryOtherID = "cccccccc-3333-3333-3333-333333333333"
)

// advisoryParams builds the action's inputs. siteID/domain are the two
// collected_data paths the STORE reads, so supplying them here is what lets the
// fallback reproduce the store's own identity resolution.
func advisoryParams(db *sql.DB, sectionType, siteID, domain string) ActionParams {
	inputData := map[string]interface{}{
		"spec": map[string]interface{}{"section_type": sectionType},
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
		ExecutionContext: &types.ExecutionContext{Action: "execute"},
		StepConfig: models.Step{Config: map[string]interface{}{
			"section_type": "input_data.spec.section_type",
		}},
		CollectedData: map[string]interface{}{"input_data": inputData},
	}
}

// expectVocabulary sets up the two vocabulary reads that now ride EVERY call.
// They run before the field lookup, so they are declared first.
func expectVocabulary(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).
		WillReturnRows(sqlmock.NewRows([]string{"aspect"}).AddRow("cta").AddRow("identity"))
	mock.ExpectQuery(`SELECT ss\.aspect, k\.key`).
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "key", "sites"}).
			AddRow("cta", "primary_url", 7).
			AddRow("identity", "company_name", 26))
}

// expectIdentityDecision sets up the pair of reads that now follow a section_type
// HIT: the store's own by-id lookup, then the dependent census (bugs_open/388).
// Before 388 a hit returned immediately and neither ran — which is exactly why
// the advisory could name a row the store would divert away from.
func expectIdentityDecision(mock sqlmock.Sqlmock, rowID, function, schema string) {
	mock.ExpectQuery(`WHERE id = \$1::uuid AND forked_from IS NULL`).WithArgs(rowID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content", "function"}).
			AddRow(rowID, "<section></section>", schema, nil, function))
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(rowID, advisorySiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))
}

func advisoryResult(t *testing.T, res interface{}) map[string]interface{} {
	t.Helper()
	m, ok := res.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type %T", res)
	}
	return m
}

// T1 — RENAMED AND REWRITTEN 2026-08-25 (bugs_open/388). It was
// TestLoadExistingComponent_SectionTypeHitDoesNotConsultTheResolver, and it
// asserted the DEFECT: that a section_type hit returned immediately without
// consulting the store's identity decision. That is precisely what let the
// advisory name one row while the store wrote another.
//
// It is rewritten rather than deleted, with this note, because a peer lane
// holding the old test as evidence must find the correction — a silently
// removed test is indistinguishable from one that never existed.
//
// What survives unchanged: the section_type query is still PRIMARY and still
// the selector's own (component_selector.go). Only what happens to its RESULT
// changed — it is now handed to the store's own decision, and the row that
// comes back is what gets advised, carrying its id.
func TestLoadExistingComponent_SectionTypeHitRoutesThroughTheIdentityDecision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("hero-banner").
		WillReturnRows(sqlmock.NewRows([]string{"id", "function", "input_schema"}).
			AddRow(advisoryBaseID, "hero-banner", `{"fields":{"heading":{},"eyebrow":{}}}`))
	expectIdentityDecision(mock, advisoryBaseID, "hero-banner", `{"fields":{"heading":{},"eyebrow":{}}}`)

	out := advisoryResult(t, mustLoad(t, db, "hero-banner", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "eyebrow, heading" {
		t.Errorf("field_names must come from the resolved row, sorted; got %v", out["field_names"])
	}
	if out["function"] != "hero-banner" {
		t.Errorf("function pin must be the resolved row's; got %v", out["function"])
	}
	// THE POINT OF THE REWRITE. Without an id the store has nothing to honour
	// and falls back to the name the LLM wrote — which is the bug.
	if out["component_id"] != advisoryBaseID {
		t.Errorf("the advisory must pin the row BY ID; got %v", out["component_id"])
	}
	// MUTATION THAT MUST TURN THIS RED: return the advice straight from the
	// primary query without resolveStorageIdentityByID — the by-id lookup and
	// the census expectations then go unmatched.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T1b — a resolved row with NO schema fields keeps its IDENTITY (bugs_open/388).
//
// This is the hole that made the whole bridge conditional on the thing it
// protects: the prompt gates its function pin on field_names, so a row that
// resolved with an empty input_schema was picked by the store and never named
// to the writer. 5 of 154 active section rows were in that state on 2026-08-25;
// 4 of them happened to carry function == section_type, so it was benign by
// luck rather than by design.
func TestLoadExistingComponent_NoSchemaFieldsStillPinsIdentity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("lobby-grid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "function", "input_schema"}).
			AddRow(advisoryBaseID, "lobby-grid", `{}`))
	expectIdentityDecision(mock, advisoryBaseID, "lobby-grid", `{}`)

	out := advisoryResult(t, mustLoad(t, db, "lobby-grid", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "" {
		t.Errorf("no schema fields means no field contract; got %v", out["field_names"])
	}
	if out["field_count"] != 0 {
		t.Errorf("field_count must be 0; got %v", out["field_count"])
	}
	// MUTATION: drop component_id from the no-fields branch -> red here.
	if out["component_id"] != advisoryBaseID {
		t.Errorf("an empty schema must not cost the IDENTITY; got %v", out["component_id"])
	}
	if out["function"] != "lobby-grid" {
		t.Errorf("nor the function pin; got %v", out["function"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T1c — a section_type hit whose row is depended on by ANOTHER site. The store
// will divert to a fresh scoped row, so there is no contract to advise, and
// advising the incumbent's fields would manufacture a refusal.
//
// Before bugs_open/388 the primary path never ran the census, so it advised the
// incumbent's contract here — the exact case this file's own header describes
// as "wrong in two live cases", cured only on the fallback path.
func TestLoadExistingComponent_ForeignDependedPrimaryRowAdvisesWhatTheStoreWillDo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("tool-widget").
		WillReturnRows(sqlmock.NewRows([]string{"id", "function", "input_schema"}).
			AddRow(advisoryBaseID, "tool-widget", `{"fields":{"base_only_field":{}}}`))
	mock.ExpectQuery(`WHERE id = \$1::uuid AND forked_from IS NULL`).WithArgs(advisoryBaseID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content", "function"}).
			AddRow(advisoryBaseID, "<section></section>", `{"fields":{"base_only_field":{}}}`, nil, "tool-widget"))
	// A foreign site depends on it -> the store diverts.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(advisoryBaseID, advisorySiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(advisoryOtherID, "other.uk"))
	// The scoped name is free -> creation, so there is no contract at all.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("tool-widget-siteb-uk").
		WillReturnError(sql.ErrNoRows)

	out := advisoryResult(t, mustLoad(t, db, "tool-widget", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "" {
		t.Errorf("the store will CREATE here, so there is no contract; got %v", out["field_names"])
	}
	if got, _ := out["field_names"].(string); strings.Contains(got, "base_only_field") {
		t.Errorf("advising the incumbent's fields on a divert-to-create manufactures a refusal; got %v", got)
	}
	if out["component_id"] != "" {
		t.Errorf("nothing is being regenerated, so nothing may be pinned; got %v", out["component_id"])
	}
	// MUTATION: advise the primary row regardless of the identity decision ->
	// field_names comes back as "base_only_field" and this fails.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T2 — the blind case, which is the bug. No row under section_type, but the
// guard will still resolve a row by FUNCTION and enforce its field names. The
// advisory must recover that contract and pin the resolved function.
//
// The mixed-case, spaced input also pins the NormaliseToKebab derivation: the
// resolver must be asked for "loans-credit-health-check", exactly as
// store_generated_component_action.go derives it. Dropping the normalisation
// sends a different argument and the expectation fails.
func TestLoadExistingComponent_BlindSectionTypeRecoversContractFromTheResolver(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	// Primary lookup: nothing under this section_type.
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("Loans Credit Health-Check").
		WillReturnError(sql.ErrNoRows)
	// Fallback: the store's own resolver, asked for the kebab-derived function.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("loans-credit-health-check").
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
			AddRow(advisoryBaseID, "<section></section>",
				`{"fields":{"button_1":{},"button_2":{},"heading_1":{}}}`, nil))
	// Dependent census: nobody foreign, so this is a plain regeneration.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(advisoryBaseID, advisorySiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))

	out := advisoryResult(t, mustLoad(t, db, "Loans Credit Health-Check", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "button_1, button_2, heading_1" {
		t.Errorf("blind section_type must still yield the guard's field contract; got %v", out["field_names"])
	}
	if out["function"] != "loans-credit-health-check" {
		t.Errorf("function pin must name the row the store will overwrite; got %v", out["function"])
	}
	if out["field_count"] != 3 {
		t.Errorf("field_count must match the recovered contract; got %v", out["field_count"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T3 — divert-to-create: the incumbent is depended on by ANOTHER site, so the
// store will write a fresh site-scoped row. There is no field contract, and
// advising the incumbent's would be actively harmful — it would instruct the
// writer to reproduce a legacy schema into a brand-new component. The advisory
// must stay dormant. Emitting the base row's fields regardless of
// IsRegeneration fails this test.
func TestLoadExistingComponent_DivertToCreateAdvisesNoContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("tool-widget").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("tool-widget").
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
			AddRow(advisoryBaseID, "<section></section>", `{"fields":{"legacy_1":{}}}`, nil))
	// A foreign site depends on it → the write diverts.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(advisoryBaseID, advisorySiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(advisoryOtherID, "other.uk"))
	// The scoped name is free → creation, no contract.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("tool-widget-siteb-uk").WillReturnError(sql.ErrNoRows)

	out := advisoryResult(t, mustLoad(t, db, "tool-widget", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "" {
		t.Errorf("a diverted CREATE has no contract; advisory must stay dormant, got %v", out["field_names"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T4 — divert-to-existing-scoped-row: the store will regenerate the SITE-SCOPED
// row, so that row's schema is the contract, not the base row's. This is the
// case a bare by-function lookup gets wrong, and getting it wrong manufactures
// a stranded-field refusal that would not otherwise have happened.
func TestLoadExistingComponent_DivertToScopedRowAdvisesTheScopedContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("tool-widget").
		WillReturnError(sql.ErrNoRows)
	// Base row: the WRONG contract to advise.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("tool-widget").
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
			AddRow(advisoryBaseID, "<section></section>", `{"fields":{"base_only_field":{}}}`, nil))
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(advisoryBaseID, advisorySiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}).AddRow(advisoryOtherID, "other.uk"))
	// The scoped row EXISTS — this is what the store will actually overwrite.
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("tool-widget-siteb-uk").
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
			AddRow(advisoryOtherID, "<section></section>", `{"fields":{"scoped_field_a":{},"scoped_field_b":{}}}`, nil))
	// Second census, on the scoped row: no foreign dependents → regenerate it.
	mock.ExpectQuery(`SELECT DISTINCT s\.id::text, s\.domain`).WithArgs(advisoryOtherID, advisorySiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain"}))

	out := advisoryResult(t, mustLoad(t, db, "tool-widget", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "scoped_field_a, scoped_field_b" {
		t.Errorf("the contract is the SCOPED row's, not the base row's; got %v", out["field_names"])
	}
	if got, ok := out["field_names"].(string); ok && strings.Contains(got, "base_only_field") {
		t.Errorf("advising the base row's fields here manufactures a refusal; got %v", got)
	}
	if out["function"] != "tool-widget-siteb-uk" {
		t.Errorf("function pin must name the scoped row; got %v", out["function"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T5 — the advisory is advisory. A resolver failure degrades to blind
// generation and a well-formed map; it must never return an error, because an
// error here would block generation on a lookup problem the guard already
// covers.
func TestLoadExistingComponent_ResolverFailureNeverBlocks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("hero-banner").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("hero-banner").
		WillReturnError(errors.New("connection reset"))

	res, err := LoadExistingComponentAction(context.Background(),
		advisoryParams(db, "hero-banner", advisorySiteID, "siteb.uk"))
	if err != nil {
		t.Fatalf("the advisory must never return an error; got %v", err)
	}
	out := advisoryResult(t, res)
	for _, k := range []string{"field_names", "function", "field_count"} {
		if _, ok := out[k]; !ok {
			t.Errorf("the map must stay well-formed so the prompt's {{if}} guards are safe; %q missing", k)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T6 — the vocabulary rides EVERY outcome, including the one where no component
// is found at all. A first-time creation invents an unresolvable source exactly
// as readily as a regeneration does — the live bugs_open/337 failure was a
// diverted create. Moving the vocabulary load inside the found-a-row branch
// fails this test.
func TestLoadExistingComponent_VocabularyRidesEvenWhenNoComponentIsFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectVocabulary(mock)
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("brand-new-section").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`WHERE function = \$1 AND forked_from IS NULL`).WithArgs("brand-new-section").WillReturnError(sql.ErrNoRows)

	out := advisoryResult(t, mustLoad(t, db, "brand-new-section", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "" {
		t.Errorf("no component means no contract; got %v", out["field_names"])
	}
	if out["known_aspects"] != "cta, identity" {
		t.Errorf("the aspect vocabulary must be offered on a creation too; got %v", out["known_aspects"])
	}
	paths, _ := out["aspect_paths"].(string)
	if !strings.Contains(paths, "site_specs.cta.primary_url (7 sites)") {
		t.Errorf("leaf paths must carry their site coverage; got %q", paths)
	}
	if out["known_query_bases"] == nil {
		t.Errorf("the compiled-in query vocabulary must always be offered")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T7 — the two fail-opens are INDEPENDENT. A vocabulary read that errors must
// cost only its own prompt block; blanking the whole map would take the field
// contract down with it and re-create the very blindness this change removes.
func TestLoadExistingComponent_VocabularyFailureDoesNotCostTheFieldContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).
		WillReturnError(errors.New("site_specs unavailable"))
	mock.ExpectQuery(`SELECT id::text, function, input_schema`).WithArgs("hero-banner").
		WillReturnRows(sqlmock.NewRows([]string{"id", "function", "input_schema"}).
			AddRow(advisoryBaseID, "hero-banner", `{"fields":{"heading":{}}}`))
	expectIdentityDecision(mock, advisoryBaseID, "hero-banner", `{"fields":{"heading":{}}}`)

	out := advisoryResult(t, mustLoad(t, db, "hero-banner", advisorySiteID, "siteb.uk"))

	if out["field_names"] != "heading" {
		t.Errorf("a vocabulary outage must not blank the field contract; got %v", out["field_names"])
	}
	if _, present := out["known_aspects"]; present {
		t.Errorf("an unreadable aspect set must leave the key ABSENT so the prompt block stays dormant")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// T8 — one vocabulary, two consumers. The list the guard's refusal prints and
// the list the writer is shown must be the same string by construction. A
// component writer shown one list and judged against another is the whole
// defect; re-inlining a separate sort or separator in either place fails here.
func TestKnownAspectsSorted_IsTheOneRenderingBothConsumersUse(t *testing.T) {
	aspects := map[string]bool{"identity": true, "cta": true, "branding": true}

	advisoryList := strings.Join(KnownAspectsSorted(aspects), ", ")
	if advisoryList != "branding, cta, identity" {
		t.Fatalf("unexpected rendering %q", advisoryList)
	}

	// The guard's refusal message must contain that exact string.
	issues := SourceVocabularyIssues(
		`{"fields":{"cta_primary_url":{"source":"site_specs.ctas.primary_url"}}}`, aspects)
	if len(issues) != 1 {
		t.Fatalf("expected one phantom-aspect issue, got %d", len(issues))
	}
	if !strings.Contains(issues[0], "aspects that exist: "+advisoryList) {
		t.Errorf("the refusal's aspect list must be byte-identical to the advisory's.\nrefusal: %s\nadvisory: %s",
			issues[0], advisoryList)
	}
}

func mustLoad(t *testing.T, db *sql.DB, sectionType, siteID, domain string) interface{} {
	t.Helper()
	res, err := LoadExistingComponentAction(context.Background(),
		advisoryParams(db, sectionType, siteID, domain))
	if err != nil {
		t.Fatalf("the advisory must never return an error; got %v", err)
	}
	return res
}
