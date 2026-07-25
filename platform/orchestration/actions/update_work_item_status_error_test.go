package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// Covers bugs_closed/040-partial-build candidate 2: update_work_item_status
// left site_work_items.error EMPTY on the one path that actually fails.
//
// The literal `error_message` config only fits a STATIC reason, so the two
// steps with a dynamic one — page-build-handler / image-build-handler's
// `mark_item_failed`, reached via error_step — carried no literal at all. The
// coordinator had already written the real message to agent_error_log from the
// same routeToErrorStep call that set __step_error, so the reason existed; it
// simply was not on the row triage reads. Live 2026-07-25: 21 of 75 failed
// items blank, 20 of them with exactly one agent_error_log row waiting.
//
// Fixtures are the real live shapes: the awaited-request timeout that stranded
// dartsonline (bare message, needs the step name) and the action error (already
// self-describing, must not be prefixed twice).

// captureStatusUpdate runs UpdateWorkItemStatusAction against sqlmock and
// returns the value it wrote to the error column.
func captureStatusUpdate(t *testing.T, config map[string]interface{}, collected map[string]interface{}) string {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	itemID := uuid.New()
	if collected == nil {
		collected = map[string]interface{}{}
	}
	if _, ok := collected["input_data"]; !ok {
		collected["input_data"] = map[string]interface{}{"work_item_id": itemID.String()}
	}

	var gotError string
	mock.ExpectExec(`UPDATE site_work_items`).
		WithArgs(itemID, sqlmock.AnyArg(), sqlmock.AnyArg(), captureArg{got: &gotError}).
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: config},
	}

	if _, err := UpdateWorkItemStatusAction(context.Background(), params); err != nil {
		t.Fatalf("UpdateWorkItemStatusAction: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	return gotError
}

func TestUpdateWorkItemStatus_RecordsRoutedStepError(t *testing.T) {
	// The live shape that produced the blank errors: an awaited-request
	// timeout. The message alone does not say WHAT timed out, so the failing
	// step is prefixed on — converging on the "step X failed: …" form the
	// column already uses rather than inventing a second format.
	routed := func(step, msg string) map[string]interface{} {
		return map[string]interface{}{
			"__step_error": map[string]interface{}{
				"failed_step": step,
				"message":     msg,
			},
		}
	}

	cases := []struct {
		name      string
		config    map[string]interface{}
		collected map[string]interface{}
		want      string
	}{
		{
			name:      "failed with no literal — records the routed error, named by step",
			config:    map[string]interface{}{"status": "failed"},
			collected: routed("deploy_page", "Request 59150fa3-2e93-4661-855a-e56abbf8012d timed out after 3 retries"),
			want:      "step deploy_page failed: Request 59150fa3-2e93-4661-855a-e56abbf8012d timed out after 3 retries",
		},
		{
			name:      "message already names its step — not prefixed twice",
			config:    map[string]interface{}{"status": "failed"},
			collected: routed("validate_content", "step validate_content failed: failed to execute action validate_page_content: content validation failed"),
			want:      "step validate_content failed: failed to execute action validate_page_content: content validation failed",
		},
		{
			// A configured literal is the workflow author's deliberate wording
			// (the two needs_human_review steps rely on this).
			name: "configured literal always wins",
			config: map[string]interface{}{
				"status":        "needs_human_review",
				"error_message": "page-build-handler no-op: content writer skipped this page",
			},
			collected: routed("write_content", "some routed error"),
			want:      "page-build-handler no-op: content writer skipped this page",
		},
		{
			// __step_error is never cleared once set, so a workflow that
			// recovers and then completes the item must not be handed a stale
			// failure. image-build-handler has exactly this literal-less
			// 'complete' step.
			name:      "complete never inherits a routed error",
			config:    map[string]interface{}{"status": "complete"},
			collected: routed("call_generator", "Request abc timed out after 3 retries"),
			want:      "",
		},
		{
			name:      "no routed error and no literal — writes nothing, as before",
			config:    map[string]interface{}{"status": "failed"},
			collected: map[string]interface{}{},
			want:      "",
		},
		{
			// Defensive: a routed entry with no failed_step still yields the
			// message rather than a bare prefix.
			name:   "routed error without a step name — message alone",
			config: map[string]interface{}{"status": "failed"},
			collected: map[string]interface{}{
				"__step_error": map[string]interface{}{"message": "handler pod died"},
			},
			want: "handler pod died",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := captureStatusUpdate(t, tc.config, tc.collected); got != tc.want {
				t.Errorf("error column:\n got  %q\n want %q", got, tc.want)
			}
		})
	}
}
