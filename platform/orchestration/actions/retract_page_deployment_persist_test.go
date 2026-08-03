// FILE: platform/orchestration/actions/retract_page_deployment_persist_test.go
//
// Debt 5 of bugs_open/098: the retraction's audit findings must SURVIVE the
// await. The action returns its findings and awaits the adapter's reply, and
// the await machinery replaces the step-name and output_field keys with that
// reply when it lands — so on the one real retraction ever run, the record
// held only {paths, success, …} and every refusal survived nowhere but pod
// logs. A retraction that refused a page refused it silently.
//
// These tests pin the two sinks that fix it, at the ACTION level rather than
// helper level on purpose: a helper test still passes after someone deletes
// the call site, and the call site — audit attached before dispatch, rows
// written before dispatch — is the behaviour.
package actions

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// failingProducer refuses every publish — the dispatch-failure arm.
type failingProducer struct{}

func (failingProducer) Produce(context.Context, string, map[string]string, []byte, []byte) error {
	return errors.New("kafka is down")
}
func (p failingProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return p.Produce(ctx, topic, headers, key, value)
}
func (failingProducer) Close() error { return nil }

func retractionPersistParams(db *sql.DB, producer kafka.Producer, config map[string]interface{}) (ActionParams, map[string]interface{}) {
	collected := map[string]interface{}{
		"input_data": map[string]interface{}{},
	}
	return ActionParams{
		ExecutionContext: &types.ExecutionContext{
			CorrelationID:   "corr-1",
			OrchestrationID: "orch-1",
			ClientID:        "client-1",
			ResponsesTopic:  "resp-topic",
			StepName:        "retract",
			StepID:          "step-1",
			Sender: types.AgentIdentity{
				AgentType: "page-retraction",
				AgentID:   "agent-1",
				PodName:   "pod-1",
			},
		},
		StepConfig:    models.Step{Name: "retract", Action: "retract_page_deployment", OutputField: "retraction", Config: config},
		CollectedData: collected,
		Producer:      producer,
		DB:            db,
		Logger:        zap.NewNop(),
	}, collected
}

