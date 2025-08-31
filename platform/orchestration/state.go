// FILE: platform/orchestration/state.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// OrchestrationStatus represents the current state of a workflow
type OrchestrationStatus string

const (
	StatusInitialized       OrchestrationStatus = "INITIALIZED" // Created but not started
	StatusRunning           OrchestrationStatus = "RUNNING"
	StatusPausedForHuman    OrchestrationStatus = "PAUSED_FOR_HUMAN"
	StatusExecutingStep     OrchestrationStatus = "EXECUTING_STEP"     // Actually running an action
	StatusAwaitingResponses OrchestrationStatus = "AWAITING_RESPONSES" // Waiting for external responses
	StatusCompleted         OrchestrationStatus = "COMPLETED"
	StatusFailed            OrchestrationStatus = "FAILED"
)

// ProcessedMessage tracks messages we've already handled
type ProcessedMessage struct {
	MessageID       string    `db:"message_id"`
	CorrelationID   string    `db:"correlation_id"`
	OrchestrationID string    `db:"orchestration_id"`
	ProcessedAt     time.Time `db:"processed_at"`
	ProcessedBy     string    `db:"processed_by"`
}

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
type OrchestrationState struct {
	// Identity
	OrchestrationID       string `db:"orchestration_id"`
	CorrelationID         string `db:"correlation_id"`
	OwnerAgentID          string `db:"owner_agent_id"`
	ParentOrchestrationID string `db:"parent_orchestration_id"`
	ClientID              string `db:"client_id"`

	// State
	Status             OrchestrationStatus `db:"status"`
	CurrentStep        string              `db:"current_step"`
	AwaitedSteps       []string            `db:"awaited_steps"`
	CurrentlyExecuting *string             `db:"currently_executing"`
	LastActivity       time.Time           `db:"last_activity"`
	ProcessingNode     string              `db:"processing_node"`
	ExecutionStartedAt *time.Time          `db:"execution_started_at"`

	// Data
	CollectedData      map[string]interface{} `db:"collected_data"`
	InitialRequestData json.RawMessage        `db:"initial_request_data"`
	FinalResult        json.RawMessage        `db:"final_result"`

	// Workflow
	WorkflowPlan models.WorkflowPlan `db:"workflow_plan"`

	// Tracking
	ExecutionPath     []ExecutionRecord `db:"execution_path"`
	ExecutionMetadata ExecutionMetadata `db:"execution_metadata"`

	FuelBudget int `db:"fuel_budget"`

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

// HasProcessedMessage checks if we've already processed this message
func (r *StateRepository) HasProcessedMessage(ctx context.Context, messageID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM processed_messages WHERE message_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, messageID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check processed message: %w", err)
	}

	return exists, nil
}

