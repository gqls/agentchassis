// FILE: platform/orchestration/state.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// OrchestrationStatus represents the current state of a workflow
type OrchestrationStatus string

const (
	StatusInitialized       OrchestrationStatus = "INITIALIZED"
	StatusRunning           OrchestrationStatus = "RUNNING"
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
	RequestID        string    `json:"request_id"`
	OrchestrationID  string    `json:"orchestration_id"`
	CorrelationID    string    `json:"correlation_id"`
	StepID           string    `json:"step_id"`
	StepName         string    `json:"step_name"`
	RetryVersion     int       `json:"retry_version"`
	TargetAgentID    string    `json:"target_agent_id,omitempty"`
	TargetAgentType  string    `json:"target_agent_type"`
	ResponsesTopic   string    `json:"responses_topic"`
	RequestsTopic    string    `json:"requests_topic"`
	SentAt           time.Time `json:"sent_at"`
	TimeoutAt        time.Time `json:"timeout_at"`
	ReplyToRequestID string    `json:"reply_to_request_id,omitempty"`

	// RequestPayload is the exact message that was produced for this request —
	// {"topic":…,"key":…,"headers":{…},"body":…}, i.e. the arguments of the
	// original producer.Produce call. A retry is a REPLAY of it; only
	// retry_version, message_id and timestamp may differ. Without it the
	// coordinator used to rebuild the request out of the AWAITING orchestration's
	// own state, which handed the child the PARENT's orchestration_id and made it
	// decline the work while logging success (bugs_open/129).
	//
	// Deliberately `json:"-"`: this struct is also serialised into
	// orchestration_states.awaited_requests, which is rewritten on every state
	// update. The payload belongs on the per-request row, not the hot one — it is
	// read back from awaited_requests.request_payload at retry time, which is the
	// only moment it is needed.
	RequestPayload json.RawMessage `json:"-"`

	Status      string     `json:"status,omitempty" db:"status"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
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
	SiteID                string `db:"site_id"` // Nullable in DB — empty string means null

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

	StepExecutionCount int      `json:"step_execution_count"`
	RecentSteps        []string `json:"recent_steps,omitempty"`

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

// HasProcessedMessage reports whether this message should be treated as a
// duplicate. bugs_open/003 F3 semantics: a message is a duplicate only if an
// EQUAL-OR-NEWER retry generation has COMPLETED processing, or another live
// pod currently holds a processing lease on it. A 'processing' row whose lease
// has expired is NOT a duplicate — its worker died mid-work, and the
// redelivered copy is how the work gets done. The same-pod exemption exists
// because the agentbase and processor dedupe layers both run on one delivery
// under the same key: the second layer must see the first layer's own live
// claim as "mine", not as a duplicate.
func (r *StateRepository) HasProcessedMessage(ctx context.Context, correlationID, requestID, agentID string, retryVersion int) (bool, error) {
	if requestID == "" {
		// No request_id means NO dedupe at all — a redelivery double-executes
		// with no record anywhere. This used to be silent (bugs_open/003).
		// Loud, not fatal: still process the message.
		r.logger.Warn("DEDUPE_SKIPPED_NO_REQUEST_ID: cannot deduplicate, processing anyway",
			zap.String("correlation_id", correlationID),
			zap.String("agent_id", agentID))
		observability.SystemErrors.WithLabelValues("dedupe", "missing_request_id").Inc()
		return false, nil
	}

	podName := os.Getenv("HOSTNAME")

	query := `
        SELECT EXISTS(
            SELECT 1 FROM processed_messages
            WHERE correlation_id = $1
              AND request_id = $2
              AND agent_id = $3
              AND retry_version >= $4
              AND ( status = 'complete'
                 OR (status = 'processing' AND lease_expires_at > NOW() AND processed_by <> $5) )
        )`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, correlationID, requestID, agentID, retryVersion, podName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check processed request: %w", err)
	}

	return exists, nil
}

// processedMessageLeaseSeconds returns the processing-lease length. It must
// comfortably exceed the worst-case handler runtime: an expired lease makes a
// redelivered copy eligible to take the work over mid-flight, so a too-short
// lease double-executes long handlers. (A rebalance while the lease is live is
// instead resolved by the parent's timeout retry, which arrives with a higher
// retry_version and takes the row over explicitly.)
func processedMessageLeaseSeconds() int {
	if v := os.Getenv("PROCESSED_MESSAGES_LEASE_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 900
}

// RecordMessageProcessing atomically claims the message for processing — the
// first half of the two-phase dedupe (bugs_open/003 F3). The row is inserted
// as status='processing' with a lease; MarkMessageComplete flips it to
// 'complete' after the handler returns. claimed=false means another worker
// owns an equal-or-newer generation right now: treat as duplicate and drop.
// The takeover arbiter is processed_messages_unique (correlation, request,
// agent — WITHOUT retry_version), deliberately: a resend with retry_version+1
// must take over the previous generation's row, not error against it.
func (r *StateRepository) RecordMessageProcessing(ctx context.Context, execCtx *types.ExecutionContext, agentID string) (bool, error) {
	if execCtx.RequestID == "" {
		// Vacuous claim — no dedupe possible; process anyway, loudly
		// (mirrors HasProcessedMessage; both fire so the WARN survives
		// whichever path a caller takes).
		r.logger.Warn("DEDUPE_SKIPPED_NO_REQUEST_ID: recording nothing, processing anyway",
			zap.String("correlation_id", execCtx.CorrelationID),
			zap.String("agent_id", agentID))
		observability.SystemErrors.WithLabelValues("dedupe", "missing_request_id").Inc()
		return true, nil
	}

	// Handle empty orchestration_id - we shouldn't have it in real life but when sending my own requests without them then:
	orchestrationID := execCtx.OrchestrationID
	if orchestrationID == "" {
		orchestrationID = "00000000-0000-0000-0214-000000000010" // NULL UUID orig - 0214
	}

	processingNode := os.Getenv("HOSTNAME")
	if processingNode == "" {
		processingNode = "unknown"
	}

	query := `
        INSERT INTO processed_messages
        (message_id, correlation_id, orchestration_id, request_id, agent_id, message_type, processed_at, processed_by, retry_version, status, lease_expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7, $8, 'processing', NOW() + make_interval(secs => $9))
        ON CONFLICT ON CONSTRAINT processed_messages_unique DO UPDATE
        SET message_id       = EXCLUDED.message_id,
            retry_version    = EXCLUDED.retry_version,
            processed_at     = NOW(),
            processed_by     = EXCLUDED.processed_by,
            status           = 'processing',
            lease_expires_at = EXCLUDED.lease_expires_at
        WHERE processed_messages.retry_version < EXCLUDED.retry_version
           OR (processed_messages.status = 'processing'
               AND (processed_messages.lease_expires_at <= NOW()
                    OR processed_messages.processed_by = EXCLUDED.processed_by))
    `

	res, err := r.db.ExecContext(ctx, query,
		execCtx.MessageID,
		execCtx.CorrelationID,
		orchestrationID,
		execCtx.RequestID,
		agentID,
		execCtx.MessageType,
		processingNode,
		execCtx.RetryVersion,
		processedMessageLeaseSeconds(),
	)

	if err != nil {
		return false, fmt.Errorf("failed to record request processing in state go: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		// Claim outcome unknowable; err on availability (process) — the worst
		// case is a duplicate execution, the same trade the error path makes.
		return true, nil
	}

	return rows == 1, nil
}

// MarkMessageComplete records that processing finished — the second half of
// the two-phase dedupe (bugs_open/003 F3). Wired as a defer immediately after
// a successful claim, so it runs on EVERY in-process disposition (success,
// handled error, validation drop); only pod death skips it, leaving a
// 'processing' lease that expires and permits redelivery to do the work.
// The retry_version guard stops a stale generation's completion clobbering a
// newer takeover that happened mid-flight.
func (r *StateRepository) MarkMessageComplete(ctx context.Context, correlationID, requestID, agentID string, retryVersion int) error {
	if requestID == "" {
		return nil // nothing was recorded
	}

	query := `
        UPDATE processed_messages
        SET status = 'complete',
            lease_expires_at = NULL,
            processed_at = NOW()
        WHERE correlation_id = $1
          AND request_id = $2
          AND agent_id = $3
          AND retry_version = $4
          AND status = 'processing'
    `

	_, err := r.db.ExecContext(ctx, query, correlationID, requestID, agentID, retryVersion)
	if err != nil {
		return fmt.Errorf("failed to mark message complete: %w", err)
	}

	return nil
}

// ReleaseMessageClaim abandons a processing claim WITHOUT completing it, so a
// re-attempt of the same message can claim it afresh.
//
// bugs_open/239: the intake pool re-runs a message whose dispatch resolution
// faulted transiently, but MarkMessageComplete's 'complete' row would make that
// re-run lose the dedupe check against its own earlier attempt (DEDUPE_CLAIM_LOST)
// and the message would be dropped for good. This is the only sanctioned way to
// give a claim back, and it is deliberately narrow: it matches the claiming
// generation (retry_version) and only status='processing', so it can never
// delete a completed record or another worker's claim.
func (r *StateRepository) ReleaseMessageClaim(ctx context.Context, correlationID, requestID, agentID string, retryVersion int) error {
	if requestID == "" {
		return nil // nothing was recorded
	}

	query := `
        DELETE FROM processed_messages
        WHERE correlation_id = $1
          AND request_id = $2
          AND agent_id = $3
          AND retry_version = $4
          AND status = 'processing'
    `

	_, err := r.db.ExecContext(ctx, query, correlationID, requestID, agentID, retryVersion)
	if err != nil {
		return fmt.Errorf("failed to release message claim: %w", err)
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
			currently_executing, last_activity, processing_node,
			site_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
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

	// ── Extract site_id for direct querying ─────────────────────────────
	var siteIDValue interface{}
	for _, path := range []string{"input_data.site_id", "site_id", "input_data.spec.site_id"} {
		if sid := datahelpers.ExtractNestedFieldString(collectedData, path); sid != "" {
			siteIDValue = sid
			break
		}
	}
	if siteIDValue == nil && unmarshalledInitialData != nil {
		if inputData, ok := unmarshalledInitialData["input_data"].(map[string]interface{}); ok {
			if sid, ok := inputData["site_id"].(string); ok && sid != "" {
				siteIDValue = sid
			}
		}
	}

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
		processingNode,    // $27 - processing_node
		siteIDValue)       // $28 - site_id

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
		fuel_budget, error, version, created_at, updated_at,
		site_id
	FROM orchestration_states
	WHERE orchestration_id = $1
`

	state := &OrchestrationState{}
	var collectedDataJSON, workflowPlanJSON, executionMetadataJSON, executionPathJSON []byte
	var awaitedRequestsJSON, awaitedStepsJSON, processingHistoryJSON, subtreeAgentsJSON []byte
	var finalResultValue, errorValue, parentOrchestrationIDNull, currentlyExecutingValue, siteIDValue sql.NullString
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
		&siteIDValue,
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
	if siteIDValue.Valid {
		state.SiteID = siteIDValue.String
	}

	return state, nil
}

