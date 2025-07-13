// FILE: platform/orchestration/state.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// OrchestrationStatus represents the current state of a workflow
type OrchestrationStatus string

const (
	StatusRunning           OrchestrationStatus = "RUNNING"
	StatusAwaitingResponses OrchestrationStatus = "AWAITING_RESPONSES"
	StatusPausedForHuman    OrchestrationStatus = "PAUSED_FOR_HUMAN_INPUT"
	StatusCompleted         OrchestrationStatus = "COMPLETED"
	StatusFailed            OrchestrationStatus = "FAILED"
)

// OrchestrationState is the database model for a Saga instance
type OrchestrationState struct {
	CorrelationID      string                 `db:"correlation_id"`
	Status             OrchestrationStatus    `db:"status"`
	CurrentStep        string                 `db:"current_step"`
	AwaitedSteps       []string               `db:"awaited_steps"`
	CollectedData      map[string]interface{} `db:"collected_data"`
	InitialRequestData json.RawMessage        `db:"initial_request_data"`
	FinalResult        json.RawMessage        `db:"final_result"`
	Error              string                 `db:"error"`
	CreatedAt          time.Time              `db:"created_at"`
	UpdatedAt          time.Time              `db:"updated_at"`
}

// StateRepository provides an interface for persisting and retrieving workflow state
type StateRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewStateRepository creates a new state repository
func NewStateRepository(db *sql.DB, logger *zap.Logger) *StateRepository {
	return &StateRepository{db: db, logger: logger}
}

// CreateInitialState creates a new record for a workflow
func (r *StateRepository) CreateInitialState(ctx context.Context, correlationID, startStep string, initialData []byte) error {
	awaitedStepsJSON, _ := json.Marshal([]string{})
	collectedDataJSON, _ := json.Marshal(map[string]interface{}{})

	query := `
        INSERT INTO orchestrator_state 
        (correlation_id, status, current_step, awaited_steps, collected_data, initial_request_data, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		correlationID, StatusRunning, startStep, awaitedStepsJSON, collectedDataJSON, initialData, now, now)

	if err != nil {
		r.logger.Error("Failed to create initial orchestration state", zap.Error(err))
		return fmt.Errorf("failed to create initial state: %w", err)
	}

	r.logger.Info("Initial orchestration state created", zap.String("correlation_id", correlationID))
	return nil
}

// GetState retrieves the current state of a workflow
func (r *StateRepository) GetState(ctx context.Context, correlationID string) (*OrchestrationState, error) {
	query := `
        SELECT correlation_id, status, current_step, awaited_steps, collected_data, 
               initial_request_data, COALESCE(final_result, '{}'), COALESCE(error, ''), created_at, updated_at
        FROM orchestrator_state
        WHERE correlation_id = $1
    `

	var state OrchestrationState
	var awaitedStepsJSON, collectedDataJSON []byte

	err := r.db.QueryRowContext(ctx, query, correlationID).Scan(
		&state.CorrelationID,
		&state.Status,
		&state.CurrentStep,
		&awaitedStepsJSON,
		&collectedDataJSON,
		&state.InitialRequestData,
		&state.FinalResult,
		&state.Error,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found for correlation_id: %s", correlationID)
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(awaitedStepsJSON, &state.AwaitedSteps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal awaited_steps: %w", err)
	}
	if err := json.Unmarshal(collectedDataJSON, &state.CollectedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collected_data: %w", err)
	}

	return &state, nil
}

// UpdateState persists changes to a workflow's state
func (r *StateRepository) UpdateState(ctx context.Context, state *OrchestrationState) error {
	awaitedStepsJSON, _ := json.Marshal(state.AwaitedSteps)
	collectedDataJSON, _ := json.Marshal(state.CollectedData)

	query := `
        UPDATE orchestrator_state 
        SET status = $2, current_step = $3, awaited_steps = $4, collected_data = $5, 
            final_result = $6, error = $7, updated_at = $8
        WHERE correlation_id = $1
    `

	_, err := r.db.ExecContext(ctx, query,
		state.CorrelationID,
		state.Status,
		state.CurrentStep,
		awaitedStepsJSON,
		collectedDataJSON,
		state.FinalResult,
		state.Error,
		time.Now().UTC(),
	)

	if err != nil {
		r.logger.Error("Failed to update orchestration state", zap.Error(err))
		return fmt.Errorf("failed to update state: %w", err)
	}

	r.logger.Debug("Orchestration state updated",
		zap.String("correlation_id", state.CorrelationID),
		zap.String("status", string(state.Status)))
	return nil
}

// GetOrchestratorStateTableSchema returns the SQL for creating the state table
func GetOrchestratorStateTableSchema() string {
	return `
CREATE TABLE IF NOT EXISTS orchestrator_state (
    correlation_id UUID PRIMARY KEY,
    status VARCHAR(50) NOT NULL,
    current_step VARCHAR(255) NOT NULL,
    awaited_steps JSONB DEFAULT '[]',
    collected_data JSONB DEFAULT '{}',
    initial_request_data JSONB,
    final_result JSONB,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orchestrator_state_status ON orchestrator_state(status);
CREATE INDEX idx_orchestrator_state_updated_at ON orchestrator_state(updated_at);
`
}