// RecordMessageProcessing records that we're processing a message
func (r *StateRepository) RecordMessageProcessing(ctx context.Context, messageID, correlationID, orchestrationID string) error {
	processingNode := os.Getenv("HOSTNAME")
	if processingNode == "" {
		processingNode = "unknown"
	}

	query := `
		INSERT INTO processed_messages 
		(message_id, correlation_id, orchestration_id, processed_at, processed_by)
		VALUES ($1, $2, $3, NOW(), $4)
		ON CONFLICT (message_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, messageID, correlationID, orchestrationID, processingNode)
	if err != nil {
		return fmt.Errorf("failed to record message processing: %w", err)
	}

	return nil
}

// CleanupOldMessages removes processed messages older than the retention period
func (r *StateRepository) CleanupOldMessages(ctx context.Context, retentionDays int) error {
	query := `DELETE FROM processed_messages WHERE processed_at < NOW() - INTERVAL '%d days'`

	_, err := r.db.ExecContext(ctx, fmt.Sprintf(query, retentionDays))
	if err != nil {
		return fmt.Errorf("failed to cleanup old messages: %w", err)
	}

	return nil
}

// CreateInitialState creates a new orchestration with the plan
func (r *StateRepository) CreateInitialState(ctx context.Context, orchestrationID, correlationID, ownerAgentID, parentOrchestrationID, clientID string, plan models.WorkflowPlan, initialData []byte) error {
	// Prepare JSON fields
	awaitedStepsJSON, _ := json.Marshal([]string{})
	collectedDataJSON, _ := json.Marshal(map[string]interface{}{})
	workflowPlanJSON, _ := json.Marshal(plan)

	processingNode := os.Getenv("HOSTNAME")
	if processingNode == "" {
		processingNode = "unknown"
	}

	r.logger.Info("Creating initial state with plan",
		zap.String("orchestration_id", orchestrationID),
		zap.String("start_step", plan.StartStep),
		zap.Int("steps", len(plan.Steps)),
		zap.String("owner_agent_id", ownerAgentID),
		zap.String("status", string(StatusInitialized))) // CHANGED: StatusInitialized

	if plan.StartStep == "" {
		return fmt.Errorf("workflow plan has empty start_step")
	}

	// Parse initial data to extract agent type
	var initData map[string]interface{}
	if initialData != nil {
		json.Unmarshal(initialData, &initData)
	} else {
		initialData = []byte("null")
	}

	// Store agent type in collected data
	collectedData := map[string]interface{}{
		"agent_type": os.Getenv("AGENT_TYPE"),
	}

	collectedDataJSON, _ = json.Marshal(collectedData)

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
	var parentOrchestrationIDValue sql.NullString
	if parentOrchestrationID != "" {
		parentOrchestrationIDValue = sql.NullString{String: parentOrchestrationID, Valid: true}
	}

	query := `
		INSERT INTO orchestration_states 
		(orchestration_id, correlation_id, owner_agent_id, parent_orchestration_id, client_id,
		 status, current_step, awaited_steps, collected_data, initial_request_data,
		 workflow_plan, execution_metadata, execution_path, version, created_at, updated_at,
		 currently_executing, last_activity, processing_node)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		orchestrationID, correlationID, ownerAgentID, parentOrchestrationIDValue, clientID,
		StatusInitialized, plan.StartStep, awaitedStepsJSON, collectedDataJSON, initialData, // CHANGED: StatusInitialized
		workflowPlanJSON, metadataJSON, executionPathJSON, 1, now, now,
		nil, now, processingNode) // Added new fields

	if err != nil {
		r.logger.Error("Failed to create initial orchestration state",
			zap.Error(err),
			zap.String("orchestration_id", orchestrationID))
		return fmt.Errorf("failed to create initial state: %w", err)
	}

	r.logger.Info("Initial orchestration state created",
		zap.String("orchestration_id", orchestrationID),
		zap.String("correlation_id", correlationID),
		zap.String("owner_agent_id", ownerAgentID),
		zap.String("start_step", plan.StartStep),
		zap.String("status", string(StatusInitialized))) // CHANGED: StatusInitialized

	return nil
}

