// FILE: platform/orchestration/actions/training_data_export.go
//
// TrainingDataExportAction — v3
//
// Exports successful LLM calls from llm_call_log as training data written to
// the training_exports Postgres schema. Replaces earlier file-writing
// behaviour that landed on ephemeral chassis pods.
//
// Storage model:
//
//   training_exports.runs — one row per export (filter + counts)
//   training_exports.rows — one row per training record, FK to runs, CASCADE
//
// Each rows.messages is a ChatML-shaped JSONB array:
//   [{"role":"user","content":<prompt_rendered>},
//    {"role":"assistant","content":<cleaned response>}]
//
// Each rows.metadata is JSONB with: source_log_id, agent_type, step_name,
// orchestration_id, model, created_at, export_version.
//
// Response cleaning reuses stripMarkdownFromResponse from ai_actions.go.
// strict_json filter drops rows whose cleaned assistant text doesn't parse
// as JSON (truncations, malformed escapes).
//
// This agent must be invoked through the training-data-export-orchestrator
// wrapper so it runs in a dedicated spawned pod. See doc 001 §"Every pod-
// running agent needs a parent that spawned it".

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Input specification
// ---------------------------------------------------------------------------
// Following the canonical pattern (doc 001 §3 Checklist for New Specialist
// Agent). ExtractActionInputs handles all three extraction strategies:
// explicit config paths, input_fields lists, direct flat lookup, and
// deprecated-field backwards compatibility.
//
// Field name collision audit (doc 001 §"Field name collisions"):
//   - agent_type       — nested lookup checks current_page, rerender_pages,
//                        site_record, input_data. None of those hold a
//                        conflicting agent_type for export purposes.
//   - step_name        — no known collision with the nested-lookup parents.
//   - model_filter     — novel name, no collision.
//   - include_fenced   — novel name, no collision.
//   - strict_json      — novel name, no collision.
//   - min_response_length — novel name, no collision.
//   - max_rows         — novel name, no collision.
//   - source_notes     — novel name, no collision.

