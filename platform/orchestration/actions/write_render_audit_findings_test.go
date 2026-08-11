package actions

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// The assertions here pin EFFECTS (which INSERT ran, with which item_type,
// status, handler and key), never the absence of a query — a vacuous guard
// passes against insertWorkItem because it swallows sqlmock errors
// (work_item_recurrence_test.go's lesson).

// argJSONHasKeys matches a JSON-string argument that contains every named
// top-level key with a non-empty value. It exists because css-patch-agent's
// prompt template renders spec.category/description/suggestion/page_name —
// an AnyArg() here would let the spec silently stop carrying the only keys
// the handler reads (council e49f5935, guardian/debug_historian).
type argJSONHasKeys struct{ keys []string }

func (a argJSONHasKeys) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		if b, okb := v.([]byte); okb {
			s = string(b)
		} else {
			return false
		}
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return false
	}
	for _, k := range a.keys {
		val, present := m[k]
		if !present || val == nil || val == "" {
			return false
		}
	}
	return true
}

// argContains matches a string argument containing a substring — used to pin
// the "[unresolved after N attempts]" label on the third-occurrence test.
type argContains struct{ substr string }

func (a argContains) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	return strings.Contains(s, a.substr)
}

func renderAuditCollected(siteID uuid.UUID, payload map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"site_id": siteID.String(),
		// The coordinator stores an awaited adapter response under
		// output_field.response — the unwrap path is part of what's under test.
		"render_audit": map[string]interface{}{
			"response":          payload,
			"response_status":   "complete",
			"response_received": "2026-08-02T00:00:00Z",
		},
	}
}

