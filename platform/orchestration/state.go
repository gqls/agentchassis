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
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// OrchestrationStatus represents the current state of a workflow
type OrchestrationStatus string

const (
	StatusInitialized       OrchestrationStatus = "INITIALIZED"
	StatusRunning           OrchestrationStatus = "RUNNING"
	StatusPausedForHuman    OrchestrationStatus = "PAUSED_FOR_HUMAN"
	StatusExecutingStep     OrchestrationStatus = "EXECUTING_STEP"
	StatusAwaitingResponses OrchestrationStatus = "AWAITING_RESPONSES"
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

// AwaitedRequest tracks async requests with retry support
type AwaitedRequest struct {
	RequestID       string    `json:"request_id"`
	StepID          string    `json:"step_id"`
	StepName        string    `json:"step_name"`
	RetryVersion    int       `json:"retry_version"`
	TargetAgentID   string    `json:"target_agent_id,omitempty"`
	TargetAgentType string    `json:"target_agent_type"`
	ResponsesTopic  string    `json:"responses_topic"`
	RequestsTopic   string    `json:"requests_topic"`
	SentAt          time.Time `json:"sent_at"`
	TimeoutAt       time.Time `json:"timeout_at"`
	ParentRequestID string    `json:"parent_request_id,omitempty"`
}

// ProcessingRecord tracks which pod processed what (for debugging)
type ProcessingRecord struct {
	PodName   string    `json:"pod_name"`
	StepID    string    `json:"step_id"`
	StepName  string    `json:"step_name"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// OrchestrationState is the database model for orchestration instances
type OrchestrationState struct {
	// Identity
	OrchestrationID       string `db:"orchestration_id"`
	OrchestrationName     string `json:"orchestration_name,omitempty"`
	CorrelationID         string `db:"correlation_id"`
	OwnerAgentID          string `db:"owner_agent_id"`
	OwnerAgentType        string `db:"owner_agent_type"`
	OwnerAgentRole        string `db:"owner_agent_role"`
	ParentOrchestrationID string `db:"parent_orchestration_id"`
	ClientID              string `db:"client_id"`

	RequestsTopic  string `db:"requests_topic"`  // Where THIS orchestration listens
	ResponsesTopic string `db:"responses_topic"` // Where THIS orchestration sends responses

	// State
	Status             OrchestrationStatus        `db:"status"`
	CurrentStep        string                     `db:"current_step"`
	AwaitedSteps       []string                   `db:"awaited_steps"`    // Legacy - kept for compatibility
	AwaitedRequests    map[string]*AwaitedRequest `db:"awaited_requests"` // NEW: Request-based tracking
	CurrentlyExecuting *string                    `db:"currently_executing"`
	LastActivity       time.Time                  `db:"last_activity"`
	ProcessingNode     string                     `db:"processing_node"`
	ExecutionStartedAt *time.Time                 `db:"execution_started_at"`

	// Data
	CollectedData      map[string]interface{} `db:"collected_data"`
	InitialRequestData json.RawMessage        `db:"initial_request_data"`
	FinalResult        json.RawMessage        `db:"final_result"`

	// Workflow
	WorkflowPlan models.WorkflowPlan `db:"workflow_plan"`

	// Tracking
	ExecutionPath     []ExecutionRecord             `db:"execution_path"`
	ExecutionMetadata ExecutionMetadata             `db:"execution_metadata"`
	ProcessingHistory []ProcessingRecord            `db:"processing_history"` // NEW: Processing audit trail
	SubtreeAgents     map[string]*types.SubtreeInfo `db:"subtree_agents"`     // NEW: Hierarchical agent tracking

	FuelBudget int `db:"fuel_budget"`

	// Error handling
	Error string `db:"error"`

	// Versioning for optimistic locking
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

// HasProcessedMessage checks if a request has been processed
func (r *StateRepository) HasProcessedMessage(ctx context.Context, correlationID, requestID, agentID string) (bool, error) {
	if requestID == "" {
		return false, nil // Can't deduplicate without request_id
	}

	query := `
        SELECT EXISTS(
            SELECT 1 FROM processed_messages 
            WHERE correlation_id = $1 
            AND request_id = $2
            AND agent_id = $3
        )`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, correlationID, requestID, agentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check processed request: %w", err)
	}

	return exists, nil
}

// RecordMessageProcessing records a request being processed
func (r *StateRepository) RecordMessageProcessing(ctx context.Context, execCtx *types.ExecutionContext, agentID string) error {
	if execCtx.RequestID == "" {
		return nil // Can't record without request_id
	}

	// Handle empty orchestration_id - we shouldn't have it in real life but when sending my own requests without them then:
	orchestrationID := execCtx.OrchestrationID
	//fmt.Printf("DEBUG orch: in RecordMessageProcessing orchestrationID %s\n", orchestrationID)
	if orchestrationID == "" {
		orchestrationID = "00000000-0000-0000-0214-000000000010" // NULL UUID orig - 0214
		// Or just skip recording if no orchestration
		// return nil
	}

	processingNode := os.Getenv("HOSTNAME")
	if processingNode == "" {
		processingNode = "unknown"
	}

	query := `
        INSERT INTO processed_messages 
        (message_id, correlation_id, orchestration_id, request_id, agent_id, message_type, processed_at, processed_by)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7)
        ON CONFLICT (correlation_id, request_id, agent_id) DO NOTHING
    `

	_, err := r.db.ExecContext(ctx, query,
		execCtx.MessageID,
		execCtx.CorrelationID,
		orchestrationID,
		execCtx.RequestID,
		agentID,
		execCtx.MessageType,
		processingNode)

	if err != nil {
		return fmt.Errorf("failed to record request processing in state go: %w", err)
	}

	return nil
}

// CreateState creates a new orchestration state
func (r *StateRepository) CreateState(ctx context.Context, state *OrchestrationState) error {
	r.logger.Info("CreateState called",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("orchestration_name", state.OrchestrationName),
		zap.String("correlation_id", state.CorrelationID))

	// Initialize maps if nil
	if state.AwaitedRequests == nil {
		state.AwaitedRequests = make(map[string]*AwaitedRequest)
	}
	if state.SubtreeAgents == nil {
		state.SubtreeAgents = make(map[string]*types.SubtreeInfo)
	}
	if state.ProcessingHistory == nil {
		state.ProcessingHistory = []ProcessingRecord{}
	}

	// Add initial processing record
	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   os.Getenv("HOSTNAME"),
		Action:    "orchestration_created",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Created by %s", state.OwnerAgentID),
	})

	// Serialize complex fields - ensure valid JSON even when empty
	awaitedStepsJSON := []byte("[]")
	if state.AwaitedSteps != nil && len(state.AwaitedSteps) > 0 {
		awaitedStepsJSON, _ = json.Marshal(state.AwaitedSteps)
	}

	collectedDataJSON := []byte("{}")
	if state.CollectedData != nil && len(state.CollectedData) > 0 {
		collectedDataJSON, _ = json.Marshal(state.CollectedData)
	}

	workflowPlanJSON := []byte("{}")
	if wfData, err := json.Marshal(state.WorkflowPlan); err == nil {
		workflowPlanJSON = wfData
	}

	executionMetadataJSON := []byte("{}")
	if emData, err := json.Marshal(state.ExecutionMetadata); err == nil {
		executionMetadataJSON = emData
	}

	executionPathJSON := []byte("[]")
	if state.ExecutionPath != nil && len(state.ExecutionPath) > 0 {
		executionPathJSON, _ = json.Marshal(state.ExecutionPath)
	}

	awaitedRequestsJSON := []byte("{}")
	if state.AwaitedRequests != nil && len(state.AwaitedRequests) > 0 {
		awaitedRequestsJSON, _ = json.Marshal(state.AwaitedRequests)
	}

	processingHistoryJSON := []byte("[]")
	if state.ProcessingHistory != nil && len(state.ProcessingHistory) > 0 {
		processingHistoryJSON, _ = json.Marshal(state.ProcessingHistory)
	}

	subtreeAgentsJSON := []byte("{}")
	if state.SubtreeAgents != nil && len(state.SubtreeAgents) > 0 {
		subtreeAgentsJSON, _ = json.Marshal(state.SubtreeAgents)
	}

	var parentOrchIDValue interface{}
	if state.ParentOrchestrationID == "" {
		parentOrchIDValue = nil
	} else {
		parentOrchIDValue = state.ParentOrchestrationID
	}

	var initialRequestDataValue interface{}
	if state.InitialRequestData != nil && len(state.InitialRequestData) > 0 {
		initialRequestDataValue = state.InitialRequestData
	} else {
		initialRequestDataValue = json.RawMessage("{}")
	}

	r.logger.Info("Attempting to insert orchestration state",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("orchestration_name", state.OrchestrationName),
		zap.String("correlation_id", state.CorrelationID),
		zap.String("owner_agent_id", state.OwnerAgentID),
		zap.Any("parent_orchestration_id", parentOrchIDValue),
		zap.String("status", string(state.Status)),
		zap.String("awaited_steps_json", string(awaitedStepsJSON)),
		zap.String("awaited_requests_json", string(awaitedRequestsJSON)),
		zap.String("collected_data_json", string(collectedDataJSON)),
		zap.String("initial_request_data_raw", string(state.InitialRequestData)),
		zap.String("workflow_plan_json", string(workflowPlanJSON)),
		zap.String("execution_metadata_json", string(executionMetadataJSON)),
	)

	query := `
        INSERT INTO orchestration_states (
            orchestration_id, orchestration_name, correlation_id, owner_agent_id, owner_agent_type, 
            parent_orchestration_id, client_id, status, current_step, awaited_steps, 
            awaited_requests, currently_executing, last_activity, processing_node, execution_started_at,
            collected_data, initial_request_data, final_result, workflow_plan,
            execution_path, execution_metadata, processing_history, subtree_agents,
            fuel_budget, error, version, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
            $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, 
            $21, $22, $23, $24, $25, $26, NOW(), NOW()
        ) ON CONFLICT (orchestration_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query,
		state.OrchestrationID,    // $1
		state.OrchestrationName,  // $2
		state.CorrelationID,      // $3
		state.OwnerAgentID,       // $4
		state.OwnerAgentType,     // $5
		parentOrchIDValue,        // $6
		state.ClientID,           // $7
		state.Status,             // $8
		state.CurrentStep,        // $9
		awaitedStepsJSON,         // $10
		awaitedRequestsJSON,      // $11
		state.CurrentlyExecuting, // $12
		state.LastActivity,       // $13
		state.ProcessingNode,     // $14
		state.ExecutionStartedAt, // $15
		collectedDataJSON,        // $16
		initialRequestDataValue,  // $17
		state.FinalResult,        // $18
		workflowPlanJSON,         // $19
		executionPathJSON,        // $20
		executionMetadataJSON,    // $21
		processingHistoryJSON,    // $22
		subtreeAgentsJSON,        // $23
		state.FuelBudget,         // $24
		state.Error,              // $25
		1,                        // $26 - version
	)

	if err != nil {
		return fmt.Errorf("failed to create state: %w", err)
	}

	state.Version = 1
	return nil
}

