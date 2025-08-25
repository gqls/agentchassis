// FILE: platform/orchestration/state.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/pkg/models"

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

// ExecutionRecord tracks each step execution
type ExecutionRecord struct {
	Step      string     `json:"step"`
	Action    string     `json:"action"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Result    string     `json:"result"`
	Error     string     `json:"error,omitempty"`
}

// ExecutionMetadata provides workflow analytics
type ExecutionMetadata struct {
	TotalSteps     int                  `json:"total_steps"`
	CompletedSteps int                  `json:"completed_steps"`
	SkippedSteps   int                  `json:"skipped_steps"`
	FailedSteps    int                  `json:"failed_steps"`
	RetryCount     map[string]int       `json:"retry_count"`
	Checkpoints    map[string]time.Time `json:"checkpoints"`
	StartTime      time.Time            `json:"start_time"`
	EndTime        *time.Time           `json:"end_time,omitempty"`
}

type PendingRequest struct {
	RequestID string
	AgentID   string
	AgentType string
	Step      string
	StartTime time.Time
}

// OrchestrationState is the database model for a Saga instance
type OrchestrationState struct {
	CorrelationID string              `db:"correlation_id"`
	ClientID      string              `db:"client_id"`
	Status        OrchestrationStatus `db:"status"`
	// Execution state
	CurrentStep   string            `db:"current_step"`
	ExecutionPath []ExecutionRecord `db:"execution_path"`
	// Async handling
	AwaitedSteps []string `db:"awaited_steps"`
	// Data management
	CollectedData      map[string]interface{} `db:"collected_data"`
	InitialRequestData json.RawMessage        `db:"initial_request_data"`
	FinalResult        json.RawMessage        `db:"final_result"`
	// Workflow definition
	WorkflowPlan models.WorkflowPlan `db:"workflow_plan"`
	// Debugging/Monitoring
	ExecutionMetadata ExecutionMetadata `db:"execution_metadata"`
	Error             string            `db:"error"`
	CreatedAt         time.Time         `db:"created_at"`
	UpdatedAt         time.Time         `db:"updated_at"`
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

// CreateInitialState creates a new workflow with the plan
func (r *StateRepository) CreateInitialState(ctx context.Context, correlationID, clientID string, plan models.WorkflowPlan, initialData []byte) error {
	awaitedStepsJSON, _ := json.Marshal([]string{})
	collectedDataJSON, _ := json.Marshal(map[string]interface{}{})
	workflowPlanJSON, _ := json.Marshal(plan)

	// Ensure we have valid initial data
	if initialData == nil {
		initialData = []byte("null")
	}

	// Initialize execution metadata
	metadata := ExecutionMetadata{
		TotalSteps:     len(plan.Steps),
		CompletedSteps: 0,
		SkippedSteps:   0,
		FailedSteps:    0,
		RetryCount:     make(map[string]int),
		Checkpoints:    make(map[string]time.Time),
		StartTime:      time.Now().UTC(),
	}
	metadataJSON, _ := json.Marshal(metadata)

	executionPathJSON, _ := json.Marshal([]ExecutionRecord{})

	query := `
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, current_step, awaited_steps, 
         collected_data, initial_request_data, workflow_plan, 
         execution_metadata, execution_path, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		correlationID, clientID, StatusRunning, plan.StartStep,
		awaitedStepsJSON, collectedDataJSON, initialData, workflowPlanJSON,
		metadataJSON, executionPathJSON, now, now)

	if err != nil {
		r.logger.Error("Failed to create initial orchestration state", zap.Error(err))
		return fmt.Errorf("failed to create initial state: %w", err)
	}

	r.logger.Info("Initial orchestration state created",
		zap.String("correlation_id", correlationID),
		zap.String("start_step", plan.StartStep),
		zap.Int("total_steps", len(plan.Steps)))

	return nil
}

