// FILE: internal/core-manager/admin/approve_fan_out_test.go
//
// Covers HandleApproveWorkItem's on_approve contract: what the owner's APPROVE
// button actually files.
//
// The defect these tests exist to prevent (bugs_open/466), met on the first
// real approval — the first paid site, 2026-09-03:
//
//  1. `include_fields` resolved its names against the REVIEW ITEM'S SPEC, and
//     checkpoint_for_review (the only producer of these items) writes a fixed
//     key set that can never contain them. So the follow-on carried
//     `copy_edit: null, page_target: null` with the real payload stranded under
//     `approved_data`. [MEASURED 2026-09-03] 42 field mentions across 21 items,
//     all history: ZERO present at spec top level. The lookup could not have
//     worked for any consumer, ever — so this is not a disagreement about where
//     content lives, it is a dead mechanism.
//
//  2. copy-editor proposes N edits, each with its own page_component_id and
//     field_updates; section-editor applies ONE. A two-edit proposal has no
//     single target, so no amount of key-renaming fixes it — the approval has
//     to fan out. Two lanes had already hand-built exactly that shape by the
//     time it was found.
//
// The invariants, and why each is easy to regress:
//
//   - An on_approve naming NEITHER new key must file exactly one item, as
//     before. The keys are opt-in with the unsafe side OFF (owner ruling
//     2026-08-02 §2); a "tidy-up" that makes fan-out the default would change
//     what every existing checkpoint consumer gets.
//   - A fanned-out element's own fields must land at the TOP of the child spec,
//     because that is where load_edit_context and apply_section_edit read them.
//   - A dead page_component_id must be REFUSED and reported, never filed. A
//     rerender replaces the component row with a new id, so a parked proposal's
//     address rots (LANDMINES, 2026-08-18). [MEASURED 2026-09-03] 3 of 31 edits
//     parked in needs_human_review already point at a row that is gone. Filing
//     one produces an item that dies at load_edit_context and tells the approver
//     nothing — which is the bug, not a fix for it.
//   - Provenance fields are applied LAST, so an element carrying its own
//     `approved_by` cannot overwrite who approved it.

package admin

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	fanReviewItemID = "1cb907ee-b05f-4193-ae2d-07b9923b4445"
	fanSiteID       = "d2aa5206-73bc-4707-a69c-2702c1eb9152"
	fanPCLive       = "322ce532-b39d-424e-b1d8-08ff0e4e13f0"
	fanPCDead       = "e5b848fa-c8d3-4997-ac38-8ec8772cbd55"
)

// capture records the spec JSON an INSERT was given, and matches anything.
// sqlmock's WithArgs cannot otherwise show us what the handler built, and the
// spec is the whole point of this handler.
type capture struct{ got []string }

func (c *capture) Match(v driver.Value) bool {
	c.got = append(c.got, fmt.Sprint(v))
	return true
}

func newApproveMock(t *testing.T) (*SiteAdminHandlers, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return NewSiteAdminHandlers(db, zap.NewNop()), mock, db
}

// reviewRow is the SELECT the handler opens with.
func reviewRow(specJSON string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"site_id", "spec", "status"}).
		AddRow(fanSiteID, []byte(specJSON), "needs_human_review")
}

func doApprove(t *testing.T, h *SiteAdminHandlers, reviewData string) (int, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "item_id", Value: fanReviewItemID}}
	body := fmt.Sprintf(`{"review_data": %s}`, reviewData)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/work-items/"+fanReviewItemID+"/approve",
		stringReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.HandleApproveWorkItem(c)

	var decoded map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &decoded)
	return w.Code, decoded
}

func stringReader(s string) *stringReadCloser { return &stringReadCloser{s: s} }

type stringReadCloser struct {
	s string
	i int
}

func (r *stringReadCloser) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

// The two-edit proposal the owner actually approved, verbatim in shape.
const twoEditReviewData = `{
  "edits": [
    {"slot_name": "content-listing", "page_component_id": "` + fanPCLive + `",
     "field_updates": {"section_subtitle": "News, previews and results from across the sport."}},
    {"slot_name": "call-to-action", "page_component_id": "` + fanPCDead + `",
     "field_updates": {"subheadline": ""}}
  ],
  "no_change_needed": false
}`

