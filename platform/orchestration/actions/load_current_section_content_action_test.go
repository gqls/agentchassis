// FILE: platform/orchestration/actions/load_current_section_content_action_test.go
//
// bugs_open/178. The load-bearing negative here is Passthrough: for every
// item that does not name mode=edit_live, this step must not touch
// section_plan at all and must not query the database — proven by sqlmock's
// refusal of an unexpected call, not by anything the action reports about
// itself (the same shape tool_content_item_test.go uses for its NoSections
// case, and for the same reason: a helper's own bookkeeping cannot prove a
// negative).
//
// bugs_open/192 — REWRITTEN. The first version of this file asserted
// result["applied"], result["reason"] and result["section_plan"], i.e. it
// encoded the WRAPPER the code happened to return as though it were the
// contract. Its own comment two lines below said the opposite ("must leave
// section_plan byte-identical"), and the wrapper is exactly what broke every
// page build in the fleet. A test that asserts the shape the code produces
// cannot catch a shape defect; these assert the shape the header PROMISES:
// the return value IS the plan. TestLoadCurrentSectionContent_NeverWraps is
// the explicit regression tripwire — it fails on the pre-192 code.
package actions

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// errQueryFailedForTest drives the query_failed return path, which is one of
// the paths that used to wrap (bugs_open/192).
var errQueryFailedForTest = errors.New("simulated page_components query failure")

func testSectionPlan() map[string]interface{} {
	return map[string]interface{}{
		"sections_ready": []interface{}{
			map[string]interface{}{"name": "generic-text-block", "status": "ready"},
			map[string]interface{}{"name": "hero", "status": "ready"},
		},
		"ready_count": 2,
	}
}

func loadCurrentSectionContentParams(db interface{ Close() error }, mode string, collected map[string]interface{}) ActionParams {
	config := map[string]interface{}{
		"site_id": "site_record.site_id",
		"page_id": "page_record.id",
	}
	if mode != "" {
		config["mode"] = "input_data.spec.mode"
	}
	return ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: config},
		CollectedData:    collected,
	}
}

// assertIsThePlan is the whole contract in one place: because the step's
// output_field is "section_plan", storeActionResult writes this return value
// straight over collected_data.section_plan. So the value must BE a plan —
// carrying sections_ready at its top level — and must never be a wrapper with
// the plan nested inside it under its own name (bugs_open/192).
func assertIsThePlan(t *testing.T, got interface{}) map[string]interface{} {
	t.Helper()
	plan, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("return value is not a map: %T", got)
	}
	if nested, has := plan["section_plan"]; has {
		t.Fatalf("bugs_open/192 regression: the plan is nested under its own key (%T). "+
			"output_field REPLACES section_plan, so the return value must be the plan itself", nested)
	}
	if _, has := plan["sections_ready"]; !has {
		t.Fatalf("return value is not a plan: no top-level sections_ready, keys = %v", reflect.ValueOf(plan).MapKeys())
	}
	return plan
}

// No mode at all — the common case, since no live emitter sets spec.mode —
// must leave section_plan byte-identical and must not query the database.
func TestLoadCurrentSectionContent_NoMode_Passthrough(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	plan := testSectionPlan()
	collected := map[string]interface{}{
		"section_plan": plan,
		"site_record":  map[string]interface{}{"site_id": uuid.New().String()},
		"page_record":  map[string]interface{}{"id": uuid.New().String()},
		"input_data":   map[string]interface{}{"spec": map[string]interface{}{}},
	}
	params := loadCurrentSectionContentParams(db, "unset", collected)
	params.DB = db

	got, err := LoadCurrentSectionContentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "byte-identical" asserted as identity, not as a field spot-check: this is
	// the same object the step was handed, untouched.
	if !reflect.DeepEqual(got, plan) {
		t.Errorf("pass-through must return section_plan unchanged.\n got: %v\nwant: %v", got, plan)
	}
	returnedPlan := assertIsThePlan(t, got)
	if returnedPlan["ready_count"] != 2 {
		t.Errorf("section_plan mutated: ready_count = %v", returnedPlan["ready_count"])
	}
	if _, has := returnedPlan["edit_live_meta"]; has {
		t.Error("a pass-through must not annotate the plan: edit_live_meta present")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (a query fired when none should have): %v", err)
	}
}

// mode=recreate — the OTHER existing value — must also pass through
// unchanged. This step only ever fires on the third, new value.
func TestLoadCurrentSectionContent_ModeRecreate_Passthrough(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	plan := testSectionPlan()
	collected := map[string]interface{}{
		"section_plan": plan,
		"site_record":  map[string]interface{}{"site_id": uuid.New().String()},
		"page_record":  map[string]interface{}{"id": uuid.New().String()},
		"input_data":   map[string]interface{}{"spec": map[string]interface{}{"mode": "recreate"}},
	}
	params := loadCurrentSectionContentParams(db, "recreate", collected)
	params.DB = db

	got, err := LoadCurrentSectionContentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, plan) {
		t.Errorf("mode=recreate must pass through unchanged.\n got: %v\nwant: %v", got, plan)
	}
	assertIsThePlan(t, got)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// mode=edit_live with a matching page_components row: the ready section's
