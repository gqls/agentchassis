// FILE: platform/orchestration/actions/prepare_training_data_action.go
//
// PrepareTrainingDataAction is the worker for the training-data-preparer
// agent (declared in migration 021_training_data_preparer_agent.sql).
//
// One Go-action invocation does three things:
//   1. INSERT a row into model_lifecycle.training_runs with status='pending'
//      capturing the hyperparameters and provenance.
//   2. Stream training_exports.rows for the given export_id as NDJSON
//      (replays the snapshot — option B from design discussion).
//   3. Upload the buffer to s3://{IMAGE_BUCKET}/{key} via params.StorageClient.
//      The bucket name comes from the IMAGE_BUCKET env var which is set to
//      'finetuning' on this agent's deployment (see agent_definitions.env_vars
//      in migration 022_training_agents_corrections.sql).
//
// Returns:
//
//	{
//	  "training_run_id": "<uuid>",
//	  "dataset_uri":     "s3://finetuning/datasets/<export_id>/training.jsonl",
//	  "row_count":       <int>,
//	  "size_bytes":      <int>
//	}
//
// This map ends up under CollectedData["preparation_result"] (per
// output_field in the workflow), where downstream callers reference it as
// preparation_result.training_run_id and preparation_result.dataset_uri.

package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// PrepareTrainingDataInputSpec declares typed-scalar inputs.
// hyperparameters is a free-form JSONB partial — read directly from
// CollectedData via ExtractNestedField, not declared here. Same pattern as
// WriteSitePlanAction's design_direction / content_strategy.
var PrepareTrainingDataInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"export_id"},
	Optional:   []string{"triggered_by", "orchestration_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("prepare_training_data", PrepareTrainingDataInputSpec)
}

