// FILE: platform/orchestration/actions/plan_sections_structural_carry_test.go
//
// bugs_open/238 — a content regeneration silently lost every resolver-sourced
// key a page already carried, and shipped five <img src=""> plus six vanished
// controls to a live homepage.
//
// The failure had TWO invisible halves, and both are pinned here because either
// one alone reads exactly like success:
//
//  1. The CARRY working and the carry finding nothing produce identical plans.
//     A section whose declared source resolves nothing looks the same in
//     `sections_ready` whether the stored row rescued it or not — so
//     TestPlanSections_RegenerationCarriesStoredStructuralKeys asserts the
//     VALUES are on the plan item, not merely that the section is ready.
//  2. A required field that resolves NOWHERE is omitted with an Info log and no
//     other trace. TestPlanSections_RequiredMissEverywhereIsRecordedNotDeferred
//     asserts the durable agent_error_log write, because "no findings row" and
//     "nothing went wrong" are the same observation otherwise.
//
// The no-op case matters as much as the damage case: a fresh build has no stored
// row, and the carry must not query for one until a source has actually missed
// (sqlmock fails on an unexpected query, which is what makes the laziness
// assertion real rather than asserted).
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// carrySchema mirrors the shape that produced bugs_open/238 on
// case-studies-grid: an llm-sourced text field beside resolver-sourced URL
// fields that are `required` and declare NO on_missing (so they default to
// skip_field), plus one optional resolver-sourced field. The image field is
// type "url" — not "image" — which is exactly why the three existing
// image-oriented checks were all silent on it.
const carrySchema = `{"fields":{
    "card1_image_alt": {"type":"text","source":"llm","required":true},
    "card1_image_url": {"type":"url","source":"site_assets.image","required":true},
    "card1_link_url":  {"type":"url","source":"site_specs.case_studies.card1_url","required":true},
    "cta_link_url":    {"type":"url","source":"site_specs.pages.contact_url","required":false}
}}`

// storedContentRows builds the lazy carry preload's result set.
func storedContentRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"slot_name", "content_data"})
}

// expectCarryComponentLoad wires the component resolution every test here does:
// the section name IS the component name, so the first pass hits.
func expectCarryComponentLoad(mock sqlmock.Sqlmock, componentID string) {
	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRowWithSchema(componentID, "case-studies-grid", "case-studies-grid",
			`<img class="csg-card-image" src="{{.card1_image_url}}" alt="{{.card1_image_alt}}" />`,
			"section", carrySchema))
}

// componentRowWithSchema is componentRow with the input_schema under test.
func componentRowWithSchema(id, name, function, htmlTemplate, level, schema string) *sqlmock.Rows {
	return sqlmock.NewRows(componentColumns).AddRow(
		id, name, name, function, "", nil, nil, htmlTemplate, schema, "template", nil, level)
}

// expectSpecsAndAssetsEmpty wires the resolver reads a site with no specs, no
// current plan imagery and no content_data hero performs — i.e. finetuning.uk's
// actual state, where all four declared sources resolve to nothing.
func expectSpecsAndAssetsEmpty(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))
}

