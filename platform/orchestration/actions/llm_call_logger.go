// FILE: platform/orchestration/actions/llm_call_logger.go
//
// Logs every LLM call to the llm_call_log table.
// Called from ExecuteLLMPromptAction after each LLM invocation.
// Fire-and-forget — errors are logged but never block the workflow.

package actions

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

// LogLLMCall logs an LLM invocation to the llm_call_log table.
// Runs in a goroutine with a 5-second timeout. Never blocks the caller.
func LogLLMCall(db *sql.DB, logger *zap.Logger, params LLMCallLogParams) {
	if db == nil {
		if logger != nil {
			logger.Warn("LogLLMCall: DB is nil, skipping LLM call log",
				zap.String("agent_type", params.AgentType),
				zap.String("model", params.Model),
				zap.Bool("success", params.Success))
		}
		return
	}

	logger.Info("LogLLMCall: writing to llm_call_log",
		zap.String("agent_type", params.AgentType),
		zap.String("model", params.Model),
		zap.Bool("success", params.Success))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Default prompt_variant to "default" if empty
		promptVariant := params.PromptVariant
		if promptVariant == "" {
			promptVariant = "default"
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO llm_call_log (
				agent_type, agent_id, step_name, orchestration_id, correlation_id,
				model, model_resolved, provider,
				prompt_template, prompt_rendered, response_text,
				input_tokens, output_tokens, latency_ms,
				success, error_message,
				work_item_id, vertical, prompt_variant, rag_context_used,
				temperature, max_tokens
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11,
				$12, $13, $14,
				$15, $16,
				$17, $18, $19, $20,
				$21, $22
			)`,
			params.AgentType, nullIfEmpty(params.AgentID),
			nullIfEmpty(params.StepName), nullIfEmpty(params.OrchestrationID),
			nullIfEmpty(params.CorrelationID),
			params.Model, nullIfEmpty(params.ModelResolved), params.Provider,
			nullIfEmpty(params.PromptTemplate), nullIfEmpty(params.PromptRendered),
			nullIfEmpty(params.ResponseText),
			nullIfZero(params.InputTokens), nullIfZero(params.OutputTokens),
			nullIfZero(params.LatencyMs),
			params.Success, nullIfEmpty(params.ErrorMessage),
			nullIfEmpty(params.WorkItemID), nullIfEmpty(params.Vertical),
			promptVariant, params.RAGContextUsed,
			params.Temperature, nullIfZero(params.MaxTokens),
		)

		if err != nil {
			logger.Warn("Failed to log LLM call (non-fatal)",
				zap.String("agent_type", params.AgentType),
				zap.Error(err))
		} else {
			logger.Info("LogLLMCall: row written successfully",
				zap.String("agent_type", params.AgentType),
				zap.String("step", params.StepName))
		}
	}()
}

// LLMCallLogParams holds all fields for an LLM call log entry
type LLMCallLogParams struct {
	AgentType       string
	AgentID         string
	StepName        string
	OrchestrationID string
	CorrelationID   string
	Model           string
	ModelResolved   string
	Provider        string
	Temperature     interface{} // actual temperature sent by the provider; nil → NULL
	MaxTokens       int         // actual max_tokens sent by the provider; 0 → NULL
	PromptTemplate  string
	PromptRendered  string
	ResponseText    string
	InputTokens     int
	OutputTokens    int
	LatencyMs       int
	Success         bool
	ErrorMessage    string
	// Flywheel fields — link calls to outcomes for training data
	WorkItemID     string // UUID string from input_data.work_item_id
	Vertical       string // Industry vertical from site identity
	PromptVariant  string // For A/B testing prompt versions
	RAGContextUsed bool   // Whether RAG retrieval was used in this call
}

// nullIfZero returns nil for zero token counts (when provider doesn't report them)
func nullIfZero(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}