func (r *StateRepository) CreateInitialState(
	ctx context.Context,
	orchestrationID string,
	orchestrationName string,
	correlationID string,
	ownerAgentID string,
	ownerAgentType string,
	ownerAgentRole string,
	parentOrchestrationID string,
	clientID string,
	plan models.WorkflowPlan,
	initialData []byte, // message value
	execCtx *types.ExecutionContext,
) error {
	r.logger.Info("CreateInitialState parameters",
		zap.String("orchestrationID", orchestrationID),
		zap.String("parentOrchestrationID", parentOrchestrationID),
		zap.String("DEBUGaa: CreateInitialState initialData look for action, request id etc", string(initialData)),
	)

	// Updated query to include topics
	query := `
		INSERT INTO orchestration_states (
			orchestration_id, orchestration_name, correlation_id, owner_agent_id, owner_agent_type, 
			owner_agent_role, parent_orchestration_id, client_id,
			requests_topic, responses_topic,
			status, current_step, awaited_steps, collected_data, initial_request_data,
			workflow_plan, execution_metadata, execution_path, 
		    processing_history, subtree_agents, fuel_budget,
		    version, created_at, updated_at,
			currently_executing, last_activity, processing_node
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
		ON CONFLICT (orchestration_id) DO NOTHING
	`

	// Store where THIS orchestration listens
	var requestsTopic, responsesTopic string
	if execCtx != nil {
		requestsTopic = execCtx.RequestsTopic
		responsesTopic = execCtx.ResponsesTopic
	} else {
		execCtx = &types.ExecutionContext{}
		requestsTopic = os.Getenv("REQUESTS_TOPIC")
		responsesTopic = os.Getenv("RESPONSES_TOPIC")
	}

	var unmarshalledInitialData map[string]interface{}
	// Parse and merge initial data
	if len(initialData) > 0 {
		if err := json.Unmarshal(initialData, &unmarshalledInitialData); err != nil {
			r.logger.Error("failed to unmarshal initial data in CreateInitialState")
		}
	}

	// Prepare collected data
	collectedData := datahelpers.NormalizeCollectedData(unmarshalledInitialData, execCtx, requestsTopic, r.logger)

	// Store where THIS orchestration listens/responds
	/*if execCtx.RequestsTopic != "" {
		collectedData["__my_requests_topic__"] = execCtx.RequestsTopic
	}
	*/
	/*	replyToTopic := execCtx.ReplyToTopic
		if replyToTopic != "" {
			// This is where I send MY responses (to parent)
			replyToTopic = os.Getenv("PARENT_RESPONSES_TOPIC")
			execCtx.ReplyToTopic = replyToTopic
			collectedData["__parent_responses_topic__"] = replyToTopic
		}*/

	/*var unmarshalledInitialData map[string]interface{}
	// Parse and merge initial data
	if len(initialData) > 0 {
		if err := json.Unmarshal(initialData, &unmarshalledInitialData); err == nil {
			for k, v := range unmarshalledInitialData {
				collectedData[k] = v
			}
		}
	}*/

	// Extract action from message
	/*if action, ok := unmarshalledInitialData["action"].(string); ok {
		collectedData["action"] = action
	}*/

	// Extract config from message (if present)
	/*if config, ok := unmarshalledInitialData["config"].(map[string]interface{}); ok {
		collectedData["config"] = config
	}*/

	// Extract agent_config if present (workflow definition)
	/*if agentConfig, ok := unmarshalledInitialData["agent_config"].(map[string]interface{}); ok {
		collectedData["agent_config"] = agentConfig
	}*/

	// Extract agent_group if present (for multi-agent spawns)
	/*if agentGroup, ok := unmarshalledInitialData["agent_group"].(map[string]interface{}); ok {
		collectedData["agent_group"] = agentGroup
	}*/

	// Extract agent_type if present
	/*if agentType, ok := unmarshalledInitialData["agent_type"].(string); ok {
		collectedData["agent_type"] = agentType
	}*/

	// Extract prompt if present (for LLM actions)
	/*if prompt, ok := unmarshalledInitialData["prompt"].(string); ok {
		collectedData["prompt"] = prompt
	}*/

	// Extract input_data to top level
	// This is the actual user/business data that flows through the system
	/*var inputData map[string]interface{}
	// Try body.input_data first (for child agents receiving from parent)
	if body, ok := unmarshalledInitialData["body"].(map[string]interface{}); ok {
		if data, ok := body["input_data"].(map[string]interface{}); ok {
			inputData = data
			r.logger.Info("Extracted input_data from body.input_data")
		}
	}

	// Fallback to message.input_data (for root agents receiving from external)
	if inputData == nil {
		if data, ok := unmarshalledInitialData["input_data"].(map[string]interface{}); ok {
			inputData = data
			r.logger.Info("Extracted input_data from message.input_data")
		}
	}

	// Store at top level for easy template access
	if inputData != nil {
		collectedData["input_data"] = inputData
	} else {
		// Initialize empty map to avoid nil pointer issues
		collectedData["input_data"] = map[string]interface{}{}
		r.logger.Warn("No input_data found in message, initialized empty map")
	}*/

	r.logger.Info("CreateInitialState collected data",
		zap.Any("collectedData", collectedData),
	)

	// Serialize fields
	awaitedStepsJSON, _ := json.Marshal([]string{})
	collectedDataJSON, _ := json.Marshal(collectedData)
	workflowPlanJSON, _ := json.Marshal(plan)

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

	now := time.Now().UTC()
	processingNode := os.Getenv("HOSTNAME")
	if processingNode == "" {
		processingNode = "unknown"
	}

	var parentOrchIDValue interface{}
	if parentOrchestrationID == "" {
		parentOrchIDValue = nil
	} else {
		parentOrchIDValue = parentOrchestrationID
	}

	// Extract topics from ExecutionContext
	/*	var requestsTopic, responsesTopic string
		if execCtx != nil {
			requestsTopic = execCtx.RequestsTopic
			responsesTopic = execCtx.ResponsesTopic
		}*/

	processingHistory := []ProcessingRecord{}
	subtreeAgents := make(map[string]*types.SubtreeInfo)
	fuelBudget := execCtx.FuelBudget

	// Execute insert with topics
	result, err := r.db.ExecContext(ctx, query,
		orchestrationID,   // $1
		orchestrationName, // $2
		correlationID,     // $3
		ownerAgentID,      // $4
		ownerAgentType,    // $5
		ownerAgentRole,    // $6
		parentOrchIDValue, // $7
		clientID,          // $8
		requestsTopic,     // $9
		responsesTopic,    // $10
		StatusInitialized, // $11
		plan.StartStep,    // $12
		awaitedStepsJSON,  // $13
		collectedDataJSON, // $14
		initialData,       // $15
		workflowPlanJSON,  // $16
		metadataJSON,      // $17
		executionPathJSON, // $18
		processingHistory, // $19
		subtreeAgents,     // $20
		fuelBudget,        // $21
		1,                 // $22 - version
		now,               // $23 - created_at
		now,               // $24 - updated_at
		nil,               // $25 - currently_executing
		now,               // $26 - last_activity
		processingNode)    // $27 - processing_node

	if err != nil {
		r.logger.Error("Failed to create initial state",
			zap.Error(err),
			zap.String("orchestration_id", orchestrationID))
		return fmt.Errorf("failed to create initial state: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Info("State already exists, proceeding",
			zap.String("orchestration_id", orchestrationID))
	} else {
		r.logger.Info("Initial Orchestration state created successfully",
			zap.String("orchestration_id", orchestrationID))
	}

	return nil
}

