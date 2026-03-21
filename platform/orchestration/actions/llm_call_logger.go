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
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := db.ExecContext(ctx, `
			INSERT INTO llm_call_log (
				agent_type, agent_id, step_name, orchestration_id, correlation_id,
				model, model_resolved, provider,
				prompt_template, prompt_rendered, response_text,
				input_tokens, output_tokens, latency_ms,
				success, error_message
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11,
				$12, $13, $14,
				$15, $16
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
		)

		if err != nil {
			logger.Warn("Failed to log LLM call (non-fatal)",
				zap.String("agent_type", params.AgentType),
				zap.Error(err))
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
	PromptTemplate  string
	PromptRendered  string
	ResponseText    string
	InputTokens     int
	OutputTokens    int
	LatencyMs       int
	Success         bool
	ErrorMessage    string
}

// nullIfZero returns nil for zero token counts (when provider doesn't report them)
func nullIfZero(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}
