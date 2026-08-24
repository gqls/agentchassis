// FILE: internal/core-manager/admin/spec_update_guard_test.go
//
// Covers HandleUpdateSiteSpec's evidence_base guard.
//
// The defect these tests exist to prevent: the handler bound the spec body as
// json.RawMessage and validated nothing about its shape, while the reading end
// (datahelpers.ParseEvidenceBase) returns nil with NO error for well-formed
// JSON of the wrong shape — a misspelt key, a pasted fragment, an object
// nested one level too deep. One such save superseded the good register with a
// 200 and a success toast, and every claims lane on the site silently no-oped
// from then on; the site then reported clean because nothing was checking.
//
// The guard's contract:
//  1. evidence_base data that fails to unmarshal → 400, nothing written.
//  2. evidence_base data that parses to nothing scannable, where the current
//     register parses non-nil → 409 EMPTY_EVIDENCE_BASE, nothing written,
//     unless confirm_empty=true accompanies the save.
//  3. A successful evidence_base save reports facts/banned-claims counts —
//     the one signal a wrong-shape save cannot fake.
//  4. Other aspects are untouched by the guard (they are prompt text, not
//     controls), but every save now reports `superseded` so a typo'd aspect
//     name (there is no allow-list) is visible as created-not-updated.

package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const guardTestSiteID = "1c6f3424-9d05-4a18-963b-72541bc19dca"

// A register that parses non-nil: one fact, two banned claims.
const goodRegister = `{
	"audit_doc": "test",
	"facts": [{"id": "f1", "claim": "established 2001", "kind": "entity"}],
	"banned_claims": [
		{"pattern": "award-winning", "reason": "no award exists"},
		{"pattern": "24/7", "reason": "not staffed 24/7"}
	]
}`

// Well-formed JSON, wrong shape: bannedClaims is not a key ParseEvidenceBase
// knows, so this parses cleanly into an empty struct and comes back nil.
const wrongShapeRegister = `{"bannedClaims": [{"pattern": "award-winning"}]}`

func doSpecUpdate(t *testing.T, h *SiteAdminHandlers, aspect, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "site_id", Value: guardTestSiteID},
		{Key: "aspect", Value: aspect},
	}
	c.Request = httptest.NewRequest(http.MethodPatch,
		"/admin/sites/"+guardTestSiteID+"/specs/"+aspect, strings.NewReader(body))
	h.HandleUpdateSiteSpec(c)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (body %s)", err, w.Body.String())
	}
	return body
}

// expectWrite arms the mock for the supersede-then-insert transaction.
func expectWrite(mock sqlmock.Sqlmock, supersededRows int64) {
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE site_specs").
		WillReturnResult(sqlmock.NewResult(0, supersededRows))
	mock.ExpectQuery("INSERT INTO site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow("7d9a1f00-0000-0000-0000-000000000001"))
	mock.ExpectCommit()
}

func TestEvidenceBaseWrongShapeSaveIsRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSiteAdminHandlers(db, zap.NewNop())

	// Only the current-register read is armed. If the handler proceeds to the
	// write transaction anyway, Begin fails and the 409 assert below catches it.
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(goodRegister)))

	w := doSpecUpdate(t, h, "evidence_base", `{"data": `+wrongShapeRegister+`}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["code"] != "EMPTY_EVIDENCE_BASE" {
		t.Errorf("code = %v, want EMPTY_EVIDENCE_BASE", body["code"])
	}
	if body["current_banned_claims_count"] != float64(2) {
		t.Errorf("current_banned_claims_count = %v, want 2", body["current_banned_claims_count"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestEvidenceBaseConfirmEmptyOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSiteAdminHandlers(db, zap.NewNop())

	// confirm_empty skips the current-register read entirely: no SELECT armed.
	expectWrite(mock, 1)

	w := doSpecUpdate(t, h, "evidence_base",
		`{"data": `+wrongShapeRegister+`, "confirm_empty": true}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["banned_claims_count"] != float64(0) || body["facts_count"] != float64(0) {
		t.Errorf("counts = %v/%v, want 0/0 — the zero is the operator's warning",
			body["facts_count"], body["banned_claims_count"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestEvidenceBaseEmptyOverAbsentProceeds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSiteAdminHandlers(db, zap.NewNop())

	// No current register: the read returns no rows, so there is nothing to
	// protect and the save proceeds (counts still report the zeros).
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}))
	expectWrite(mock, 0)

	w := doSpecUpdate(t, h, "evidence_base", `{"data": `+wrongShapeRegister+`}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["superseded"] != false {
		t.Errorf("superseded = %v, want false (nothing was current)", body["superseded"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestEvidenceBaseGoodSaveReturnsCounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSiteAdminHandlers(db, zap.NewNop())

	expectWrite(mock, 1)

	w := doSpecUpdate(t, h, "evidence_base", `{"data": `+goodRegister+`}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["facts_count"] != float64(1) || body["banned_claims_count"] != float64(2) {
		t.Errorf("counts = %v facts / %v bans, want 1/2",
			body["facts_count"], body["banned_claims_count"])
	}
	if body["superseded"] != true {
		t.Errorf("superseded = %v, want true", body["superseded"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestEvidenceBaseUnparseableIs400(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSiteAdminHandlers(db, zap.NewNop())

	// A JSON array is valid JSON but cannot unmarshal into EvidenceBase.
	// Nothing is armed: no read, no write.
	w := doSpecUpdate(t, h, "evidence_base", `{"data": [1, 2, 3]}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestOtherAspectsBypassTheGuard(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSiteAdminHandlers(db, zap.NewNop())

	// Prompt-text aspects go straight to the write: no guard SELECT armed.
	expectWrite(mock, 1)

	w := doSpecUpdate(t, h, "content_direction", `{"data": {"tone": "plain"}}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if _, present := body["facts_count"]; present {
		t.Errorf("facts_count present on a non-evidence_base save")
	}
	if body["superseded"] != true {
		t.Errorf("superseded = %v, want true", body["superseded"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