// GetState retrieves orchestration state by ID
func (r *StateRepository) GetState(ctx context.Context, orchestrationID string) (*OrchestrationState, error) {

	query := `
	SELECT 
		orchestration_id, orchestration_name, correlation_id, owner_agent_id, owner_agent_type, 
		owner_agent_role, parent_orchestration_id, client_id, 
		requests_topic, responses_topic, 
		status, current_step, awaited_steps, awaited_requests,
		currently_executing, last_activity, processing_node, execution_started_at,
		collected_data, initial_request_data, final_result, workflow_plan,
		execution_path, execution_metadata, processing_history, subtree_agents,
		fuel_budget, error, version, created_at, updated_at
	FROM orchestration_states
	WHERE orchestration_id = $1
`

	state := &OrchestrationState{}
	var collectedDataJSON, workflowPlanJSON, executionMetadataJSON, executionPathJSON []byte
	var awaitedRequestsJSON, awaitedStepsJSON, processingHistoryJSON, subtreeAgentsJSON []byte
	var finalResultValue, errorValue, parentOrchestrationIDNull, currentlyExecutingValue sql.NullString
	var executionStartedAtValue sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orchestrationID).Scan(
		&state.OrchestrationID,
		&state.OrchestrationName,
		&state.CorrelationID,
		&state.OwnerAgentID,
		&state.OwnerAgentType,
		&state.OwnerAgentRole,
		&parentOrchestrationIDNull,
		&state.ClientID,
		&state.RequestsTopic,
		&state.ResponsesTopic,
		&state.Status,
		&state.CurrentStep,
		&awaitedStepsJSON,
		&awaitedRequestsJSON,
		&currentlyExecutingValue,
		&state.LastActivity,
		&state.ProcessingNode,
		&executionStartedAtValue,
		&collectedDataJSON,
		&state.InitialRequestData,
		&finalResultValue,
		&workflowPlanJSON,
		&executionPathJSON,
		&executionMetadataJSON,
		&processingHistoryJSON,
		&subtreeAgentsJSON,
		&state.FuelBudget,
		&errorValue,
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

	// Initialize maps if nil
	if state.AwaitedRequests == nil {
		state.AwaitedRequests = make(map[string]*AwaitedRequest)
	}
	if state.SubtreeAgents == nil {
		state.SubtreeAgents = make(map[string]*types.SubtreeInfo)
	}

	// Deserialize complex fields
	json.Unmarshal(collectedDataJSON, &state.CollectedData)
	json.Unmarshal(workflowPlanJSON, &state.WorkflowPlan)
	json.Unmarshal(executionMetadataJSON, &state.ExecutionMetadata)
	json.Unmarshal(executionPathJSON, &state.ExecutionPath)
	json.Unmarshal(awaitedRequestsJSON, &state.AwaitedRequests)
	json.Unmarshal(processingHistoryJSON, &state.ProcessingHistory)
	json.Unmarshal(subtreeAgentsJSON, &state.SubtreeAgents)

	// Handle nullable fields
	if finalResultValue.Valid {
		state.FinalResult = json.RawMessage(finalResultValue.String)
	}
	if errorValue.Valid {
		state.Error = errorValue.String
	}
	if currentlyExecutingValue.Valid {
		state.CurrentlyExecuting = &currentlyExecutingValue.String
	}
	if executionStartedAtValue.Valid {
		state.ExecutionStartedAt = &executionStartedAtValue.Time
	}

	// deserialize it:
	if len(awaitedStepsJSON) > 0 {
		json.Unmarshal(awaitedStepsJSON, &state.AwaitedSteps)
	}
	if parentOrchestrationIDNull.Valid {
		state.ParentOrchestrationID = parentOrchestrationIDNull.String
	}

	return state, nil
}