var TrainingDataExportInputSpec = datahelpers.ActionInputSpec{
	Required: []string{
		"agent_type",
		"step_name",
	},
	Optional: []string{
		"model_filter",
		"include_fenced",
		"strict_json",
		"min_response_length",
		"max_rows",
		"source_notes",
	},
	Defaults: map[string]interface{}{
		"include_fenced":      true,
		"strict_json":         true,
		"min_response_length": 10,
		"max_rows":            100000,
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("training_data_export", TrainingDataExportInputSpec)
}

// ---------------------------------------------------------------------------
// Batch size
// ---------------------------------------------------------------------------
// Postgres parameter limit is $65535 per query. Each row contributes 4
// parameters (export_id, row_index, messages, metadata) so 500 rows = 2000
// parameters, well under the ceiling. Big enough to amortise per-row
// insert overhead without fragmenting into too many round-trips.

const exportBatchSize = 500

// ---------------------------------------------------------------------------
// Action
// ---------------------------------------------------------------------------

func TrainingDataExportAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// -- Extract inputs (canonical pattern) --------------------------------

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		TrainingDataExportInputSpec,
		logger,
	)
	if err != nil {
		return nil, err
	}

	agentType := inputs.Get("agent_type")
	stepName := inputs.Get("step_name")
	modelFilter := inputs.Get("model_filter")
	sourceNotes := inputs.Get("source_notes")

	includeFenced := inputs.GetBool("include_fenced", true)
	strictJSON := inputs.GetBool("strict_json", true)
	minResponseLength := inputs.GetInt("min_response_length", 10)
	maxRows := inputs.GetInt("max_rows", 100000)

	if params.DB == nil {
		return nil, fmt.Errorf("training_data_export requires a database connection")
	}

	orchestrationID := params.ExecutionContext.OrchestrationID

	logger.Info("training_data_export v3: starting",
		zap.String("agent_type", agentType),
		zap.String("step_name", stepName),
		zap.String("model_filter", modelFilter),
		zap.Bool("include_fenced", includeFenced),
		zap.Bool("strict_json", strictJSON),
		zap.Int("min_response_length", minResponseLength),
		zap.Int("max_rows", maxRows),
	)

	// -- Begin transaction -------------------------------------------------

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				logger.Warn("training_data_export: rollback failed", zap.Error(rbErr))
			}
		}
	}()

	// -- Insert runs row, get export_id ------------------------------------

	var exportID string
	insertRunSQL := `
        INSERT INTO training_exports.runs
            (agent_type, step_name, model_filter, format, export_version,
             orchestration_id, source_notes)
        VALUES ($1, $2, NULLIF($3, ''), 'chatml', '1',
                NULLIF($4, '')::uuid, NULLIF($5, ''))
        RETURNING id::text
    `
	if err := tx.QueryRowContext(ctx, insertRunSQL,
		agentType, stepName, modelFilter, orchestrationID, sourceNotes,
	).Scan(&exportID); err != nil {
		return nil, fmt.Errorf("insert runs row: %w", err)
	}

	logger.Info("training_data_export: runs row created",
		zap.String("export_id", exportID))

	// -- Query llm_call_log ------------------------------------------------

	queryRows, err := tx.QueryContext(ctx, buildExportQuery(), queryExportArgs{
		AgentType:         agentType,
		StepName:          stepName,
		ModelFilter:       modelFilter,
		IncludeFenced:     includeFenced,
		MinResponseLength: minResponseLength,
		MaxRows:           maxRows,
	}.asSlice()...)
	if err != nil {
		return nil, fmt.Errorf("query llm_call_log: %w", err)
	}
	defer queryRows.Close()

	// -- Stream rows, clean, batch-insert into training_exports.rows -------

	stats := exportStats{}
	var batch []trainingRowInsert
	rowIndex := 0
	var totalSizeBytes int64 = 0

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := bulkInsertRows(ctx, tx, exportID, batch); err != nil {
			return fmt.Errorf("bulk insert %d rows: %w", len(batch), err)
		}
		batch = batch[:0]
		return nil
	}

	for queryRows.Next() {
		var (
			id             string
			logAgentType   string
			logStepName    string
			orchIDFromLog  sql.NullString
			model          string
			createdAt      time.Time
			promptRendered string
			responseText   string
		)
		if err := queryRows.Scan(&id, &logAgentType, &logStepName, &orchIDFromLog,
			&model, &createdAt, &promptRendered, &responseText); err != nil {
			stats.RowsSkippedScanError++
			logger.Warn("training_data_export: row scan failed", zap.Error(err))
			continue
		}

		stats.RowsSeen++

		cleaned := stripMarkdownFromResponse(responseText)

		if strictJSON && !isValidTrainingJSON(cleaned) {
			stats.RowsSkippedInvalidJSON++
			logger.Warn("training_data_export: cleaned response is not valid JSON, skipping",
				zap.String("source_log_id", id),
				zap.Int("assistant_len", len(cleaned)),
			)
			continue
		}

		messages := []chatMessage{
			{Role: "user", Content: promptRendered},
			{Role: "assistant", Content: cleaned},
		}
		metadata := trainingMetadata{
			SourceLogID:     id,
			AgentType:       logAgentType,
			StepName:        logStepName,
			OrchestrationID: orchIDFromLog.String,
			Model:           model,
			CreatedAt:       createdAt.UTC().Format("2006-01-02T15:04:05Z"),
			ExportVersion:   "1",
		}

		messagesJSON, err := json.Marshal(messages)
		if err != nil {
			stats.RowsSkippedMarshalError++
			logger.Warn("training_data_export: messages marshal failed",
				zap.String("source_log_id", id), zap.Error(err))
			continue
		}
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			stats.RowsSkippedMarshalError++
			logger.Warn("training_data_export: metadata marshal failed",
				zap.String("source_log_id", id), zap.Error(err))
			continue
		}

		batch = append(batch, trainingRowInsert{
			RowIndex: rowIndex,
			Messages: messagesJSON,
			Metadata: metadataJSON,
		})
		rowIndex++
		totalSizeBytes += int64(len(messagesJSON) + len(metadataJSON))
		stats.RowsExported++

		if len(batch) >= exportBatchSize {
			if err := flushBatch(); err != nil {
				return nil, err
			}
		}
	}
	if err := queryRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating llm_call_log rows: %w", err)
	}
	if err := flushBatch(); err != nil {
		return nil, err
	}
	queryRows.Close()

	// -- Update runs row with final counts ---------------------------------

	rowsSkippedJSON, _ := json.Marshal(map[string]int{
		"invalid_json":  stats.RowsSkippedInvalidJSON,
		"scan_error":    stats.RowsSkippedScanError,
		"marshal_error": stats.RowsSkippedMarshalError,
	})

	updateRunSQL := `
        UPDATE training_exports.runs
        SET rows_seen = $1,
            rows_exported = $2,
            rows_skipped = $3::jsonb,
            size_bytes = $4,
            completed_at = NOW()
        WHERE id = $5::uuid
    `
	if _, err := tx.ExecContext(ctx, updateRunSQL,
		stats.RowsSeen, stats.RowsExported, string(rowsSkippedJSON),
		totalSizeBytes, exportID,
	); err != nil {
		return nil, fmt.Errorf("update runs row: %w", err)
	}

	// -- Commit ------------------------------------------------------------

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	committed = true

	result := map[string]interface{}{
		"export_id":                  exportID,
		"rows_seen":                  stats.RowsSeen,
		"rows_exported":              stats.RowsExported,
		"rows_skipped_invalid_json":  stats.RowsSkippedInvalidJSON,
		"rows_skipped_scan_error":    stats.RowsSkippedScanError,
		"rows_skipped_marshal_error": stats.RowsSkippedMarshalError,
		"size_bytes":                 totalSizeBytes,
		"agent_type":                 agentType,
		"step_name":                  stepName,
		"model_filter":               modelFilter,
		"format":                     "chatml",
		"export_version":             "1",
	}

	logger.Info("training_data_export v3: complete",
		zap.String("export_id", exportID),
		zap.Int("rows_seen", stats.RowsSeen),
		zap.Int("rows_exported", stats.RowsExported),
		zap.Int("rows_skipped_invalid_json", stats.RowsSkippedInvalidJSON),
		zap.Int64("size_bytes", totalSizeBytes),
	)

	return result, nil
}