const fanOutSpec = `{
  "domain": "boxingonline.com",
  "checkpoint": true,
  "on_approve": {
    "item_type": "section_edit",
    "handler_agent": "section-editor",
    "fan_out_from": "edits",
    "defaults": {"edit_type": "content_edit", "page_name": "index"},
    "include_fields": ["domain"]
  }
}`

// One follow-on per edit, each carrying its OWN target and field_updates at the
// top of the spec — the shape section-editor reads.
func TestApproveFansOutOnePerEdit(t *testing.T) {
	h, mock, db := newApproveMock(t)
	defer db.Close()

	specs := &capture{}
	mock.ExpectQuery("SELECT site_id, spec, status").WillReturnRows(reviewRow(fanOutSpec))
	// Both addresses live.
	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("INSERT INTO site_work_items").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), specs, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-1111-1111-111111111111"))
	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("INSERT INTO site_work_items").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), specs, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("22222222-2222-2222-2222-222222222222"))
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	code, body := doApprove(t, h, twoEditReviewData)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	ids, _ := body["follow_on_item_ids"].([]interface{})
	if len(ids) != 2 {
		t.Fatalf("follow_on_item_ids = %v, want 2 (one per edit)", body["follow_on_item_ids"])
	}
	if len(specs.got) != 2 {
		t.Fatalf("captured %d specs, want 2", len(specs.got))
	}

	var first map[string]interface{}
	if err := json.Unmarshal([]byte(specs.got[0]), &first); err != nil {
		t.Fatalf("spec 0 is not JSON: %v", err)
	}
	// The element's own fields, at the top level.
	if first["page_component_id"] != fanPCLive {
		t.Errorf("page_component_id = %v, want the edit's own %s", first["page_component_id"], fanPCLive)
	}
	if first["slot_name"] != "content-listing" {
		t.Errorf("slot_name = %v, want content-listing", first["slot_name"])
	}
	fu, ok := first["field_updates"].(map[string]interface{})
	if !ok || fu["section_subtitle"] != "News, previews and results from across the sport." {
		t.Errorf("field_updates = %v, want this edit's own", first["field_updates"])
	}
	// defaults supply what the proposal does not carry.
	if first["edit_type"] != "content_edit" || first["page_name"] != "index" {
		t.Errorf("defaults not merged: edit_type=%v page_name=%v", first["edit_type"], first["page_name"])
	}
	// include_fields still resolves from the review item's own spec first.
	if first["domain"] != "boxingonline.com" {
		t.Errorf("domain = %v, want boxingonline.com from the review spec", first["domain"])
	}
	// Provenance, applied last.
	if first["source_item_id"] != fanReviewItemID || first["approved_by"] != "admin" {
		t.Errorf("provenance lost: %v / %v", first["source_item_id"], first["approved_by"])
	}
	// The whole payload is NOT duplicated into each child — the review row holds it.
	if _, dup := first["approved_data"]; dup {
		t.Errorf("child carries approved_data; the review row is the audit trail")
	}

	var second map[string]interface{}
	_ = json.Unmarshal([]byte(specs.got[1]), &second)
	if second["page_component_id"] != fanPCDead {
		t.Errorf("second child targets %v, want the second edit's own address", second["page_component_id"])
	}
}

// A dead address is refused and REPORTED. Filing it would produce an item that
// dies at load_edit_context with nothing said to the approver.
func TestApproveSkipsDeadComponentAddressAndSaysSo(t *testing.T) {
	h, mock, db := newApproveMock(t)
	defer db.Close()

	mock.ExpectQuery("SELECT site_id, spec, status").WillReturnRows(reviewRow(fanOutSpec))
	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("INSERT INTO site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-1111-1111-111111111111"))
	// Second element's component row is gone.
	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	code, body := doApprove(t, h, twoEditReviewData)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	ids, _ := body["follow_on_item_ids"].([]interface{})
	if len(ids) != 1 {
		t.Fatalf("follow_on_item_ids = %v, want 1 (the dead one must not be filed)", body["follow_on_item_ids"])
	}
	skipped, _ := body["skipped_edits"].([]interface{})
	if len(skipped) != 1 {
		t.Fatalf("skipped_edits = %v, want 1 — a silent drop is the defect", body["skipped_edits"])
	}
	s0, _ := skipped[0].(map[string]interface{})
	if s0["page_component_id"] != fanPCDead {
		t.Errorf("skipped entry names %v, want the dead address %s", s0["page_component_id"], fanPCDead)
	}
	if s0["reason"] == nil || s0["reason"] == "" {
		t.Errorf("skipped entry carries no reason; the approver cannot act on that")
	}
}