// UpdateState updates an existing orchestration state with optimistic locking
func (r *StateRepository) UpdateState(ctx context.Context, state *OrchestrationState) error {
	return r.UpdateStateWithVersion(ctx, state, state.Version)
}

// UpdateStateWithVersion updates state with version check for optimistic locking
func (r *StateRepository) UpdateStateWithVersion(ctx context.Context, state *OrchestrationState, expectedVersion int) error {
	r.logger.Info("UpdateStateWithVersion in state.go 623")

	// Add processing record for this update
	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   os.Getenv("HOSTNAME"),
		StepID:    state.CurrentStep,
		Action:    "state_updated",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Status: %s", state.Status),
	})

	// Serialize complex fields
	awaitedStepsJSON, _ := json.Marshal(state.AwaitedSteps)
	collectedDataJSON, _ := json.Marshal(state.CollectedData)
	workflowPlanJSON, _ := json.Marshal(state.WorkflowPlan)
	executionMetadataJSON, _ := json.Marshal(state.ExecutionMetadata)
	executionPathJSON, _ := json.Marshal(state.ExecutionPath)
	awaitedRequestsJSON, _ := json.Marshal(state.AwaitedRequests)
	processingHistoryJSON, _ := json.Marshal(state.ProcessingHistory)
	subtreeAgentsJSON, _ := json.Marshal(state.SubtreeAgents)

	// Handle nullable fields
	var finalResultValue, errorValue, currentlyExecutingValue sql.NullString
	if state.FinalResult != nil {
		finalResultValue = sql.NullString{String: string(state.FinalResult), Valid: true}
	}
	if state.Error != "" {
		errorValue = sql.NullString{String: state.Error, Valid: true}
	}
	if state.CurrentlyExecuting != nil {
		currentlyExecutingValue = sql.NullString{String: *state.CurrentlyExecuting, Valid: true}
	}

	now := time.Now()

	query := `
		UPDATE orchestration_states SET
			status = $1, current_step = $2, awaited_steps = $3, awaited_requests = $4,
			currently_executing = $5, collected_data = $6, initial_request_data = $7,
			final_result = $8, error = $9, workflow_plan = $10, execution_metadata = $11,
			execution_path = $12, processing_history = $13, subtree_agents = $14,
			last_activity = $15, updated_at = $16, owner_agent_id = $17, owner_agent_type = $18, owner_agent_role = $19, version = $20
		WHERE orchestration_id = $21 AND version = $22
	`

	result, err := r.db.ExecContext(ctx, query,
		state.Status,
		state.CurrentStep,
		awaitedStepsJSON,
		awaitedRequestsJSON,
		currentlyExecutingValue,
		collectedDataJSON,
		state.InitialRequestData,
		finalResultValue,
		errorValue,
		workflowPlanJSON,
		executionMetadataJSON,
		executionPathJSON,
		processingHistoryJSON,
		subtreeAgentsJSON,
		now, // last_activity
		now, // updated_at
		state.OwnerAgentID,
		state.OwnerAgentType,
		state.OwnerAgentRole,
		expectedVersion+1, // New version
		state.OrchestrationID,
		expectedVersion, // Expected current version
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
	state.Version = expectedVersion + 1
	return nil
}

