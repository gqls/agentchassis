// FILE: platform/orchestration/actions/plan_sections_renderer_carry_test.go
//
// bugs_open/268 — a content regeneration dropped every renderer/static-sourced
// key from page_components.content_data (cta_url, primary_cta_url,
// secondary_cta_url), deleting 214 CTA anchors across 19 live sites while five
// of six checks stayed green.
//
// The route was NOT the bugs_open/238 miss path: a renderer/static field takes
// the early branch in planSection ("resolved at render time, not now"), which
// used to write only a declared fallback and continue — so the field never
// reached resolver.resolve, handleMissingField, or carryStored, and
// save_page_sections then replaced the row wholesale without the key. That
// branch predates the 238 carry (it is original design from the file's first
// commit); nothing ever tested it — no test in the repo had ever declared a
// renderer- or static-sourced field until this file.
//
// Two properties of the branch are load-bearing and pinned here alongside the
// fix, because SQL migrations rely on them in prose:
//   - required:true on a renderer field must NOT block section readiness and
//     must NOT record a structural miss (098_broaden_cta_recompute_source_
//     unlock.sql relies on this for gauntlet-cta's primary urls);
//   - a declared fallback still writes when nothing is stored
//     (181_class_e_live_cta_url_integrity.sql / 097b's deliberate pin).
//
// The one deliberate behaviour change beyond the carry itself: a stored value
// now beats a declared fallback (the fallback is a default; the stored value
// is the page). TestPlanSections_StoredValueBeatsStaticFallback pins it.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// rendererCTASchema mirrors the live hero/call-to-action shape after
// migrations 091/098: llm label beside renderer-sourced destination fields,
// required=false, on_missing=skip_field, no fallback — verified against the
// live schema in bugs_open/268 §1.
const rendererCTASchema = `{"fields":{
    "cta_text":        {"type":"text","source":"llm","required":true},
    "cta_url":         {"type":"url","source":"renderer","required":false,"on_missing":"skip_field"},
    "primary_cta_url":  {"type":"url","source":"renderer","required":false,"on_missing":"skip_field"}
}}`

func expectRendererComponentLoad(mock sqlmock.Sqlmock, componentID, schema string) {
	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRowWithSchema(componentID, "hero", "hero",
			`{{if and .cta_text .cta_url}}<a href="{{.cta_url}}">{{.cta_text}}</a>{{end}}`,
			"section", schema))
}