func TestWriteRenderAuditFindings_FilesFirmContrastSkipsOverImage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pricingPageID := uuid.New()

	// No locked components on the site.
	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))
	// The site's pages, for first-class page_id resolution.
	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).
			AddRow(pricingPageID.String(), "/pricing.html"))

	mock.ExpectBegin()
	// Two-strike pre-check for the ONE firm finding (the over_image one must
	// never reach here — a second pre-check would fail ExpectationsWereMet).
	mock.ExpectQuery("INTERVAL '7 days'").
		WithArgs(siteID, "contrast_failure:/pricing.html#h2.card-title").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	// The effect: one INSERT with the routed shape. The spec matcher pins the
	// handler-contract keys; page_id ($8) must be the resolved pages row.
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "render-audit", "build", "contrast_failure", "high",
			sqlmock.AnyArg(),
			argJSONHasKeys{keys: []string{"category", "description", "suggestion", "page_name", "affected_url", "selector", "acceptance_test"}},
			pricingPageID, sqlmock.AnyArg(),
			60, "css-patch-agent", "detected", sqlmock.AnyArg(),
			"contrast_failure:/pricing.html#h2.card-title", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id": "run-1",
			"contrast": []map[string]interface{}{
				{
					"url": "https://example.com/pricing.html", "tag": "h2",
					"class": "card-title muted", "text": "Plans",
					"fg": "#111111", "bg": "#0f0f0f",
					"ratio": 1.2, "need": 4.5, "font_px": 20, "over_image": false,
				},
				{
					"url": "https://example.com/index.html", "tag": "p",
					"class": "hero-sub", "fg": "#ffffff", "bg": "#888888",
					"ratio": 2.9, "need": 4.5, "over_image": true,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["inserted"] != 1 || m["over_image_reported"] != 1 {
		t.Fatalf("want inserted=1 over_image_reported=1, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestWriteRenderAuditFindings_LockedCulpritIsSkippedAndCounted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	// A locked component whose markup carries the finding's class token.
	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).
			AddRow(`<section><h2 class="card-title">Locked</h2></section>`))
	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}))
	// No pre-check, no INSERT — the tx opens and commits empty. Asserting the
	// commit (an effect) rather than "no insert happened" keeps this non-vacuous:
	// an unexpected INSERT fails ExpectationsWereMet loudly.
	mock.ExpectBegin()
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id": "run-2",
			"contrast": []map[string]interface{}{
				{
					"url": "https://example.com/about.html", "tag": "h2",
					"class": "card-title", "fg": "#222222", "bg": "#111111",
					"ratio": 1.1, "need": 4.5, "over_image": false,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["skipped_locked"] != 1 || m["inserted"] != 0 {
		t.Fatalf("want skipped_locked=1 inserted=0, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestWriteRenderAuditFindings_BrokenImagesAttributedAndNot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	assetID := uuid.New()

	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))
	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}))

	// hero.jpg resolves to an assets row; ghost.jpg does not.
	mock.ExpectQuery("FROM assets").
		WithArgs(siteID, "hero.jpg").
		WillReturnRows(sqlmock.NewRows([]string{"id", "purpose"}).AddRow(assetID.String(), "content_hero"))
	mock.ExpectQuery("FROM assets").
		WithArgs(siteID, "ghost.jpg").
		WillReturnRows(sqlmock.NewRows([]string{"id", "purpose"}))

	mock.ExpectBegin()
	mock.ExpectQuery("INTERVAL '7 days'").
		WithArgs(siteID, "undeployed_asset:"+assetID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "render-audit", "build", "undeployed_asset", "medium",
			sqlmock.AnyArg(),
			argJSONHasKeys{keys: []string{"asset_id", "purpose", "affected_url", "src"}},
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			60, "asset-deployer", "detected", sqlmock.AnyArg(),
			"undeployed_asset:"+assetID.String(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id": "run-3",
			"broken_images": []map[string]interface{}{
				{"url": "https://example.com/index.html", "src": "/assets/images/hero.jpg"},
				{"url": "https://example.com/index.html", "src": "/assets/images/ghost.jpg"},
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["inserted"] != 1 || m["unattributed_images"] != 1 {
		t.Fatalf("want inserted=1 unattributed_images=1, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestWriteRenderAuditFindings_ThirdOccurrenceIsBornUnresolvedNotDropped is
// the direct answer to council e49f5935's gating objection (bug_historian,
// high): "omitting recurrenceExpected silently drops the third occurrence."
// It does not: for a DETECTED DEFECT the two-strike rule INSERTS the third
// occurrence with status 'unresolved' and a loud label — parked for the admin
// dashboard's attention queues, not vanished. The test pins both halves:
//
//   - the two-strike pre-check RUNS (recurrenceExpected must stay false: if a
//     session "fixes" the objection by setting it, the INTERVAL query is never
//     issued and this test fails on an unmet expectation), and
//   - the third occurrence is INSERTED, born 'unresolved' with the
//     "[unresolved after 2 attempts]" prefix — not suppressed.
//
// Mutation-proven 2026-08-03: setting recurrenceExpected=true on the contrast
// item fails this test twice over (unmet INTERVAL expectation; status arrives
// 'detected'); deleting the two-strike block fails it the same way.
func TestWriteRenderAuditFindings_ThirdOccurrenceIsBornUnresolvedNotDropped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))
	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}))

	mock.ExpectBegin()
	// Two completed fix attempts inside the 7-day window, the newest 24h old:
	// strike two is spent, and this re-file is the third occurrence.
	mock.ExpectQuery("INTERVAL '7 days'").
		WithArgs(siteID, "contrast_failure:/faq.html#p.answer").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(2, 24.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "render-audit", "build", "contrast_failure", "medium",
			argContains{substr: "[unresolved after 2 attempts]"},
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			60, "css-patch-agent", "unresolved", sqlmock.AnyArg(),
			"contrast_failure:/faq.html#p.answer", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id": "run-4",
			"contrast": []map[string]interface{}{
				{
					"url": "https://example.com/faq.html", "tag": "p",
					"class": "answer", "fg": "#777777", "bg": "#999999",
					"ratio": 2.2, "need": 4.5, "font_px": 16, "over_image": false,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["inserted"] != 1 {
		t.Fatalf("third occurrence must be INSERTED (born unresolved), got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestWriteRenderAuditFindings_AbsentAuditIsAnError(t *testing.T) {
	// Absent ≠ malformed ≠ clean: a run whose audit never landed must FAIL the
	// step, not report zero findings written.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	_, err = WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    map[string]interface{}{"site_id": uuid.New().String()},
	})
	if err == nil || !strings.Contains(err.Error(), "has not run") {
		t.Fatalf("want a loud 'has not run' error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no queries should have run: %v", err)
	}
}

func TestWriteRenderAuditFindings_StillAwaitedIsAnError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	_, err = WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id": uuid.New().String(),
			// The await-signal shape the action's own request step returns,
			// with no .response yet.
			"render_audit": map[string]interface{}{"success": true, "request_id": "r-1"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "awaited or failed") {
		t.Fatalf("want a loud still-awaited error, got %v", err)
	}
}

// bugs_open/242: a truncated sweep must not report as a whole-site verdict.
// The summary's cap-bite echo is stamped into this action's durable result —
// parity with findings_capped/findings_dropped, the max_items cap's own honest
// reporting.
func TestWriteRenderAuditFindings_TruncatedSweepIsStampedInResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))
	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}))
	mock.ExpectBegin()
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id":   "run-t",
			"contrast": []map[string]interface{}{},
			"summary": map[string]interface{}{
				"pages": 25, "pages_total": 27, "truncated": true,
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["truncated"] != true {
		t.Fatalf("truncated sweep must stamp truncated=true, got %#v", m)
	}
	if m["pages_total"] != 27 || m["pages_audited"] != 25 {
		t.Fatalf("want pages_total=27 pages_audited=25, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// The control: an untruncated (or old-shape) summary leaves the result shape
// unchanged for every existing consumer — no keys added.
func TestWriteRenderAuditFindings_UntruncatedSweepAddsNoKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))
	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}))
	mock.ExpectBegin()
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id":   "run-u",
			"contrast": []map[string]interface{}{},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	for _, key := range []string{"truncated", "pages_total", "pages_audited"} {
		if _, present := m[key]; present {
			t.Errorf("untruncated result must not carry %q, got %#v", key, m)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
