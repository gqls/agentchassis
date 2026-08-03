// FILE: platform/orchestration/actions/load_current_section_content_action_test.go
//
// bugs_open/178. The load-bearing negative here is Passthrough: for every
// item that does not name mode=edit_live, this step must not touch
// section_plan at all and must not query the database — proven by sqlmock's
// refusal of an unexpected call, not by anything the action reports about
// itself (the same shape tool_content_item_test.go uses for its NoSections
// case, and for the same reason: a helper's own bookkeeping cannot prove a
// negative).
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

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
	result := got.(map[string]interface{})
	if result["applied"] != false {
		t.Errorf("applied = %v, want false", result["applied"])
	}
	if reason := result["reason"]; reason != "not_edit_live" {
		t.Errorf("reason = %v, want not_edit_live", reason)
	}
	returnedPlan, ok := result["section_plan"].(map[string]interface{})
	if !ok {
		t.Fatalf("section_plan not returned as a map: %T", result["section_plan"])
	}
	if returnedPlan["ready_count"] != 2 {
		t.Errorf("section_plan mutated: ready_count = %v", returnedPlan["ready_count"])
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

	collected := map[string]interface{}{
		"section_plan": testSectionPlan(),
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
	result := got.(map[string]interface{})
	if result["applied"] != false || result["reason"] != "not_edit_live" {
		t.Errorf("got applied=%v reason=%v, want applied=false reason=not_edit_live", result["applied"], result["reason"])
	}
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
	result := got.(map[string]interface{})
	if result["applied"] != true {
		t.Fatalf("applied = %v, want true", result["applied"])
	}
	if result["matched"] != 1 {
		t.Errorf("matched = %v, want 1 (only generic-text-block has a page_components row; unrelated-slot matches no ready section, hero has no row)", result["matched"])
	}

	plan := result["section_plan"].(map[string]interface{})
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
	result := got.(map[string]interface{})
	if result["applied"] != false || result["reason"] != "no_ready_sections" {
		t.Errorf("got applied=%v reason=%v, want applied=false reason=no_ready_sections", result["applied"], result["reason"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (a query fired with nothing to match): %v", err)
	}
}