// The 13-arg INSERT into agent_error_log, asserting the columns that make the
// row findable: action, error_code, severity. Everything else is identity
// plumbing.
func expectConditionInsert(mock sqlmock.Sqlmock, code string) {
	mock.ExpectExec(`INSERT INTO agent_error_log`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"retract_page_deployment",
			sqlmock.AnyArg(), code, "warning", sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func auditFromCollected(t *testing.T, collected map[string]interface{}) map[string]interface{} {
	t.Helper()
	audit, ok := collected[retractionAuditKey].(map[string]interface{})
	if !ok {
		t.Fatalf("collected_data[%q] missing or not a map: %T — the audit did not outlive the action", retractionAuditKey, collected[retractionAuditKey])
	}
	return audit
}

func refusedCount(t *testing.T, audit map[string]interface{}) int {
	t.Helper()
	cands, ok := audit["candidates"].([]retractionCandidate)
	if !ok {
		t.Fatalf("audit candidates missing or wrong type: %T", audit["candidates"])
	}
	n := 0
	for _, c := range cands {
		if c.Refused != "" {
			n++
		}
	}
	return n
}

// TestRetractionAuditSurvivesTheAwaitOverwrite is the defect's repro, inverted.
// A real run: one eligible page, two refused (one active, one with an
// underivable url), one nav row retired, dispatch succeeds. Then the await
// overwrite is applied exactly as applyResponseToState's default branch does —
// both the step-name key and the output_field key replaced wholesale by the
// adapter's reply — and the audit, refusals included, must still be there.
func TestRetractionAuditSurvivesTheAwaitOverwrite(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()

	mock.ExpectQuery(`SELECT domain FROM sites`).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("robot-hands.com"))
	mock.ExpectQuery(`deployed_at IS NOT NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "status"}).
			AddRow(uuid.New().String(), "gone-page", "/gone.html", "archived").
			AddRow(uuid.New().String(), "live-page", "/live.html", "active").
			AddRow(uuid.New().String(), "frag-page", "/tools.html#frag", "archived"))
	mock.ExpectQuery(`= 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}))
	// Graph audit for the one eligible page.
	mock.ExpectQuery(`FROM unnest`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "slot", "url"}))
	mock.ExpectQuery(`site_components`).
		WillReturnRows(sqlmock.NewRows([]string{"slot", "url"}))
	mock.ExpectQuery(`site_nav_items`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "url"}).
			AddRow("nav-1", "Gone", "/gone.html"))
	mock.ExpectQuery(`WITH outbound`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "status"}))
	// The two refusals become durable rows BEFORE the nav write and dispatch.
	expectConditionInsert(mock, "RETRACTION_REFUSED")
	expectConditionInsert(mock, "RETRACTION_REFUSED")
	mock.ExpectExec(`UPDATE site_nav_items`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	producer := &capturingProducer{}
	params, collected := retractionPersistParams(db, producer,
		map[string]interface{}{"site_id": siteID.String(), "repo_name": "sites"})

	res, err := RetractPageDeploymentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("RetractPageDeploymentAction: %v", err)
	}
	result := res.(map[string]interface{})
	if result["await_response"] != true {
		t.Fatalf("expected an awaited dispatch, got %+v", result)
	}
	if producer.value == nil {
		t.Fatal("nothing was produced to the git adapter")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (refusal rows or nav update missing): %v", err)
	}

	audit := auditFromCollected(t, collected)
	if got := refusedCount(t, audit); got != 2 {
		t.Errorf("audit carries %d refusals, want 2", got)
	}
	if audit["dispatched"] != true {
		t.Errorf("audit.dispatched = %v, want true", audit["dispatched"])
	}
	if audit["nav_retired"] != 1 {
		t.Errorf("audit.nav_retired = %v, want 1", audit["nav_retired"])
	}
	if rid, _ := audit["request_id"].(string); rid == "" {
		t.Error("audit.request_id is empty — the record cannot be joined to the adapter request")
	}
	if audit["conditions_recorded"] != 2 {
		t.Errorf("audit.conditions_recorded = %v, want 2", audit["conditions_recorded"])
	}

	// The await overwrite, verbatim in effect: applyResponseToState's default
	// branch does state.CollectedData[stepName] = reply and
	// state.CollectedData[outputField] = reply. Nothing else.
	adapterReply := map[string]interface{}{"success": true, "paths": []string{"robot-hands.com/gone.html"}}
	collected["retract"] = adapterReply
	collected["retraction"] = adapterReply

	audit = auditFromCollected(t, collected)
	if got := refusedCount(t, audit); got != 2 {
		t.Errorf("after the await overwrite the audit lost its refusals: %d, want 2", got)
	}
}

// TestRetractionDryRunRecordsNothingDurable — a dry run must not create
// fleet-visible warning rows, and its audit must say it was a dry run. The
// recorder's absence is proven by the audit itself: conditions_recorded is
// only ever set by a real run, so a dry run that starts writing rows fails
// here on the key appearing.
func TestRetractionDryRunRecordsNothingDurable(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()

	mock.ExpectQuery(`SELECT domain FROM sites`).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("robot-hands.com"))
	mock.ExpectQuery(`deployed_at IS NOT NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "status"}).
			AddRow(uuid.New().String(), "gone-page", "/gone.html", "archived").
			AddRow(uuid.New().String(), "live-page", "/live.html", "active"))
	mock.ExpectQuery(`= 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}))
	mock.ExpectQuery(`FROM unnest`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "slot", "url"}))
	mock.ExpectQuery(`site_components`).
		WillReturnRows(sqlmock.NewRows([]string{"slot", "url"}))
	mock.ExpectQuery(`site_nav_items`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "url"}))
	mock.ExpectQuery(`WITH outbound`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "status"}))

	producer := &capturingProducer{}
	params, collected := retractionPersistParams(db, producer,
		map[string]interface{}{"site_id": siteID.String(), "dry_run": true})

	if _, err := RetractPageDeploymentAction(context.Background(), params); err != nil {
		t.Fatalf("RetractPageDeploymentAction: %v", err)
	}
	if producer.value != nil {
		t.Fatal("a dry run dispatched to the adapter")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}

	audit := auditFromCollected(t, collected)
	if audit["dry_run"] != true {
		t.Errorf("audit.dry_run = %v, want true", audit["dry_run"])
	}
	if _, present := audit["conditions_recorded"]; present {
		t.Error("a dry run wrote durable condition rows — dry means dry")
	}
}