// TestPlanSections_RegenerationCarriesStoredStructuralKeys is the motivating
// case, reduced: every declared source misses, and the page's own deployed row
// holds the values. Before the fix all three URL keys were dropped from
// resolved_data, save_page_sections replaced the row without them, and the
// template rendered src="" / no anchor at all.
//
// This is the pin for the live 58→47 key drop: any future change that lets a
// declared non-llm key present in the stored row fall out of resolved_data
// fails here.
func TestPlanSections_RegenerationCarriesStoredStructuralKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema's fields MAP, so the order in which
	// sources are resolved — and therefore the order these queries arrive in —
	// is randomised per run. Ordered expectations here are a coin flip that
	// passes until it does not.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectCarryComponentLoad(mock, componentID)
	// Slot-identity read (Path 0) — no stored slot identity in this fixture.
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)

	// site_assets.image → role alias → ensureAssets. Empty: no plan imagery.
	mock.ExpectQuery("site_plan_imagery").WillReturnRows(sqlmock.NewRows([]string{"key", "url", "scope", "scope_ref", "kind"}))
	mock.ExpectQuery("FROM sites").WillReturnRows(sqlmock.NewRows([]string{"content_data"}))

	// THE CARRY PRELOAD: the page's deployed row still holds all three URLs.
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows().AddRow("case-studies-grid", []byte(`{
            "card1_image_url": "/assets/images/case-study-facilities.jpg",
            "card1_link_url":  "/case-studies.html",
            "cta_link_url":    "/contact.html",
            "card1_image_alt": "stale alt the LLM will rewrite"
        }`)))

	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"case-studies-grid"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected one ready section, got %d", len(items))
	}
	got := items[0]

	want := map[string]string{
		"card1_image_url": "/assets/images/case-study-facilities.jpg",
		"card1_link_url":  "/case-studies.html",
		"cta_link_url":    "/contact.html",
	}
	for field, wantValue := range want {
		value, ok := got.ResolvedData[field]
		if !ok {
			t.Errorf("%s was NOT carried — this is the bugs_open/238 drop: the key is absent from resolved_data, so the regenerated row ships without it and the template renders empty", field)
			continue
		}
		if value != wantValue {
			t.Errorf("%s carried %v, want %v", field, value, wantValue)
		}
	}

	// The alt text is LLM-sourced: carrying it would silently defeat the
	// rewrite the regeneration exists to perform.
	if _, carried := got.ResolvedData["card1_image_alt"]; carried {
		t.Error("card1_image_alt was carried — an llm-sourced field must never be taken from the stored row, or a tone_shift can never change the copy")
	}
	if len(got.CarriedFields) != 3 {
		t.Errorf("expected 3 carried fields recorded on the plan item, got %v", got.CarriedFields)
	}
	// Everything resolved (from the stored row), so nothing resolved NOWHERE.
	if len(got.StructuralMisses) != 0 {
		t.Errorf("expected no structural misses when the carry succeeded, got %v", got.StructuralMisses)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_FreshBuildDoesNotQueryStoredContent is the no-op case, and
// the laziness pin. When every declared source resolves, the carry must never
// run its query — the cost of this fix on the common path has to be zero, and
// "I only added a query in the miss path" is a claim a test can either make good
// or expose. sqlmock fails on an unexpected query, so the ABSENCE of a
// build_status='deployed' expectation below IS the assertion.
func TestPlanSections_FreshBuildDoesNotQueryStoredContent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema's fields MAP, so the order in which
	// sources are resolved — and therefore the order these queries arrive in —
	// is randomised per run. Ordered expectations here are a coin flip that
	// passes until it does not.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	// An all-llm schema: no field ever reaches the source resolver.
	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRowWithSchema(componentID, "hero", "hero", "<section>{{.headline}}</section>",
			"section", `{"fields":{"headline":{"type":"text","source":"llm","required":true}}}`))
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 || items[0].Status != "ready" {
		t.Fatalf("expected one ready section, got %+v", items)
	}
	if len(items[0].CarriedFields) != 0 {
		t.Errorf("a build with nothing to carry recorded carried fields: %v", items[0].CarriedFields)
	}
	m := out.(map[string]interface{})
	// A nil map stored in an interface is NOT == nil, so assert on the typed
	// value: the key marshals to JSON null, which is what "nothing was carried"
	// looks like downstream.
	carried, _ := m["structural_keys_carried"].(map[string][]string)
	if len(carried) != 0 {
		t.Errorf("structural_keys_carried should be empty when nothing was carried, got %v", carried)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations (an unexpected query here means the carry preload ran on a path where every source resolved): %v", err)
	}
}