// AddAwaitedRequest adds a request to the awaited list
func (r *StateRepository) AddAwaitedRequest(ctx context.Context, orchestrationID string, request *AwaitedRequest) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	// Initialize if needed
	if state.AwaitedRequests == nil {
		state.AwaitedRequests = make(map[string]*AwaitedRequest)
	}

	// Store the child's response topic so we know where to listen
	if request.ResponsesTopic == "" {
		// This shouldn't happen with new architecture
		r.logger.Warn("No responses topic in awaited request",
			zap.String("request_id", request.RequestID))
	}

	// Check if already awaiting this request
	if existing, exists := state.AwaitedRequests[request.RequestID]; exists {
		r.logger.Debug("Request already in awaited list, updating retry version",
			zap.String("orchestration_id", orchestrationID),
			zap.String("request_id", request.RequestID),
			zap.Int("old_retry", existing.RetryVersion),
			zap.Int("new_retry", request.RetryVersion))
		existing.RetryVersion = request.RetryVersion
		existing.SentAt = request.SentAt
		existing.TimeoutAt = request.TimeoutAt
	} else {
		// Add new request
		state.AwaitedRequests[request.RequestID] = request

		// Also add to legacy AwaitedSteps for compatibility
		state.AwaitedSteps = append(state.AwaitedSteps, request.RequestID)
	}

	state.Status = StatusAwaitingResponses

	r.logger.Info("Adding request to awaited list",
		zap.String("orchestration_id", orchestrationID),
		zap.String("request_id", request.RequestID),
		zap.Int("retry_version", request.RetryVersion),
		zap.Int("total_awaited", len(state.AwaitedRequests)))

	return r.UpdateState(ctx, state)
}