// ---------------------------------------------------------------------------
// SQL / batch insert helpers (private to this file)
// ---------------------------------------------------------------------------

type queryExportArgs struct {
	AgentType         string
	StepName          string
	ModelFilter       string
	IncludeFenced     bool
	MinResponseLength int
	MaxRows           int
}

func (q queryExportArgs) asSlice() []interface{} {
	return []interface{}{
		q.AgentType,         // $1
		q.StepName,          // $2
		q.MinResponseLength, // $3
		q.ModelFilter,       // $4 (empty string = no filter)
		q.MaxRows,           // $5
		q.IncludeFenced,     // $6
	}
}

func buildExportQuery() string {
	return `
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
}

type trainingRowInsert struct {
	RowIndex int
	Messages []byte
	Metadata []byte
}

// bulkInsertRows writes a batch of training rows via a single multi-VALUE
// INSERT. 4 parameters per row (export_id, row_index, messages, metadata).
// With batch size 500 that's 2000 parameters, well under Postgres' 65535 limit.
func bulkInsertRows(ctx context.Context, tx *sql.Tx, exportID string, batch []trainingRowInsert) error {
	if len(batch) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO training_exports.rows (export_id, row_index, messages, metadata) VALUES ")

	args := make([]interface{}, 0, 4*len(batch))
	for i, r := range batch {
		if i > 0 {
			sb.WriteString(", ")
		}
		base := i*4 + 1
		fmt.Fprintf(&sb, "($%d::uuid, $%d, $%d::jsonb, $%d::jsonb)",
			base, base+1, base+2, base+3)
		args = append(args, exportID, r.RowIndex, string(r.Messages), string(r.Metadata))
	}

	_, err := tx.ExecContext(ctx, sb.String(), args...)
	return err
}

// ---------------------------------------------------------------------------
// Local types and one helper
// ---------------------------------------------------------------------------

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

// isValidTrainingJSON is unexported and specific to this file. Named with a
// prefix to avoid collision with anything else in the actions package.
func isValidTrainingJSON(s string) bool {
	var v interface{}
	return json.Unmarshal([]byte(s), &v) == nil
}
