// FILE: platform/orchestration/actions/training_data_export.go
//
// TrainingDataExportAction exports successful LLM calls from llm_call_log as
// training data in NDJSON format, suitable for fine-tuning.
//
// Format: ChatML messages + metadata sidecar. One JSON object per line.
//
//   {
//     "messages": [
//       {"role": "user", "content": "<prompt_rendered>"},
//       {"role": "assistant", "content": "<cleaned response>"}
//     ],
//     "metadata": {
//       "source_log_id": "...",
//       "agent_type": "page-content-writer",
//       "step_name": "process_sections_loop_iter_0_generate_content",
//       "orchestration_id": "...",
//       "model": "claude-sonnet-4-6",
//       "created_at": "2026-04-20T10:15:06Z",
//       "export_version": "1"
//     }
//   }
//
// Response cleaning reuses stripMarkdownFromResponse (ai_actions.go).
//
// Parameters are read from orchestration CollectedData["input_data"], not from
// static step config. This matches the existing convention (datahelpers.Extract...)
// used elsewhere in the chassis and lets one agent definition handle exports
// for any agent/step target by varying the input_data payload.
//
// Expected input_data fields:
//   agent_type           (required) — filter on llm_call_log.agent_type
//   step_name            (required) — filter on llm_call_log.step_name
//   output_path          (required) — path inside the agent-chassis pod
//   model_filter         (optional) — filter on llm_call_log.model
//   include_fenced       (optional, default true)
//   strict_json          (optional, default true)
//   min_response_length  (optional, default 10)
//   max_rows             (optional, default 100000)

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func TrainingDataExportAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// -- Read parameters from input_data -----------------------------------

	agentType := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.agent_type")
	if agentType == "" {
		return nil, fmt.Errorf("training_data_export: required input_data.agent_type is missing or empty")
	}
	stepName := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.step_name")
	if stepName == "" {
		return nil, fmt.Errorf("training_data_export: required input_data.step_name is missing or empty")
	}
	outputPath := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.output_path")
	if outputPath == "" {
		return nil, fmt.Errorf("training_data_export: required input_data.output_path is missing or empty")
	}

	modelFilter := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.model_filter")
	// Optional fields with defaults. ExtractNestedField returns interface{}; type-assert.
	includeFenced := extractBoolWithDefault(params.CollectedData, "input_data.include_fenced", true)
	strictJSON := extractBoolWithDefault(params.CollectedData, "input_data.strict_json", true)
	minResponseLength := extractIntWithDefault(params.CollectedData, "input_data.min_response_length", 10)
	maxRows := extractIntWithDefault(params.CollectedData, "input_data.max_rows", 100000)

	if params.DB == nil {
		return nil, fmt.Errorf("training_data_export requires a database connection")
	}

	logger.Info("training_data_export: starting",
		zap.String("agent_type", agentType),
		zap.String("step_name", stepName),
		zap.String("output_path", outputPath),
		zap.String("model_filter", modelFilter),
		zap.Bool("include_fenced", includeFenced),
		zap.Bool("strict_json", strictJSON),
		zap.Int("min_response_length", minResponseLength),
		zap.Int("max_rows", maxRows),
	)

	// -- Prepare output file ----------------------------------------------

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating output dir: %w", err)
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("creating output file %s: %w", outputPath, err)
	}
	defer outFile.Close()

	// -- Query + stream ----------------------------------------------------

	rows, queryErr := queryTrainingRows(ctx, params.DB, trainingQueryParams{
		AgentType:         agentType,
		StepName:          stepName,
		ModelFilter:       modelFilter,
		IncludeFenced:     includeFenced,
		MinResponseLength: minResponseLength,
		MaxRows:           maxRows,
	})
	if queryErr != nil {
		return nil, fmt.Errorf("training_data_export query failed: %w", queryErr)
	}
	defer rows.Close()

	stats := exportStats{}
	encoder := json.NewEncoder(outFile)

	for rows.Next() {
		var (
			id              string
			logAgentType    string
			logStepName     string
			orchestrationID sql.NullString
			model           string
			createdAt       time.Time
			promptRendered  string
			responseText    string
		)
		if err := rows.Scan(&id, &logAgentType, &logStepName, &orchestrationID,
			&model, &createdAt, &promptRendered, &responseText); err != nil {
			stats.RowsSkippedScanError++
			logger.Warn("training_data_export: row scan failed", zap.Error(err))
			continue
		}

		stats.RowsSeen++

		cleaned := stripMarkdownFromResponse(responseText)

		if strictJSON && !isValidJSON(cleaned) {
			stats.RowsSkippedInvalidJSON++
			logger.Warn("training_data_export: cleaned response is not valid JSON, skipping",
				zap.String("source_log_id", id),
				zap.Int("assistant_len", len(cleaned)),
			)
			continue
		}

		record := trainingRecord{
			Messages: []chatMessage{
				{Role: "user", Content: promptRendered},
				{Role: "assistant", Content: cleaned},
			},
			Metadata: trainingMetadata{
				SourceLogID:     id,
				AgentType:       logAgentType,
				StepName:        logStepName,
				OrchestrationID: orchestrationID.String,
				Model:           model,
				CreatedAt:       createdAt.UTC().Format("2006-01-02T15:04:05Z"),
				ExportVersion:   "1",
			},
		}

		if err := encoder.Encode(record); err != nil {
			stats.RowsSkippedMarshalError++
			logger.Warn("training_data_export: marshal/write failed",
				zap.String("source_log_id", id),
				zap.Error(err),
			)
			continue
		}
		stats.RowsExported++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating training rows: %w", err)
	}

	if err := outFile.Sync(); err != nil {
		logger.Warn("training_data_export: sync failed", zap.Error(err))
	}

	fileInfo, statErr := os.Stat(outputPath)
	var fileSize int64 = 0
	if statErr == nil {
		fileSize = fileInfo.Size()
	}

	result := map[string]interface{}{
		"rows_seen":                  stats.RowsSeen,
		"rows_exported":              stats.RowsExported,
		"rows_skipped_invalid_json":  stats.RowsSkippedInvalidJSON,
		"rows_skipped_scan_error":    stats.RowsSkippedScanError,
		"rows_skipped_marshal_error": stats.RowsSkippedMarshalError,
		"output_path":                outputPath,
		"file_size_bytes":            fileSize,
		"agent_type":                 agentType,
		"step_name":                  stepName,
		"model_filter":               modelFilter,
		"format":                     "chatml",
		"export_version":             "1",
	}

	logger.Info("training_data_export: complete",
		zap.Int("rows_seen", stats.RowsSeen),
		zap.Int("rows_exported", stats.RowsExported),
		zap.Int("rows_skipped_invalid_json", stats.RowsSkippedInvalidJSON),
		zap.Int64("file_size_bytes", fileSize),
		zap.String("output_path", outputPath),
	)

	return result, nil
}