// TestRetractionAllRefusedStillRecords — the loudest silent case: every
// candidate refused, nothing dispatched, and BEFORE this fix the refusals'
// only trace was a return value nobody overwrote but whose record expires
// with the orchestration row. The refusal row must be written even though the
// run ends at nothing_to_retract.
func TestRetractionAllRefusedStillRecords(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()

	mock.ExpectQuery(`SELECT domain FROM sites`).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("robot-hands.com"))
	mock.ExpectQuery(`deployed_at IS NOT NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "status"}).
			AddRow(uuid.New().String(), "live-page", "/live.html", "active"))
	mock.ExpectQuery(`= 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}))
	// No graph-audit queries: the eligible set is empty.
	expectConditionInsert(mock, "RETRACTION_REFUSED")

	producer := &capturingProducer{}
	params, collected := retractionPersistParams(db, producer,
		map[string]interface{}{"site_id": siteID.String()})

	res, err := RetractPageDeploymentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("RetractPageDeploymentAction: %v", err)
	}
	result := res.(map[string]interface{})
	if result["status"] != "nothing_to_retract" {
		t.Fatalf("status = %v, want nothing_to_retract", result["status"])
	}
	if producer.value != nil {
		t.Fatal("an all-refused run dispatched to the adapter")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the refusal row was not written: %v", err)
	}

	audit := auditFromCollected(t, collected)
	if audit["status"] != "nothing_to_retract" {
		t.Errorf("audit.status = %v, want nothing_to_retract", audit["status"])
	}
	if audit["conditions_recorded"] != 1 {
		t.Errorf("audit.conditions_recorded = %v, want 1", audit["conditions_recorded"])
	}
}

// TestRetractionDispatchFailureDoesNotUnrecord — the rows and the audit are
// written BEFORE the send, so a dispatch failure returns an error and leaves
// the record standing, with dispatched=false telling the reader exactly how
// far the run got.
func TestRetractionDispatchFailureDoesNotUnrecord(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()

	mock.ExpectQuery(`SELECT domain FROM sites`).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("robot-hands.com"))
	mock.ExpectQuery(`deployed_at IS NOT NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "status"}).
			AddRow(uuid.New().String(), "gone-page", "/gone.html", "archived").
			AddRow(uuid.New().String(), "frag-page", "/tools.html#frag", "archived"))
	mock.ExpectQuery(`= 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}))
	mock.ExpectQuery(`FROM unnest`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "slot", "url"}))
	mock.ExpectQuery(`site_components`).
		WillReturnRows(sqlmock.NewRows([]string{"slot", "url"}))
	mock.ExpectQuery(`site_nav_items`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "url"}))
	mock.ExpectQuery(`WITH outbound`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "status"}))
	expectConditionInsert(mock, "RETRACTION_REFUSED")

	params, collected := retractionPersistParams(db, failingProducer{},
		map[string]interface{}{"site_id": siteID.String(), "repo_name": "sites"})

	if _, err := RetractPageDeploymentAction(context.Background(), params); err == nil {
		t.Fatal("a failed dispatch must fail the action")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the refusal row vanished with the failed dispatch: %v", err)
	}

	audit := auditFromCollected(t, collected)
	if audit["dispatched"] != false {
		t.Errorf("audit.dispatched = %v, want false — the send never happened", audit["dispatched"])
	}
	if audit["conditions_recorded"] != 1 {
		t.Errorf("audit.conditions_recorded = %v, want 1", audit["conditions_recorded"])
	}
}
