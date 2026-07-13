// FILE: platform/orchestration/actions/mark_training_run_terminal_action.go
//
// mark_training_run_terminal: flips a model_lifecycle.training_runs row from
// 'running' to a terminal state — 'complete' or 'failed' — chosen by the step's
// config `status`. The counterpart to mark_training_run_running (pending→running);
// together they own the run's lifecycle transitions. Used by the
// thunder-training-monitor's mark_complete (status=complete) and mark_failed
// (status=failed) steps, both of which then route to decommission.
//
// Why this exists rather than reusing markTrainingRunFailed: that is an UNEXPORTED
// helper inside prepare_training_data_action.go (the data-preparer's own error
// path), not a registered workflow action — the monitor needs a registered one,
// and it needs the 'complete' direction too. This action mirrors that helper's
// column shape (status, completed_at, error_message) so the two stay consistent.
//
// Schema facts (019_model_lifecycle_schema.sql, verified):
//   - status CHECK IN ('pending','running','complete','failed') — the literal is
//     'complete' (not 'completed').
//   - completed_at TIMESTAMPTZ — stamped on BOTH terminal outcomes (matching the
//     existing markTrainingRunFailed).
//   - error_message TEXT — populated only when status='failed'; cleared to NULL
//     on 'complete'.
//
// Input resolution (per 001 §Config value patterns — literals direct, paths via spec):
//   - status:        config literal "complete"|"failed" — read directly from
//                    params.StepConfig.Config via datahelpers.GetStringField.
//   - training_run_id: a value from input_data (input_mapping at spawn) — read via
//                    ExtractActionInputs, mirroring mark_training_run_running.
//   - error_message: optional config literal (failed only) via GetStringField; a
//                    default is used if absent.
//
// Idempotent: the WHERE clause only transitions rows still 'running', so a
// re-drive once the row is already terminal is a harmless no-op and an existing
// terminal state is never clobbered (no complete→failed flip).

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

var MarkTrainingRunTerminalInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"training_run_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("mark_training_run_terminal", MarkTrainingRunTerminalInputSpec)
}

// MarkTrainingRunTerminalAction performs the running→complete|failed UPDATE.
func MarkTrainingRunTerminalAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "mark_training_run_terminal"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("mark_training_run_terminal: database connection required")
	}

	status := strings.ToLower(strings.TrimSpace(datahelpers.GetStringField(params.StepConfig.Config, "status", "")))
	if status != "complete" && status != "failed" {
		return nil, fmt.Errorf("mark_training_run_terminal: status must be \"complete\" or \"failed\", got %q", status)
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		MarkTrainingRunTerminalInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("mark_training_run_terminal: input extraction failed: %w", err)
	}
	runID, err := uuid.Parse(inputs.Get("training_run_id"))
	if err != nil {
		return nil, fmt.Errorf("mark_training_run_terminal: invalid training_run_id: %w", err)
	}

	// Only transition rows still 'running' (idempotent; never clobbers terminal).
	// completed_at stamped for both outcomes; error_message NULL on complete,
	// populated on failed.
	var n int64
	if status == "complete" {
		r, e := params.DB.ExecContext(ctx, `
			UPDATE model_lifecycle.training_runs
			   SET status = 'complete',
			       completed_at = NOW(),
			       error_message = NULL
			 WHERE id = $1 AND status = 'running'
		`, runID)
		if e != nil {
			return nil, fmt.Errorf("mark_training_run_terminal: update to complete failed: %w", e)
		}
		n, _ = r.RowsAffected()
	} else {
		errMsg := strings.TrimSpace(datahelpers.GetStringField(params.StepConfig.Config, "error_message", ""))
		if errMsg == "" {
			errMsg = "training run marked failed by thunder-training-monitor"
		}
		r, e := params.DB.ExecContext(ctx, `
			UPDATE model_lifecycle.training_runs
			   SET status = 'failed',
			       completed_at = NOW(),
			       error_message = $2
			 WHERE id = $1 AND status = 'running'
		`, runID, errMsg)
		if e != nil {
			return nil, fmt.Errorf("mark_training_run_terminal: update to failed failed: %w", e)
		}
		n, _ = r.RowsAffected()
	}

	transitioned := n > 0
	logger.Info("Marked training run terminal",
		zap.String("training_run_id", runID.String()),
		zap.String("target_status", status),
		zap.Int64("rows_affected", n),
		zap.Bool("transitioned", transitioned),
	)

	// rows_affected == 0 means the row wasn't 'running' (already terminal or
	// missing) — not an error; report it so the monitor can see the no-op.
	return map[string]interface{}{
		"training_run_id": runID.String(),
		"status":          status,
		"transitioned":    transitioned,
	}, nil
}