// RemoveAwaitedRequest removes a request from the awaited list
func (r *StateRepository) RemoveAwaitedRequest(ctx context.Context, orchestrationID string, requestID string) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	// Remove from AwaitedRequests map
	if _, exists := state.AwaitedRequests[requestID]; !exists {
		r.logger.Warn("Request was not in awaited list",
			zap.String("orchestration_id", orchestrationID),
			zap.String("request_id", requestID))
		return nil
	}

	delete(state.AwaitedRequests, requestID)

	// Also remove from legacy AwaitedSteps
	updatedAwaitedSteps := []string{}
	for _, awaited := range state.AwaitedSteps {
		if awaited != requestID {
			updatedAwaitedSteps = append(updatedAwaitedSteps, awaited)
		}
	}
	state.AwaitedSteps = updatedAwaitedSteps

	// Update status if no more awaited requests
	if len(state.AwaitedRequests) == 0 {
		state.CurrentlyExecuting = nil
		state.Status = StatusExecutingStep // Ready to continue
	}

	r.logger.Info("Removed request from awaited list",
		zap.String("orchestration_id", orchestrationID),
		zap.String("request_id", requestID),
		zap.Int("remaining_awaited", len(state.AwaitedRequests)))

	return r.UpdateState(ctx, state)
}