// GetState retrieves the current state of the workflow with the full plan
func (r *StateRepository) GetState(ctx context.Context, correlationID string) (*OrchestrationState, error) {
	query := `
        SELECT correlation_id, client_id, status, current_step, awaited_steps, 
               collected_data, initial_request_data, final_result, error, 
               workflow_plan, execution_metadata, execution_path, created_at, updated_at
        FROM orchestrator_state
        WHERE correlation_id = $1
    `

	var state OrchestrationState
	var awaitedStepsJSON, collectedDataJSON, workflowPlanJSON []byte
	var executionMetadataJSON, executionPathJSON []byte
	var initialRequestDataNull, finalResultNull, errorNull sql.NullString

	err := r.db.QueryRowContext(ctx, query, correlationID).Scan(
		&state.CorrelationID,
		&state.ClientID,
		&state.Status,
		&state.CurrentStep,
		&awaitedStepsJSON,
		&collectedDataJSON,
		&initialRequestDataNull,
		&finalResultNull,
		&errorNull,
		&workflowPlanJSON,
		&executionMetadataJSON,
		&executionPathJSON,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found for correlation_id: %s", correlationID)
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Handle nullable fields
	if initialRequestDataNull.Valid {
		state.InitialRequestData = json.RawMessage(initialRequestDataNull.String)
	}
	if finalResultNull.Valid {
		state.FinalResult = json.RawMessage(finalResultNull.String)
	}
	if errorNull.Valid {
		state.Error = errorNull.String
	}

	// Unmarshal JSON fields
	json.Unmarshal(awaitedStepsJSON, &state.AwaitedSteps)
	json.Unmarshal(collectedDataJSON, &state.CollectedData)
	json.Unmarshal(workflowPlanJSON, &state.WorkflowPlan)
	json.Unmarshal(executionMetadataJSON, &state.ExecutionMetadata)
	json.Unmarshal(executionPathJSON, &state.ExecutionPath)

	return &state, nil
}

// UpdateState persists all changes to a workflow's state including execution tracking
func (r *StateRepository) UpdateState(ctx context.Context, state *OrchestrationState) error {
	// Ensure all JSON fields are properly initialized
	if state.AwaitedSteps == nil {
		state.AwaitedSteps = []string{}
	}
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}
	if state.ExecutionPath == nil {
		state.ExecutionPath = []ExecutionRecord{}
	}

	awaitedStepsJSON, err := json.Marshal(state.AwaitedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal awaited_steps: %w", err)
	}

	collectedDataJSON, err := json.Marshal(state.CollectedData)
	if err != nil {
		return fmt.Errorf("failed to marshal collected_data: %w", err)
	}

	workflowPlanJSON, err := json.Marshal(state.WorkflowPlan)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow_plan: %w", err)
	}

	executionMetadataJSON, err := json.Marshal(state.ExecutionMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal execution_metadata: %w", err)
	}

	// Special handling for ExecutionPath to ensure proper time serialization
	executionPathJSON, err := json.Marshal(state.ExecutionPath)
	if err != nil {
		return fmt.Errorf("failed to marshal execution_path: %w", err)
	}

	// Ensure FinalResult is valid JSON or NULL
	var finalResultValue interface{}
	if state.FinalResult != nil && len(state.FinalResult) > 0 {
		finalResultValue = state.FinalResult
	} else {
		finalResultValue = nil
	}

	// Ensure Error is not empty for NULL handling
	var errorValue interface{}
	if state.Error != "" {
		errorValue = state.Error
	} else {
		errorValue = nil
	}

	query := `
        UPDATE orchestrator_state 
        SET status = $2, 
            current_step = $3, 
            awaited_steps = $4::jsonb, 
            collected_data = $5::jsonb, 
            final_result = $6::jsonb, 
            error = $7, 
            workflow_plan = $8::jsonb, 
            execution_metadata = $9::jsonb, 
            execution_path = $10::jsonb, 
            updated_at = $11
        WHERE correlation_id = $1
        AND updated_at = $12  -- Optimistic locking
    `

	result, err := r.db.ExecContext(ctx, query,
		state.CorrelationID,
		state.Status,
		state.CurrentStep,
		awaitedStepsJSON,
		collectedDataJSON,
		finalResultValue,
		errorValue,
		workflowPlanJSON,
		executionMetadataJSON,
		executionPathJSON,
		time.Now().UTC(),
		state.UpdatedAt,
	)

	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to update orchestration state",
				zap.Error(err),
				zap.String("correlation_id", state.CorrelationID))
		}
		return fmt.Errorf("failed to update state: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("state was modified by another process")
	}

	return nil
}

// AddExecutionRecord adds a step execution to the history
func (r *StateRepository) AddExecutionRecord(ctx context.Context, state *OrchestrationState, record ExecutionRecord) error {
	// Ensure the ExecutionPath is initialized
	if state.ExecutionPath == nil {
		state.ExecutionPath = []ExecutionRecord{}
	}

	state.ExecutionPath = append(state.ExecutionPath, record)
	return r.UpdateState(ctx, state)
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