// current rendered_html is attached, keyed by slot_name == section name.
func TestLoadCurrentSectionContent_EditLive_AttachesMatchingSlot(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	mock.ExpectQuery("SELECT pc.slot_name, COALESCE\\(pc.rendered_html, ''\\)").
		WithArgs(pageID, siteID).
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}).
			AddRow("generic-text-block", "<section>full prior prose</section>").
			AddRow("unrelated-slot", "<section>other content</section>"))

	collected := map[string]interface{}{
		"section_plan": testSectionPlan(),
		"site_record":  map[string]interface{}{"site_id": siteID.String()},
		"page_record":  map[string]interface{}{"id": pageID.String()},
		"input_data":   map[string]interface{}{"spec": map[string]interface{}{"mode": "edit_live"}},
	}
	params := loadCurrentSectionContentParams(db, "edit_live", collected)
	params.DB = db

	got, err := LoadCurrentSectionContentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Even on the path that DOES something, the return value is still the plan.
	plan := assertIsThePlan(t, got)

	meta, ok := plan["edit_live_meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("edit_live_meta missing or not a map: %T", plan["edit_live_meta"])
	}
	if meta["applied"] != true {
		t.Errorf("edit_live_meta.applied = %v, want true", meta["applied"])
	}
	if meta["matched"] != 1 {
		t.Errorf("edit_live_meta.matched = %v, want 1 (only generic-text-block has a page_components row; unrelated-slot matches no ready section, hero has no row)", meta["matched"])
	}

	ready := plan["sections_ready"].([]interface{})

	var gotBlock, gotHero map[string]interface{}
	for _, raw := range ready {
		s := raw.(map[string]interface{})
		switch s["name"] {
		case "generic-text-block":
			gotBlock = s
		case "hero":
			gotHero = s
		}
	}
	if gotBlock == nil || gotBlock["existing_content_html"] != "<section>full prior prose</section>" {
		t.Errorf("generic-text-block existing_content_html = %v, want the matching page_components row", gotBlock["existing_content_html"])
	}
	if gotHero != nil {
		if _, has := gotHero["existing_content_html"]; has {
			t.Errorf("hero has no page_components row and must not gain existing_content_html, got %v", gotHero["existing_content_html"])
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// No ready sections: nothing to join against, so the query must not fire —
// same negative-proof shape as the passthrough cases.
func TestLoadCurrentSectionContent_NoReadySections_SkipsQuery(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	emptyPlan := map[string]interface{}{
		"sections_ready": []interface{}{},
		"ready_count":    0,
	}
	collected := map[string]interface{}{
		"section_plan": emptyPlan,
		"site_record":  map[string]interface{}{"site_id": siteID.String()},
		"page_record":  map[string]interface{}{"id": pageID.String()},
		"input_data":   map[string]interface{}{"spec": map[string]interface{}{"mode": "edit_live"}},
	}
	params := loadCurrentSectionContentParams(db, "edit_live", collected)
	params.DB = db

	got, err := LoadCurrentSectionContentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, emptyPlan) {
		t.Errorf("no ready sections must pass through unchanged.\n got: %v\nwant: %v", got, emptyPlan)
	}
	assertIsThePlan(t, got)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (a query fired with nothing to match): %v", err)
	}
}

// bugs_open/192, the regression tripwire stated as its own case rather than
// left implicit in the four above. Every reachable return path is walked, and
// each one must yield a value that IS the plan. This test FAILS on the pre-192
// code on every one of its sub-cases — which is what makes it worth having;
// its predecessor passed on the broken code because it asserted the break.
func TestLoadCurrentSectionContent_NeverWraps(t *testing.T) {
	newCollected := func(mode string, plan interface{}) map[string]interface{} {
		return map[string]interface{}{
			"section_plan": plan,
			"site_record":  map[string]interface{}{"site_id": uuid.New().String()},
			"page_record":  map[string]interface{}{"id": uuid.New().String()},
			"input_data":   map[string]interface{}{"spec": map[string]interface{}{"mode": mode}},
		}
	}

	cases := []struct {
		name      string
		mode      string
		plan      interface{}
		configure func(mock sqlmock.Sqlmock)
	}{
		{name: "not_edit_live", mode: "recreate", plan: testSectionPlan()},
		{name: "no_ready_sections", mode: "edit_live", plan: map[string]interface{}{"sections_ready": []interface{}{}}},
		{
			name: "query_failed",
			mode: "edit_live",
			plan: testSectionPlan(),
			configure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT pc.slot_name").WillReturnError(errQueryFailedForTest)
			},
		},
		{
			name: "applied",
			mode: "edit_live",
			plan: testSectionPlan(),
			configure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT pc.slot_name").WillReturnRows(
					sqlmock.NewRows([]string{"slot_name", "rendered_html"}).
						AddRow("generic-text-block", "<section>prose</section>"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newInsertMock(t)
			defer db.Close()
			if tc.configure != nil {
				tc.configure(mock)
			}

			params := loadCurrentSectionContentParams(db, tc.mode, newCollected(tc.mode, tc.plan))
			params.DB = db

			got, err := LoadCurrentSectionContentAction(context.Background(), params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertIsThePlan(t, got)
		})
	}
}