// UpdateState updates an existing orchestration state with optimistic locking
func (r *StateRepository) UpdateState(ctx context.Context, state *OrchestrationState) error {
	return r.UpdateStateWithVersion(ctx, state, state.Version)
}

// The jsonb NUL-escape sanitiser lives in datahelpers (shared home, so other
// jsonb writers can reuse it) — see datahelpers.SanitiseJSONBNulEscapes and
// bugs_open/056. Policy council-approved as its own change (corr d8e844ac).

// Size tripwire for collected_data — bugs_open/289, residual (5).
//
// A loop bug grew tool-auditor's collected_data to 22 MB and killed 62 of its 63
// runs, and it ran that way from 2026-07-29 to 2026-08-17 with NOTHING noticing:
// there was no threshold anywhere, so an orchestration carrying a thousand times
// its siblings' payload looked exactly like one carrying a normal one. This is
// the missing instrument, not a fix for that bug.
//
// Thresholds are set from a fleet census taken 2026-08-17, so they sit above real
// traffic rather than at a round number: internal-linker 22 kB, tool-suggester
// 447 kB, page-content-writer 886 kB avg / 1.8 MB max, build-dispatch-loop 2.8 MB
// avg / 14 MB max, tool-auditor 22 MB avg / 29 MB max. 8 MiB is therefore above
// everything healthy and below both pathological cases.
//
// Deliberately fires on EVERY oversized persist rather than once per run. A stuck
// orchestration updates rarely (it is stuck), a healthy one never crosses the
// line, and a tripwire that under-reports to stay quiet is the failure mode this
// exists to end.
const (
	collectedDataWarnBytes  = 8 << 20  // 8 MiB — above all healthy traffic
	collectedDataAlarmBytes = 24 << 20 // 24 MiB — tool-auditor's dead runs sat here
)

