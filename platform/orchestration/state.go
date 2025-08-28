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

// PendingRequest tracks async requests
type PendingRequest struct {
	RequestID string
	AgentID   string
	AgentType string
	Step      string
	StartTime time.Time
}

// OrchestrationState is the database model for orchestration instances
// Using the NEW schema from your architecture document
type OrchestrationState struct {
	// Identity
	OrchestrationID       string `db:"orchestration_id"`
	CorrelationID         string `db:"correlation_id"`
	OwnerAgentID          string `db:"owner_agent_id"`
	ParentOrchestrationID string `db:"parent_orchestration_id"`
	ClientID              string `db:"client_id"`

	// State
	Status       OrchestrationStatus `db:"status"`
	CurrentStep  string              `db:"current_step"`
	AwaitedSteps []string            `db:"awaited_steps"`

	// Data
	CollectedData      map[string]interface{} `db:"collected_data"`
	InitialRequestData json.RawMessage        `db:"initial_request_data"`
	FinalResult        json.RawMessage        `db:"final_result"`

	// Workflow
	WorkflowPlan models.WorkflowPlan `db:"workflow_plan"`

	// Tracking
	ExecutionPath     []ExecutionRecord `db:"execution_path"`
	ExecutionMetadata ExecutionMetadata `db:"execution_metadata"`

	// Error handling
	Error string `db:"error"`

	// Versioning
	Version int `db:"version"`

	// Timestamps
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// StateRepository provides persistence for orchestration state
type StateRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewStateRepository creates a new state repository
func NewStateRepository(db *sql.DB, logger *zap.Logger) *StateRepository {
	return &StateRepository{db: db, logger: logger}
}

// CreateInitialState creates a new orchestration with the plan
func (r *StateRepository) CreateInitialState(ctx context.Context, orchestrationID, correlationID, ownerAgentID, ParentOrchestrationID, clientID string, plan models.WorkflowPlan, initialData []byte) error {
	// Prepare JSON fields
	awaitedStepsJSON, _ := json.Marshal([]string{})
	collectedDataJSON, _ := json.Marshal(map[string]interface{}{})
	workflowPlanJSON, _ := json.Marshal(plan)

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

	// Handle parent_orchestration_id - convert empty string to nil for database
	var ParentOrchestrationIDValue interface{}
	if ParentOrchestrationID != "" {
		ParentOrchestrationIDValue = ParentOrchestrationID
	} else {
		ParentOrchestrationIDValue = nil // This will be inserted as NULL
	}

	r.logger.Info("Initial orchestration state variables",
		zap.String("DEBUG_STATE_1: orchestration_id", orchestrationID),
		zap.String("DEBUG_STATE_1: correlation_id", correlationID),
		zap.String("DEBUG_STATE_1: owner_agent_id", ownerAgentID),
		zap.String("DEBUG_STATE_1: start_step", plan.StartStep))

	query := `
        INSERT INTO orchestration_states 
        (orchestration_id, correlation_id, owner_agent_id, parent_orchestration_id, client_id,
         status, current_step, awaited_steps, collected_data, initial_request_data,
         workflow_plan, execution_metadata, execution_path, version, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
    `

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		orchestrationID, correlationID, ownerAgentID, ParentOrchestrationIDValue, clientID,
		StatusRunning, plan.StartStep, awaitedStepsJSON, collectedDataJSON, initialData,
		workflowPlanJSON, metadataJSON, executionPathJSON, 1, now, now)

	if err != nil {
		r.logger.Error("Failed to create initial orchestration state",
			zap.Error(err),
			zap.String("DEBUG_STATE_2: orchestration_id", orchestrationID))
		return fmt.Errorf("failed to create initial state: %w", err)
	}

	r.logger.Info("Initial orchestration state created",
		zap.String("DEBUG_STATE_3: orchestration_id", orchestrationID),
		zap.String("DEBUG_STATE_3: correlation_id", correlationID),
		zap.String("DEBUG_STATE_3: owner_agent_id", ownerAgentID),
		zap.String("DEBUG_STATE_3: start_step", plan.StartStep))

	return nil
}