// TestPlanSections_RegenerationCarriesRendererSourcedURLs is the motivating
// case, reduced: the page's deployed row still holds both destinations, the
// regeneration replans the section, and before the fix both keys silently
// fell out of resolved_data — so the replacement row shipped label-only and
// the template rendered no anchor at all.
func TestPlanSections_RegenerationCarriesRendererSourcedURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema's fields MAP, so query arrival order is
	// randomised per run — ordered expectations are a coin flip.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectRendererComponentLoad(mock, componentID, rendererCTASchema)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)

	// THE CARRY PRELOAD: the deployed row still holds both destinations.
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows().AddRow("hero", []byte(`{
            "cta_text": "Get in touch",
            "cta_url": "/contact.html",
            "primary_cta_url": "/services.html"
        }`)))

	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected one ready section, got %d", len(items))
	}
	got := items[0]

	want := map[string]string{
		"cta_url":         "/contact.html",
		"primary_cta_url": "/services.html",
	}
	for field, wantValue := range want {
		value, ok := got.ResolvedData[field]
		if !ok {
			t.Errorf("%s was NOT carried — the bugs_open/268 drop: the renderer/static branch continues before carryStored, the key is absent from resolved_data, and save_page_sections replaces the row without it", field)
			continue
		}
		if value != wantValue {
			t.Errorf("%s carried %v, want %v", field, value, wantValue)
		}
	}
	// The label is LLM-sourced: carrying it would defeat the rewrite.
	if _, carried := got.ResolvedData["cta_text"]; carried {
		t.Error("cta_text was carried — an llm-sourced field must never be taken from the stored row")
	}
	if len(got.CarriedFields) != 2 {
		t.Errorf("expected 2 carried fields recorded on the plan item, got %v", got.CarriedFields)
	}
	if len(got.StructuralMisses) != 0 {
		t.Errorf("expected no structural misses when the carry succeeded, got %v", got.StructuralMisses)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_RendererFieldWithNothingStoredStaysAbsent pins
// correct-or-absent (LNK-005) AND the 098 readiness contract in one fixture:
// nothing is stored for this slot, one destination field is required:true —
// the section must still come out ready, with no fabricated value, no
// structural miss, and no durable error row. The carry can only ever restore
// a real prior value; a first build has nothing to restore and that is not a
// defect.
func TestPlanSections_RendererFieldWithNothingStoredStaysAbsent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	// required:true on a renderer source is the 098 gauntlet-cta shape.
	expectRendererComponentLoad(mock, componentID, `{"fields":{
        "cta_text": {"type":"text","source":"llm","required":true},
        "cta_url":  {"type":"url","source":"renderer","required":true}
    }}`)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	// The preload runs (the carry now looks) and finds a different slot only.
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows().AddRow("footer-cta", []byte(`{"cta_url":"/contact.html"}`)))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))
	// No agent_error_log expectation: a renderer field must not take the
	// handleMissingField route, required or not — sqlmock fails on an
	// unexpected exec, so the ABSENCE of that expectation is the assertion.

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 || items[0].Status != "ready" {
		t.Fatalf("a required renderer field with nothing stored must not block readiness (098's contract); got %+v", items)
	}
	if _, present := items[0].ResolvedData["cta_url"]; present {
		t.Error("cta_url was fabricated — nothing was stored for this slot, so the field must stay absent (correct-or-absent)")
	}
	if len(items[0].CarriedFields) != 0 {
		t.Errorf("nothing was available to carry, yet fields were recorded as carried: %v", items[0].CarriedFields)
	}
	if len(items[0].StructuralMisses) != 0 {
		t.Errorf("a renderer field must not record structural misses (it never takes the miss path), got %v", items[0].StructuralMisses)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_StaticFallbackStillWritesWhenNothingStored pins 097b's
// deliberate pin (181): a static source with a declared fallback and nothing
// stored still gets the fallback, unconditionally — required/on_missing do
// not apply on this branch.
func TestPlanSections_StaticFallbackStillWritesWhenNothingStored(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectRendererComponentLoad(mock, componentID, `{"fields":{
        "cta_text": {"type":"text","source":"llm","required":true},
        "cta_url":  {"type":"url","source":"static","required":true,"fallback":"/tools/arena/index.html"}
    }}`)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	// Preload runs, finds nothing at all.
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows())
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected one ready section, got %d", len(items))
	}
	if got := items[0].ResolvedData["cta_url"]; got != "/tools/arena/index.html" {
		t.Errorf("declared fallback must still write when nothing is stored (181/097b's pin); got %v", got)
	}
	if len(items[0].CarriedFields) != 0 {
		t.Errorf("a fallback write is not a carry, got %v", items[0].CarriedFields)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_SpecSourcedFallbackWritesWhenCarryMisses is the sibling
// branch to TestPlanSections_StaticFallbackStillWritesWhenNothingStored, asked
// for by the council round that approved the 268 fix (corr e6c1e4eb,
// bug_historian): a resolver-path field (site_specs.*) with
// on_missing=use_fallback, a declared fallback, and nothing stored. The route
// is resolve() miss → handleMissingField → carryStored (offered, finds
// nothing) → on_missing. Pins the pre-existing asymmetry between the two
// fallback arms, which nothing had ever tested: a resolver-path fallback
// applies ONLY under on_missing=use_fallback, while a renderer/static
// fallback writes unconditionally (181/097b's pin, the test above).
func TestPlanSections_SpecSourcedFallbackWritesWhenCarryMisses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	// A non-identity aspect, so the alias chain (nested shape, sites row)
	// misses without issuing further queries.
	expectRendererComponentLoad(mock, componentID, `{"fields":{
        "cta_text": {"type":"text","source":"llm","required":true},
        "cta_url":  {"type":"url","source":"site_specs.conversion.primary_target","required":false,"on_missing":"use_fallback","fallback":"/contact.html"}
    }}`)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	// The carry IS consulted on this branch — the preload runs and finds
	// nothing stored. (Contrast: before the 268 fix the renderer/static branch
	// never consulted it at all.)
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows())
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 || items[0].Status != "ready" {
		t.Fatalf("an optional field under use_fallback must not affect readiness; got %+v", items)
	}
	if got := items[0].ResolvedData["cta_url"]; got != "/contact.html" {
		t.Errorf("declared fallback must write when the source and the carry both miss under on_missing=use_fallback; got %v", got)
	}
	if len(items[0].CarriedFields) != 0 {
		t.Errorf("a fallback write is not a carry, got %v", items[0].CarriedFields)
	}
	if len(items[0].StructuralMisses) != 0 {
		t.Errorf("an optional field must not record a structural miss, got %v", items[0].StructuralMisses)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_StoredValueBeatsStaticFallback pins the one deliberate
// behaviour change beyond the carry itself. Before the fix a declared
// fallback wrote UNCONDITIONALLY, so a page whose static field had been
// retargeted (an authored edit, stored in content_data) was silently reverted
// to the default on every regeneration — the same revert class migration 091
// closed for spec-derived sources. The stored value is the page; the fallback
// is a default for pages that have nothing.
func TestPlanSections_StoredValueBeatsStaticFallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectRendererComponentLoad(mock, componentID, `{"fields":{
        "cta_text": {"type":"text","source":"llm","required":true},
        "cta_url":  {"type":"url","source":"static","required":false,"fallback":"/tools/arena/index.html"}
    }}`)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectSpecsAndAssetsEmpty(mock)
	mock.ExpectQuery("build_status = 'deployed'").
		WithArgs(siteID, "index").
		WillReturnRows(storedContentRows().AddRow("hero", []byte(`{"cta_url":"/retargeted-by-editor.html"}`)))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected one ready section, got %d", len(items))
	}
	if got := items[0].ResolvedData["cta_url"]; got != "/retargeted-by-editor.html" {
		t.Errorf("the stored (authored) value must beat the declared fallback; got %v — a regeneration just reverted an edit to the default", got)
	}
	if len(items[0].CarriedFields) != 1 {
		t.Errorf("expected the carried field recorded on the plan item, got %v", items[0].CarriedFields)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
