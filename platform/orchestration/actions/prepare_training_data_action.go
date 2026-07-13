// FILE: platform/orchestration/actions/prepare_training_data_action.go
//
// CHANGES from v3:
//   - Fix size_bytes=0 reporting bug. bytes.Buffer is drained when the AWS
//     SDK reads from it during Upload, so buf.Len() returns 0 after the
//     upload call. Capture buffer size into sizeBytes BEFORE Upload, use
//     that variable in both the log line and the return map.
//   - Functional behaviour unchanged. Upload itself was always working —
//     only the reported size was wrong.

package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	cfgpkg "github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var PrepareTrainingDataInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"export_id"},
	Optional:   []string{"triggered_by", "orchestration_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("prepare_training_data", PrepareTrainingDataInputSpec)
}

func PrepareTrainingDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "prepare_training_data"))
	logger.Info("PrepareTrainingDataAction: starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// ── Inputs ──────────────────────────────────────────────────────────
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		PrepareTrainingDataInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	exportID, err := uuid.Parse(inputs.Get("export_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid export_id: %w", err)
	}

	hpRaw := datahelpers.ExtractNestedField(params.CollectedData, "input_data.hyperparameters")
	if hpRaw == nil {
		return nil, fmt.Errorf("hyperparameters not found at input_data.hyperparameters")
	}
	hpJSON, err := json.Marshal(hpRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hyperparameters: %w", err)
	}

	triggeredBy := datahelpers.NullableString(inputs.Get("triggered_by"))
	orchestrationID := datahelpers.NullableString(inputs.Get("orchestration_id"))

	s3KeyTemplate, _ := params.StepConfig.Config["s3_key_template"].(string)
	if s3KeyTemplate == "" {
		return nil, fmt.Errorf("s3_key_template required in step config")
	}
	s3Key := strings.ReplaceAll(s3KeyTemplate, "{export_id}", exportID.String())

	// ── Construct S3 client (same idiom as storage_actions.go::storeToB2) ─
	bucket := "finetuning"
	if b, ok := params.StepConfig.Config["bucket"].(string); ok && b != "" {
		bucket = b
	}
	storageConfig := cfgpkg.ObjectStorageConfig{
		Provider:        "s3",
		Bucket:          bucket,
		AccessKeyEnvVar: "B2_APPLICATION_KEY_ID",
		SecretKeyEnvVar: "B2_APPLICATION_KEY",
	}
	s3Client, err := storage.NewS3Client(ctx, storageConfig, *logger)
	if err != nil {
		return nil, fmt.Errorf("failed to construct S3 client: %w", err)
	}

	// ── INSERT training_runs row ────────────────────────────────────────
	var trainingRunID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO model_lifecycle.training_runs (
			export_id, status, hyperparameters, triggered_by, orchestration_id
		) VALUES ($1, 'pending', $2::jsonb, $3, $4)
		RETURNING id
	`, exportID, string(hpJSON), triggeredBy, orchestrationID).Scan(&trainingRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert training_runs row: %w", err)
	}
	logger.Info("Inserted training_runs row",
		zap.String("training_run_id", trainingRunID.String()),
	)

	// ── Stream training_exports.rows as NDJSON into a buffer ────────────
	cursor, err := params.DB.QueryContext(ctx, `
		SELECT messages, metadata FROM training_exports.rows
		WHERE export_id = $1 ORDER BY row_index
	`, exportID)
	if err != nil {
		markTrainingRunFailed(ctx, params, trainingRunID,
			fmt.Sprintf("query failed: %v", err))
		return nil, fmt.Errorf("failed to query training_exports.rows: %w", err)
	}
	defer cursor.Close()

	var buf bytes.Buffer
	rowCount := 0
	for cursor.Next() {
		var messagesRaw, metadataRaw []byte
		if err := cursor.Scan(&messagesRaw, &metadataRaw); err != nil {
			markTrainingRunFailed(ctx, params, trainingRunID,
				fmt.Sprintf("scan failed at row %d: %v", rowCount, err))
			return nil, fmt.Errorf("scan row %d: %w", rowCount, err)
		}
		line := struct {
			Messages json.RawMessage `json:"messages"`
			Metadata json.RawMessage `json:"metadata"`
		}{json.RawMessage(messagesRaw), json.RawMessage(metadataRaw)}
		encoded, err := json.Marshal(line)
		if err != nil {
			markTrainingRunFailed(ctx, params, trainingRunID,
				fmt.Sprintf("encode row %d: %v", rowCount, err))
			return nil, fmt.Errorf("encode row %d: %w", rowCount, err)
		}
		buf.Write(encoded)
		buf.WriteByte('\n')
		rowCount++
	}
	if err := cursor.Err(); err != nil {
		markTrainingRunFailed(ctx, params, trainingRunID,
			fmt.Sprintf("cursor: %v", err))
		return nil, fmt.Errorf("cursor: %w", err)
	}
	if rowCount == 0 {
		markTrainingRunFailed(ctx, params, trainingRunID,
			fmt.Sprintf("export %s has no rows", exportID))
		return nil, fmt.Errorf("export %s has no training rows", exportID)
	}

	// Capture buffer size BEFORE Upload — bytes.Buffer.Read is consuming,
	// so buf.Len() drops to 0 after the AWS SDK reads it during PutObject.
	sizeBytes := buf.Len()

	datasetURI, err := s3Client.Upload(ctx, s3Key, "application/x-ndjson", &buf)
	if err != nil {
		markTrainingRunFailed(ctx, params, trainingRunID,
			fmt.Sprintf("upload failed: %v", err))
		return nil, fmt.Errorf("failed to upload dataset: %w", err)
	}
	logger.Info("Uploaded training dataset",
		zap.String("dataset_uri", datasetURI),
		zap.Int("rows", rowCount),
		zap.Int("bytes", sizeBytes),
	)

	return map[string]interface{}{
		"training_run_id": trainingRunID.String(),
		"dataset_uri":     datasetURI,
		"row_count":       rowCount,
		"size_bytes":      sizeBytes,
	}, nil
}

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
