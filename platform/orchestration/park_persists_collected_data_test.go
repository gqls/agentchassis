// FILE: platform/orchestration/park_persists_collected_data_test.go
//
// End-to-end cover for the park path: drives the REAL persistAwaitingStateWithRetry
// against a mocked DB and asserts on the collected_data JSON it actually writes.
//
// Why this exists alongside park_carries_collected_data_test.go: those tests call
// the carry helper directly, so they would pass whether or not the park is wired
// to it — a helper with no live caller looks exactly like a finished refactor.
// This test fails if the call site is removed, which is the only way to catch it.
package orchestration

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// updateArgCount is the number of placeholders in UpdateStateWithVersion's
// UPDATE; collectedDataArgIndex is collected_data's zero-based position ($6).
const (
	updateArgCount        = 22
	collectedDataArgIndex = 5
)

// argCaptor records the value it is matched against and accepts anything, so a
// single positional expectation can both pass and capture.
type argCaptor struct{ dest *string }

func (c argCaptor) Match(v driver.Value) bool {
	switch typed := v.(type) {
	case []byte:
		*c.dest = string(typed)
	case string:
		*c.dest = typed
	}
	return true
}

// freshRowColumns is GetState's SELECT list, in its scan order.
var freshRowColumns = []string{
	"orchestration_id", "orchestration_name", "correlation_id", "owner_agent_id", "owner_agent_type",
	"owner_agent_role", "parent_orchestration_id", "client_id",
	"requests_topic", "responses_topic",
	"status", "current_step", "awaited_steps", "awaited_requests",
	"currently_executing", "last_activity", "processing_node", "execution_started_at",
	"collected_data", "initial_request_data", "final_result", "workflow_plan",
	"execution_path", "execution_metadata", "processing_history", "subtree_agents",
	"fuel_budget", "error", "version", "created_at", "updated_at",
	"site_id",
}

// freshRow is the DB copy the park reloads: earlier steps' keys are on it, but
// nothing from the step currently dispatching.
func freshRow(collectedData string) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(freshRowColumns).AddRow(
		"orch-1", "site-work-orchestrator", "corr-1", "agent-1", "site-work-orchestrator",
		"owner", nil, "client-1",
		"requests.topic", "responses.topic",
		"processing", "deploy_logo_image", []byte(`[]`), []byte(`{}`),
		nil, now, "pod-1", nil,
		[]byte(collectedData), []byte(`{}`), nil, []byte(`{}`),
		[]byte(`[]`), []byte(`{}`), []byte(`[]`), []byte(`{}`),
		0, nil, 7, now, now,
		nil,
	)
}

func parkingState() *OrchestrationState {
	return &OrchestrationState{
		OrchestrationID: "orch-1",
		CurrentStep:     "deploy_logo_image",
		Version:         7,
		CollectedData: map[string]interface{}{
			"logo_deployed": logoDeployedInMemory(),
		},
		AwaitedRequests: map[string]*AwaitedRequest{
			"req-logo-1": {RequestID: "req-logo-1", StepName: "deploy_logo_image"},
		},
	}
}

// The defect, end to end: the action's image_url must reach the PERSISTED
// collected_data, not merely the in-memory map.
func TestParkPersistsTheActionsOwnKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	var persisted string
	args := make([]driver.Value, updateArgCount)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[collectedDataArgIndex] = argCaptor{dest: &persisted}

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(freshRow(`{"input_data":{"domain":"cookly.uk"}}`))
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewStateRepository(db, zap.NewNop())
	if err := persistAwaitingStateWithRetry(context.Background(), parkingState(), repo, zap.NewNop()); err != nil {
		t.Fatalf("park failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	var round map[string]interface{}
	if err := json.Unmarshal([]byte(persisted), &round); err != nil {
		t.Fatalf("persisted collected_data is not JSON: %v\n%s", err, persisted)
	}

	container, ok := round["logo_deployed"].(map[string]interface{})
	if !ok {
		t.Fatalf("logo_deployed was not persisted at park; collected_data=%s", persisted)
	}
	if container["image_url"] != "/assets/images/logo.png" {
		t.Errorf("persisted image_url = %v, want /assets/images/logo.png", container["image_url"])
	}
	if round["input_data"] == nil {
		t.Error("the fresh row's own keys were lost - the carry must be additive, not a replacement")
	}
}

// The reply-beats-park race: when the fresh row already shows an arrived
// response, the park returns early and must not write at all.
func TestParkDoesNotWriteWhenTheReplyAlreadyArrived(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(freshRow(`{"deploy_logo_image":{"response":{"data":{"success":true}}}}`))
	// Deliberately no ExpectExec: any UPDATE here is the failure.

	repo := NewStateRepository(db, zap.NewNop())
	if err := persistAwaitingStateWithRetry(context.Background(), parkingState(), repo, zap.NewNop()); err != nil {
		t.Fatalf("park failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("park issued an unexpected write after the reply had landed: %v", err)
	}
}