// TestPlanSections_RequiredMissEverywhereIsRecordedNotDeferred pins the second
// invisible half. The source resolves nothing AND the stored row has nothing —
// the leopardess/oufe "never had the key" class, not a regression. The section
// must still build (deferring here would be RFC_009's option A, which the owner
// did not take), and the miss must leave a durable row, because a plan missing a
// field it never declared and a plan missing a field it lost are otherwise the
// same plan.
func TestPlanSections_RequiredMissEverywhereIsRecordedNotDeferred(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema's fields MAP, so the order in which
	// sources are resolved — and therefore the order these queries arrive in —
	// is randomised per run. Ordered expectations here are a coin flip that
	// passes until it does not.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectCarryComponentLoad(mock, componentID)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	mock.ExpectQuery("site_plan_imagery").WillReturnRows(sqlmock.NewRows([]string{"key", "url", "scope", "scope_ref", "kind"}))
	mock.ExpectQuery("FROM sites").WillReturnRows(sqlmock.NewRows([]string{"content_data"}))
	// Stored row exists for a DIFFERENT slot — so the preload runs and finds
	// nothing for this one. (A no-rows result would prove the same thing; this
	// shape also proves the lookup is slot-scoped rather than page-wide.)
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows().AddRow("hero", []byte(`{"headline":"unrelated"}`)))

	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))
	// The durable record. Its absence from this list would let a silent
	// omission pass as a clean build.
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"case-studies-grid"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected the section to build anyway (skip_field is honoured), got ready=%d", len(items))
	}
	got := items[0]
	if len(got.CarriedFields) != 0 {
		t.Errorf("nothing was available to carry, yet fields were recorded as carried: %v", got.CarriedFields)
	}
	// Only the two REQUIRED non-llm fields are recorded; the optional
	// cta_link_url resolving to nothing is its declared contract, not a defect.
	if len(got.StructuralMisses) != 2 {
		t.Fatalf("expected the 2 required non-llm fields recorded as structural misses, got %v", got.StructuralMisses)
	}
	m := out.(map[string]interface{})
	misses, ok := m["structural_key_misses"].(map[string][]missingField)
	if !ok || len(misses["case-studies-grid"]) != 2 {
		t.Errorf("expected the misses surfaced on the action result, got %v", m["structural_key_misses"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_LiveResolutionBeatsStoredCarry pins the precedence that
// bounds staleness: the carry is a fallback, never a preference. A site whose
// spec has since been corrected must get the corrected value, or a carried value
// would outlive the reason it was carried.
func TestPlanSections_LiveResolutionBeatsStoredCarry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema's fields MAP, so the order in which
	// sources are resolved — and therefore the order these queries arrive in —
	// is randomised per run. Ordered expectations here are a coin flip that
	// passes until it does not.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRowWithSchema(componentID, "case-studies-grid", "case-studies-grid",
			`<a href="{{.cta_link_url}}">go</a>`, "section",
			`{"fields":{"cta_link_url":{"type":"url","source":"site_specs.pages.contact_url","required":true}}}`))
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	// The spec aspect now EXISTS and holds the corrected value.
	mock.ExpectQuery("FROM site_specs").WillReturnRows(
		sqlmock.NewRows([]string{"aspect", "data"}).AddRow("pages", []byte(`{"contact_url":"/get-in-touch.html"}`)))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))
	// No carry preload expected: the source resolved, so the miss path is never
	// reached. sqlmock would fail the test if it ran.

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"case-studies-grid"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected one ready section, got %d", len(items))
	}
	if got := items[0].ResolvedData["cta_link_url"]; got != "/get-in-touch.html" {
		t.Errorf("live resolution must win over the stored value; got %v", got)
	}
	if len(items[0].CarriedFields) != 0 {
		t.Errorf("carry fired despite the source resolving: %v", items[0].CarriedFields)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_EmptyStoredValueIsNotCarried: a stored key holding "" is the
// DEFECT, not a value. Carrying it would reproduce src="" faithfully and report
// a successful carry while doing it — the failure mode this whole fix exists to
// remove, re-entering through its own front door.
func TestPlanSections_EmptyStoredValueIsNotCarried(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema's fields MAP, so the order in which
	// sources are resolved — and therefore the order these queries arrive in —
	// is randomised per run. Ordered expectations here are a coin flip that
	// passes until it does not.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRowWithSchema(componentID, "case-studies-grid", "case-studies-grid",
			`<img src="{{.card1_image_url}}" />`, "section",
			`{"fields":{"card1_image_url":{"type":"url","source":"site_assets.image","required":true}}}`))
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	mock.ExpectQuery("site_plan_imagery").WillReturnRows(sqlmock.NewRows([]string{"key", "url", "scope", "scope_ref", "kind"}))
	mock.ExpectQuery("FROM sites").WillReturnRows(sqlmock.NewRows([]string{"content_data"}))
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows().AddRow("case-studies-grid", []byte(`{"card1_image_url":""}`)))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"case-studies-grid"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected one ready section, got %d", len(items))
	}
	if _, carried := items[0].ResolvedData["card1_image_url"]; carried {
		t.Error("an empty stored value was carried — that ships src=\"\" and reports it as a successful carry")
	}
	if len(items[0].StructuralMisses) != 1 {
		t.Errorf("an unusable stored value must still count as a miss, got %v", items[0].StructuralMisses)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_ConflictedSlotIsNotACarrySource: two rows share a slot_name
// with different content_data, so which one holds "the" value is unanswerable.
// Resolving it arbitrarily would make the carry non-deterministic — the same
// judgement loadPageSlotComponentIDs already makes about slot identity, kept
// deliberately identical so the two cannot drift.
func TestPlanSections_ConflictedSlotIsNotACarrySource(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema's fields MAP, so the order in which
	// sources are resolved — and therefore the order these queries arrive in —
	// is randomised per run. Ordered expectations here are a coin flip that
	// passes until it does not.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRowWithSchema(componentID, "case-studies-grid", "case-studies-grid",
			`<img src="{{.card1_image_url}}" />`, "section",
			`{"fields":{"card1_image_url":{"type":"url","source":"site_assets.image","required":true}}}`))
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	mock.ExpectQuery("site_plan_imagery").WillReturnRows(sqlmock.NewRows([]string{"key", "url", "scope", "scope_ref", "kind"}))
	mock.ExpectQuery("FROM sites").WillReturnRows(sqlmock.NewRows([]string{"content_data"}))
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows().
			AddRow("case-studies-grid", []byte(`{"card1_image_url":"/assets/images/a.jpg"}`)).
			AddRow("case-studies-grid", []byte(`{"card1_image_url":"/assets/images/b.jpg"}`)))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"case-studies-grid"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected one ready section, got %d", len(items))
	}
	if v, carried := items[0].ResolvedData["card1_image_url"]; carried {
		t.Errorf("an ambiguous slot was used as a carry source, picking %v arbitrarily", v)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
