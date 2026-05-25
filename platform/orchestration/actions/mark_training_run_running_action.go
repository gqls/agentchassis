// FILE: platform/orchestration/actions/mark_training_run_running_action.go
//
// mark_training_run_running: flips a model_lifecycle.training_runs row from
// 'pending' to 'running' and stamps started_at. The sibling of
// markTrainingRunFailed (prepare_training_data_action.go) — same table, same
// params.DB.ExecContext idiom, opposite terminal direction.
//
// Used by the training-launcher (Phase 5) as the step AFTER ssh_exec has
// successfully launched the backgrounded training process on the VM. At that
// point the run is genuinely underway, so the row should reflect 'running'.
// (A later training-monitor / artefact-collector moves it to 'complete' or
// 'failed'; this action only owns the pending→running transition.)
//
// Input (from CollectedData["input_data"]):
//   - training_run_id: UUID of the training_runs row (required)
//   - thunder_instance_id: optional audit — which instance is running it
//     (stored in training_runs.thunder_instance_id, TEXT)
//
// Idempotent: the WHERE clause only transitions rows still in 'pending', so a
// re-drive is a harmless no-op once the row is already 'running' or terminal.

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

var MarkTrainingRunRunningInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"training_run_id"},
	Optional:   []string{"thunder_instance_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("mark_training_run_running", MarkTrainingRunRunningInputSpec)
}

// MarkTrainingRunRunningAction performs the pending→running UPDATE.
func MarkTrainingRunRunningAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "mark_training_run_running"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("mark_training_run_running: database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		MarkTrainingRunRunningInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("mark_training_run_running: input extraction failed: %w", err)
	}

	runID, err := uuid.Parse(inputs.Get("training_run_id"))
	if err != nil {
		return nil, fmt.Errorf("mark_training_run_running: invalid training_run_id: %w", err)
	}
	thunderInstanceID := datahelpers.NullableString(inputs.Get("thunder_instance_id"))

	// Only transition rows still 'pending' (idempotent re-drive). Record the
	// instance id that's running it when supplied. started_at stamped NOW().
	res, err := params.DB.ExecContext(ctx, `
		UPDATE model_lifecycle.training_runs
		SET status = 'running',
		    started_at = NOW(),
		    thunder_instance_id = COALESCE($2, thunder_instance_id),
		    updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, runID, thunderInstanceID)
	if err != nil {
		return nil, fmt.Errorf("mark_training_run_running: update failed: %w", err)
	}

	n, _ := res.RowsAffected()
	logger.Info("Marked training run running",
		zap.String("training_run_id", runID.String()),
		zap.Int64("rows_affected", n),
	)

	// rows_affected == 0 means the row wasn't 'pending' (already running/terminal,
	// or missing). Not an error — report it so the workflow/monitor can see it.
	return map[string]interface{}{
		"training_run_id": runID.String(),
		"transitioned":    n > 0,
		"status":          "running",
	}, nil
}