// FindByAwaitedRequestID finds orchestration waiting for a specific request
func (r *StateRepository) FindByAwaitedRequestID(ctx context.Context, requestID string) (*OrchestrationState, error) {
	r.logger.Info("Searching for orchestration awaiting request",
		zap.String("request_id", requestID))

	// Query using JSONB contains for the awaited_requests map
	query := `
		SELECT orchestration_id
		FROM orchestration_states
		WHERE status = 'AWAITING_RESPONSES' 
		  AND awaited_requests ? $1
		LIMIT 1
	`

	var orchestrationID string
	err := r.db.QueryRowContext(ctx, query, requestID).Scan(&orchestrationID)

	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback to legacy awaited_steps array search
			alternativeQuery := `
				SELECT orchestration_id
				FROM orchestration_states
				WHERE status = 'AWAITING_RESPONSES'
				  AND $1 = ANY(awaited_steps)
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

	r.logger.Info("Found orchestration waiting for request",
		zap.String("orchestration_id", orchestrationID),
		zap.String("request_id", requestID))

	// Now get the full state
	return r.GetState(ctx, orchestrationID)
}

// AddSubtreeAgent adds an agent to the subtree hierarchy
func (r *StateRepository) AddSubtreeAgent(ctx context.Context, orchestrationID string, agentInfo *types.SubtreeInfo) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	if state.SubtreeAgents == nil {
		state.SubtreeAgents = make(map[string]*types.SubtreeInfo)
	}

	state.SubtreeAgents[agentInfo.AgentID] = agentInfo

	// If this agent has a parent, add it to parent's children
	if agentInfo.ParentAgentID != "" {
		if parent, exists := state.SubtreeAgents[agentInfo.ParentAgentID]; exists {
			if parent.Children == nil {
				parent.Children = make(map[string]*types.SubtreeInfo)
			}
			parent.Children[agentInfo.AgentID] = agentInfo
		}
	}

	r.logger.Info("Added agent to subtree",
		zap.String("orchestration_id", orchestrationID),
		zap.String("agent_id", agentInfo.AgentID),
		zap.String("agent_type", agentInfo.AgentType),
		zap.String("parent_id", agentInfo.ParentAgentID))

	return r.UpdateState(ctx, state)
}

// UpdateAgentPerformance updates performance metrics for an agent
func (r *StateRepository) UpdateAgentPerformance(ctx context.Context, orchestrationID, agentID string, metrics *types.PerformanceMetrics) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	if agent, exists := state.SubtreeAgents[agentID]; exists {
		agent.Performance = metrics
		agent.LastActiveAt = time.Now()
		return r.UpdateState(ctx, state)
	}

	return fmt.Errorf("agent %s not found in subtree", agentID)
}

// GetSubtreeForAgent returns the entire subtree rooted at the given agent
func (r *StateRepository) GetSubtreeForAgent(ctx context.Context, orchestrationID, agentID string) (*types.SubtreeInfo, error) {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	if agent, exists := state.SubtreeAgents[agentID]; exists {
		return agent, nil
	}

	return nil, fmt.Errorf("agent %s not found in subtree", agentID)
}

// AddExecutionRecord adds a step execution to the history
func (r *StateRepository) AddExecutionRecord(ctx context.Context, state *OrchestrationState, record ExecutionRecord) error {
	if state.ExecutionPath == nil {
		state.ExecutionPath = []ExecutionRecord{}
	}

	state.ExecutionPath = append(state.ExecutionPath, record)
	return r.UpdateState(ctx, state)
}

// SetExecutingStep atomically sets the currently executing step
func (r *StateRepository) SetExecutingStep(ctx context.Context, orchestrationID string, step string) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return err
	}

	state.CurrentlyExecuting = &step
	state.Status = StatusExecutingStep
	state.ProcessingNode = os.Getenv("HOSTNAME")

	return r.UpdateState(ctx, state)
}

// ClearExecutingStep clears the currently executing step
func (r *StateRepository) ClearExecutingStep(ctx context.Context, orchestrationID string) error {
	state, err := r.GetState(ctx, orchestrationID)
	if err != nil {
		return err
	}

	state.CurrentlyExecuting = nil

	// Only change status if not waiting for responses
	if state.Status == StatusExecutingStep {
		state.Status = StatusRunning
	}

	return r.UpdateState(ctx, state)
}

// ExecuteWithOptimisticLocking executes a function with retry on version conflicts
func (r *StateRepository) ExecuteWithOptimisticLocking(ctx context.Context, orchestrationID string, fn func(*OrchestrationState) error) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// Load current state with version
		state, err := r.GetState(ctx, orchestrationID)
		if err != nil {
			return err
		}

		currentVersion := state.Version

		// Execute the function
		if err := fn(state); err != nil {
			return err
		}

		// Try to save with version check
		err = r.UpdateStateWithVersion(ctx, state, currentVersion)
		if err == nil {
			return nil // Success!
		}

		if strings.Contains(err.Error(), "optimistic lock failure") {
			r.logger.Info("Version conflict, retrying",
				zap.String("orchestration_id", orchestrationID),
				zap.Int("attempt", i+1))
			time.Sleep(time.Millisecond * time.Duration(100*(i+1)))
			continue
		}

		return err
	}

	return fmt.Errorf("max retries exceeded for orchestration %s", orchestrationID)
}

func (r *StateRepository) GetStateByCorrelation(ctx context.Context, correlationID string) (*OrchestrationState, error) {
	query := `
        SELECT orchestration_id, orchestration_name, correlation_id, owner_agent_id, owner_agent_type, owner_agent_role, parent_orchestration_id,
               client_id, status, current_step, awaited_steps, awaited_requests,
               currently_executing, last_activity, processing_node, execution_started_at,
               collected_data, initial_request_data, final_result, workflow_plan,
               execution_path, execution_metadata, processing_history, subtree_agents,
               fuel_budget, error, version, created_at, updated_at
        FROM orchestration_states
        WHERE correlation_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `

	state := &OrchestrationState{}
	var collectedDataJSON, workflowPlanJSON, executionMetadataJSON, executionPathJSON []byte
	var awaitedRequestsJSON, awaitedStepsJSON, processingHistoryJSON, subtreeAgentsJSON []byte
	var finalResultValue, errorValue, currentlyExecutingValue sql.NullString
	var executionStartedAtValue sql.NullTime

	err := r.db.QueryRowContext(ctx, query, correlationID).Scan(
		&state.OrchestrationID,
		&state.OrchestrationName,
		&state.CorrelationID,
		&state.OwnerAgentID,
		&state.OwnerAgentType,
		&state.OwnerAgentRole,
		&state.ParentOrchestrationID,
		&state.ClientID,
		&state.Status,
		&state.CurrentStep,
		&awaitedStepsJSON,
		&awaitedRequestsJSON,
		&currentlyExecutingValue,
		&state.LastActivity,
		&state.ProcessingNode,
		&executionStartedAtValue,
		&collectedDataJSON,
		&state.InitialRequestData,
		&finalResultValue,
		&workflowPlanJSON,
		&executionPathJSON,
		&executionMetadataJSON,
		&processingHistoryJSON,
		&subtreeAgentsJSON,
		&state.FuelBudget,
		&errorValue,
		&state.Version,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	// deserialize it:
	if len(awaitedStepsJSON) > 0 {
		json.Unmarshal(awaitedStepsJSON, &state.AwaitedSteps)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found for correlation_id: %s", correlationID)
		}
		return nil, fmt.Errorf("failed to get state by correlation: %w", err)
	}

	// Deserialize complex fields
	if len(collectedDataJSON) > 0 {
		json.Unmarshal(collectedDataJSON, &state.CollectedData)
	}
	if len(workflowPlanJSON) > 0 {
		json.Unmarshal(workflowPlanJSON, &state.WorkflowPlan)
	}
	if len(executionMetadataJSON) > 0 {
		json.Unmarshal(executionMetadataJSON, &state.ExecutionMetadata)
	}
	if len(executionPathJSON) > 0 {
		json.Unmarshal(executionPathJSON, &state.ExecutionPath)
	}
	if len(awaitedRequestsJSON) > 0 {
		json.Unmarshal(awaitedRequestsJSON, &state.AwaitedRequests)
	}
	if len(processingHistoryJSON) > 0 {
		json.Unmarshal(processingHistoryJSON, &state.ProcessingHistory)
	}
	if len(subtreeAgentsJSON) > 0 {
		json.Unmarshal(subtreeAgentsJSON, &state.SubtreeAgents)
	}

	// Handle nullable fields
	if finalResultValue.Valid {
		state.FinalResult = json.RawMessage(finalResultValue.String)
	}
	if errorValue.Valid {
		state.Error = errorValue.String
	}
	if currentlyExecutingValue.Valid {
		state.CurrentlyExecuting = &currentlyExecutingValue.String
	}
	if executionStartedAtValue.Valid {
		state.ExecutionStartedAt = &executionStartedAtValue.Time
	}

	// Initialize maps if nil (for safety)
	if state.AwaitedRequests == nil {
		state.AwaitedRequests = make(map[string]*AwaitedRequest)
	}
	if state.SubtreeAgents == nil {
		state.SubtreeAgents = make(map[string]*types.SubtreeInfo)
	}
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}
	if state.ExecutionMetadata.RetryCount == nil {
		state.ExecutionMetadata.RetryCount = make(map[string]int)
	}
	if state.ExecutionMetadata.Checkpoints == nil {
		state.ExecutionMetadata.Checkpoints = make(map[string]time.Time)
	}

	return state, nil
}