// SetExecutingStep marks a step as currently executing
func (r *StateRepository) SetExecutingStep(ctx context.Context, orchestrationID string, stepName string) error {
	processingNode := os.Getenv("HOSTNAME")
	if processingNode == "" {
		processingNode = "unknown"
	}

	now := time.Now().UTC()
	query := `
		UPDATE orchestration_states 
		SET status = $2,
		    currently_executing = $3,
		    last_activity = $4,
		    processing_node = $5,
		    execution_started_at = $6,
		    version = version + 1,
		    updated_at = $7
		WHERE orchestration_id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		orchestrationID,
		StatusExecutingStep,
		stepName,
		now,
		processingNode,
		now,
		now)

	if err != nil {
		return fmt.Errorf("failed to set executing step: %w", err)
	}

	r.logger.Info("Set step as executing",
		zap.String("orchestration_id", orchestrationID),
		zap.String("step", stepName),
		zap.String("node", processingNode))

	return nil
}

// ClearExecutingStep clears the currently executing step
func (r *StateRepository) ClearExecutingStep(ctx context.Context, orchestrationID string) error {
	query := `
		UPDATE orchestration_states 
		SET currently_executing = NULL,
		    last_activity = $2,
		    version = version + 1,
		    updated_at = $3
		WHERE orchestration_id = $1
	`

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query, orchestrationID, now, now)

	if err != nil {
		return fmt.Errorf("failed to clear executing step: %w", err)
	}

	return nil
}

// CheckStuckOrchestrations finds orchestrations that appear to be stuck
func (r *StateRepository) CheckStuckOrchestrations(ctx context.Context, timeout time.Duration) ([]string, error) {
	query := `
		SELECT orchestration_id 
		FROM orchestration_states 
		WHERE status = $1 
		  AND currently_executing IS NOT NULL
		  AND last_activity < $2
	`

	cutoff := time.Now().UTC().Add(-timeout)
	rows, err := r.db.QueryContext(ctx, query, StatusExecutingStep, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to check stuck orchestrations: %w", err)
	}
	defer rows.Close()

	var stuckIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		stuckIDs = append(stuckIDs, id)
	}

	return stuckIDs, nil
}

// GetState retrieves state by orchestrationID (primary lookup)
func (r *StateRepository) GetState(ctx context.Context, orchestrationID string) (*OrchestrationState, error) {
	query := `
		SELECT orchestration_id, correlation_id, owner_agent_id, parent_orchestration_id, client_id,
		       status, current_step, awaited_steps, collected_data, initial_request_data,
		       final_result, error, workflow_plan, execution_metadata, execution_path,
		       version, created_at, updated_at, currently_executing, last_activity, 
		       processing_node, execution_started_at
		FROM orchestration_states
		WHERE orchestration_id = $1
	`

	var state OrchestrationState
	var awaitedStepsJSON, collectedDataJSON, workflowPlanJSON []byte
	var executionMetadataJSON, executionPathJSON []byte
	var parentOrchestrationIDNull, initialRequestDataNull, finalResultNull, errorNull sql.NullString
	var currentlyExecutingNull, processingNodeNull sql.NullString
	var executionStartedAtNull sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orchestrationID).Scan(
		&state.OrchestrationID,
		&state.CorrelationID,
		&state.OwnerAgentID,
		&parentOrchestrationIDNull,
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
		&currentlyExecutingNull,
		&state.LastActivity,
		&processingNodeNull,
		&executionStartedAtNull,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found for orchestration_id: %s", orchestrationID)
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Handle nullable fields
	if parentOrchestrationIDNull.Valid {
		state.ParentOrchestrationID = parentOrchestrationIDNull.String
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
	if currentlyExecutingNull.Valid {
		state.CurrentlyExecuting = &currentlyExecutingNull.String
	}
	if processingNodeNull.Valid {
		state.ProcessingNode = processingNodeNull.String
	}
	if executionStartedAtNull.Valid {
		state.ExecutionStartedAt = &executionStartedAtNull.Time
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
		SELECT orchestration_id
		FROM orchestration_states
		WHERE correlation_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var orchestrationID string
	err := r.db.QueryRowContext(ctx, query, correlationID).Scan(&orchestrationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found for correlation_id: %s", correlationID)
		}
		return nil, fmt.Errorf("failed to get state by correlation: %w", err)
	}

	// Now get the full state using the orchestration ID
	return r.GetState(ctx, orchestrationID)
}

// UpdateState persists changes with optimistic locking and retry
func (r *StateRepository) UpdateState(ctx context.Context, state *OrchestrationState) error {
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.attemptUpdate(ctx, state)
		if err == nil {
			return nil
		}

		if !strings.Contains(err.Error(), "optimistic lock failure") {
			return err
		}

		if attempt < maxRetries-1 {
			// Reload and merge
			freshState, err := r.GetState(ctx, state.OrchestrationID)
			if err != nil {
				return fmt.Errorf("failed to reload state: %w", err)
			}

			// Merge critical updates
			freshState.Status = state.Status
			freshState.CurrentStep = state.CurrentStep
			freshState.AwaitedSteps = state.AwaitedSteps
			freshState.CurrentlyExecuting = state.CurrentlyExecuting
			freshState.LastActivity = state.LastActivity

			// Merge CollectedData selectively
			for k, v := range state.CollectedData {
				freshState.CollectedData[k] = v
			}

			state = freshState
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}
	}

	return fmt.Errorf("failed after %d retries: optimistic lock failure", maxRetries)
}

// attemptUpdate persists changes with optimistic locking
func (r *StateRepository) attemptUpdate(ctx context.Context, state *OrchestrationState) error {
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

	var currentlyExecutingValue interface{} = nil
	if state.CurrentlyExecuting != nil {
		currentlyExecutingValue = *state.CurrentlyExecuting
	}

	// Validate owner_agent_id
	if state.OwnerAgentID == "" {
		r.logger.Error("Cannot update state with empty owner_agent_id",
			zap.String("orchestration_id", state.OrchestrationID))
		return fmt.Errorf("owner_agent_id is required for update")
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
		    currently_executing = $11,
		    last_activity = $12,
		    version = version + 1,
		    updated_at = $13
		WHERE orchestration_id = $1
		  AND owner_agent_id = $14
		  AND version = $15
	`

	now := time.Now().UTC()
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
		currentlyExecutingValue,
		now, // last_activity
		now, // updated_at
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

// FindByAwaitedRequestID finds orchestration waiting for a specific request
func (r *StateRepository) FindByAwaitedRequestID(ctx context.Context, requestID string) (*OrchestrationState, error) {
	r.logger.Info("CRITICAL: Searching for orchestration awaiting request",
		zap.String("request_id", requestID))

	// Try the JSONB contains operator for array search
	query := `
		SELECT orchestration_id
		FROM orchestration_states
		WHERE status = 'AWAITING_RESPONSES' 
		  AND awaited_steps @> $1::jsonb
		LIMIT 1
	`

	// Create JSON array with the request ID
	jsonArray := fmt.Sprintf(`["%s"]`, requestID)

	var orchestrationID string
	err := r.db.QueryRowContext(ctx, query, jsonArray).Scan(&orchestrationID)

	if err != nil {
		if err == sql.ErrNoRows {
			// Try alternative query using JSONB array element search
			alternativeQuery := `
				SELECT orchestration_id
				FROM orchestration_states
				WHERE status = 'AWAITING_RESPONSES'
				  AND EXISTS (
					SELECT 1 FROM jsonb_array_elements_text(awaited_steps) AS step
					WHERE step = $1
				  )
				LIMIT 1
			`

			err = r.db.QueryRowContext(ctx, alternativeQuery, requestID).Scan(&orchestrationID)
			if err != nil {
				r.logger.Debug("No state found for awaited request_id",
					zap.String("request_id", requestID),
					zap.Error(err))
				return nil, fmt.Errorf("state not found for awaited request_id: %s", requestID)
			}
		} else {
			r.logger.Error("Failed to find orchestration state by awaited request ID",
				zap.Error(err),
				zap.String("request_id", requestID))
			return nil, fmt.Errorf("failed to find state by awaited request ID: %w", err)
		}
	}

	if err == nil {
		r.logger.Info("CRITICAL: Found orchestration waiting",
			zap.String("orchestration_id", orchestrationID),
			zap.String("for_request_id", requestID))
	}
	
	// Now get the full state
	return r.GetState(ctx, orchestrationID)
}

// AddAwaitedRequest adds a request to awaited list
func (r *StateRepository) AddAwaitedRequest(ctx context.Context, orchestrationID string, requestID string) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	// Ensure AwaitedSteps is initialized
	if state.AwaitedSteps == nil {
		state.AwaitedSteps = []string{}
	}

	// Check if already awaiting this request
	for _, awaited := range state.AwaitedSteps {
		if awaited == requestID {
			r.logger.Debug("Request already in awaited list",
				zap.String("orchestration_id", orchestrationID),
				zap.String("request_id", requestID))
			return nil
		}
	}

	// Add to awaited steps
	state.AwaitedSteps = append(state.AwaitedSteps, requestID)
	state.Status = StatusAwaitingResponses

	r.logger.Info("Adding request to awaited steps",
		zap.String("orchestration_id", orchestrationID),
		zap.String("request_id", requestID),
		zap.Int("total_awaited", len(state.AwaitedSteps)))

	return r.UpdateState(ctx, state)
}

// RemoveAwaitedRequest removes a request from awaited list
func (r *StateRepository) RemoveAwaitedRequest(ctx context.Context, orchestrationID string, requestID string) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	updatedAwaitedSteps := []string{}
	found := false
	for _, awaited := range state.AwaitedSteps {
		if awaited != requestID {
			updatedAwaitedSteps = append(updatedAwaitedSteps, awaited)
		} else {
			found = true
		}
	}

	if !found {
		r.logger.Warn("Request was not in awaited list",
			zap.String("orchestration_id", orchestrationID),
			zap.String("request_id", requestID))
		return nil
	}

	state.AwaitedSteps = updatedAwaitedSteps

	// Update status if no more awaited requests
	if len(state.AwaitedSteps) == 0 {
		// Clear executing state when done waiting
		state.CurrentlyExecuting = nil
		state.Status = StatusExecutingStep // Ready to continue
	}

	r.logger.Info("Removed request from awaited steps",
		zap.String("orchestration_id", orchestrationID),
		zap.String("request_id", requestID),
		zap.Int("remaining_awaited", len(state.AwaitedSteps)))

	return r.UpdateState(ctx, state)
}