// -- Helpers --------------------------------------------------------------

// extractBoolWithDefault reads a bool from CollectedData using dot-path,
// returning the default if missing or wrong type.
func extractBoolWithDefault(data map[string]interface{}, path string, def bool) bool {
	v := datahelpers.ExtractNestedField(data, path)
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

// extractIntWithDefault reads an int from CollectedData using dot-path,
// tolerating int / int64 / float64 (JSON numbers typically arrive as float64).
func extractIntWithDefault(data map[string]interface{}, path string, def int) int {
	v := datahelpers.ExtractNestedField(data, path)
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return def
}

// -- Query and types ------------------------------------------------------

type trainingQueryParams struct {
	AgentType         string
	StepName          string
	ModelFilter       string
	IncludeFenced     bool
	MinResponseLength int
	MaxRows           int
}

type exportStats struct {
	RowsSeen                int
	RowsExported            int
	RowsSkippedInvalidJSON  int
	RowsSkippedScanError    int
	RowsSkippedMarshalError int
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type trainingMetadata struct {
	SourceLogID     string `json:"source_log_id"`
	AgentType       string `json:"agent_type"`
	StepName        string `json:"step_name"`
	OrchestrationID string `json:"orchestration_id,omitempty"`
	Model           string `json:"model"`
	CreatedAt       string `json:"created_at"`
	ExportVersion   string `json:"export_version"`
}

type trainingRecord struct {
	Messages []chatMessage    `json:"messages"`
	Metadata trainingMetadata `json:"metadata"`
}

func queryTrainingRows(ctx context.Context, db *sql.DB, p trainingQueryParams) (*sql.Rows, error) {
	query := `
        SELECT id::text,
               agent_type,
               step_name,
               orchestration_id,
               model,
               created_at,
               prompt_rendered,
               response_text
        FROM llm_call_log
        WHERE agent_type = $1
          AND step_name = $2
          AND success = true
          AND prompt_rendered IS NOT NULL
          AND length(prompt_rendered) > 100
          AND length(response_text) >= $3
          AND ($4 = '' OR model = $4)
          AND (
                response_text LIKE '{%'
                OR response_text LIKE '[%'
                OR ($6 AND response_text LIKE E'` + "```" + `%')
              )
        ORDER BY created_at ASC
        LIMIT $5
    `
	return db.QueryContext(ctx, query,
		p.AgentType, p.StepName, p.MinResponseLength,
		p.ModelFilter, p.MaxRows, p.IncludeFenced,
	)
}

func isValidJSON(s string) bool {
	var v interface{}
	return json.Unmarshal([]byte(s), &v) == nil
}