// GetState retrieves state by orchestrationID (primary lookup)
func (r *StateRepository) GetState(ctx context.Context, orchestrationID string) (*OrchestrationState, error) {
	query := `
		SELECT orchestration_id, correlation_id, owner_agent_id, parent_orchestration_id, client_id,
		       status, current_step, awaited_steps, collected_data, initial_request_data,
		       final_result, error, workflow_plan, execution_metadata, execution_path,
		       version, created_at, updated_at
		FROM orchestration_states
		WHERE orchestration_id = $1
	`

	var state OrchestrationState
	var awaitedStepsJSON, collectedDataJSON, workflowPlanJSON []byte
	var executionMetadataJSON, executionPathJSON []byte
	var ParentOrchestrationIDNull, initialRequestDataNull, finalResultNull, errorNull sql.NullString

	err := r.db.QueryRowContext(ctx, query, orchestrationID).Scan(
		&state.OrchestrationID,
		&state.CorrelationID,
		&state.OwnerAgentID,
		&ParentOrchestrationIDNull,
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
		&state.Version,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found for orchestration_id: %s", orchestrationID)
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Handle nullable fields
	if ParentOrchestrationIDNull.Valid {
		state.ParentOrchestrationID = ParentOrchestrationIDNull.String
	}
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

// GetStateByCorrelation retrieves state by correlationID (for backward compatibility)
func (r *StateRepository) GetStateByCorrelation(ctx context.Context, correlationID string) (*OrchestrationState, error) {
	query := `
		SELECT orchestration_id, correlation_id, owner_agent_id, parent_orchestration_id, client_id,
		       status, current_step, awaited_steps, collected_data, initial_request_data,
		       final_result, error, workflow_plan, execution_metadata, execution_path,
		       version, created_at, updated_at
		FROM orchestration_states
		WHERE correlation_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var state OrchestrationState
	var awaitedStepsJSON, collectedDataJSON, workflowPlanJSON []byte
	var executionMetadataJSON, executionPathJSON []byte
	var ParentOrchestrationIDNull, initialRequestDataNull, finalResultNull, errorNull sql.NullString

	err := r.db.QueryRowContext(ctx, query, correlationID).Scan(
		&state.OrchestrationID,
		&state.CorrelationID,
		&state.OwnerAgentID,
		&ParentOrchestrationIDNull,
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
		&state.Version,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found for correlation_id: %s", correlationID)
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Handle nullable fields and unmarshal JSON (same as GetState)
	if ParentOrchestrationIDNull.Valid {
		state.ParentOrchestrationID = ParentOrchestrationIDNull.String
	}
	if initialRequestDataNull.Valid {
		state.InitialRequestData = json.RawMessage(initialRequestDataNull.String)
	}
	if finalResultNull.Valid {
		state.FinalResult = json.RawMessage(finalResultNull.String)
	}
	if errorNull.Valid {
		state.Error = errorNull.String
	}

	json.Unmarshal(awaitedStepsJSON, &state.AwaitedSteps)
	json.Unmarshal(collectedDataJSON, &state.CollectedData)
	json.Unmarshal(workflowPlanJSON, &state.WorkflowPlan)
	json.Unmarshal(executionMetadataJSON, &state.ExecutionMetadata)
	json.Unmarshal(executionPathJSON, &state.ExecutionPath)

	return &state, nil
}

// UpdateState persists changes with optimistic locking
func (r *StateRepository) UpdateState(ctx context.Context, state *OrchestrationState) error {
	// Ensure JSON fields are initialized
	if state.AwaitedSteps == nil {
		state.AwaitedSteps = []string{}
	}
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}
	if state.ExecutionPath == nil {
		state.ExecutionPath = []ExecutionRecord{}
	}

	// Marshal JSON fields
	awaitedStepsJSON, _ := json.Marshal(state.AwaitedSteps)
	collectedDataJSON, _ := json.Marshal(state.CollectedData)
	workflowPlanJSON, _ := json.Marshal(state.WorkflowPlan)
	executionMetadataJSON, _ := json.Marshal(state.ExecutionMetadata)
	executionPathJSON, _ := json.Marshal(state.ExecutionPath)

	// Handle nullable fields
	var finalResultValue interface{} = nil
	if state.FinalResult != nil && len(state.FinalResult) > 0 {
		finalResultValue = state.FinalResult
	}

	var errorValue interface{} = nil
	if state.Error != "" {
		errorValue = state.Error
	}

	query := `
		UPDATE orchestration_states 
		SET status = $2, 
		    current_step = $3, 
		    awaited_steps = $4::jsonb, 
		    collected_data = $5::jsonb, 
		    final_result = $6::jsonb, 
		    error = $7, 
		    workflow_plan = $8::jsonb, 
		    execution_metadata = $9::jsonb, 
		    execution_path = $10::jsonb,
		    version = version + 1,
		    updated_at = $11
		WHERE orchestration_id = $1
		  AND owner_agent_id = $12
		  AND version = $13
	`

	result, err := r.db.ExecContext(ctx, query,
		state.OrchestrationID,
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
		state.OwnerAgentID,
		state.Version,
	)

	if err != nil {
		r.logger.Error("Failed to update orchestration state",
			zap.Error(err),
			zap.String("orchestration_id", state.OrchestrationID))
		return fmt.Errorf("failed to update state: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("state was modified by another process (optimistic lock failure)")
	}

	// Increment version for the caller
	state.Version++

	return nil
}

// AddExecutionRecord adds a step execution to the history
func (r *StateRepository) AddExecutionRecord(ctx context.Context, state *OrchestrationState, record ExecutionRecord) error {
	if state.ExecutionPath == nil {
		state.ExecutionPath = []ExecutionRecord{}
	}

	state.ExecutionPath = append(state.ExecutionPath, record)
	return r.UpdateState(ctx, state)
}