// largestCollectedDataKey names the single biggest entry in collected_data.
// Reported alongside the total because the total says "something is wrong" and
// the key says WHERE — on 289 the key name (`create_items_loop_iter_9_done`)
// identified the mechanism immediately, while the total alone did not.
//
// Only called once a run is already over the threshold, so the per-key marshal
// cost is paid on abnormal states and never on the hot path.
func largestCollectedDataKey(collected map[string]interface{}) (string, int) {
	biggestKey, biggestSize := "", 0
	for k, v := range collected {
		encoded, err := json.Marshal(v)
		if err != nil {
			continue
		}
		if len(encoded) > biggestSize {
			biggestKey, biggestSize = k, len(encoded)
		}
	}
	return biggestKey, biggestSize
}

// reportOversizedCollectedData is the tripwire itself. It only ever logs — an
// orchestration is not failed for being large, because the size is a symptom of
// somebody else's defect and killing the run would destroy the evidence.
func (r *StateRepository) reportOversizedCollectedData(state *OrchestrationState, totalBytes int) {
	if totalBytes < collectedDataWarnBytes {
		return
	}
	biggestKey, biggestSize := largestCollectedDataKey(state.CollectedData)
	fields := []zap.Field{
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("owner_agent_type", state.OwnerAgentType),
		zap.String("current_step", state.CurrentStep),
		zap.Int("collected_data_bytes", totalBytes),
		zap.String("largest_key", biggestKey),
		zap.Int("largest_key_bytes", biggestSize),
		zap.Int("collected_data_keys", len(state.CollectedData)),
	}
	if totalBytes >= collectedDataAlarmBytes {
		r.logger.Error("collected_data is past the point where an orchestration can reliably be carried — see bugs_open/289", fields...)
		return
	}
	r.logger.Warn("collected_data is unusually large and growing unchecked — see bugs_open/289", fields...)
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

	// Serialize complex fields. Every jsonb-bound value of this UPDATE — the
	// eight marshalled below plus FinalResult and InitialRequestData — passes
	// through datahelpers.SanitiseJSONBNulEscapes: one stray NUL anywhere in
	// the state otherwise fails the whole UPDATE (22P05) and silently kills the
	// run (bugs_open/056; the substitute-don't-die policy was council-reviewed
	// as its own change, corr d8e844ac). Substitution is deliberately LOUD —
	// the WARN below fires whenever it actually happened.
	awaitedStepsJSON, _ := json.Marshal(state.AwaitedSteps)
	collectedDataJSON, _ := json.Marshal(state.CollectedData)
	workflowPlanJSON, _ := json.Marshal(state.WorkflowPlan)
	executionMetadataJSON, _ := json.Marshal(state.ExecutionMetadata)
	executionPathJSON, _ := json.Marshal(state.ExecutionPath)
	awaitedRequestsJSON, _ := json.Marshal(state.AwaitedRequests)
	processingHistoryJSON, _ := json.Marshal(state.ProcessingHistory)
	subtreeAgentsJSON, _ := json.Marshal(state.SubtreeAgents)
	nulReplaced := 0
	for _, j := range []*[]byte{
		&awaitedStepsJSON, &collectedDataJSON, &workflowPlanJSON,
		&executionMetadataJSON, &executionPathJSON, &awaitedRequestsJSON,
		&processingHistoryJSON, &subtreeAgentsJSON,
	} {
		var n int
		*j, n = datahelpers.SanitiseJSONBNulEscapes(*j)
		nulReplaced += n
	}
	initialRequestData, n := datahelpers.SanitiseJSONBNulEscapes(state.InitialRequestData)
	nulReplaced += n

	// Measured here because the bytes are already in hand — the marshal above is
	// the one this UPDATE writes, so the tripwire costs a length check and never
	// a second serialisation.
	r.reportOversizedCollectedData(state, len(collectedDataJSON))

	// Handle nullable fields
	var finalResultValue, errorValue, currentlyExecutingValue sql.NullString
	if state.FinalResult != nil {
		fr, frN := datahelpers.SanitiseJSONBNulEscapes(state.FinalResult)
		nulReplaced += frN
		finalResultValue = sql.NullString{String: string(fr), Valid: true}
	}
	if nulReplaced > 0 {
		r.logger.Warn("jsonb persist: NUL escape(s) substituted with U+FFFD — content carried a NUL that jsonb cannot store (bugs_open/056)",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.Int("replaced", nulReplaced))
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
		json.RawMessage(initialRequestData),
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

// UpdateStateWithRetry attempts to update state with automatic retry on optimistic lock failures
func (r *StateRepository) UpdateStateWithRetry(ctx context.Context, state *OrchestrationState, maxRetries int) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.UpdateState(ctx, state)
		if err == nil {
			return nil
		}

		// Check if it's an optimistic lock failure
		if strings.Contains(err.Error(), "optimistic lock failure") {
			r.logger.Warn("Optimistic lock failure, retrying",
				zap.Int("attempt", attempt+1),
				zap.Int("max_retries", maxRetries),
				zap.String("orchestration_id", state.OrchestrationID),
				zap.Int("stale_version", state.Version))

			// Reload state to get latest version
			reloaded, reloadErr := r.GetState(ctx, state.OrchestrationID)
			if reloadErr != nil {
				return fmt.Errorf("failed to reload state on retry attempt %d: %w", attempt+1, reloadErr)
			}

			r.logger.Info("Reloaded state for retry",
				zap.Int("attempt", attempt+1),
				zap.Int("new_version", reloaded.Version),
				zap.Int("old_version", state.Version))

			// Update our state reference with the reloaded version
			*state = *reloaded

			// Continue to next attempt
			continue
		}

		// Not an optimistic lock failure - return the error
		return err
	}

	return fmt.Errorf("failed to update state after %d retries due to optimistic lock failures", maxRetries)
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
	r.logger.Info("FindByAwaitedRequestID Searching for orchestration awaiting request",
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

	// DEBUGaa what is in db, where is it being overwritten
	queryDebug := `
		SELECT orchestration_id
		FROM orchestration_states
		WHERE status = 'AWAITING_RESPONSES' ;
	`

	rows := r.db.QueryRowContext(ctx, queryDebug, requestID)
	r.logger.Info("FindByAwaitedRequestID DEBUGaa: what is in db now",
		zap.String("orchestration_id", orchestrationID),
		zap.String("request_id", requestID),
		zap.Any("DEBUGaa: rows", rows),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback to legacy awaited_steps array search
			alternativeQuery := `
				SELECT orchestration_id
				FROM orchestration_states
				WHERE status = 'AWAITING_RESPONSES'
				  AND awaited_steps ? $1
				LIMIT 1
			`

			err = r.db.QueryRowContext(ctx, alternativeQuery, requestID).Scan(&orchestrationID)
			if err != nil {
				r.logger.Error("No state found for awaited request_id",
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
	// INERT, deliberately left visible (bugs_open/075): UpdateStateWithVersion's
	// UPDATE does not list processing_node among its columns, so this assignment
	// never reaches the database. Consequence: processing_node records the pod
	// that CREATED the row, not the pod driving it — which is why an ownership
	// gate built on it could never distinguish a dead owner from a live one.
	// The only writer that actually moves the column after creation is
	// TakeOverOrchestration.
	state.ProcessingNode = os.Getenv("HOSTNAME")

	return r.UpdateState(ctx, state)
}

// TakeOverOrchestration re-stamps processing_node from a named previous holder
// to this pod, returning whether the handover was won (bugs_open/075).
//
// WHY. The stamp is written once at row creation and never refreshed (see
// SetExecutingStep above), so an orchestration whose creating pod has died
// carries a dead pod's name for ever and no living consumer can ever match it.
// ProcessResponse used to DISCARD responses on that mismatch, and a discard is
// permanent: AgentClient.processResponse commits the Kafka offset when
// ProcessResponse returns nil. Once bug-003's F2 retry driver existed, the
// stranding stopped being silent and became a ~3-minute loop that re-executed
// the step for ever, with real external side effects per cycle.
//
// The CAS is guarded on the previous holder so the log can name who it was
// taken from and two pods cannot both claim the same handover. It deliberately
// leaves `version` alone: this is bookkeeping, not a state transition, and must
// never collide with UpdateStateWithVersion's optimistic lock.
//
// INVARIANT — orchestration_states now has TWO guarded-update mechanisms, and
// they must never both govern the same column. UpdateStateWithVersion's
// version-CAS owns every workflow field (status, current_step, collected_data,
// awaited_requests, …); this pod-name CAS owns processing_node and nothing
// else. If a future change makes either write the other's columns, the two
// locks start racing on one field and neither is authoritative — so add fields
// to one side, never to both (council objection, corr 4a227ed9, reuse_agent).
//
// Callers proceed
// whether or not they win — losing means another live pod took the handover
// microseconds ago, which is no reason to throw away a response only this pod
// received.
func (r *StateRepository) TakeOverOrchestration(ctx context.Context, orchestrationID, newPod, previousPod string) (bool, error) {
	query := `
		UPDATE orchestration_states
		SET processing_node = $2,
		    updated_at = NOW()
		WHERE orchestration_id = $1
		  AND processing_node = $3
	`

	result, err := r.db.ExecContext(ctx, query, orchestrationID, newPod, previousPod)
	if err != nil {
		return false, fmt.Errorf("failed to take over orchestration: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to read takeover result: %w", err)
	}

	return rows > 0, nil
}

// ExecuteWithOptimisticLocking executes a function with retry on version conflicts
func (r *StateRepository) ExecuteWithOptimisticLocking(ctx context.Context, orchestrationID string, fn func(*OrchestrationState) error) error {
	maxRetries := 12
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
			// Exponential backoff: 50ms, 100ms, 200ms, 400ms... capped at 2s
			backoff := time.Duration(50*(1<<uint(i))) * time.Millisecond
			if backoff > 2*time.Second {
				backoff = 2 * time.Second
			}
			// Add jitter (0-100ms)
			jitter := time.Duration(rand.Intn(100)) * time.Millisecond

			r.logger.Info("Version conflict, retrying",
				zap.String("orchestration_id", orchestrationID),
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
				zap.Duration("backoff", backoff+jitter))

			time.Sleep(backoff + jitter)
			continue
		}

		return err
	}

	return fmt.Errorf("max retries (%d) exceeded for orchestration %s", maxRetries, orchestrationID)
}

// ClaimStaleOrchestration is the check-and-claim behind handleOrchestrationStatus's
// two takeover arms (bugs_open/329).
//
// THE DEFECT IT CLOSES is not a missing lock — it is a CHECK-THEN-ACT ACROSS TWO
// READS. The arm judged "stuck" from the caller's SNAPSHOT; the write that used to
// follow did its own fresh GetState → mutate → version-CAS and never re-tested the
// predicate, so two takers arriving seconds apart BOTH won, each CASing against the
// version it had just read. (⚠ Note the corollary before writing any test: exactly
// SIMULTANEOUS takers never double-executed — the loser's CAS failed. The
// disconfirming case is the SEQUENTIAL interleaving.)
//
// ⚠ bugs_open/329 and bugs_closed/294 both state that these writes are unversioned
// ("ends in r.UpdateState(...), not UpdateStateWithVersion"). That is FALSE and was
// false when written: UpdateState is a one-line wrapper for UpdateStateWithVersion
// (see above). The version CAS was always there; it was answering a different
// question.
//
// So: re-judge the FRESH row INSIDE the version-CAS and write only if it is still
// stale. The write IS the claim — UpdateStateWithVersion stamps last_activity = now
// and version+1 unconditionally, so the next caller's fresh read sees a row that is
// no longer stale and gets ErrTakeoverLost. A lost version race is re-read by
// ExecuteWithOptimisticLocking and RE-JUDGED; it never retries the claim blind.
//
// WHAT THIS DOES NOT CLOSE, so no caller reads more into it: a live driver versus a
// taker. defaultLocalActionTimeout is 7200s and NOTHING refreshes last_activity
// during a local action, so a driver inside a long step is "stuck" by the 300s clock
// while behaving correctly. A driver holds nothing, and no claim on the takeover
// side can exclude it. This bounds concurrency at 2 (driver + exactly one taker),
// down from unbounded; closing the rest needs a driver heartbeat, a separate seam.
//
// Composed from the version-CAS ONLY: no SQL of its own, and processing_node is
// neither read nor written — that column belongs to TakeOverOrchestration, and the
// two guarded mechanisms must never govern the same column (state_locks_test.go,
// council objection corr 4a227ed9). Deliberately NOT built on TakeOverOrchestration:
// its CAS is `WHERE processing_node = $3` from the OBSERVED value, so where the row
// already carries the acting pod's own name two callers in that pod both match and
// both report rowsAffected = 1 — no exclusion at all.
func (r *StateRepository) ClaimStaleOrchestration(ctx context.Context, orchestrationID string,
	expectStatus OrchestrationStatus, staleAfter time.Duration, claimedBy string) (*OrchestrationState, error) {

	var claimed *OrchestrationState

	err := r.ExecuteWithOptimisticLocking(ctx, orchestrationID, func(fresh *OrchestrationState) error {
		if fresh.Status != expectStatus {
			return ErrTakeoverLost
		}
		if time.Since(fresh.LastActivity) <= staleAfter {
			return ErrTakeoverLost
		}

		idleFor := time.Since(fresh.LastActivity)

		// EXECUTING_STEP additionally clears the executing step — what
		// ClearExecutingStep used to do in a separate, unguarded read-modify-write.
		if expectStatus == StatusExecutingStep {
			if fresh.CurrentlyExecuting == nil {
				return ErrTakeoverLost
			}
			fresh.CurrentlyExecuting = nil
			fresh.Status = StatusRunning
		}

		if fresh.ProcessingHistory == nil {
			fresh.ProcessingHistory = []ProcessingRecord{}
		}
		fresh.ProcessingHistory = append(fresh.ProcessingHistory, ProcessingRecord{
			PodName:   claimedBy,
			StepID:    fresh.CurrentStep,
			StepName:  fresh.CurrentStep,
			Action:    "stale_takeover_claimed",
			Timestamp: time.Now(),
			Details: fmt.Sprintf("idle %s > %s in %s",
				idleFor.Round(time.Second), staleAfter, expectStatus),
		})

		claimed = fresh
		return nil
	})
	if err != nil {
		return nil, err
	}

	// UpdateStateWithVersion bumped Version and LastActivity on this pointer.
	return claimed, nil
}

func (r *StateRepository) GetStateByCorrelation(ctx context.Context, correlationID string) (*OrchestrationState, error) {
	query := `
        SELECT orchestration_id, orchestration_name, correlation_id, owner_agent_id, owner_agent_type, owner_agent_role, parent_orchestration_id,
               client_id, status, current_step, awaited_steps, awaited_requests,
               currently_executing, last_activity, processing_node, execution_started_at,
               collected_data, initial_request_data, final_result, workflow_plan,
               execution_path, execution_metadata, processing_history, subtree_agents,
               fuel_budget, error, version, created_at, updated_at,
               site_id
        FROM orchestration_states
        WHERE correlation_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `

	state := &OrchestrationState{}
	var collectedDataJSON, workflowPlanJSON, executionMetadataJSON, executionPathJSON []byte
	var awaitedRequestsJSON, awaitedStepsJSON, processingHistoryJSON, subtreeAgentsJSON []byte
	var finalResultValue, errorValue, currentlyExecutingValue, siteIDCorr sql.NullString
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
		&siteIDCorr,
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
	if siteIDCorr.Valid {
		state.SiteID = siteIDCorr.String
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

// InsertAwaitedRequest atomically inserts a new awaited request into the table
// Returns error if request_id already exists
func (r *StateRepository) InsertAwaitedRequest(ctx context.Context, req *AwaitedRequest) error {
	query := `
		INSERT INTO awaited_requests (
			request_id, orchestration_id, correlation_id, step_id, step_name,
			retry_version, target_agent_id, target_agent_type,
			responses_topic, requests_topic, sent_at, timeout_at,
			reply_to_request_id, request_payload, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 'waiting')
		ON CONFLICT (request_id) DO NOTHING
	`

	result, err := r.db.ExecContext(ctx, query,
		req.RequestID,
		req.OrchestrationID,
		req.CorrelationID,
		req.StepID,
		req.StepName,
		req.RetryVersion,
		req.TargetAgentID,
		req.TargetAgentType,
		req.ResponsesTopic,
		req.RequestsTopic,
		req.SentAt,
		req.TimeoutAt,
		req.ReplyToRequestID,
		nullableJSON(req.RequestPayload),
	)

	if err != nil {
		return fmt.Errorf("failed to insert awaited request: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check insert result: %w", err)
	}

	if rows == 0 {
		// The row already exists — the "already exists" error below is
		// load-bearing (its caller uses it to decide whether to arm a timeout
		// handler), so the conflict clause stays DO NOTHING rather than becoming
		// DO UPDATE. But spawn_agent pre-registers its row BEFORE it builds and
		// sends the message (preRegisterAwaitedRequest), so on that path this
		// insert is the FIRST time the payload is known. Back-fill it, and only
		// when the row has none — never overwrite a recorded payload, or a
		// retry could replay something other than what was sent.
		if len(req.RequestPayload) > 0 {
			// Check BOTH the error and rows-affected. A guarded UPDATE that
			// matches nothing succeeds with err == nil, so the error alone
			// cannot tell "backed-fill" from "did nothing" — and the only
			// other place this surfaces is RETRY_PAYLOAD_UNAVAILABLE minutes
			// later, by which time the write-time cause is gone. Same shape as
			// 016b §9 "a field assigned in memory before an UPDATE that omits
			// its column is a silent no-op" and "ON CONFLICT DO NOTHING
			// succeeds while inserting nothing".
			res, upErr := r.db.ExecContext(ctx,
				`UPDATE awaited_requests SET request_payload = $2
				  WHERE request_id = $1 AND request_payload IS NULL`,
				req.RequestID, []byte(req.RequestPayload))
			switch {
			case upErr != nil:
				r.logger.Error("RETRY_PAYLOAD_BACKFILL_FAILED: pre-registered awaited request kept no payload — its retry will refuse rather than poison",
					zap.String("request_id", req.RequestID),
					zap.Error(upErr))
			default:
				if n, rErr := res.RowsAffected(); rErr != nil {
					r.logger.Warn("RETRY_PAYLOAD_BACKFILL_UNCOUNTED: driver would not report rows affected",
						zap.String("request_id", req.RequestID), zap.Error(rErr))
				} else if n == 0 {
					// Benign when a payload is already recorded (this insert is
					// a duplicate of one that carried it); a real gap when the
					// row is gone or the column stayed NULL. Say which.
					var present bool
					if qErr := r.db.QueryRowContext(ctx,
						`SELECT request_payload IS NOT NULL FROM awaited_requests WHERE request_id = $1`,
						req.RequestID).Scan(&present); qErr != nil || !present {
						r.logger.Error("RETRY_PAYLOAD_BACKFILL_MISSED: no row took the payload — its retry will refuse rather than poison",
							zap.String("request_id", req.RequestID),
							zap.Bool("payload_present", present),
							zap.Error(qErr))
					}
				}
			}
		}
		return fmt.Errorf("awaited request already exists: %s", req.RequestID)
	}

	r.logger.Info("Inserted awaited request into database",
		zap.String("request_id", req.RequestID),
		zap.String("orchestration_id", req.OrchestrationID),
	)

	return nil
}

// GetAwaitedRequest retrieves an awaited request by request_id
// Returns nil if not found or already processed.
// 'retrying' is included (bugs_open/003 F2): a row claimed by the retry
// driver is still an awaited request. Note the predicate EXCLUDES
// 'processing' — which is why the retry paths must not read here: a claimed
// row is passed through from the claim's own RETURNING instead, because a
// re-read of a row you hold in 'processing' misses by construction
// (bugs_open/216).
func (r *StateRepository) GetAwaitedRequest(ctx context.Context, requestID string) (*AwaitedRequest, error) {
	query := `
		SELECT ` + awaitedRequestColumns + `
		FROM awaited_requests
		WHERE request_id = $1
		  AND status IN ('waiting', 'retrying')
	`

	record, err := scanAwaitedRequestRow(r.db.QueryRowContext(ctx, query, requestID).Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get awaited request: %w", err)
	}

	return record, nil
}

// GetAwaitedRequestWithRetry tries multiple times to find an awaited request
// Handles race condition where INSERT may not be visible yet
// OutstandingAwaitedRequests returns the awaited_requests rows for an
// orchestration that are still outstanding — everything the response consumer
// could still legitimately be waiting on — excluding one request id.
//
// This is the TABLE's answer to "what is outstanding". The advance decision
// (handleCompleteResponse) reads the AwaitedRequests JSONB map instead, and
// nothing reconciles the two: bug 343 (silent post-abandonment freeze)'s wedge is exactly a state where the
// map says "all done" while a row here says otherwise, and the orchestration
// advances past work it still owes. Use this as a CROSS-CHECK, never as a naive
// replacement — the map's optimistic-lock CAS is what serialises two pods racing
// to advance, and a table-only decision reintroduces a mutual back-off where each
// pod sees the other's row 'processing' and neither proceeds.
//
// excludeRequestID is the request being completed by the caller: it legitimately
// sits in 'processing' under the consumer's own claim at that instant, because the
// row is marked complete only after the state save succeeds.
//
// Rows, not a count: detection needs only len() > 0, but enforcement needs the
// rows themselves to re-adopt, and one method means a count and a fetch can never
// drift apart.
func (r *StateRepository) OutstandingAwaitedRequests(ctx context.Context, orchestrationID, excludeRequestID string) ([]*AwaitedRequest, error) {
	query := `
		SELECT ` + awaitedRequestColumns + `
		FROM awaited_requests
		WHERE orchestration_id = $1
		  AND request_id <> $2
		  AND status IN ('waiting', 'processing', 'retrying')
	`

	rows, err := r.db.QueryContext(ctx, query, orchestrationID, excludeRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to query outstanding awaited requests: %w", err)
	}
	defer rows.Close()

	var outstanding []*AwaitedRequest
	for rows.Next() {
		record, scanErr := scanAwaitedRequestRow(rows.Scan)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan outstanding awaited request: %w", scanErr)
		}
		outstanding = append(outstanding, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read outstanding awaited requests: %w", err)
	}

	return outstanding, nil
}

func (r *StateRepository) GetAwaitedRequestWithRetry(ctx context.Context, requestID string, maxRetries int) (*AwaitedRequest, error) {
	retryDelay := 50 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		record, err := r.GetAwaitedRequest(ctx, requestID)
		if err != nil {
			return nil, err
		}

		if record != nil {
			if attempt > 0 {
				r.logger.Info("Found awaited request after retry",
					zap.String("request_id", requestID),
					zap.Int("attempts", attempt+1),
				)
			}
			return record, nil
		}

		// Not found, retry if not last attempt
		if attempt < maxRetries-1 {
			r.logger.Debug("Awaited request not found, retrying",
				zap.String("request_id", requestID),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", retryDelay),
			)
			time.Sleep(retryDelay)
		}
	}

	r.logger.Warn("Awaited request not found after all retries",
		zap.String("request_id", requestID),
		zap.Int("max_retries", maxRetries),
	)
	return nil, nil
}

// awaitedRequestColumns is the canonical column order shared by every claim/get
// of an awaited_requests row in this file, and by scanAwaitedRequestRow. Adding
// a column means adding it HERE and in scanAwaitedRequestRow — the two are one
// contract, and every SELECT below interpolates this constant so they cannot
// drift apart.
const awaitedRequestColumns = `request_id, orchestration_id, correlation_id, step_id, step_name,
		retry_version, target_agent_id, target_agent_type,
		responses_topic, requests_topic, sent_at, timeout_at,
		reply_to_request_id, request_payload, status, processed_at`

// nullableJSON binds a possibly-absent JSON document as SQL NULL rather than as
// an empty string, which a jsonb column rejects.
func nullableJSON(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return []byte(raw)
}

// scanAwaitedRequestRow scans one awaited_requests row in the canonical column
// order shared by every claim/get in this file (awaitedRequestColumns).
func scanAwaitedRequestRow(scan func(dest ...interface{}) error) (*AwaitedRequest, error) {
	record := &AwaitedRequest{}
	var processedAt sql.NullTime
	var requestPayload []byte

	err := scan(
		&record.RequestID,
		&record.OrchestrationID,
		&record.CorrelationID,
		&record.StepID,
		&record.StepName,
		&record.RetryVersion,
		&record.TargetAgentID,
		&record.TargetAgentType,
		&record.ResponsesTopic,
		&record.RequestsTopic,
		&record.SentAt,
		&record.TimeoutAt,
		&record.ReplyToRequestID,
		&requestPayload,
		&record.Status,
		&processedAt,
	)
	if err != nil {
		return nil, err
	}
	if processedAt.Valid {
		record.ProcessedAt = &processedAt.Time
	}
	if len(requestPayload) > 0 {
		record.RequestPayload = json.RawMessage(requestPayload)
	}
	return record, nil
}

// ClaimAwaitedRequestForRetry atomically claims ONE timed-out request for the
// retry driver (bugs_open/003 F2 fast path — the in-process timer). Claims
// from 'waiting' whose timeout has passed (the timer beat the cleanup sweep)
// or from 'expired' (the sweep got there first). Clearing processed_at
// re-opens the late-response path: a response arriving mid-retry hits
// CLAIM_RECOVERY instead of DUPLICATE_SKIPPED. Returns nil if the row is
// gone, already answered, or claimed by another actor — exactly one actor
// drives any given expiry.
func (r *StateRepository) ClaimAwaitedRequestForRetry(ctx context.Context, requestID, podName string) (*AwaitedRequest, error) {
	query := `
		UPDATE awaited_requests
		SET status = 'retrying',
		    processing_started_at = NOW(),
		    processing_pod = $2,
		    processed_at = NULL
		WHERE request_id = $1
		  AND ( (status = 'waiting' AND timeout_at <= NOW()) OR status = 'expired' )
		RETURNING ` + awaitedRequestColumns + `
	`

	record, err := scanAwaitedRequestRow(r.db.QueryRowContext(ctx, query, requestID, podName).Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to claim awaited request for retry: %w", err)
	}
	return record, nil
}

// ClaimExpiredAwaitedRequestsForRetry atomically claims a batch of expired
// requests for the durable retry driver (bugs_open/003 F2 — the per-minute
// ticker in every chassis pod). This is what makes the timeout guarantee
// survive restarts: the in-process timers die with their pod, the DB rows do
// not. FOR UPDATE SKIP LOCKED makes concurrent pods cooperate instead of
// double-driving (the same claim shape site_work_items uses). The join
// confines claims to orchestrations still AWAITING_RESPONSES, so requests of
// completed/failed/reaped orchestrations are never resurrected. The two arms:
// fresh 'expired' rows (processed_at is the expiry stamp; the 60-minute
// window skips the pre-deploy backlog), and stale 'retrying' rows whose
// claiming pod died mid-drive (>5 min; a live drive takes seconds).
func (r *StateRepository) ClaimExpiredAwaitedRequestsForRetry(ctx context.Context, podName string, limit int) ([]*AwaitedRequest, error) {
	query := `
		UPDATE awaited_requests ar
		SET status = 'retrying',
		    processing_started_at = NOW(),
		    processing_pod = $1,
		    processed_at = NULL
		WHERE ar.request_id IN (
		    SELECT a.request_id
		    FROM awaited_requests a
		    JOIN orchestration_states os ON os.orchestration_id = a.orchestration_id
		    WHERE os.status = 'AWAITING_RESPONSES'
		      AND ( (a.status = 'expired'  AND a.processed_at > NOW() - INTERVAL '60 minutes')
		         OR (a.status = 'retrying' AND a.processing_started_at < NOW() - INTERVAL '5 minutes') )
		    ORDER BY a.timeout_at
		    FOR UPDATE OF a SKIP LOCKED
		    LIMIT $2
		)
		RETURNING ` + awaitedRequestColumns + `
	`

	rows, err := r.db.QueryContext(ctx, query, podName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to claim expired awaited requests: %w", err)
	}
	defer rows.Close()

	var claimed []*AwaitedRequest
	for rows.Next() {
		record, scanErr := scanAwaitedRequestRow(rows.Scan)
		if scanErr != nil {
			return claimed, fmt.Errorf("failed to scan claimed awaited request: %w", scanErr)
		}
		claimed = append(claimed, record)
	}
	if err := rows.Err(); err != nil {
		return claimed, fmt.Errorf("failed reading claimed awaited requests: %w", err)
	}
	return claimed, nil
}

// MarkAwaitedRequestFailed terminally releases a claimed row whose retries are
// exhausted (bugs_open/003 F2). Guarded on 'retrying' so a concurrent
// completion by a late response is never clobbered. Setting it FIRST — before
// routing to the error step — is what makes retry exhaustion wedge-proof: no
// downstream failure can leave the row parked in 'retrying'.
func (r *StateRepository) MarkAwaitedRequestFailed(ctx context.Context, requestID string) error {
	query := `
		UPDATE awaited_requests
		SET status = 'error',
		    processed_at = NOW()
		WHERE request_id = $1
		  AND status = 'retrying'
	`
	_, err := r.db.ExecContext(ctx, query, requestID)
	if err != nil {
		return fmt.Errorf("failed to mark awaited request failed: %w", err)
	}
	return nil
}

// CompleteAwaitedRequest marks a claimed request as fully processed
func (r *StateRepository) CompleteAwaitedRequest(ctx context.Context, requestID string) error {
	query := `
		UPDATE awaited_requests
		SET status = 'processed',
		    processed_at = NOW()
		WHERE request_id = $1
		  AND status = 'processing'
	`

	result, err := r.db.ExecContext(ctx, query, requestID)
	if err != nil {
		return fmt.Errorf("failed to complete awaited request: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check complete result: %w", err)
	}

	if rows == 0 {
		r.logger.Warn("CompleteAwaitedRequest: no request was marked complete (unexpected state)",
			zap.String("request_id", requestID),
		)
	} else {
		r.logger.Info("CompleteAwaitedRequest: request marked as processed",
			zap.String("request_id", requestID),
		)
	}

	return nil
}

// UpdateAwaitedRequestForRetry updates the retry version and timeout for a retry attempt
func (r *StateRepository) UpdateAwaitedRequestForRetry(ctx context.Context, requestID string, retryVersion int, timeoutAt time.Time) error {
	query := `
		UPDATE awaited_requests 
		SET retry_version = $2, 
		    timeout_at = $3, 
		    sent_at = NOW(),
		    status = 'waiting'
		WHERE request_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, requestID, retryVersion, timeoutAt)
	if err != nil {
		return fmt.Errorf("failed to update awaited request for retry: %w", err)
	}

	rows, _ := result.RowsAffected()
	r.logger.Debug("Updated awaited request for retry",
		zap.String("request_id", requestID),
		zap.Int("retry_version", retryVersion),
		zap.Int64("rows_affected", rows))

	return nil
}

// MarkAwaitedRequestProcessed marks a request as processed
func (r *StateRepository) oldMarkAwaitedRequestProcessed(ctx context.Context, requestID string) error {
	query := `
		UPDATE awaited_requests
		SET status = 'processed',
		    processed_at = NOW()
		WHERE request_id = $1
		  AND status = 'waiting'
	`

	result, err := r.db.ExecContext(ctx, query, requestID)
	if err != nil {
		return fmt.Errorf("failed to mark awaited request as processed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}

	if rows == 0 {
		r.logger.Warn("No awaited request was marked as processed (already processed or not found)",
			zap.String("request_id", requestID),
		)
	} else {
		r.logger.Info("Marked awaited request as processed",
			zap.String("request_id", requestID),
		)
	}

	return nil
}

// CancelAwaitedRequestsForOrchestration cancels all waiting requests for an orchestration
// Called when orchestration completes or fails.
// 'retrying' and 'expired' are swept too (bugs_open/003 F2): once the
// orchestration is finished, a claimed or expired row must not be resurrected
// by the retry ticker.
func (r *StateRepository) CancelAwaitedRequestsForOrchestration(ctx context.Context, orchestrationID string) error {
	query := `
		UPDATE awaited_requests
		SET status = 'cancelled',
		    processed_at = NOW()
		WHERE orchestration_id = $1
		  AND status IN ('waiting', 'retrying', 'expired')
	`

	result, err := r.db.ExecContext(ctx, query, orchestrationID)
	if err != nil {
		return fmt.Errorf("failed to cancel awaited requests: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check cancel result: %w", err)
	}

	if rows > 0 {
		r.logger.Info("Cancelled awaited requests for orchestration",
			zap.String("orchestration_id", orchestrationID),
			zap.Int64("count", rows),
		)
	}

	return nil
}

// CleanupExpiredAwaitedRequests marks expired requests and deletes old ones
// Should be called periodically (e.g., every minute)
func (r *StateRepository) CleanupExpiredAwaitedRequests(ctx context.Context) (int, error) {
	query := `SELECT cleanup_expired_awaited_requests()`

	var expiredCount int
	err := r.db.QueryRowContext(ctx, query).Scan(&expiredCount)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired awaited requests: %w", err)
	}

	if expiredCount > 0 {
		r.logger.Info("Cleaned up expired awaited requests",
			zap.Int("count", expiredCount),
		)
	}

	return expiredCount, nil
}