// PrepareTrainingDataAction is the action handler.
// Registered as "prepare_training_data", IsLocal=true.
func PrepareTrainingDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "prepare_training_data"))
	logger.Info("PrepareTrainingDataAction: starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Chassis convention — initialize call returns immediately.
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}
	if params.StorageClient == nil {
		return nil, fmt.Errorf("storage client required")
	}

	// ── 1. Extract typed inputs ─────────────────────────────────────────
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		PrepareTrainingDataInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	exportIDStr := inputs.Get("export_id")
	exportID, err := uuid.Parse(exportIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid export_id %q: %w", exportIDStr, err)
	}

	// ── 2. Pull hyperparameters JSONB via canonical path resolver ───────
	// Use ExtractNestedField (datahelpers' canonical resolver) rather than
	// indexing CollectedData directly — it handles dot paths + .response
	// auto-unwrap consistently with the rest of the codebase.
	hpRaw := datahelpers.ExtractNestedField(params.CollectedData, "input_data.hyperparameters")
	if hpRaw == nil {
		return nil, fmt.Errorf("hyperparameters not found at input_data.hyperparameters")
	}
	hpJSON, err := json.Marshal(hpRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hyperparameters: %w", err)
	}

	// Optional UUIDs — use the canonical NullableString helper.
	triggeredBy := datahelpers.NullableString(inputs.Get("triggered_by"))
	orchestrationID := datahelpers.NullableString(inputs.Get("orchestration_id"))

	// ── 3. Read step-config knobs (key template only — bucket is env-driven) ─
	// IMAGE_BUCKET env var on this agent's deployment routes the storage
	// client to 'finetuning' (see agent_definitions.env_vars).
	s3KeyTemplate, _ := params.StepConfig.Config["s3_key_template"].(string)
	if s3KeyTemplate == "" {
		return nil, fmt.Errorf("s3_key_template required in step config")
	}
	s3Key := strings.ReplaceAll(s3KeyTemplate, "{export_id}", exportID.String())

	// ── 4. INSERT model_lifecycle.training_runs row ─────────────────────
	var trainingRunID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO model_lifecycle.training_runs (
			export_id, status, hyperparameters, triggered_by, orchestration_id
		) VALUES (
			$1, 'pending', $2::jsonb, $3, $4
		)
		RETURNING id
	`, exportID, string(hpJSON), triggeredBy, orchestrationID).Scan(&trainingRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert training_runs row: %w", err)
	}
	logger.Info("Inserted training_runs row",
		zap.String("training_run_id", trainingRunID.String()),
		zap.String("export_id", exportID.String()),
	)

	// ── 5. Stream training_exports.rows as NDJSON into a buffer ─────────
	cursor, err := params.DB.QueryContext(ctx, `
		SELECT messages, metadata
		FROM training_exports.rows
		WHERE export_id = $1
		ORDER BY row_index
	`, exportID)
	if err != nil {
		markTrainingRunFailed(ctx, params, trainingRunID, fmt.Sprintf("query training_exports.rows failed: %v", err))
		return nil, fmt.Errorf("failed to query training_exports.rows: %w", err)
	}
	defer cursor.Close()

	var buf bytes.Buffer
	rowCount := 0
	for cursor.Next() {
		var messagesRaw, metadataRaw []byte
		if err := cursor.Scan(&messagesRaw, &metadataRaw); err != nil {
			markTrainingRunFailed(ctx, params, trainingRunID, fmt.Sprintf("scan failed at row %d: %v", rowCount, err))
			return nil, fmt.Errorf("failed to scan row %d: %w", rowCount, err)
		}
		// json.RawMessage prevents double-encoding of already-JSON jsonb bytes.
		line := struct {
			Messages json.RawMessage `json:"messages"`
			Metadata json.RawMessage `json:"metadata"`
		}{
			Messages: json.RawMessage(messagesRaw),
			Metadata: json.RawMessage(metadataRaw),
		}
		encoded, err := json.Marshal(line)
		if err != nil {
			markTrainingRunFailed(ctx, params, trainingRunID, fmt.Sprintf("encode failed at row %d: %v", rowCount, err))
			return nil, fmt.Errorf("failed to encode row %d: %w", rowCount, err)
		}
		buf.Write(encoded)
		buf.WriteByte('\n')
		rowCount++
	}
	if err := cursor.Err(); err != nil {
		markTrainingRunFailed(ctx, params, trainingRunID, fmt.Sprintf("cursor iteration failed: %v", err))
		return nil, fmt.Errorf("cursor iteration failed: %w", err)
	}
	if rowCount == 0 {
		markTrainingRunFailed(ctx, params, trainingRunID, fmt.Sprintf("export_id %s has no rows", exportID))
		return nil, fmt.Errorf("export_id %s has no training rows", exportID)
	}
	logger.Info("Streamed training rows",
		zap.Int("rows", rowCount),
		zap.Int("bytes", buf.Len()),
	)

	// ── 6. Upload via the storage client (bucket = IMAGE_BUCKET env var) ─
	datasetURI, err := params.StorageClient.Upload(
		ctx, s3Key, "application/x-ndjson", &buf,
	)
	if err != nil {
		markTrainingRunFailed(ctx, params, trainingRunID, fmt.Sprintf("dataset upload failed: %v", err))
		return nil, fmt.Errorf("failed to upload dataset to s3: %w", err)
	}
	logger.Info("Uploaded training dataset",
		zap.String("dataset_uri", datasetURI),
		zap.String("training_run_id", trainingRunID.String()),
	)

	// ── 7. Return — workflow stores under CollectedData["preparation_result"] ─
	return map[string]interface{}{
		"training_run_id": trainingRunID.String(),
		"dataset_uri":     datasetURI,
		"row_count":       rowCount,
		"size_bytes":      buf.Len(),
	}, nil
}

// markTrainingRunFailed is a small private helper — keeping it in this file
// per the guideline that private helpers reduce repetition without splitting
// the action into wrapper+core. When 022 (gpu-provisioner), 023 (launcher),
// 024 (status-checker) etc. need the same operation, lift this to
// training_runs_helpers.go alongside the other workers.
func markTrainingRunFailed(ctx context.Context, params ActionParams, runID uuid.UUID, msg string) {
	if params.DB == nil {
		return
	}
	_, _ = params.DB.ExecContext(ctx, `
		UPDATE model_lifecycle.training_runs
		SET status = 'failed', error_message = $2, completed_at = NOW()
		WHERE id = $1
	`, runID, msg)
}