// The regression that started it: include_fields naming something the review
// item's spec cannot hold must now resolve from the approved body.
func TestApproveIncludeFieldsFallsBackToTheApprovedBody(t *testing.T) {
	h, mock, db := newApproveMock(t)
	defer db.Close()

	const spec = `{
	  "checkpoint": true,
	  "on_approve": {"item_type": "section_edit", "handler_agent": "section-editor",
	                 "include_fields": ["page_target", "copy_edit"]}
	}`
	specs := &capture{}
	mock.ExpectQuery("SELECT site_id, spec, status").WillReturnRows(reviewRow(spec))
	mock.ExpectQuery("INSERT INTO site_work_items").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), specs, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("33333333-3333-3333-3333-333333333333"))
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	code, body := doApprove(t, h, `{"page_target": "index", "no_change_needed": false}`)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	var got map[string]interface{}
	_ = json.Unmarshal([]byte(specs.got[0]), &got)
	if got["page_target"] != "index" {
		t.Errorf("page_target = %v, want index resolved from the approved body", got["page_target"])
	}
	// Absent in BOTH sources: the key must not appear at all. Writing it as an
	// explicit null is what made the old failure look like real, empty content.
	if v, present := got["copy_edit"]; present {
		t.Errorf("copy_edit present as %v; a field absent from both sources must be absent, not null", v)
	}
}

// Opt-in means opt-in: an on_approve naming neither new key files exactly one
// item, carrying the whole payload, exactly as it did before.
func TestApproveWithoutFanOutIsUnchanged(t *testing.T) {
	h, mock, db := newApproveMock(t)
	defer db.Close()

	const spec = `{"checkpoint": true,
	  "on_approve": {"item_type": "ready_to_build", "handler_agent": "pageflow-builder"}}`
	specs := &capture{}
	mock.ExpectQuery("SELECT site_id, spec, status").WillReturnRows(reviewRow(spec))
	mock.ExpectQuery("INSERT INTO site_work_items").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), specs, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("44444444-4444-4444-4444-444444444444"))
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	code, body := doApprove(t, h, twoEditReviewData)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if body["follow_on_item_id"] != "44444444-4444-4444-4444-444444444444" {
		t.Errorf("follow_on_item_id = %v", body["follow_on_item_id"])
	}

	var got map[string]interface{}
	_ = json.Unmarshal([]byte(specs.got[0]), &got)
	if _, ok := got["approved_data"]; !ok {
		t.Errorf("the single follow-on must still carry approved_data; got %v", got)
	}
	if got["original_pipeline"] != "build" || got["approved_by"] != "admin" {
		t.Errorf("base fields changed: %v", got)
	}
}

// fan_out_from naming something the approved data does not hold must not
// silently file nothing — it falls back to the single item AND says why.
func TestApproveFanOutFromMissingKeyFallsBackAndReports(t *testing.T) {
	h, mock, db := newApproveMock(t)
	defer db.Close()

	const spec = `{"checkpoint": true,
	  "on_approve": {"item_type": "section_edit", "handler_agent": "section-editor",
	                 "fan_out_from": "changes"}}`
	mock.ExpectQuery("SELECT site_id, spec, status").WillReturnRows(reviewRow(spec))
	mock.ExpectQuery("INSERT INTO site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("55555555-5555-5555-5555-555555555555"))
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	code, body := doApprove(t, h, twoEditReviewData)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if body["fan_out_note"] == nil {
		t.Errorf("no fan_out_note; a misnamed fan_out_from must be visible, not silent")
	}
}
