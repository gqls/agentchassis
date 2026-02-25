// FILE: platform/orchestration/coordinator.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/governance"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

const (
	// Topic for notifications to the UI
	NotificationTopic = "system.notifications.ui"
	// Topic for receiving resume commands
	ResumeWorkflowTopic = "system.commands.workflow.resume"
	// Timeout for stuck orchestrations
	StuckOrchestrationTimeout = 5 * time.Minute
	// Default request timeout
	//DefaultRequestTimeout = 180 * time.Second
	// MaxStepExecutions is the circuit breaker limit - if exceeded, abort
	MaxStepExecutions = 200
	// RecentStepsWindow is how many steps to track for cycle detection
	RecentStepsWindow = 15
	// Optimistic lock retry settings for response processing
	maxOptimisticLockRetries     = 15
	optimisticLockBaseRetryDelay = 50 * time.Millisecond
)

var (
	ErrWaitingForResponse = errors.New("orchestration is waiting for responses")
	ErrVersionMismatch    = errors.New("optimistic lock failure: version mismatch")
)

// backoffWithJitter calculates exponential backoff with random jitter.
// Jitter prevents thundering herd problems when multiple processes retry simultaneously.
// Returns a duration between 50% and 100% of the calculated exponential delay.
func backoffWithJitter(baseDelay time.Duration, attempt int) time.Duration {
	// Calculate exponential delay: baseDelay * 2^(attempt-1)
	expDelay := baseDelay * time.Duration(1<<(attempt-1))

	// Add jitter: random value between 50% and 100% of expDelay
	// This spreads out retries so they don't all happen at the same instant
	jitterRange := int64(expDelay / 2)
	if jitterRange <= 0 {
		return expDelay
	}

	jitter := time.Duration(rand.Int63n(jitterRange))
	return expDelay/2 + jitter
}

// SagaCoordinator manages the execution of complex workflows
type SagaCoordinator struct {
	db            *sql.DB
	producer      kafka.Producer
	storageClient storage.Client
	logger        *zap.Logger
	fuelManager   *governance.FuelManager
	tracer        *types.TraceLogger

	// For stateless operation
	isStateless bool
	podName     string

	// Retry tracking
	retryCounters map[string]int
	retryMutex    sync.RWMutex
}

// NewSagaCoordinator creates a new coordinator instance
func NewSagaCoordinator(db *sql.DB, producer kafka.Producer, storageClient storage.Client, logger *zap.Logger) *SagaCoordinator {
	podName := os.Getenv("HOSTNAME")
	if podName == "" {
		podName = fmt.Sprintf("coordinator-local-%d", os.Getpid())
	}

	coordinator := &SagaCoordinator{
		db:            db,
		producer:      producer,
		storageClient: storageClient,
		logger:        logger,
		fuelManager:   governance.NewFuelManager(),
		tracer:        types.NewTraceLogger(logger),
		isStateless:   os.Getenv("ENABLE_STATELESS_MODE") == "true",
		podName:       podName,
	}

	// Start cleanup goroutine
	go coordinator.cleanupExpiredAwaitedRequests()

	return coordinator
}

// ExecuteWorkflow executes a workflow with stateless support
func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
	// fmt.Fprintf(os.Stderr, "DEBUG uuid: ExecuteWorkflow START printf - headers: %+v\n", headers)
	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG uuid: ExecuteWorkflow failed to parse headers printf: %v\n", err)
		// Create minimal context if parsing fails
		execCtx = &types.ExecutionContext{
			CorrelationID:   headers["correlation_id"],
			ClientID:        headers["client_id"],
			MessageID:       headers["message_id"],
			OrchestrationID: headers["orchestration_id"],
			Timestamp:       time.Now(),
		}
	}

	l := s.logger.With(
		zap.String("correlation_id", execCtx.CorrelationID),
		zap.String("message_id", execCtx.MessageID),
		zap.Bool("stateless", s.isStateless),
		zap.String("pod_name", s.podName))

	l.Info("ExecuteWorkflow called coordinator.go 101 PLAN",
		zap.Any("plan", plan),
		zap.String("start_step_", plan.StartStep),
		zap.Int("total_steps", len(plan.Steps)))

	if execCtx.ClientID == "" {
		return fmt.Errorf("client_id is required to execute a workflow")
	}

	// Get or create orchestration state
	state, orchestrationID, isNew, err := s.getOrCreateState(ctx, execCtx, plan, initialData)
	if err != nil {
		l.Error("Failed to get or create state", zap.Error(err))
		return err
	}

	l.Info("Orchestration state retrieved",
		zap.String("orchestration_id", orchestrationID),
		zap.Bool("is_new", isNew),
		zap.String("status", string(state.Status)))

	// Update context with orchestration ID
	execCtx.OrchestrationID = orchestrationID
	headers["orchestration_id"] = orchestrationID

	// Handle based on current status
	return s.handleOrchestrationStatus(ctx, state, execCtx, isNew)
}

func (s *SagaCoordinator) ProcessResponse(ctx context.Context, execCtx *types.ExecutionContext, response types.ResponseMessage) error {
	s.logger.Info("ProcessResponse in coordinator.go",
		zap.Any("response", response),
		zap.Any("DEBUGaa: incoming execCtx", execCtx),
	)

	// Prepare values for tracing
	var inResponseToParentOrchestrationID string
	var inResponseToRequest string

	if execCtx.InResponseTo != nil {
		inResponseToParentOrchestrationID = execCtx.InResponseTo.ParentOrchestrationID
		inResponseToRequest = execCtx.InResponseTo.RequestID
	}

	s.tracer.TraceMessage(execCtx, "RECEIVE_RESPONSE", execCtx.ResponsesTopic,
		map[string]interface{}{
			"consuming_agent_type":       os.Getenv("AGENT_TYPE"),
			"consuming_agent_id":         os.Getenv("AGENT_ID"),
			"response_orch_id":           execCtx.OrchestrationID,
			"in_response_to_parent_orch": inResponseToParentOrchestrationID,
			"in_response_to_request":     inResponseToRequest,
		})

	// Extract request ID from InResponseTo
	var requestID string
	if execCtx.InResponseTo != nil {
		requestID = execCtx.InResponseTo.RequestID
	}
	if requestID == "" {
		requestID = execCtx.RequestID
	}
	if requestID == "" {
		return fmt.Errorf("no request ID in response")
	}

	// Create response preview safely
	responsePreview := createResponsePreview(response.Body.Body)

	s.logger.Info("RESPONSE_RECEIVED: Processing response in coordinator")
	contextLogger := s.logger.With(
		zap.String("request_id", requestID),
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("execCtx.InResponseTo.ParentOrchestrationID", execCtx.InResponseTo.ParentOrchestrationID),
		zap.String("from_agent", execCtx.FromAgentID),
		zap.String("status", execCtx.Status),
		zap.Int("retry_version", execCtx.RetryVersion),
		zap.String("response_preview", responsePreview),
	)

	current, caller := getFuncInfo(1)

	contextLogger.Info("In file coordinator.go ProcessResponse",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	repo := NewStateRepository(s.db, s.logger)

	// =========================================================================
	// ATOMIC CLAIM - This single operation prevents duplicate processing
	// =========================================================================
	contextLogger.Info("Attempting atomic claim for request",
		zap.String("request_id", requestID),
		zap.String("claiming_pod", s.podName))

	// For progress updates, DON'T claim - just update state
	if execCtx.Status == "awaiting" || execCtx.Status == "processing" {
		// Get the awaited request without claiming
		awaitedReq, err := repo.GetAwaitedRequest(ctx, requestID)
		if err != nil || awaitedReq == nil {
			return nil
		}
		state, err := repo.GetState(ctx, awaitedReq.OrchestrationID)
		if err != nil || state == nil {
			return nil
		}
		return s.handleProgressUpdate(ctx, state, execCtx)
	}

	awaitedReq, err := processResponseClaimWithRetry(ctx, repo, requestID, s.podName, contextLogger)
	if err != nil {
		return err
	}
	if awaitedReq == nil {
		return nil // Duplicate or orphaned - already logged
	}

	contextLogger.Info("CLAIM_SUCCESS: exclusively claimed request",
		zap.String("request_id", requestID),
		zap.String("orchestration_id", awaitedReq.OrchestrationID),
		zap.String("step_name", awaitedReq.StepName))

	// Load the orchestration state
	state, err := repo.GetState(ctx, awaitedReq.OrchestrationID)
	if err != nil {
		return fmt.Errorf("failed to load orchestration state: %w", err)
	}

	if state == nil {
		return fmt.Errorf("no state found for orchestration_id=%s", awaitedReq.OrchestrationID)
	}

	// Additional check: verify this orchestrator owns this orchestration
	if state.ProcessingNode != "" && state.ProcessingNode != s.podName {
		contextLogger.Info("Response for orchestration owned by different pod, ignoring",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("owner_pod", state.ProcessingNode),
			zap.String("my_pod", s.podName))
		return nil
	}

	contextLogger.Info("RESPONSE_MATCHED: Found orchestration for response",
		zap.String("state_orch_id", state.OrchestrationID),
		zap.String("state_status", string(state.Status)),
		zap.Int("awaited_requests", len(state.AwaitedRequests)))

	// Handle based on response status
	switch execCtx.Status {
	case "awaiting", "processing":
		return s.handleProgressUpdate(ctx, state, execCtx)
	case "complete", "success":
		return s.handleCompleteResponse(ctx, state, requestID, execCtx, response, awaitedReq)
	case "error_recoverable":
		return s.handleRecoverableError(ctx, state, requestID, execCtx, response)
	case "error_unrecoverable", "failed", "error":
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	default:
		contextLogger.Warn("Unknown response status, treating as unrecoverable error",
			zap.String("status", execCtx.Status),
			zap.String("request_id", requestID),
			zap.String("step_name", awaitedReq.StepName))
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	}
}

func processResponseClaimWithRetry(ctx context.Context, repo *StateRepository, requestID string, podName string, contextLogger *zap.Logger) (*AwaitedRequest, error) {
	// Attempt to claim with retry for race condition
	// (response may arrive before awaited_request is inserted)
	maxRetries := 5

	for attempt := 0; attempt <= maxRetries; attempt++ {
		awaitedReq, err := repo.ClaimAwaitedRequest(ctx, requestID, podName)
		if err != nil {
			contextLogger.Error("Failed to claim awaited request",
				zap.String("request_id", requestID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to claim awaited request: %w", err)
		}

		if awaitedReq != nil {
			// Successfully claimed
			contextLogger.Info("CLAIM_SUCCESS: exclusively claimed request",
				zap.String("request_id", requestID),
				zap.String("orchestration_id", awaitedReq.OrchestrationID),
				zap.String("step_name", awaitedReq.StepName),
				zap.Int("attempt", attempt+1))
			return awaitedReq, nil
		}

		// awaitedReq is nil - could be:
		// 1. Already claimed AND processed by another worker (actual duplicate - skip)
		// 2. Already claimed but NOT processed (failed after claim - allow retry)
		// 3. Not inserted yet (race condition - response arrived before insert)

		if attempt < maxRetries {
			// Check request status to distinguish the cases
			status, processedAt, statusErr := repo.GetAwaitedRequestStatus(ctx, requestID)
			if statusErr != nil {
				contextLogger.Warn("Failed to get awaited request status",
					zap.String("request_id", requestID),
					zap.Error(statusErr))
			}

			if status != "" {
				// Request exists - check if it was actually processed
				if processedAt != nil {
					// Request was claimed AND successfully processed - true duplicate
					contextLogger.Info("DUPLICATE_SKIPPED: request already processed",
						zap.String("request_id", requestID),
						zap.String("status", status),
						zap.Time("processed_at", *processedAt),
						zap.String("my_pod", podName))
					return nil, nil // Safe to skip
				}

				// Request was claimed but NOT processed (processed_at is nil)
				// This means a previous attempt failed after claiming
				// We should allow re-processing by resetting to 'waiting'
				contextLogger.Warn("CLAIM_RECOVERY: request was claimed but not processed, resetting for retry",
					zap.String("request_id", requestID),
					zap.String("current_status", status),
					zap.String("my_pod", podName))

				// Reset the request to 'waiting' so it can be re-claimed
				resetErr := repo.ResetAwaitedRequestForRetry(ctx, requestID)
				if resetErr != nil {
					contextLogger.Error("Failed to reset awaited request for retry",
						zap.String("request_id", requestID),
						zap.Error(resetErr))
					// Continue to next retry attempt anyway
				}
				// Loop will retry the claim
				continue
			}

			// Request doesn't exist yet - wait for it to be inserted (race condition)
			delay := backoffWithJitter(50*time.Millisecond, attempt+1)
			contextLogger.Debug("Awaited request not found yet, retrying",
				zap.String("request_id", requestID),
				zap.Int("attempt", attempt+1),
				zap.Int("max_retries", maxRetries),
				zap.Duration("retry_delay_with_jitter", delay))
			time.Sleep(delay)
		}
	}

	// If still nil after all retries
	contextLogger.Info("RESPONSE_ORPHANED: awaited_request never found after retries",
		zap.String("request_id", requestID),
		zap.String("my_pod", podName),
		zap.Int("total_wait_ms", 50+100+200+400+800)) // ~1.5 seconds total
	return nil, nil
}

// ClaimAwaitedRequest atomically claims a request for processing.
// Returns the AwaitedRequest if successfully claimed, nil if already claimed/processed.
// This prevents the race condition in the original two-step check-then-mark pattern.
func (r *StateRepository) ClaimAwaitedRequest(ctx context.Context, requestID string, claimerPodName string) (*AwaitedRequest, error) {
	// DEBUG BEFORE claim: check current status
	var beforeStatus string
	checkQuery := `SELECT status FROM awaited_requests WHERE request_id = $1`
	r.db.QueryRowContext(ctx, checkQuery, requestID).Scan(&beforeStatus)
	r.logger.Info("ClaimAwaitedRequest: status before claim attempt",
		zap.String("request_id", requestID),
		zap.String("status_before", beforeStatus),
		zap.String("claimer", claimerPodName))

	query := `
    UPDATE awaited_requests
    SET status = 'processing',
        processing_started_at = NOW(),
        processing_pod = $2
    WHERE request_id = $1
      AND status = 'waiting'
    RETURNING 
        request_id, orchestration_id, correlation_id, step_id, step_name,
        retry_version, target_agent_id, target_agent_type,
        responses_topic, requests_topic, sent_at, timeout_at,
        reply_to_request_id, status, processed_at
`

	record := &AwaitedRequest{}
	var processedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, requestID, claimerPodName).Scan(
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
		&record.Status,
		&processedAt,
	)

	if err == sql.ErrNoRows {
		// Not found OR already claimed by another worker - both cases return nil
		r.logger.Debug("ClaimAwaitedRequest: not claimed (already processed or not found)",
			zap.String("request_id", requestID),
		)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to claim awaited request: %w", err)
	}

	if processedAt.Valid {
		record.ProcessedAt = &processedAt.Time
	}

	// DEBUG AFTER claim: log result
	r.logger.Info("ClaimAwaitedRequest: claim result",
		zap.String("request_id", requestID),
		zap.String("status_before", beforeStatus),
		zap.Bool("claimed", record != nil),
		zap.String("claimer", claimerPodName))

	// back to normal logs
	r.logger.Info("ClaimAwaitedRequest: successfully claimed",
		zap.String("request_id", requestID),
		zap.String("orchestration_id", record.OrchestrationID),
	)

	return record, nil
}

// HandleResponse processes responses using headers (compatibility method)
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response types.ResponseMessage) error {
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		s.logger.Error("Failed to create ExecutionContext from headers", zap.Error(err))

		// Try manual extraction for critical fields
		requestID := headers["in_response_to"]
		if requestID == "" {
			requestID = headers["causation_id"] // Legacy fallback
		}
		if requestID == "" {
			return fmt.Errorf("no request ID in headers")
		}

		// Create minimal ExecutionContext
		execCtx = &types.ExecutionContext{
			OrchestrationID: headers["orchestration_id"],
			CorrelationID:   headers["correlation_id"],
			ClientID:        headers["client_id"],
			Status:          headers["status"],
			InResponseTo: &types.ResponseContext{
				RequestID: requestID,
			},
		}
	}
	s.logger.Info("in HandleResponse making execCtx from headers looks old",
		zap.Any("execCtx", execCtx),
		zap.Any("from the headers", headers),
	)

	return s.ProcessResponse(ctx, execCtx, response)
}

// createResponsePreview safely creates a preview string from response body
func createResponsePreview(body interface{}) string {
	if body == nil {
		return "<nil>"
	}

	switch v := body.(type) {
	case string:
		if len(v) > 500 {
			return v[:500] + "..."
		}
		return v

	case map[string]interface{}:
		// Convert map to JSON string for preview
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("<map with %d keys>", len(v))
		}
		preview := string(jsonBytes)
		if len(preview) > 500 {
			return preview[:500] + "..."
		}
		return preview

	case []byte:
		preview := string(v)
		if len(preview) > 500 {
			return preview[:500] + "..."
		}
		return preview

	default:
		// For any other type, use fmt.Sprintf
		preview := fmt.Sprintf("%v", v)
		if len(preview) > 500 {
			return preview[:500] + "..."
		}
		return preview
	}
}

// getOrCreateState gets or creates orchestration state
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, execCtx *types.ExecutionContext, plan models.WorkflowPlan, initialData []byte) (*OrchestrationState, string, bool, error) {
	repo := NewStateRepository(s.db, s.logger)

	// fmt.Fprintf(os.Stderr, "DEBUG uuid: getOrCreateState START printf - orch=%s, parent=%s\n",
	// execCtx.OrchestrationID, execCtx.ParentOrchestrationID)

	// Generate orchestration name if not provided
	orchestrationName := execCtx.OrchestrationName
	if orchestrationName == "" {
		// Generate readable name based on agent type and timestamp
		orchestrationName = fmt.Sprintf("%s-%s-%s",
			s.determineOwnerAgentType(execCtx),
			execCtx.Action,
			time.Now().Format("0102-1504"))
	}

	// Priority 1: Use explicit orchestration_id
	if execCtx.OrchestrationID != "" {
		state, err := repo.GetState(ctx, execCtx.OrchestrationID)
		if err == nil {
			s.logger.Info("Found existing state by orchestration_id",
				zap.String("orchestration_id", execCtx.OrchestrationID))
			return state, execCtx.OrchestrationID, false, nil
		}

		// If we have a parent, this is likely a new child orchestration
		if execCtx.ParentOrchestrationID != "" {
			// Generate name for child orchestration if not provided
			childOrchestrationName := execCtx.OrchestrationName
			if childOrchestrationName == "" {
				childOrchestrationName = fmt.Sprintf("%s-%s-%s",
					s.determineOwnerAgentType(execCtx),
					execCtx.Action,
					time.Now().Format("0102-1504"))
			}

			s.logger.Info("Creating child orchestration",
				zap.String("orchestration_id", execCtx.OrchestrationID),
				zap.String("orchestration_name", childOrchestrationName),
				zap.String("parent_orchestration_id", execCtx.ParentOrchestrationID))

			ownerAgentID := s.determineOwnerAgentID(execCtx)
			ownerAgentType := s.determineOwnerAgentType(execCtx)
			ownerAgentRole := execCtx.Sender.Role

			// When creating initial state, pass both ID and Type
			err := repo.CreateInitialState(ctx, execCtx.OrchestrationID, execCtx.OrchestrationName, execCtx.CorrelationID,
				ownerAgentID, ownerAgentType, ownerAgentRole, execCtx.ParentOrchestrationID, execCtx.ClientID, plan, initialData, execCtx)

			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					state, err = repo.GetState(ctx, execCtx.OrchestrationID)
					if err == nil {
						return state, execCtx.OrchestrationID, false, nil
					}
				}
				return nil, "", false, fmt.Errorf("failed to create child orchestration: %w", err)
			}

			state, err = repo.GetState(ctx, execCtx.OrchestrationID)
			if err != nil {
				return nil, "", false, fmt.Errorf("failed to get newly created state: %w", err)
			}
			return state, execCtx.OrchestrationID, true, nil
		}
	}

	// Priority 2: For root orchestrations, try correlation_id lookup
	if execCtx.ParentOrchestrationID == "" {
		state, err := repo.GetStateByCorrelation(ctx, execCtx.CorrelationID)
		if err == nil {
			s.logger.Info("Found existing root orchestration by correlation_id",
				zap.String("correlation_id", execCtx.CorrelationID))
			return state, state.OrchestrationID, false, nil
		}
	}

	// Priority 3: Create new orchestration
	newOrchestrationID := execCtx.OrchestrationID
	if newOrchestrationID == "" {
		newOrchestrationID = uuid.New().String()
	}

	// Generate name for new orchestration if not provided
	newOrchestrationName := execCtx.OrchestrationName
	if newOrchestrationName == "" {
		newOrchestrationName = fmt.Sprintf("%s-%s-%s",
			s.determineOwnerAgentType(execCtx),
			execCtx.Action,
			time.Now().Format("0102-1504"))
	}

	ownerAgentID := s.determineOwnerAgentID(execCtx)
	ownerAgentType := s.determineOwnerAgentType(execCtx)
	ownerAgentRole := execCtx.Sender.Role

	s.logger.Info("Creating new orchestration",
		zap.String("orchestration_id", newOrchestrationID),
		zap.String("orchestration_name", newOrchestrationName),
		zap.String("correlation_id", execCtx.CorrelationID))

	err := repo.CreateInitialState(ctx, newOrchestrationID, newOrchestrationName, execCtx.CorrelationID,
		ownerAgentID, ownerAgentType, ownerAgentRole, execCtx.ParentOrchestrationID, execCtx.ClientID, plan, initialData, execCtx)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			// Race condition - try to find existing
			if execCtx.ParentOrchestrationID == "" {
				state, err := repo.GetStateByCorrelation(ctx, execCtx.CorrelationID)
				if err == nil {
					return state, state.OrchestrationID, false, nil
				}
			} else {
				state, err := repo.GetState(ctx, newOrchestrationID)
				if err == nil {
					return state, newOrchestrationID, false, nil
				}
			}
		}
		return nil, "", false, fmt.Errorf("failed to create orchestration: %w", err)
	}

	state, err := repo.GetState(ctx, newOrchestrationID)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to get newly created state: %w", err)
	}

	return state, newOrchestrationID, true, nil
}

// handleOrchestrationStatus handles orchestration based on its current status
func (s *SagaCoordinator) handleOrchestrationStatus(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext, isNew bool) error {
	callstack := s.tracer.GetCallStack(12)

	s.logger.Info("Handling orchestration status handleOrchestrationStatus coordinator.go 512",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.Any("state.Status", state.Status),
		zap.Any("state.CurrentlyExecuting", state.CurrentlyExecuting),
		zap.Any("state.CurrentStep", state.CurrentStep),
		zap.Any("call stack handleOrchestrationStatus", callstack),
	)

	repo := NewStateRepository(s.db, s.logger)

	switch state.Status {
	case StatusInitialized:
		// New orchestration, start execution
		s.logger.Info("Status Initialized Starting new orchestration execution in handleOrchestrationStatus",
			zap.String("about to execute step", state.CurrentStep))

		if err := repo.SetExecutingStep(ctx, state.OrchestrationID, state.CurrentStep); err != nil {
			return err
		}

		// Reload state to get updated version
		state, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to reload state: %w", err)
		}

		return s.continueExecution(ctx, state, execCtx)

	case StatusExecutingStep:
		s.logger.Info("Status Executing Step in handleOrchestrationStatus")
		// Check if stuck
		if state.CurrentlyExecuting != nil && time.Since(state.LastActivity) > StuckOrchestrationTimeout {
			s.logger.Warn("Found stuck orchestration, taking over",
				zap.String("stuck_step", *state.CurrentlyExecuting))

			if err := repo.ClearExecutingStep(ctx, state.OrchestrationID); err != nil {
				return err
			}

			state, err := repo.GetState(ctx, state.OrchestrationID)
			if err != nil {
				return fmt.Errorf("failed to reload state: %w", err)
			}

			return s.continueExecution(ctx, state, execCtx)
		}

		s.logger.Info("Orchestration is actively executing")
		return nil

	case StatusAwaitingResponses:
		s.logger.Info("Orchestration is awaiting responses in handleOrchestrationStatus",
			zap.Int("awaited_count", len(state.AwaitedRequests)))
		return ErrWaitingForResponse

	case StatusCompleted:
		s.logger.Info("Workflow already completed")
		return nil

	case StatusFailed:
		s.logger.Info("Workflow previously failed",
			zap.String("error", state.Error))
		return nil

	default:
		return fmt.Errorf("unknown orchestration status: %s", state.Status)
	}
}

// continueExecution executes from the current step
func (s *SagaCoordinator) continueExecution(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext) error {
	s.logger.Info("in continueExecution",
		zap.Any("DEBUGaa: orchestration state", state),
		//zap.Any("Exec context", execCtx),
	)

	repo := NewStateRepository(s.db, s.logger)

	// check for loop
	if err := s.checkCircuitBreaker(state, s.logger); err != nil {
		state.Status = "FAILED"
		state.Error = err.Error()
		if saveErr := repo.UpdateState(ctx, state); saveErr != nil {
			s.logger.Error("Failed to save circuit breaker state", zap.Error(saveErr))
		}
		return err
	}

	// This initial check remains outside the loop. If the function is called
	// on a state that's already waiting, we should exit immediately.
	if state.Status == StatusAwaitingResponses {
		s.logger.Info("Already in waiting state", zap.String("orchestration_id", state.OrchestrationID))
		return nil
	}

	// Main execution loop. It will run for each step in the workflow
	// until a 'return' statement is hit.
	for {
		l := s.logger.With(
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("current_step", state.CurrentStep),
			zap.String("exec_ctx_sender_type", execCtx.Sender.AgentType),
			zap.String("exec_ctx_message_type", execCtx.MessageType),
			zap.Any("exec_ctx_in_response_to", execCtx.InResponseTo),
			zap.Any("state.Status", state.Status),
		)

		l.Info("Continuing workflow execution continueExecution coordinator.go",
			zap.Int("total_steps", len(state.WorkflowPlan.Steps)))

		// Mark the current step as executing for this iteration
		if err := repo.SetExecutingStep(ctx, state.OrchestrationID, state.CurrentStep); err != nil {
			return err
		}

		// Reload the state to ensure we have the latest version for this step
		reloadedState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to reload state: %w", err)
		}
		state = reloadedState
		currentStepName := state.CurrentStep

		l = s.logger.With(
			zap.Any("state loaded from repo - in continueExecution", state),
			zap.String("current_step", state.CurrentStep),
			zap.Any("state.Status", state.Status),
			zap.Any("currentStepConfig", state.WorkflowPlan.Steps[state.CurrentStep]),
		)

		currentStepConfig, exists := state.WorkflowPlan.Steps[state.CurrentStep]
		if !exists {
			l.Error("Current step doesnt exist in the plan", zap.Any("state.WorkflowPlan", state.WorkflowPlan))
			return s.failWorkflow(ctx, state, fmt.Sprintf("step '%s' not found", state.CurrentStep))
		}

		if !s.dependenciesMet(currentStepConfig.Dependencies, state) {
			return s.skipStep(ctx, state, "dependencies not met")
		}

		// Execute the single step for this iteration
		// NOTE: This stores result in state.CollectedData IN MEMORY
		err = s.executeStep(ctx, state, currentStepConfig, execCtx)
		if err != nil {
			// Check if this is a loop iteration with continue_on_error
			if shouldContinueLoopOnError(state, l) {
				if skipErr := s.skipToNextLoopIteration(ctx, state, err.Error(), l); skipErr != nil {
					l.Error("Failed to skip loop iteration", zap.Error(skipErr))
					return s.failWorkflow(ctx, state, fmt.Sprintf("step %s failed: %v (loop skip also failed: %v)", state.CurrentStep, err, skipErr))
				}
				return nil // Loop continues from next iteration
			}
			// Check if the failed step has an error_step to route to
			return s.routeToErrorStepOrFail(ctx, state, state.CurrentStep, fmt.Sprintf("step %s failed: %v", state.CurrentStep, err))
		}

		// Check if a recursive call (from loop expansion) set the status to AWAITING_RESPONSES
		// We need to check the DATABASE status, not our in-memory state, because recursive
		// continueExecution calls operate on their own loaded state objects.
		// However, we must NOT reload the entire state because that would lose the step result
		// that was stored in memory by executeStep.
		dbState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to check state status after step execution: %w", err)
		}
		if dbState.Status == StatusAwaitingResponses {
			// A recursive call set the status - workflow is now waiting
			l.Info("Execution paused - waiting for responses (set by recursive call)",
				zap.String("current_step", dbState.CurrentStep),
				zap.String("orchestration_id", state.OrchestrationID),
				zap.Int("awaited_count", len(dbState.AwaitedRequests)),
			)
			return nil
		}

		// If awaiting responses, state was already persisted in processAwaitResponse
		// Just return - don't save again
		if state.Status == StatusAwaitingResponses {
			l.Info("Execution paused - waiting for responses (state already persisted)",
				zap.String("current_step", state.CurrentStep),
				zap.String("orchestration_id", state.OrchestrationID),
				zap.Int("awaited_count", len(state.AwaitedRequests)),
			)
			return nil
		}

		// For non-awaiting steps, save step result with retry
		if err := s.saveStepResultWithRetry(ctx, state, currentStepName, l); err != nil {
			l.Error("Failed to save step result after retries",
				zap.String("step", currentStepName),
				zap.Error(err))
			return fmt.Errorf("failed to persist step result for '%s': %w", currentStepName, err)
		}

		// --- This is the core step-transition logic ---

		// First, check if the action result specified a next_step override
		nextStep := currentStepConfig.NextStep
		if overrideStep, found := getNextStepFromResult(state.CollectedData, state.CurrentStep); found {
			l.Info("Action result specified next_step override",
				zap.String("configured_next_step", currentStepConfig.NextStep),
				zap.String("override_next_step", overrideStep))
			nextStep = overrideStep
		}

		if nextStep != "" {
			l.Info("Transitioning to next step", zap.String("next_step", nextStep))
			state.CurrentStep = nextStep

			// Advance to next step with retry
			if err := s.advanceToNextStepWithRetry(ctx, state.OrchestrationID, nextStep, l); err != nil {
				return err
			}

			// Reload state for next iteration
			state, err = repo.GetState(ctx, state.OrchestrationID)
			if err != nil {
				return fmt.Errorf("failed to reload state after advancing: %w", err)
			}

			// Go to the top of the for loop to run the next step
			continue
		} else {
			l.Info("No next step defined, completing workflow.")
			// No next step, so complete the workflow and exit the loop
			return s.completeWorkflow(ctx, state)
		}
	}
}

// advanceToNextStepWithRetry updates CurrentStep with optimistic lock retry
func (s *SagaCoordinator) advanceToNextStepWithRetry(ctx context.Context, orchestrationID string, nextStep string, logger *zap.Logger) error {
	repo := NewStateRepository(s.db, s.logger)

	maxRetries := 10
	baseDelay := 50 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Load fresh state
		freshState, err := repo.GetState(ctx, orchestrationID)
		if err != nil {
			return fmt.Errorf("failed to load state for step advance: %w", err)
		}

		// Update current step
		freshState.CurrentStep = nextStep
		freshState.LastActivity = time.Now()

		// Try to save
		err = repo.UpdateState(ctx, freshState)
		if err == nil {
			if attempt > 1 {
				logger.Info("Step advanced after retry",
					zap.String("next_step", nextStep),
					zap.Int("attempts", attempt))
			}
			return nil
		}

		// Check if optimistic lock error
		if !IsOptimisticLockError(err) {
			return fmt.Errorf("failed to advance step: %w", err)
		}

		// Max retries exceeded
		if attempt >= maxRetries {
			return fmt.Errorf("failed to advance step after %d attempts: %w", attempt, err)
		}

		// Backoff with jitter
		delay := backoffWithJitter(baseDelay, attempt)
		logger.Warn("Optimistic lock failure advancing step, retrying",
			zap.String("next_step", nextStep),
			zap.Int("attempt", attempt),
			zap.Duration("backoff", delay))
		time.Sleep(delay)
	}

	return nil
}

// saveStepResultWithRetry saves the step result to CollectedData with optimistic lock retry
// This ensures step results are persisted even when concurrent modifications occur
func (s *SagaCoordinator) saveStepResultWithRetry(ctx context.Context, state *OrchestrationState, stepName string, logger *zap.Logger) error {
	repo := NewStateRepository(s.db, s.logger)

	// Extract the step result from in-memory state
	stepResult := state.CollectedData[stepName]

	// Also check for output_field
	var outputField string
	var outputResult interface{}
	if stepConfig, exists := state.WorkflowPlan.Steps[stepName]; exists {
		if stepConfig.OutputField != "" {
			outputField = stepConfig.OutputField
			outputResult = state.CollectedData[outputField]
		}
	}

	if stepResult == nil && outputResult == nil {
		// No result to save
		logger.Debug("No step result to save", zap.String("step", stepName))
		return nil
	}

	maxRetries := 10
	baseDelay := 50 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Load fresh state
		freshState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to load state for step result save: %w", err)
		}

		// Apply step result to fresh state
		if stepResult != nil {
			freshState.CollectedData[stepName] = stepResult
		}
		if outputField != "" && outputResult != nil {
			freshState.CollectedData[outputField] = outputResult
		}

		// Also preserve the awaited requests if the step set them
		if len(state.AwaitedRequests) > 0 {
			if freshState.AwaitedRequests == nil {
				freshState.AwaitedRequests = make(map[string]*AwaitedRequest)
			}
			for k, v := range state.AwaitedRequests {
				freshState.AwaitedRequests[k] = v
			}
		}

		// Preserve status if it changed (e.g., to StatusAwaitingResponses)
		if state.Status == StatusAwaitingResponses {
			freshState.Status = StatusAwaitingResponses
		}

		// Try to save
		err = repo.UpdateState(ctx, freshState)
		if err == nil {
			if attempt > 1 {
				logger.Info("Step result saved after retry",
					zap.String("step", stepName),
					zap.Int("attempts", attempt))
			}

			// Update our in-memory state to match
			state.Version = freshState.Version
			return nil
		}

		// Check if optimistic lock error
		if !IsOptimisticLockError(err) {
			return fmt.Errorf("failed to save step result: %w", err)
		}

		// Max retries exceeded
		if attempt >= maxRetries {
			return fmt.Errorf("failed to save step result after %d attempts: %w", attempt, err)
		}

		// Backoff with jitter
		delay := backoffWithJitter(baseDelay, attempt)
		logger.Warn("Optimistic lock failure saving step result, retrying",
			zap.String("step", stepName),
			zap.Int("attempt", attempt),
			zap.Duration("backoff", delay))
		time.Sleep(delay)
	}

	return nil
}

// getNextStepFromResult extracts a next_step override from an action result if present.
// Returns the override step name and true if found, or empty string and false if not.
func getNextStepFromResult(collectedData map[string]interface{}, currentStep string) (string, bool) {
	stepResult, ok := collectedData[currentStep]
	if !ok {
		return "", false
	}

	resultMap, ok := stepResult.(map[string]interface{})
	if !ok {
		return "", false
	}

	// Check for next_step (preferred)
	if nextStep, ok := resultMap["next_step"].(string); ok && nextStep != "" {
		return nextStep, true
	}

	// Also check for next_step_override (legacy compatibility)
	if nextStep, ok := resultMap["next_step_override"].(string); ok && nextStep != "" {
		return nextStep, true
	}

	return "", false
}

// executeStep executes a single workflow step
func (s *SagaCoordinator) executeStep(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	s.logger.Info("Executing step in executeStep before executeLocalAction",
		zap.String("step", state.CurrentStep),
		zap.String("action", step.Action),
		zap.Any("step before timeout change", step),
	)

	// Convert config.timeout_seconds to step.Timeout
	datahelpers.ConvertStepTimeout(&step, s.logger)

	s.logger.Info("ExecuteStep before executeLocalAction",
		zap.Any("step aftertimeout change", step),
	)

	datahelpers.LogCollectedDataStructure(state.CollectedData, s.logger, "before_"+state.CurrentStep)

	// Route to appropriate handler
	if isLocalAction(step.Action) {
		return s.executeLocalAction(ctx, state, step, execCtx)
	} else if step.Topic != "" {
		return s.executeRemoteAction(ctx, state, step, execCtx)
	}

	return fmt.Errorf("unknown action: %s", step.Action)
}

// executeLocalAction - Main entry point, orchestrates local action execution
func (s *SagaCoordinator) executeLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	s.logger.Info("just into executeLocalAction look for execCtx before it gets changed",
		zap.String("step", state.CurrentStep),
		zap.String("action", step.Action),
		//zap.Any("DEBUGaa: executeLocalAction execCtx before", execCtx),
		zap.Any("before state", state),
	)

	// 1. Prepare execution context for this step
	// 1a. Ensure execution context is fully populated from state
	ensureFullExecutionContext(execCtx, state, s.podName, s.logger)

	// 1b. Then continue with prepareExecutionContext which handles step-specific fields
	prepareExecutionContext(execCtx, state, step, s.podName)

	// 2. Create contextual logger
	contextLogger := createActionLogger(s.logger, execCtx, state.CurrentStep, step.Action)

	contextLogger.Info("Executing local action in executelocalaction",
		zap.Any("config", step.Config),
		zap.Any("DEBUGaa: executeLocalAction step", step),
		zap.Any("DEBUGaa: executeLocalAction state after", state),
		//zap.Any("DEBUGaa: executeLocalAction execCtx after", execCtx),
	)

	// 3. Handle retry logic for spawn actions
	if shouldRetry := s.handleSpawnRetry(state, step, contextLogger); !shouldRetry {
		return fmt.Errorf("spawn action failed after max retries I think this is a maximum number of spawned agents")
	}

	// 4. Get and validate action handler
	handler, err := getActionHandler(step.Action)
	if err != nil {
		return err
	}

	// 5a. Set loop variable if this is a loop iteration step
	// Must happen before buildActionParams so propagated outputs are included
	s.setLoopVariable(ctx, state, step, contextLogger)

	// 5. Build action parameters - get input data
	params := buildActionParams(ctx, execCtx, state, step, s, contextLogger)

	s.logger.Info("Executing local action",
		zap.Any("DEBUGaa: params sent to action handler", params),
		zap.Any("action in handler", step.Action),
	)

	// 6. Execute the action
	result, err := executeAction(ctx, handler, params, contextLogger)
	if err != nil {
		datahelpers.LogCollectedDataStructure(state.CollectedData, s.logger, "error_"+step.Name)
		return handleActionError(err, step, contextLogger)
	}

	s.logger.Info("Executing local action - result back is: look for request id",
		zap.Any("DEBUGaa: result", result),
		zap.Any("action handler - (back from)", step.Action),
		zap.Any("state awaiting status", state.Status),
		zap.String("orchestration_id", state.OrchestrationID),
		zap.Any("DEBUGaa: state", state),
	)

	// 6a. Check if result is a loop expansion
	if resultMap, ok := result.(map[string]interface{}); ok {
		if isLoop, _ := resultMap["loop_action"].(bool); isLoop {
			contextLogger.Info("Detected loop action, expanding workflow")

			// Handle loop expansion - this injects all iteration steps
			if err := s.handleLoopExpansion(state, resultMap, contextLogger); err != nil {
				return fmt.Errorf("failed to expand loop: %w", err)
			}

			// Loop expansion sets state.CurrentStep to first iteration step
			// Save state immediately
			repo := NewStateRepository(s.db, s.logger)
			if err := repo.UpdateState(ctx, state); err != nil {
				contextLogger.Error("Failed to save state after loop expansion", zap.Error(err))
				return fmt.Errorf("failed to save state after loop expansion: %w", err)
			}

			// Continue workflow with first iteration step
			contextLogger.Info("Loop expanded, continuing to first iteration",
				zap.String("next_step", state.CurrentStep),
			)
			return s.continueExecution(ctx, state, execCtx)
		}
	}

	// 7. Process action result
	if err := processActionResult(state, result, step, execCtx, s, contextLogger); err != nil {
		return err
	}

	// 7a. If awaiting responses, state was already persisted in processAwaitResponse
	// Just return early
	if state.Status == StatusAwaitingResponses {
		contextLogger.Info("Step set up awaited request, state already persisted",
			zap.Int("awaited_count", len(state.AwaitedRequests)))
		return nil
	}

	// 8. Record processing history
	recordActionExecution(state, execCtx, step, s.podName)

	/*// 9. Save state if needed - pass the logger
	return saveStateIfNeeded(ctx, state, s.db, s.logger)*/
	// 9. State saving is now handled by continueExecution with retry logic
	return nil
}

// Prepare the execution context for this step
func prepareExecutionContext(execCtx *types.ExecutionContext, state *OrchestrationState, step models.Step, podName string) {
	// Generate step ID if not present
	if execCtx.StepID == "" {
		execCtx.StepID = uuid.New().String()
	}

	// Set step information
	execCtx.StepName = state.CurrentStep
	execCtx.Action = step.Action
	//execCtx.RequestID = uuid.New().String()
	//execCtx.MessageID = uuid.New().String()
	execCtx.Timestamp = time.Now().UTC()

	// Set reply-to topic from environment
	if parentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC"); parentResponsesTopic != "" {
		execCtx.ReplyToTopic = parentResponsesTopic
	}

	// Ensure sender is populated
	if execCtx.Sender.AgentType == "" {
		execCtx.Sender = types.AgentIdentity{
			AgentType:    state.OwnerAgentType,
			AgentID:      state.OwnerAgentID,
			PodName:      podName,
			AgentVersion: os.Getenv("AGENT_VERSION"),
			Role:         state.OwnerAgentRole,
		}
	}

	// Ensure this is marked as a request
	execCtx.MessageType = "request"

	// Clear any response-specific fields
	execCtx.InResponseTo = nil
	execCtx.Status = ""
	execCtx.IsComplete = false

	// Propagate step timeout to execution context
	// This ensures call_agent and other actions use the configured timeout
	if step.Timeout > 0 {
		execCtx.TimeoutSeconds = int(step.Timeout.Seconds())
	} else if timeout, ok := step.Config["timeout_seconds"].(float64); ok && timeout > 0 {
		execCtx.TimeoutSeconds = int(timeout)
	} else if timeout, ok := step.Config["timeout_seconds"].(int); ok && timeout > 0 {
		execCtx.TimeoutSeconds = timeout
	}
	// If still 0 or negative, leave it for the action to set a default
}

// ensureFullExecutionContext populates missing fields in ExecutionContext from state
// This should be called before executing any action to ensure all required fields are present
func ensureFullExecutionContext(execCtx *types.ExecutionContext, state *OrchestrationState, podName string, logger *zap.Logger) {
	updated := false

	// CorrelationID
	if execCtx.CorrelationID == "" && state.CorrelationID != "" {
		execCtx.CorrelationID = state.CorrelationID
		updated = true
	}

	// OrchestrationID
	if execCtx.OrchestrationID == "" && state.OrchestrationID != "" {
		execCtx.OrchestrationID = state.OrchestrationID
		updated = true
	}

	// OrchestrationName
	if execCtx.OrchestrationName == "" && state.OrchestrationName != "" {
		execCtx.OrchestrationName = state.OrchestrationName
		updated = true
	}

	// ParentOrchestrationID
	if execCtx.ParentOrchestrationID == "" && state.ParentOrchestrationID != "" {
		execCtx.ParentOrchestrationID = state.ParentOrchestrationID
		updated = true
	}

	// ClientID
	if execCtx.ClientID == "" && state.ClientID != "" {
		execCtx.ClientID = state.ClientID
		updated = true
	}

	// ResponsesTopic - try multiple sources
	if execCtx.ResponsesTopic == "" {
		if myResponses, ok := state.CollectedData["__my_responses_topic__"].(string); ok && myResponses != "" {
			execCtx.ResponsesTopic = myResponses
			updated = true
		} else if parentResponses, ok := state.CollectedData["__parent_responses_topic__"].(string); ok && parentResponses != "" {
			execCtx.ResponsesTopic = parentResponses
			updated = true
		} else if os.Getenv("RESPONSES_TOPIC") != "" {
			execCtx.ResponsesTopic = os.Getenv("RESPONSES_TOPIC")
			updated = true
		} else if os.Getenv("PARENT_RESPONSES_TOPIC") != "" {
			execCtx.ResponsesTopic = os.Getenv("PARENT_RESPONSES_TOPIC")
			updated = true
		}
	}

	// Sender
	if execCtx.Sender.AgentType == "" {
		execCtx.Sender = types.AgentIdentity{
			AgentType:    state.OwnerAgentType,
			AgentID:      state.OwnerAgentID,
			PodName:      podName,
			AgentVersion: os.Getenv("AGENT_VERSION"),
			Role:         state.OwnerAgentRole,
		}
		updated = true
	}

	// Basics
	if execCtx.MessageType == "" {
		execCtx.MessageType = "request"
	}
	if execCtx.Version == "" {
		execCtx.Version = "2.0"
	}
	if execCtx.Timestamp.IsZero() {
		execCtx.Timestamp = time.Now()
	}
	if execCtx.FuelBudget == 0 && state.FuelBudget > 0 {
		execCtx.FuelBudget = state.FuelBudget
	}

	if updated {
		logger.Info("ensureFullExecutionContext: populated missing fields",
			zap.String("correlation_id", execCtx.CorrelationID),
			zap.String("orchestration_id", execCtx.OrchestrationID),
			zap.String("client_id", execCtx.ClientID),
			zap.String("responses_topic", execCtx.ResponsesTopic))
	}
}

// Create a contextual logger for the action
func createActionLogger(baseLogger *zap.Logger, execCtx *types.ExecutionContext, stepName, action string) *zap.Logger {
	return baseLogger.With(
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", stepName),
		zap.String("action", action),
		zap.String("step_id", execCtx.StepID))
}

// Handle retry logic for spawn actions
func (s *SagaCoordinator) handleSpawnRetry(state *OrchestrationState, step models.Step, logger *zap.Logger) bool {
	if step.Action != "spawn_agent" {
		return true // Not a spawn action, continue
	}

	retryKey := fmt.Sprintf("spawn_retry_%s_%s", state.OrchestrationID, step.Name)
	retryCount := s.getRetryCount(retryKey)

	if retryCount >= 30 {
		logger.Error("Spawn action exceeded max retries",
			zap.String("step_name", state.CurrentStep),
			zap.Int("retry_count", retryCount))
		return false
	}

	s.incrementRetryCount(retryKey)
	return true
}

// Get and validate action handler
func getActionHandler(action string) (actions.ActionFunc, error) {
	handler, exists := actions.GetAction(action)
	if !exists {
		return nil, fmt.Errorf("unknown action: %s, not found in registry", action)
	}
	return handler, nil
}

// Build action parameters
func buildActionParams(ctx context.Context, execCtx *types.ExecutionContext, state *OrchestrationState,
	step models.Step, coordinator *SagaCoordinator, logger *zap.Logger) actions.ActionParams {

	logger.Info("in buildActionParams",
		zap.Any("DEBUGaa: in buildActionParams state.CollectedData", state.CollectedData), // good here in generic but not in hero
		zap.Any("DEBUGaa: in buildActionParams headers are execCtx.ToHeaders", execCtx.ToHeaders()),
		zap.Any("DEBUGaa: in buildActionParams current step", state.CurrentStep),
	)
	return actions.ActionParams{
		Context:          ctx,
		ExecutionContext: execCtx,
		StepConfig:       step,
		Headers:          execCtx.ToHeaders(),
		CollectedData:    state.CollectedData,
		//SagaCoordinator:  coordinator,
		SagaCoordinator: nil,
		Logger:          logger,
		Producer:        coordinator.producer,
		DB:              coordinator.db,
		StorageClient:   coordinator.storageClient,
		Tracer:          coordinator.tracer,
		CurrentStep:     state.CurrentStep,
	}
}

// Execute the action handler
func executeAction(ctx context.Context, handler actions.ActionFunc, params actions.ActionParams, logger *zap.Logger) (interface{}, error) {
	logger.Info("Calling action handler",
		zap.String("action", params.StepConfig.Action))

	result, err := handler(ctx, params)
	if err != nil {
		return nil, err
	}

	logger.Info("Action handler completed - in executeAction",
		zap.String("action", params.StepConfig.Action),
		zap.Any("DEBUGaa: result", result),
	)

	return result, nil
}

// Handle action execution errors
func handleActionError(err error, step models.Step, logger *zap.Logger) error {
	logger.Error("Local action failed in handleActionError",
		zap.String("step_name", step.Name),
		zap.String("action", step.Action),
		zap.Error(err))

	// Special handling for spawn failures with topic issues
	if step.Action == "spawn_agent" && strings.Contains(err.Error(), "Unknown Topic") {
		logger.Warn("Spawn failed due to missing topic, adding delay for retry")
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("failed to execute action %s: %w", step.Action, err)
}

// Process the result from the action
func processActionResult(state *OrchestrationState, result interface{}, step models.Step,
	execCtx *types.ExecutionContext, coordinator *SagaCoordinator, logger *zap.Logger) error {

	logger.Info("in processActionResult just",
		zap.String("step_name", step.Name),
		zap.Any("step - whats in step", step),
	)

	// Store result in collected data
	if err := storeActionResult(state, result, step, logger); err != nil {
		logger.Info("in processActionResult error when storing action result",
			zap.String("step_name", step.Name),
			zap.Any("result", result),
		)
		return err
	}

	logger.Info("in processActionResult what is result",
		zap.String("step_name", step.Name),
		zap.Any("result", result),
		zap.String("result_type", fmt.Sprintf("%T", result)),
	)

	// Process result by converting it to a map[string]interface{}
	// This robustly handles both map[string]interface{} and struct types
	// Array results (from database queries) don't need spawn/await checking
	var resultMap map[string]interface{}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		logger.Error("Failed to marshal action result", zap.Error(err), zap.Any("result_type", fmt.Sprintf("%T", result)))
		return err // Can't proceed
	}
	if err := json.Unmarshal(jsonBytes, &resultMap); err != nil {
		// Result is not a map (likely an array from database query)
		// Data was already stored by storeActionResult above, so this is OK
		// Arrays don't have await_response or subtree_info, so no further processing needed
		logger.Info("Action result is not a map type - spawn/await checks skipped (data already stored)",
			zap.String("step_name", step.Name),
			zap.String("result_type", fmt.Sprintf("%T", result)))
		return nil // Success - workflow continues
	}

	// Process result based on type
	if resultMap != nil {
		// Handle subtree information (from spawn actions)
		processSubtreeInfo(state, resultMap, logger)

		logger.Info("in processActionResult was ok creating resultMap",
			zap.Any("result map is", resultMap),
		)

		// Check if action requires waiting for response
		// Note: processAwaitResponse now sets StatusAwaitingResponses and persists state
		if needsWaiting := processAwaitResponse(state, resultMap, execCtx, step, coordinator, logger); needsWaiting {
			logger.Info("in processActionResult - action requires waiting (state already persisted)",
				zap.String("step_name", step.Name),
				zap.String("status", string(state.Status)),
			)

			// State needs to wait for response
			// state.Status = StatusAwaitingResponses
			// Status already set by processAwaitResponse, no need to set again
		} else {
			logger.Info("in processActionResult was ok creating resultMap but didnt consider it needsWaiting",
				zap.String("step_name", step.Name),
			)
		}
	} else {
		logger.Info("in processActionResult did not create resultMap successfully",
			zap.String("step_name", step.Name),
		)
	}

	return nil
}

// Store action result in collected data
func storeActionResult(state *OrchestrationState, result interface{}, step models.Step, logger *zap.Logger) error {
	if state.CurrentStep == "" {
		logger.Error("Cannot store result - CurrentStep is empty")
		return fmt.Errorf("current step is empty")
	}

	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}

	state.CollectedData[state.CurrentStep] = result

	// ALSO store under output_field if specified (matching handleCompleteResponse behavior)
	if step.OutputField != "" {
		state.CollectedData[step.OutputField] = result
		logger.Info("Stored action result under output_field",
			zap.Any("step", step),
			zap.String("output_field", step.OutputField),
		)
	}

	logger.Info("Stored action result in storeActionResult",
		zap.String("step", state.CurrentStep),
		zap.Any("result_keys", getMapKeys(result)),
		zap.Any("DEBUGaa: result_value", result),
		zap.Any("DEBUGaa: state look at collected data", result),
	)

	return nil
}

// Process subtree information from spawn actions
func processSubtreeInfo(state *OrchestrationState, result map[string]interface{}, logger *zap.Logger) {
	subtreeInfo, ok := result["subtree_info"].(*types.SubtreeInfo)
	if !ok {
		return
	}

	if state.SubtreeAgents == nil {
		state.SubtreeAgents = make(map[string]*types.SubtreeInfo)
	}

	state.SubtreeAgents[subtreeInfo.AgentID] = subtreeInfo

	logger.Info("Added agent to subtree",
		zap.String("agent_id", subtreeInfo.AgentID),
		zap.String("agent_type", subtreeInfo.AgentType),
		zap.String("agent_name", subtreeInfo.AgentName))
}

// Process await_response flag and setup waiting if needed
func processAwaitResponse(state *OrchestrationState, result map[string]interface{},
	execCtx *types.ExecutionContext, step models.Step, coordinator *SagaCoordinator, logger *zap.Logger) bool {

	logger.Info("in processAwaitResponse",
		zap.String("step_name", step.Name),
		zap.Any("what type was await response - it is true but being seen as false perhaps?", result["await_response"]),
		zap.Any("how deep is the await response key - keyed by action I think. result is:", result),
		zap.String("result[await_response]_type in processAwaitResponse", fmt.Sprintf("%T", result["await_response"])),
	)

	// Check if action requires waiting
	awaitResponse, ok := result["await_response"].(bool)
	logger.Info("in processAwaitResponse",
		zap.Any("what is await response is it true or false", result["await_response"]),
	)

	// extra checking
	awaitVal, ok := result["await_response"]
	if !ok {
		return false // Key doesn't exist
	}

	switch v := awaitVal.(type) {
	case bool:
		logger.Info("in processAwaitResponse all good? recognised as bool",
			zap.Any("what is await response", awaitResponse),
		)
		awaitResponse = v
	case string:
		logger.Info("in processAwaitResponse all good? recognised as string",
			zap.Any("what is await response", awaitResponse),
		)
		awaitResponse = (v == "true" || v == "True")
	default:
		// Don't know how to handle this type, assume false
		logger.Warn("Unknown type for 'await_response' in action result",
			zap.String("step_name", step.Name),
			zap.Any("value", awaitVal))
		return false
	}

	if !awaitResponse {
		return false // Value was false or not "true"
	}

	// Extract request ID
	requestID := extractRequestID(result, execCtx)
	if requestID == "" {
		logger.Error("No request ID for awaited response")
		return false
	}

	// Determine response topic
	responsesTopic := determineResponsesTopic(result, execCtx, logger)
	if responsesTopic == "" {
		logger.Error("No responses topic available for awaited request",
			zap.String("request_id", requestID))
		return false
	}

	// Determine requests topic (where to send the request)
	requestsTopic := determineRequestsTopic(result, execCtx, logger)
	if requestsTopic == "" {
		logger.Warn("No requests topic available for awaited request - retry will fail",
			zap.String("request_id", requestID))
	}

	// Create awaited request entry
	awaitedReq := createAwaitedRequest(requestID, execCtx, state, step, result, responsesTopic, requestsTopic)

	// Add to state FIRST (in memory)
	if state.AwaitedRequests == nil {
		state.AwaitedRequests = make(map[string]*AwaitedRequest)
	}
	state.AwaitedRequests[requestID] = awaitedReq

	// Set status to awaiting (must happen before persist)
	state.Status = StatusAwaitingResponses

	// PERSIST STATE BEFORE TABLE INSERT to prevent race condition
	// The JSONB must be persisted before the table insert, otherwise:
	// 1. Response arrives after table insert but before JSONB persist
	// 2. Response handler loads state with empty AwaitedRequests
	// 3. Workflow incorrectly thinks all responses are in
	ctx := context.Background()
	repo := NewStateRepository(coordinator.db, logger)
	if err := persistAwaitingStateWithRetry(ctx, state, repo, logger); err != nil {
		logger.Error("Failed to persist state before table insert",
			zap.String("request_id", requestID),
			zap.Error(err))
		// Remove from in-memory state since we failed
		delete(state.AwaitedRequests, requestID)
		state.Status = "" // Reset status
		return false
	}

	// NOW insert into database table (JSONB is already persisted)
	err := repo.InsertAwaitedRequest(ctx, awaitedReq)
	alreadyExisted := false
	if err != nil {
		// Check if this is an "already exists" error - this is actually OK
		// It means a previous attempt (optimistic lock retry) already inserted the row
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("Awaited request already exists in table (previous attempt succeeded)",
				zap.String("request_id", requestID),
				zap.String("orchestration_id", state.OrchestrationID))
			alreadyExisted = true
			// Continue as success - state is persisted, timeout handler was started by previous attempt
		} else {
			logger.Error("Failed to insert awaited request into database",
				zap.String("request_id", requestID),
				zap.Error(err))
			// State is already persisted with the awaited request, which is fine
			// The request will eventually timeout if table insert keeps failing
			// But we should try a few times
			for retry := 0; retry < 3; retry++ {
				time.Sleep(50 * time.Millisecond)
				if err = repo.InsertAwaitedRequest(ctx, awaitedReq); err == nil {
					break
				}
				if strings.Contains(err.Error(), "already exists") {
					// Another goroutine succeeded
					logger.Info("Awaited request already exists (concurrent insert succeeded)",
						zap.String("request_id", requestID))
					alreadyExisted = true
					err = nil
					break
				}
			}
			if err != nil {
				return false
			}
		}
	}

	logger.Info("Action requires waiting for response - Added awaited request (state persisted first)",
		zap.String("request_id", requestID),
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("target_agent_type", awaitedReq.TargetAgentType),
		zap.String("target_agent_id", awaitedReq.TargetAgentID),
		zap.String("responses_topic", responsesTopic),
		zap.Int("total_awaited", len(state.AwaitedRequests)))

	// Setup timeout handler only if we actually created the row (not if it already existed)
	if !alreadyExisted {
		go coordinator.handleRequestTimeout(context.Background(), state.OrchestrationID, requestID, awaitedReq.TimeoutAt)
	}

	return true
}

// persistAwaitingStateWithRetry saves state with awaited request before table insert
// This prevents the race condition where responses arrive before JSONB is persisted
func persistAwaitingStateWithRetry(ctx context.Context, state *OrchestrationState, repo *StateRepository, logger *zap.Logger) error {
	maxRetries := 10
	baseDelay := 50 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Load fresh state
		freshState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}

		// Check if response already arrived (in case of retry loops)
		for reqID := range state.AwaitedRequests {
			if existingData, exists := freshState.CollectedData[state.AwaitedRequests[reqID].StepName].(map[string]interface{}); exists {
				if _, hasResponse := existingData["response"]; hasResponse {
					logger.Info("Response already arrived during state persist - continuing",
						zap.String("request_id", reqID))
					return nil
				}
			}
		}

		// Apply awaited requests to fresh state
		if freshState.AwaitedRequests == nil {
			freshState.AwaitedRequests = make(map[string]*AwaitedRequest)
		}
		for k, v := range state.AwaitedRequests {
			freshState.AwaitedRequests[k] = v
		}

		// Apply status
		freshState.Status = StatusAwaitingResponses
		freshState.LastActivity = time.Now()

		// Try to save
		err = repo.UpdateState(ctx, freshState)
		if err == nil {
			if attempt > 1 {
				logger.Info("Awaiting state persisted after retry",
					zap.Int("attempts", attempt),
					zap.Int("awaited_count", len(freshState.AwaitedRequests)))
			}
			// Update in-memory version
			state.Version = freshState.Version
			return nil
		}

		// Check if optimistic lock error
		if !IsOptimisticLockError(err) {
			return fmt.Errorf("failed to persist awaiting state: %w", err)
		}

		if attempt >= maxRetries {
			return fmt.Errorf("failed to persist awaiting state after %d attempts: %w", attempt, err)
		}

		// Backoff
		delay := backoffWithJitter(baseDelay, attempt)
		logger.Warn("Optimistic lock failure persisting awaiting state, retrying",
			zap.Int("attempt", attempt),
			zap.Duration("backoff", delay))
		time.Sleep(delay)
	}

	return nil
}

// Extract request ID from result or context
func extractRequestID(result map[string]interface{}, execCtx *types.ExecutionContext) string {
	// Try to get from result first
	if reqID, ok := result["request_id"].(string); ok && reqID != "" {
		return reqID
	}
	// Fall back to execution context
	return execCtx.RequestID
}

// Determine the responses topic for waiting
func determineResponsesTopic(result map[string]interface{}, execCtx *types.ExecutionContext, logger *zap.Logger) string {
	// Priority order:
	// 1. Environment variable (most reliable)
	if topic := os.Getenv("RESPONSES_TOPIC"); topic != "" {
		logger.Info("Using RESPONSES_TOPIC from environment",
			zap.String("topic", topic))
		return topic
	}

	// 2. Result from action
	if topic, ok := result["responses_topic"].(string); ok && topic != "" {
		logger.Info("Using responses_topic from action result",
			zap.String("topic", topic))
		return topic
	}

	// 3. Execution context
	if execCtx.ResponsesTopic != "" {
		logger.Info("Using ResponsesTopic from execution context",
			zap.String("topic", execCtx.ResponsesTopic))
		return execCtx.ResponsesTopic
	}

	return ""
}

func determineRequestsTopic(result map[string]interface{}, execCtx *types.ExecutionContext, logger *zap.Logger) string {
	// Priority order:
	// 1. Result from action (spawn returns this)
	if topic, ok := result["requests_topic"].(string); ok && topic != "" {
		logger.Info("Using requests_topic from action result",
			zap.String("topic", topic))
		return topic
	}

	// 2. Check if result has topics map (from spawn)
	if topics, ok := result["topics"].(map[string]interface{}); ok {
		if topic, ok := topics["requests"].(string); ok && topic != "" {
			logger.Info("Using requests topic from result.topics",
				zap.String("topic", topic))
			return topic
		}
	}

	// 3. Environment variable (fallback)
	if topic := os.Getenv("REQUESTS_TOPIC"); topic != "" {
		logger.Info("Using REQUESTS_TOPIC from environment",
			zap.String("topic", topic))
		return topic
	}

	// 4. Execution context
	if execCtx.RequestsTopic != "" {
		logger.Info("Using RequestsTopic from execution context",
			zap.String("topic", execCtx.RequestsTopic))
		return execCtx.RequestsTopic
	}

	return ""
}

// Create an awaited request entry
func createAwaitedRequest(requestID string, execCtx *types.ExecutionContext, state *OrchestrationState,
	step models.Step, result map[string]interface{}, responsesTopic string, requestsTopic string) *AwaitedRequest {

	return &AwaitedRequest{
		RequestID:       requestID,
		OrchestrationID: state.OrchestrationID,
		CorrelationID:   state.CorrelationID,
		StepID:          execCtx.StepID,
		StepName:        state.CurrentStep,
		RetryVersion:    0,
		TargetAgentType: extractTargetAgentType(step, result),
		TargetAgentID:   extractTargetAgentID(result),
		ResponsesTopic:  responsesTopic,
		RequestsTopic:   requestsTopic,
		SentAt:          time.Now(),
		TimeoutAt:       time.Now().Add(getTimeout(step)),
	}
}

// Extract target agent type from step or result
func extractTargetAgentType(step models.Step, result map[string]interface{}) string {
	// Try result first
	if agentType, ok := result["target_agent_type"].(string); ok && agentType != "" {
		return agentType
	}
	// Try step config
	if agentType, ok := step.Config["agent_type"].(string); ok && agentType != "" {
		return agentType
	}
	return "unknown"
}

// Extract target agent ID from result
func extractTargetAgentID(result map[string]interface{}) string {
	// Try multiple possible keys
	keys := []string{"agent_id", "target_agent_id", "agent_called"}
	for _, key := range keys {
		if id, ok := result[key].(string); ok && id != "" {
			return id
		}
	}
	return ""
}

// Record the action execution in processing history
func recordActionExecution(state *OrchestrationState, execCtx *types.ExecutionContext, step models.Step, podName string) {
	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   podName,
		StepID:    execCtx.StepID,
		StepName:  state.CurrentStep,
		Action:    "executed",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Action %s executed by %s", step.Action, podName),
	})
}

// Save state if it needs to be persisted (e.g., waiting for responses)
func saveStateIfNeeded(ctx context.Context, state *OrchestrationState, db *sql.DB, logger *zap.Logger) error {
	if state.Status != StatusAwaitingResponses {
		// State will be saved by continueExecution
		return nil
	}

	// We're waiting for responses, need to save now
	repo := NewStateRepository(db, logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to save state before waiting: %w", err)
	}

	return nil
}

// executeRemoteAction sends work to another agent
func (s *SagaCoordinator) executeRemoteAction(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	contextLogger := s.logger.With(
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", state.CurrentStep),
		zap.String("action", step.Action))

	contextLogger.Info("Executing remote action",
		zap.String("topic", step.Topic),
		zap.String("target_agent_type", step.TargetAgentType))

	// Update execution context
	execCtx.RequestID = uuid.New().String()
	execCtx.StepID = uuid.New().String()
	execCtx.StepName = state.CurrentStep
	execCtx.Action = step.Action
	execCtx.ToAgentType = step.TargetAgentType

	// ResponsesTopic should already be set in execCtx from the process() function
	if execCtx.ResponsesTopic == "" {
		contextLogger.Error("No responses topic configured",
			zap.String("step", state.CurrentStep))
		return fmt.Errorf("no responses topic configured for remote action")
	}

	// Prepare the message
	requestMsg := &types.RequestMessage{
		Headers: execCtx.ToRequestHeaders(),
		Body: models.TaskRequest{
			Action: step.Action,
			Data:   state.CollectedData,
		},
	}

	msgBytes, err := json.Marshal(requestMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send the message
	if err := s.producer.Produce(ctx, step.Topic, requestMsg.Headers.ToMap(),
		[]byte(execCtx.CorrelationID), msgBytes); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	contextLogger.Info("Remote request sent",
		zap.String("request_id", execCtx.RequestID),
		zap.String("topic", step.Topic),
		zap.String("responses_topic", execCtx.ResponsesTopic))

	// Create awaited request
	awaitedReq := &AwaitedRequest{
		RequestID:       execCtx.RequestID,
		StepID:          execCtx.StepID,
		StepName:        state.CurrentStep,
		RetryVersion:    0,
		TargetAgentType: step.TargetAgentType,
		ResponsesTopic:  execCtx.ResponsesTopic,
		RequestsTopic:   step.Topic, // Where we sent the request
		SentAt:          time.Now(),
		TimeoutAt:       time.Now().Add(getTimeout(step)),
	}

	// Update state
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.AddAwaitedRequest(ctx, state.OrchestrationID, awaitedReq); err != nil {
		return err
	}

	state.Status = StatusAwaitingResponses
	state.CurrentStep = step.NextStep

	if err := repo.UpdateState(ctx, state); err != nil {
		return err
	}

	contextLogger.Info("Remote action initiated, waiting for response",
		zap.String("request_id", execCtx.RequestID),
		zap.String("awaiting_on_topic", execCtx.ResponsesTopic))

	// Set up timeout
	go s.handleRequestTimeout(ctx, state.OrchestrationID, execCtx.RequestID, awaitedReq.TimeoutAt)

	return nil
}

func (s *SagaCoordinator) getResponsesTopicFromResult(result map[string]interface{}, execCtx *types.ExecutionContext) string {
	// First check if the action result includes a responses_topic
	// This would be set by spawn_agent or call_agent actions
	if topic, ok := result["responses_topic"].(string); ok && topic != "" {
		return topic
	}

	// For child agents, they might include the child's response topic
	if topic, ok := result["child_responses_topic"].(string); ok && topic != "" {
		return topic
	}

	// Otherwise use the execution context's topic
	// This is where THIS orchestration expects to receive responses
	if execCtx.ResponsesTopic != "" {
		return execCtx.ResponsesTopic
	}

	// This shouldn't happen with proper setup
	s.logger.Error("No responses topic available",
		zap.Any("result_keys", getMapKeys(result)),
		zap.String("exec_responses_topic", execCtx.ResponsesTopic))
	return ""
}

func (s *SagaCoordinator) getTargetAgentID(result map[string]interface{}) string {
	if agentID, ok := result["agent_id"].(string); ok {
		return agentID
	}
	if agentID, ok := result["target_agent_id"].(string); ok {
		return agentID
	}
	if agentID, ok := result["agent_called"].(string); ok {
		return agentID
	}
	return ""
}

func getMapKeys(m interface{}) []string {
	if m == nil {
		return []string{}
	}

	switch v := m.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys
	default:
		return []string{}
	}
}

// handleProgressUpdate handles progress updates from agents
func (s *SagaCoordinator) handleProgressUpdate(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext) error {
	s.logger.Info("In handleProgressUpdate. Progress update received",
		zap.String("status", execCtx.Status),
		zap.String("from_agent", execCtx.Sender.AgentID))

	repo := NewStateRepository(s.db, s.logger)

	maxRetries := 5
	baseDelay := 25 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Load fresh state each attempt
		freshState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to load state for progress update: %w", err)
		}

		// Add processing record to fresh state
		freshState.ProcessingHistory = append(freshState.ProcessingHistory, ProcessingRecord{
			PodName:   s.podName,
			StepID:    execCtx.StepID,
			StepName:  execCtx.StepName,
			Action:    fmt.Sprintf("progress_%s", execCtx.Status),
			Timestamp: time.Now(),
			Details:   fmt.Sprintf("Progress from %s", execCtx.Sender.AgentID),
		})

		err = repo.UpdateState(ctx, freshState)
		if err == nil {
			if attempt > 1 {
				s.logger.Info("Progress update saved after retry",
					zap.Int("attempts", attempt),
					zap.String("orchestration_id", state.OrchestrationID))
			}
			return nil
		}

		if !IsOptimisticLockError(err) {
			return fmt.Errorf("failed to save progress update: %w", err)
		}

		if attempt >= maxRetries {
			s.logger.Warn("Progress update failed after max retries, dropping update",
				zap.Int("attempts", attempt),
				zap.String("orchestration_id", state.OrchestrationID),
				zap.Error(err))
			// Progress updates are non-critical, so we can drop them rather than fail
			return nil
		}

		delay := backoffWithJitter(baseDelay, attempt)
		s.logger.Debug("Optimistic lock failure in progress update, retrying",
			zap.Int("attempt", attempt),
			zap.Duration("backoff", delay))
		time.Sleep(delay)
	}

	return nil
}

// handleCompleteResponse processes a successful response
func (s *SagaCoordinator) handleCompleteResponse(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response types.ResponseMessage, awaitedReq *AwaitedRequest) error {
	s.logger.Info("Processing complete response",
		zap.String("step_name", awaitedReq.StepName),
		zap.String("request_id", requestID))

	// 1. Parse response body
	normalisedData, err := s.parseResponseBody(response)
	if err != nil {
		return err
	}

	stepName := awaitedReq.StepName
	step, stepExists := state.WorkflowPlan.Steps[stepName]

	// 2. Create repo (but DON'T mark complete yet - wait for successful state save)
	repo := NewStateRepository(s.db, s.logger)

	// 3. Save response data with retry loop (with exponential backoff)
	maxRetries := 15
	baseDelay := 50 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		freshState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}

		// Apply response to collected data
		s.applyResponseToState(freshState, stepName, step, stepExists, normalisedData, awaitedReq)

		// Remove from awaited requests
		delete(freshState.AwaitedRequests, requestID)

		// Check if this was the last awaited response
		allDone := len(freshState.AwaitedRequests) == 0
		if allDone {
			// Advance to next step
			if currentStep, exists := freshState.WorkflowPlan.Steps[freshState.CurrentStep]; exists && currentStep.NextStep != "" {
				freshState.CurrentStep = currentStep.NextStep
			}
			freshState.Status = StatusExecutingStep
			freshState.LastActivity = time.Now()
		}

		// Save state
		err = repo.UpdateState(ctx, freshState)
		if err == nil {
			// SUCCESS - NOW mark the awaited request as complete
			// This is the key fix: only mark complete AFTER state is saved
			if markErr := repo.MarkAwaitedRequestComplete(ctx, requestID); markErr != nil {
				s.logger.Warn("Failed to mark awaited request complete", zap.Error(markErr))
				// Don't fail the whole operation - state was saved successfully
			}

			if attempt > 1 {
				s.logger.Info("Response saved after retry",
					zap.Int("attempts", attempt))
			}
			if allDone {
				s.logger.Info("All responses received, continuing workflow")
				freshExecCtx := s.createContinuationContext(freshState)
				return s.continueExecution(ctx, freshState, freshExecCtx)
			}
			s.logger.Info("Response saved, still awaiting more responses",
				zap.Int("remaining", len(freshState.AwaitedRequests)))
			return nil
		}

		// Check if it's an optimistic lock error
		if !IsOptimisticLockError(err) {
			return fmt.Errorf("failed to save response data: %w", err)
		}

		// Max retries exceeded
		if attempt >= maxRetries {
			return fmt.Errorf("failed to save response after %d attempts (optimistic lock): %w", attempt, err)
		}

		// Exponential backoff with jitter to prevent thundering herd
		delay := backoffWithJitter(baseDelay, attempt)
		s.logger.Warn("Optimistic lock failure in response processing, retrying",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", maxRetries),
			zap.Duration("backoff_with_jitter", delay))
		time.Sleep(delay)
	}

	return nil
}

// retryStateUpdate retries a state update operation when optimistic lock failures occur.
// This is used during response processing which can race with loop execution.
func (s *SagaCoordinator) retryStateUpdate(
	ctx context.Context,
	orchestrationID string,
	operationName string,
	applyChanges func(state *OrchestrationState) error,
	logger *zap.Logger,
) error {
	repo := NewStateRepository(s.db, s.logger)

	for attempt := 1; attempt <= maxOptimisticLockRetries; attempt++ {
		// Load fresh state from DB
		state, err := repo.GetState(ctx, orchestrationID)
		if err != nil {
			return fmt.Errorf("failed to load state for retry: %w", err)
		}

		// Apply the changes
		if err := applyChanges(state); err != nil {
			return fmt.Errorf("failed to apply changes: %w", err)
		}

		// Try to save
		if err := repo.UpdateState(ctx, state); err != nil {
			if IsOptimisticLockError(err) && attempt < maxOptimisticLockRetries {
				delay := backoffWithJitter(optimisticLockBaseRetryDelay, attempt)
				logger.Warn("Optimistic lock failure, retrying state update",
					zap.String("operation", operationName),
					zap.String("orchestration_id", orchestrationID),
					zap.Int("attempt", attempt),
					zap.Int("max_attempts", maxOptimisticLockRetries),
					zap.Duration("backoff_with_jitter", delay))

				time.Sleep(delay)
				continue
			}
			return fmt.Errorf("state update failed after %d attempts: %w", attempt, err)
		}

		if attempt > 1 {
			logger.Info("State update succeeded after retry",
				zap.String("operation", operationName),
				zap.String("orchestration_id", orchestrationID),
				zap.Int("attempts", attempt))
		}
		return nil
	}

	return fmt.Errorf("max retries (%d) exceeded for optimistic lock", maxOptimisticLockRetries)
}

// applyResponseToState merges response data into state's CollectedData
func (s *SagaCoordinator) applyResponseToState(state *OrchestrationState, stepName string, step models.Step, stepExists bool, normalisedData map[string]interface{}, awaitedReq *AwaitedRequest) {

	// =========================================================================
	// NEW: Check for output_mapping in step config
	// If defined, extract only the mapped fields and store WITHOUT .response wrapper
	// This makes downstream access much simpler (e.g., hero_result.image_uri)
	// =========================================================================
	var outputMapping map[string]interface{}
	if stepExists {
		outputMapping, _ = step.Config["output_mapping"].(map[string]interface{})
	}

	if len(outputMapping) > 0 {
		mappedResult := applyOutputMapping(normalisedData, outputMapping, s.logger)

		// Add metadata
		mappedResult["response_received_at"] = time.Now().Format(time.RFC3339)
		mappedResult["response_status"] = "complete"

		// Store mapped result DIRECTLY (no .response wrapper)
		state.CollectedData[stepName] = mappedResult

		// Also store at output_field if specified
		outputField := ""
		if stepExists {
			outputField = step.OutputField
		}
		if outputField != "" {
			state.CollectedData[outputField] = mappedResult
		}

		s.logger.Info("Applied output_mapping to response",
			zap.String("step_name", stepName),
			zap.String("output_field", outputField),
			zap.Int("mapped_fields", len(mappedResult)-2), // -2 for metadata fields
			zap.Any("mapped_keys", getMapKeys(mappedResult)))
		return
	}

	// Determine if this is a call_agent response that needs .response wrapper.
	// Key fix: use awaitedReq.TargetAgentType when step doesn't exist in WorkflowPlan
	isAgentResponse := false
	if stepExists && (step.Action == "spawn_agent" || step.Action == "call_agent") {
		isAgentResponse = true
	} else if awaitedReq != nil && awaitedReq.TargetAgentType != "" {
		// Step doesn't exist (dynamically expanded), but we have TargetAgentType
		// from the awaited request - this was a call_agent
		isAgentResponse = true
		s.logger.Info("Detected call_agent response for dynamic step (stepExists=false)",
			zap.String("step_name", stepName),
			zap.String("target_agent_type", awaitedReq.TargetAgentType))
	}

	// Get output_field - from step if exists, otherwise derive from stepName
	outputField := ""
	if stepExists {
		outputField = step.OutputField
	} else {
		// For loop steps like "build_pages_loop_iter_2_write_page_content"
		// derive the indexed output field like "page_content_2"
		outputField = deriveOutputFieldFromLoopStepName(stepName)
		if outputField != "" {
			s.logger.Debug("Derived output_field from step name",
				zap.String("step_name", stepName),
				zap.String("output_field", outputField))
		}
	}

	// Handle agent responses - wrap in .response
	if isAgentResponse {
		// Get or create existing data map
		existingData, exists := state.CollectedData[stepName].(map[string]interface{})
		if !exists {
			existingData = make(map[string]interface{})
		}

		existingData["response"] = normalisedData
		existingData["response_received_at"] = time.Now().Format(time.RFC3339)
		existingData["response_status"] = "complete"

		if stepExists && step.Action == "spawn_agent" {
			existingData["initialized"] = true
		}

		state.CollectedData[stepName] = existingData

		// Also store at output_field if specified
		if outputField != "" {
			outputData, exists := state.CollectedData[outputField].(map[string]interface{})
			if !exists {
				outputData = make(map[string]interface{})
			}
			outputData["response"] = normalisedData
			outputData["response_received_at"] = time.Now().Format(time.RFC3339)
			outputData["response_status"] = "complete"
			if stepExists && step.Action == "spawn_agent" {
				outputData["initialized"] = true
			}
			state.CollectedData[outputField] = outputData
		}

		s.logger.Info("Stored agent response with .response wrapper",
			zap.String("step_name", stepName),
			zap.String("output_field", outputField),
			zap.Bool("step_exists", stepExists))
		return
	}

	// Handle spawn_agent without existing data
	if stepExists && step.Action == "spawn_agent" {
		spawnData := s.extractSpawnData(normalisedData, step)
		state.CollectedData[stepName] = spawnData
		if step.OutputField != "" {
			state.CollectedData[step.OutputField] = spawnData
		}
		return
	}

	// Default: store response directly (for non-agent actions)
	state.CollectedData[stepName] = normalisedData
	if outputField != "" {
		dataToStore := normalisedData
		if stepExists && shouldExtractFormFields(step) {
			if formFields := extractHITLFormFields(normalisedData, step.Config, s.logger); len(formFields) > 0 {
				dataToStore = formFields
			}
		}
		state.CollectedData[outputField] = dataToStore
	}
}

// applyOutputMapping extracts fields from response using dot-notation paths
// mapping format: {"target_key": "source.path.to.field", ...}
func applyOutputMapping(data map[string]interface{}, mapping map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	result := make(map[string]interface{})

	for targetKey, sourcePath := range mapping {
		pathStr, ok := sourcePath.(string)
		if !ok {
			logger.Warn("output_mapping: invalid path type",
				zap.String("target", targetKey),
				zap.String("type", fmt.Sprintf("%T", sourcePath)))
			continue
		}

		value := datahelpers.ExtractNestedField(data, pathStr)
		if value != nil {
			result[targetKey] = value
			logger.Debug("output_mapping: extracted field",
				zap.String("target", targetKey),
				zap.String("source", pathStr))
		} else {
			logger.Debug("output_mapping: field not found",
				zap.String("target", targetKey),
				zap.String("source", pathStr))
		}
	}

	return result
}

// deriveOutputFieldFromLoopStepName extracts indexed output_field from loop step names.
// Example: "build_pages_loop_iter_2_write_page_content" -> "page_content_2"
// Pattern: {loop_name}_iter_{N}_{substep} where substep has output_field {base}
// Result: {base}_{N}
func deriveOutputFieldFromLoopStepName(stepName string) string {
	// Match pattern: *_iter_{N}_{substep}
	re := regexp.MustCompile(`_iter_(\d+)_(\w+)$`)
	matches := re.FindStringSubmatch(stepName)
	if len(matches) != 3 {
		return ""
	}

	iterNum := matches[1]
	substep := matches[2]

	// Map known substeps to their base output_field names
	// These should match the output_field values in the sub_workflow steps
	substepToBase := map[string]string{
		"write_page_content":  "page_content",
		"review_page_content": "reviewed_content",
		"assemble_page":       "assembled_page",
		"deploy_page":         "page_deployed",
		"generate_content":    "page_content",
		"create_html":         "page_html",
	}

	if base, ok := substepToBase[substep]; ok {
		return fmt.Sprintf("%s_%s", base, iterNum)
	}

	return ""
}

// parseResponseBody extracts and normalizes response body data
func (s *SagaCoordinator) parseResponseBody(response types.ResponseMessage) (map[string]interface{}, error) {
	var responseBodyData map[string]interface{}

	switch bodyData := response.Body.Body.(type) {
	case map[string]interface{}:
		responseBodyData = bodyData
	case []byte:
		if err := json.Unmarshal(bodyData, &responseBodyData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body bytes: %w", err)
		}
	case string:
		if err := json.Unmarshal([]byte(bodyData), &responseBodyData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body string: %w", err)
		}
	default:
		jsonBytes, err := json.Marshal(bodyData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response body: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &responseBodyData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return datahelpers.CleanDataMap(responseBodyData), nil
}

// extractSpawnData pulls spawn info from response when existing data is missing
func (s *SagaCoordinator) extractSpawnData(normalisedData map[string]interface{}, step models.Step) map[string]interface{} {
	spawnData := make(map[string]interface{})

	// Try nested body first
	if nestedBody, ok := normalisedData["body"].(map[string]interface{}); ok {
		for k, v := range nestedBody {
			spawnData[k] = v
		}
	} else {
		// Fallback to top level
		for _, key := range []string{"agent_id", "agent_type", "role", "topics", "initialized"} {
			if val, exists := normalisedData[key]; exists {
				spawnData[key] = val
			}
		}
	}

	spawnData["response"] = normalisedData
	spawnData["response_received_at"] = time.Now().Format(time.RFC3339)
	spawnData["initialized"] = true

	if _, hasRole := spawnData["role"]; !hasRole {
		if configRole, ok := step.Config["role"].(string); ok {
			spawnData["role"] = configRole
		}
	}

	return spawnData
}

// createContinuationContext builds an ExecutionContext for continuing the workflow
func (s *SagaCoordinator) createContinuationContext(state *OrchestrationState) *types.ExecutionContext {
	responsesTopic := state.ResponsesTopic
	if responsesTopic == "" {
		responsesTopic = os.Getenv("MY_RESPONSES_TOPIC")
	}

	requestsTopic := state.RequestsTopic
	if requestsTopic == "" {
		requestsTopic = os.Getenv("MY_REQUESTS_TOPIC")
	}

	return &types.ExecutionContext{
		CorrelationID:   state.CorrelationID,
		OrchestrationID: state.OrchestrationID,
		ClientID:        state.ClientID,
		MessageType:     "request",
		MessageID:       uuid.New().String(),
		Sender: types.AgentIdentity{
			AgentType:    state.OwnerAgentType,
			AgentID:      state.OwnerAgentID,
			PodName:      s.podName,
			AgentVersion: os.Getenv("AGENT_VERSION"),
		},
		FuelBudget:     state.FuelBudget,
		TimeoutSeconds: 30,
		Timestamp:      time.Now(),
		Version:        "2.0",
		ResponsesTopic: responsesTopic,
		RequestsTopic:  requestsTopic,
	}
}

// handleRecoverableError handles errors that can be retried
func (s *SagaCoordinator) handleRecoverableError(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response types.ResponseMessage) error {
	s.logger.Warn("Recoverable error received",
		zap.String("request_id", requestID),
		zap.Int("retry_version_from_execCtx", execCtx.RetryVersion),
	)

	// Get awaited request from DATABASE to get current retry_version
	// The in-memory state.AwaitedRequests may have stale retry_version
	repo := NewStateRepository(s.db, s.logger)
	awaited, err := repo.GetAwaitedRequest(ctx, requestID)
	if err != nil || awaited == nil {
		// Fall back to in-memory if DB lookup fails
		awaited = state.AwaitedRequests[requestID]
		if awaited == nil {
			return fmt.Errorf("no awaited request found for %s", requestID)
		}
		s.logger.Warn("Using in-memory awaited request (DB lookup failed)",
			zap.String("request_id", requestID),
			zap.Error(err))
	}

	s.logger.Info("Loaded awaited request for retry",
		zap.String("request_id", requestID),
		zap.Int("retry_version_from_db", awaited.RetryVersion),
		zap.String("step_name", awaited.StepName))

	// Check retry count - max 3 retries to prevent infinite loops
	if awaited.RetryVersion >= 3 {
		s.logger.Error("Max retries exceeded",
			zap.String("request_id", requestID),
			zap.Int("retry_version", awaited.RetryVersion))
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	}

	// Adapter actions (git_commit, etc.) need to re-execute the step, not send
	// a retry message, because adapters need the full payload (files, etc.)
	isAdapterAction := awaited.TargetAgentID == "" &&
		strings.HasPrefix(awaited.RequestsTopic, "system.adapter")

	if isAdapterAction {
		s.logger.Info("Re-executing adapter action for retry",
			zap.String("step_name", awaited.StepName),
			zap.String("request_topic", awaited.RequestsTopic),
			zap.Int("retry_attempt", awaited.RetryVersion+1))

		// Remove the failed awaited request
		delete(state.AwaitedRequests, requestID)

		// Mark it processed in DB so it doesn't linger
		if err := repo.MarkAwaitedRequestComplete(ctx, requestID); err != nil {
			s.logger.Warn("Failed to mark awaited request complete for re-execution",
				zap.String("request_id", requestID),
				zap.Error(err))
		}

		// Track retry count in execution metadata
		if state.ExecutionMetadata.RetryCount == nil {
			state.ExecutionMetadata.RetryCount = make(map[string]int)
		}
		state.ExecutionMetadata.RetryCount[awaited.StepName] = awaited.RetryVersion + 1

		// Reset state to re-execute the step
		state.CurrentStep = awaited.StepName
		state.Status = StatusExecutingStep
		state.CurrentlyExecuting = &awaited.StepName
		state.LastActivity = time.Now()

		// Save state
		if err := repo.UpdateState(ctx, state); err != nil {
			s.logger.Error("Failed to update state for adapter retry",
				zap.String("request_id", requestID),
				zap.Error(err))
			return fmt.Errorf("failed to update state for adapter retry: %w", err)
		}

		s.logger.Info("Re-executing step after adapter timeout",
			zap.String("step_name", awaited.StepName),
			zap.Int("retry_attempt", awaited.RetryVersion+1))

		// Continue execution - this will re-run the step action with full context
		return s.continueExecution(ctx, state, execCtx)
	}

	// Use the stored RequestsTopic to send retry to target agent
	if awaited.RequestsTopic == "" {
		s.logger.Error("Cannot retry - no requests_topic stored in awaited request",
			zap.String("request_id", requestID),
			zap.String("step_name", awaited.StepName),
			zap.String("target_agent_type", awaited.TargetAgentType))
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	}

	// Increment retry version
	awaited.RetryVersion++

	// Use original timeout duration, not hardcoded 60s
	// Calculate from the original TimeoutAt - SentAt if available
	originalDuration := awaited.TimeoutAt.Sub(awaited.SentAt)
	newTimeout := originalDuration
	if newTimeout <= 0 || newTimeout > 30*time.Minute {
		// Fallback to reasonable default if calculation fails
		newTimeout = 3 * time.Minute
	}
	// For retries, cap at 5 minutes to avoid very long waits
	if newTimeout > 5*time.Minute {
		newTimeout = 5 * time.Minute
	}

	awaited.SentAt = time.Now()
	awaited.TimeoutAt = time.Now().Add(newTimeout)

	// Update in-memory state too (for consistency)
	state.AwaitedRequests[requestID] = awaited

	s.logger.Info("Retrying request",
		zap.String("request_id", requestID),
		zap.Int("retry_version", awaited.RetryVersion),
		zap.String("target_requests_topic", awaited.RequestsTopic),
		zap.String("responses_topic", awaited.ResponsesTopic),
		zap.String("target_agent_type", awaited.TargetAgentType),
		zap.String("target_agent_id", awaited.TargetAgentID),
	)

	// Persist the increment to database BEFORE sending retry
	if err := repo.UpdateAwaitedRequestRetry(ctx, requestID, awaited.RetryVersion, awaited.TimeoutAt); err != nil {
		s.logger.Error("Failed to update retry version in DB", zap.Error(err))
		// Continue anyway - better to retry than to fail
	}

	// Determine fuel budget
	fuelBudget := state.FuelBudget
	if fuelBudget <= 0 {
		fuelBudget = 100
	}

	// Build sender info
	sender := execCtx.Sender
	if sender.AgentID == "" {
		sender = types.AgentIdentity{
			AgentID:   os.Getenv("AGENT_ID"),
			AgentType: os.Getenv("AGENT_TYPE"),
			PodName:   os.Getenv("HOSTNAME"),
		}
	}

	// Create retry request
	retryRequest := &types.RequestMessage{
		Headers: types.RequestHeaders{
			CorrelationID:     state.CorrelationID,
			ClientID:          state.ClientID,
			Sender:            sender,
			OrchestrationID:   state.OrchestrationID,
			OrchestrationName: state.OrchestrationName,
			RequestID:         requestID,
			RetryVersion:      awaited.RetryVersion,
			StepID:            awaited.StepID,
			StepName:          awaited.StepName,
			ToAgentType:       awaited.TargetAgentType,
			ToAgent:           awaited.TargetAgentID,
			RequestsTopic:     awaited.RequestsTopic,
			ResponsesTopic:    awaited.ResponsesTopic,
			MessageID:         uuid.New().String(),
			MessageType:       "request",
			Timestamp:         time.Now(),
			Action:            "execute",
			FuelBudget:        fuelBudget,
			TimeoutSeconds:    int(newTimeout.Seconds()),
		},
		Body: map[string]interface{}{
			"is_retry":      true,
			"retry_version": awaited.RetryVersion,
		},
	}

	retryBytes, err := json.Marshal(retryRequest)
	if err != nil {
		s.logger.Error("Failed to marshal retry request",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}

	s.logger.Info("Sending retry to target agent requests topic",
		zap.String("request_id", requestID),
		zap.String("topic", awaited.RequestsTopic),
		zap.String("target_agent_id", awaited.TargetAgentID))

	// Send the retry
	err = s.producer.Produce(ctx, awaited.RequestsTopic, retryRequest.Headers.ToMap(), []byte(requestID), retryBytes)
	if err != nil {
		s.logger.Error("Failed to send retry request",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}

	// Start a new timeout goroutine for the retry
	go s.handleRequestTimeout(context.Background(), state.OrchestrationID, requestID, awaited.TimeoutAt)

	return nil
}

/*func (r *StateRepository) UpdateAwaitedRequestRetry(ctx context.Context, requestID string, retryVersion int, timeoutAt time.Time) error {
	query := `
        UPDATE awaited_requests
        SET retry_version = $1,
            timeout_at = $2,
            sent_at = NOW()
        WHERE request_id = $3`

	_, err := r.db.ExecContext(ctx, query, retryVersion, timeoutAt, requestID)
	return err
}*/

func (r *StateRepository) UpdateAwaitedRequestRetry(ctx context.Context, requestID string, retryVersion int, timeoutAt time.Time) error {
	query := `
        UPDATE awaited_requests 
        SET retry_version = $1, 
            timeout_at = $2,
            sent_at = NOW(),
            status = 'waiting',
            processed_at = NULL,
            processing_started_at = NULL,
            processing_pod = NULL
        WHERE request_id = $3`

	_, err := r.db.ExecContext(ctx, query, retryVersion, timeoutAt, requestID)
	return err
}

// handleUnrecoverableError handles fatal errors
func (s *SagaCoordinator) handleUnrecoverableError(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response types.ResponseMessage) error {
	s.logger.Error("Unrecoverable error received",
		zap.String("request_id", requestID))

	// Extract error message from the ResponseMessage structure
	errorMsg := "Unknown error"

	// Check if there's an Error field in the ResponseBody
	if response.Body.Error != nil {
		errorMsg = response.Body.Error.Message
		if response.Body.Error.Code != "" {
			errorMsg = fmt.Sprintf("%s (code: %s)", errorMsg, response.Body.Error.Code)
		}
	} else if response.Body.Body != nil {
		// If no Error field, check the Body field for error information
		switch body := response.Body.Body.(type) {
		case map[string]interface{}:
			if errStr, ok := body["error"].(string); ok {
				errorMsg = errStr
			}
		case string:
			errorMsg = body
		}
	}

	// Check if this is a loop iteration with continue_on_error
	if shouldContinueLoopOnError(state, s.logger) {
		return s.skipToNextLoopIterationForAsync(ctx, state, requestID, errorMsg, s.logger)
	}

	// Check if the failed step has an error_step to route to
	failedStepName := ""
	if ar, exists := state.AwaitedRequests[requestID]; exists {
		failedStepName = ar.StepName
	}
	if failedStepName != "" {
		if step, exists := state.WorkflowPlan.Steps[failedStepName]; exists {
			if errorStep, ok := step.Config["error_step"].(string); ok && errorStep != "" {
				// Clean up the awaited request before routing to error_step
				delete(state.AwaitedRequests, requestID)
				repo := NewStateRepository(s.db, s.logger)
				if markErr := repo.MarkAwaitedRequestComplete(ctx, requestID); markErr != nil {
					s.logger.Warn("Failed to mark awaited request complete for error_step routing",
						zap.String("request_id", requestID), zap.Error(markErr))
				}
				return s.routeToErrorStep(ctx, state, failedStepName, errorStep, errorMsg)
			}
		}
	}

	return s.failWorkflow(ctx, state, fmt.Sprintf("Request %s failed: %s", requestID, errorMsg))

}

// handleRequestTimeout handles request timeouts
func (s *SagaCoordinator) handleRequestTimeout(ctx context.Context, orchestrationID, requestID string, timeoutAt time.Time) {
	time.Sleep(time.Until(timeoutAt))

	repo := NewStateRepository(s.db, s.logger)

	// Get fresh awaited request from DB (not from state which may be stale)
	awaited, err := repo.GetAwaitedRequest(ctx, requestID)
	if err != nil || awaited == nil {
		return // Already completed or doesn't exist
	}

	// Check retry count from DB
	if awaited.RetryVersion >= 3 {
		s.logger.Error("Max retries exceeded",
			zap.String("request_id", requestID),
			zap.Int("retry_version", awaited.RetryVersion))
		state, _ := repo.GetState(ctx, orchestrationID)
		if state != nil {
			// Check if this is a loop iteration with continue_on_error
			if shouldContinueLoopOnError(state, s.logger) {
				timeoutMsg := fmt.Sprintf("Request %s timed out after %d retries", requestID, awaited.RetryVersion)
				if err := s.skipToNextLoopIterationForAsync(ctx, state, requestID, timeoutMsg, s.logger); err != nil {
					s.logger.Error("Failed to skip loop iteration on timeout", zap.Error(err))
					s.failWorkflow(ctx, state, timeoutMsg)
				}
				return
			}
			s.routeToErrorStepOrFail(ctx, state, awaited.StepName, fmt.Sprintf("Request %s timed out after %d retries", requestID, awaited.RetryVersion))
		}
		return
	}

	state, err := repo.GetState(ctx, orchestrationID)
	if err != nil {
		return
	}

	// Check if still waiting
	if awaited, exists := state.AwaitedRequests[requestID]; exists {
		s.logger.Error("Request timed out",
			zap.String("request_id", requestID),
			zap.String("awaited step id", awaited.StepID),
			zap.String("awaited step name", awaited.StepName),
			zap.Int("retry_version", awaited.RetryVersion),
		)

		// Retry or fail
		if awaited.RetryVersion < 3 {
			// Create full ExecutionContext for retry - pull values from state and CollectedData
			// Try to get responses topic from various sources
			responsesTopic := ""
			if myResponses, ok := state.CollectedData["__my_responses_topic__"].(string); ok && myResponses != "" {
				responsesTopic = myResponses
			} else if parentResponses, ok := state.CollectedData["__parent_responses_topic__"].(string); ok && parentResponses != "" {
				responsesTopic = parentResponses
			} else if os.Getenv("RESPONSES_TOPIC") != "" {
				responsesTopic = os.Getenv("RESPONSES_TOPIC")
			} else if os.Getenv("PARENT_RESPONSES_TOPIC") != "" {
				responsesTopic = os.Getenv("PARENT_RESPONSES_TOPIC")
			}

			execCtx := &types.ExecutionContext{
				// Identity
				CorrelationID:         state.CorrelationID,
				OrchestrationID:       orchestrationID,
				OrchestrationName:     state.OrchestrationName,
				ParentOrchestrationID: state.ParentOrchestrationID,
				ClientID:              state.ClientID,

				// Request info
				RequestID:   requestID,
				MessageType: "request",
				Status:      "error_recoverable",

				// Topics
				ResponsesTopic: responsesTopic,
				ReplyToTopic:   responsesTopic,

				// Sender
				Sender: types.AgentIdentity{
					AgentType:    state.OwnerAgentType,
					AgentID:      state.OwnerAgentID,
					PodName:      s.podName,
					AgentVersion: os.Getenv("AGENT_VERSION"),
					Role:         state.OwnerAgentRole,
				},

				// Resources
				FuelBudget:     state.FuelBudget,
				TimeoutSeconds: 30,
				Timestamp:      time.Now(),
				Version:        "2.0",
			}

			s.logger.Info("Created ExecutionContext for timeout retry",
				zap.String("correlation_id", execCtx.CorrelationID),
				zap.String("orchestration_id", execCtx.OrchestrationID),
				zap.String("client_id", execCtx.ClientID),
				zap.String("responses_topic", execCtx.ResponsesTopic))

			// Create a timeout response message instead of raw bytes
			timeoutResponse := types.ResponseMessage{
				Headers: types.ResponseHeaders{
					InResponseToRequestID: requestID,
					InResponseToStepID:    awaited.StepID,
					InResponseToStepName:  awaited.StepName,
					InResponseToAction:    "unknown",
					Status:                "error_recoverable",
					IsError:               true,
					MessageType:           "response",
					TimeSent:              time.Now(),
					OrchestrationID:       orchestrationID,
					CorrelationID:         state.CorrelationID,
				},
				Body: types.ResponseBody{
					Success: false,
					Error: &types.ErrorInfo{
						Code:        "TIMEOUT",
						Message:     "Request timed out",
						Recoverable: true,
						Details: map[string]interface{}{
							"timeout_after": timeoutAt.Sub(awaited.TimeoutAt).String(),
							"retry_count":   awaited.RetryVersion,
						},
					},
				},
			}

			s.handleRecoverableError(ctx, state, requestID, execCtx, timeoutResponse)
		} else {
			s.routeToErrorStepOrFail(ctx, state, awaited.StepName, fmt.Sprintf("Request %s timed out after %d retries", requestID, awaited.RetryVersion))
		}
	}
}

// Helper methods
func (s *SagaCoordinator) determineOwnerAgentType(execCtx *types.ExecutionContext) string {
	if execCtx.Sender.AgentType != "" {
		return execCtx.Sender.AgentType
	}
	if agentType := os.Getenv("AGENT_TYPE"); agentType != "" {
		return agentType
	}
	s.logger.Error("Did not find agent type in determineOwnerAgentType (coordinator 887)")
	return "generic"
}

func (s *SagaCoordinator) determineOwnerAgentID(execCtx *types.ExecutionContext) string {
	if execCtx.Sender.AgentID != "" {
		return execCtx.Sender.AgentID
	}
	if ownerID := os.Getenv("AGENT_ID"); ownerID != "" {
		return ownerID
	}
	return "00000000-0000-0000-0000-000000000001"
}

func (s *SagaCoordinator) dependenciesMet(dependencies []string, state *OrchestrationState) bool {
	for _, dep := range dependencies {
		if _, ok := state.CollectedData[dep]; !ok {
			return false
		}
	}
	return true
}

func (s *SagaCoordinator) skipStep(ctx context.Context, state *OrchestrationState, reason string) error {
	s.logger.Info("Skipping step",
		zap.String("step", state.CurrentStep),
		zap.String("reason", reason))

	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		StepName:  state.CurrentStep,
		Action:    "skipped",
		Timestamp: time.Now(),
		Details:   reason,
	})

	repo := NewStateRepository(s.db, s.logger)
	return repo.UpdateState(ctx, state)
}

// routeToErrorStepOrFail checks if the failed step has an error_step configured.
// If so, stores error info in collected_data and routes the workflow to that step.
// Otherwise, falls through to failWorkflow.
//
// Used by: continueExecution (local step failures), handleRequestTimeout (timeout failures).
// For async child failures (handleUnrecoverableError), the caller must clean up the
// awaited request first, then call routeToErrorStep directly.
func (s *SagaCoordinator) routeToErrorStepOrFail(ctx context.Context, state *OrchestrationState, failedStepName string, errorMsg string) error {
	if failedStepName != "" {
		if step, exists := state.WorkflowPlan.Steps[failedStepName]; exists {
			if errorStep, ok := step.Config["error_step"].(string); ok && errorStep != "" {
				return s.routeToErrorStep(ctx, state, failedStepName, errorStep, errorMsg)
			}
		}
	}
	return s.failWorkflow(ctx, state, errorMsg)
}

// routeToErrorStep routes the workflow to the given error step, storing error
// context in collected_data["__step_error"]. The workflow continues from errorStep
// instead of failing. This allows workflows to handle errors gracefully — e.g. the
// dispatch loop can mark a work item as failed and continue to the next item.
//
// Workflow config usage:
//
//	"call_handler": {
//	    "action": "call_agent",
//	    "config": {
//	        "target_role": "handler",
//	        "error_step": "mark_failed",
//	        ...
//	    },
//	    "next_step": "mark_complete"
//	}
func (s *SagaCoordinator) routeToErrorStep(ctx context.Context, state *OrchestrationState, failedStepName string, errorStep string, errorMsg string) error {
	s.logger.Info("Routing to error_step instead of failing workflow",
		zap.String("failed_step", failedStepName),
		zap.String("error_step", errorStep),
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("error_msg", errorMsg))

	// Store error context so downstream steps can read it if needed
	state.CollectedData["__step_error"] = map[string]interface{}{
		"failed_step": failedStepName,
		"message":     errorMsg,
	}

	state.CurrentStep = errorStep
	state.Status = StatusExecutingStep
	state.LastActivity = time.Now()

	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		StepName:  failedStepName,
		Action:    "error_routed",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("routed to %s: %s", errorStep, errorMsg),
	})

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		s.logger.Error("Failed to save state for error_step routing",
			zap.String("error_step", errorStep),
			zap.Error(err))
		// Fall back to failing the workflow if we can't save
		return s.failWorkflow(ctx, state, errorMsg)
	}

	return s.continueExecution(ctx, state, s.createContinuationContext(state))
}

func (s *SagaCoordinator) failWorkflow(ctx context.Context, state *OrchestrationState, errorMsg string) error {
	// Check if this is an optimistic lock failure - these are recoverable
	if strings.Contains(errorMsg, "optimistic lock failure") {
		s.logger.Info("Optimistic lock failure detected - checking if workflow recovered",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("error", errorMsg))

		// Reload fresh state to see if another process already recovered
		repo := NewStateRepository(s.db, s.logger)
		freshState, err := repo.GetState(ctx, state.OrchestrationID)
		if err == nil && freshState != nil {
			// If workflow is still running or completed, don't fail it
			if freshState.Status == StatusRunning ||
				freshState.Status == StatusAwaitingResponses ||
				freshState.Status == StatusCompleted {
				s.logger.Info("Workflow recovered by another process - not sending error",
					zap.String("orchestration_id", state.OrchestrationID),
					zap.String("fresh_status", string(freshState.Status)),
					zap.String("fresh_step", freshState.CurrentStep))
				// Return error locally but don't notify parent
				return fmt.Errorf("workflow recovered: %s", errorMsg)
			}

			// If workflow already failed, check if error was already sent
			if freshState.Status == StatusFailed && freshState.Error != "" {
				s.logger.Info("Workflow already failed - not sending duplicate error",
					zap.String("orchestration_id", state.OrchestrationID),
					zap.String("existing_error", freshState.Error))
				return fmt.Errorf("workflow already failed: %s", freshState.Error)
			}
		}
	}

	// Proceed with normal failure handling
	state.Status = StatusFailed
	state.Error = errorMsg

	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		Action:    "workflow_failed",
		Timestamp: time.Now(),
		Details:   errorMsg,
	})

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		// If we can't save the failed state, check if it's because workflow already completed
		if strings.Contains(err.Error(), "optimistic lock") {
			freshState, loadErr := repo.GetState(ctx, state.OrchestrationID)
			if loadErr == nil && freshState.Status == StatusCompleted {
				s.logger.Info("Workflow completed while we tried to fail it - not sending error",
					zap.String("orchestration_id", state.OrchestrationID))
				return fmt.Errorf("workflow completed concurrently: %s", errorMsg)
			}
		}
		s.logger.Error("Failed to save failed state", zap.Error(err))
	}

	// Notify parent of failure
	s.notifyParentOfFailure(ctx, state, errorMsg)

	return fmt.Errorf("workflow failed: %s", errorMsg)
}

func (s *SagaCoordinator) notifyParentOfSuccess(ctx context.Context, state *OrchestrationState) {
	// Get parent response topic
	parentTopic, _ := state.CollectedData["__parent_responses_topic__"].(string)
	replyToRequestID, _ := state.CollectedData["__reply_to_request_id__"].(string)

	if parentTopic == "" || replyToRequestID == "" {
		s.logger.Debug("No parent to notify of success",
			zap.String("parent_topic", parentTopic),
			zap.String("reply_to_request_id", replyToRequestID))
		return
	}

	s.logger.Info("Notifying parent of child orchestration success",
		zap.String("parent_topic", parentTopic),
		zap.String("reply_to_request_id", replyToRequestID),
		zap.String("orchestration_id", state.OrchestrationID))

	// Build result from CollectedData - extract workflow outputs
	resultData := s.extractWorkflowResultWithSizeLimit(state)

	successResponse := types.ResponseMessage{
		Headers: types.ResponseHeaders{
			InResponseToRequestID: replyToRequestID,
			Status:                "complete",
			IsComplete:            true,
			MessageType:           "response",
			//MessageID:             uuid.New().String(),
			TimeSent:        time.Now(),
			OrchestrationID: state.OrchestrationID,
			CorrelationID:   state.CorrelationID,
			ClientID:        state.ClientID,
		},
		Body: types.ResponseBody{
			Success: true,
			Body:    resultData,
		},
	}

	responseBytes, err := json.Marshal(successResponse)
	if err != nil {
		s.logger.Error("Failed to marshal success response", zap.Error(err))
		return
	}

	if err := s.producer.Produce(ctx, parentTopic, successResponse.Headers.ToMap(), []byte(replyToRequestID), responseBytes); err != nil {
		s.logger.Error("Failed to notify parent of success", zap.Error(err))
	} else {
		s.logger.Info("Successfully notified parent of workflow completion",
			zap.String("parent_topic", parentTopic),
			zap.String("reply_to_request_id", replyToRequestID))
	}
}

// extractWorkflowResult builds the result payload to send back to parent
// respects output_fields config from complete_workflow step
func (s *SagaCoordinator) extractWorkflowResult(state *OrchestrationState) map[string]interface{} {
	result := make(map[string]interface{})

	// Try to get output_fields from the complete step's config
	// state.WorkflowPlan is models.WorkflowPlan with Steps map[string]Step
	var outputFields []string
	if state.WorkflowPlan.Steps != nil {
		if completeStep, ok := state.WorkflowPlan.Steps["complete"]; ok {
			if config, ok := completeStep.Config["output_fields"].([]interface{}); ok {
				for _, f := range config {
					if fieldName, ok := f.(string); ok {
						outputFields = append(outputFields, fieldName)
					}
				}
			}
		}
	}

	// If output_fields is configured, only include those fields
	if len(outputFields) > 0 {
		s.logger.Info("extractWorkflowResult: using configured output_fields",
			zap.Strings("output_fields", outputFields))

		for _, fieldName := range outputFields {
			if value := datahelpers.ExtractNestedField(state.CollectedData, fieldName); value != nil {
				// Extract step data to remove .response wrapper if present
				result[fieldName] = datahelpers.ExtractStepData(value)
			}
		}
	} else {
		// Fallback: include non-internal, non-large fields
		// Skip fields that are typically large and not needed by parent
		skipPatterns := []string{
			"page_content_",     // Individual page content (can be large)
			"reviewed_content_", // Reviewed versions
			"build_pages_loop_", // Loop iteration data
			"assembled_page",    // Full HTML
			"page_deployed_",    // Deployment results per page
		}

		for key, value := range state.CollectedData {
			// Skip internal fields
			if strings.HasPrefix(key, "__") {
				continue
			}
			// Skip loop metadata
			if key == "loop_metadata" {
				continue
			}
			// Skip large pattern matches
			skip := false
			for _, pattern := range skipPatterns {
				if strings.HasPrefix(key, pattern) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			// Extract step data to remove .response wrapper
			result[key] = datahelpers.ExtractStepData(value)
		}
	}

	// Add completion metadata (small, always useful)
	result["orchestration_id"] = state.OrchestrationID
	result["completed_steps"] = state.ExecutionMetadata.CompletedSteps
	result["completed_at"] = time.Now().Format(time.RFC3339)

	// Log the final size for monitoring
	if resultBytes, err := json.Marshal(result); err == nil {
		s.logger.Info("extractWorkflowResult: result size",
			zap.Int("size_bytes", len(resultBytes)),
			zap.Int("field_count", len(result)))
	}

	return result
}

const MaxResultSizeBytes = 900000 // Leave headroom below 1MB Kafka limit

// extractWorkflowResultWithSizeLimit builds result respecting size limits
func (s *SagaCoordinator) extractWorkflowResultWithSizeLimit(state *OrchestrationState) map[string]interface{} {
	// First try with configured output_fields
	result := s.extractWorkflowResult(state)

	// Check size
	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.logger.Error("Failed to marshal result for size check", zap.Error(err))
		return s.extractMinimalResult(state)
	}

	if len(resultBytes) <= MaxResultSizeBytes {
		return result
	}

	// Result too large - try removing large string fields
	s.logger.Warn("extractWorkflowResult: result too large, trimming",
		zap.Int("original_size", len(resultBytes)))

	for key, value := range result {
		if str, ok := value.(string); ok && len(str) > 10000 {
			// Truncate large strings
			result[key] = str[:1000] + "...[truncated]"
		}
		// Could also recurse into maps/arrays if needed
	}

	// Re-check size
	resultBytes, _ = json.Marshal(result)
	if len(resultBytes) <= MaxResultSizeBytes {
		return result
	}

	// Still too large - return minimal result
	s.logger.Warn("extractWorkflowResult: still too large after trimming, using minimal result",
		zap.Int("size_after_trim", len(resultBytes)))
	return s.extractMinimalResult(state)
}

// extractMinimalResult returns just the essential completion info
func (s *SagaCoordinator) extractMinimalResult(state *OrchestrationState) map[string]interface{} {
	return map[string]interface{}{
		"orchestration_id": state.OrchestrationID,
		"completed_steps":  state.ExecutionMetadata.CompletedSteps,
		"completed_at":     time.Now().Format(time.RFC3339),
		"status":           "completed",
		"message":          "Workflow completed successfully. Full result exceeded size limit.",
	}
}

func (s *SagaCoordinator) notifyParentOfFailure(ctx context.Context, state *OrchestrationState, errorMsg string) {
	// Get parent response topic
	parentTopic, _ := state.CollectedData["__parent_responses_topic__"].(string)
	replyToRequestID, _ := state.CollectedData["__reply_to_request_id__"].(string)

	if parentTopic == "" || replyToRequestID == "" {
		s.logger.Debug("No parent to notify of failure",
			zap.String("parent_topic", parentTopic),
			zap.String("reply_to_request_id", replyToRequestID))
		return
	}

	s.logger.Info("Notifying parent of child orchestration failure",
		zap.String("parent_topic", parentTopic),
		zap.String("reply_to_request_id", replyToRequestID),
		zap.String("error", errorMsg))

	failureResponse := types.ResponseMessage{
		Headers: types.ResponseHeaders{
			InResponseToRequestID: replyToRequestID,
			Status:                "error_unrecoverable",
			IsError:               true,
			MessageType:           "response",
			//MessageID:             uuid.New().String(),
			TimeSent:        time.Now(),
			OrchestrationID: state.OrchestrationID,
			CorrelationID:   state.CorrelationID,
		},
		Body: types.ResponseBody{
			Success: false,
			Error: &types.ErrorInfo{
				Code:        "CHILD_ORCHESTRATION_FAILED",
				Message:     errorMsg,
				Recoverable: false,
			},
		},
	}

	responseBytes, _ := json.Marshal(failureResponse)
	if err := s.producer.Produce(ctx, parentTopic, failureResponse.Headers.ToMap(), []byte(replyToRequestID), responseBytes); err != nil {
		s.logger.Error("Failed to notify parent of failure", zap.Error(err))
	}
}

func (s *SagaCoordinator) completeWorkflow(ctx context.Context, state *OrchestrationState) error {

	safeDataKeys := make([]string, 0, len(state.CollectedData))
	for k := range state.CollectedData {
		safeDataKeys = append(safeDataKeys, k)
	}

	s.logger.Info("WORKFLOW_COMPLETION: Completing workflow completeWorkflow coordinator.go",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("parent_orchestration_id", state.ParentOrchestrationID),
		zap.Strings("collected_data_keys", safeDataKeys), // Just log the keys
		zap.String("owner_agent_type", state.OwnerAgentType))

	state.Status = StatusCompleted

	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		Action:    "workflow_completed",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Completed after %d steps", state.ExecutionMetadata.CompletedSteps),
	})

	// After status update
	s.logger.Info("WORKFLOW_COMPLETION: Status updated",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("new_status", string(state.Status)))

	repo := NewStateRepository(s.db, s.logger)

	// Cancel any remaining awaited requests
	err := repo.CancelAwaitedRequestsForOrchestration(ctx, state.OrchestrationID)
	if err != nil {
		s.logger.Error("Failed to cancel awaited requests",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.Error(err))
		// Continue anyway
	}

	// Notify parent of successful completion before updating state
	// This ensures parent receives notification even if state update fails
	s.notifyParentOfSuccess(ctx, state)

	return repo.UpdateState(ctx, state)
}

func isLocalAction(action string) bool {
	return actioncheck.IsLocalAction(action)
}

func getTargetAgentType(step models.Step, result map[string]interface{}) string {
	if agentType, ok := result["target_agent_type"].(string); ok {
		return agentType
	}
	if agentType, ok := step.Config["agent_type"].(string); ok {
		return agentType
	}
	return "unknown"
}

func getTimeout(step models.Step) time.Duration {
	return datahelpers.GetStepTimeout(step)
}

func extractFuelBudget(state *OrchestrationState) int {
	if fuel, ok := state.CollectedData["__fuel_budget__"].(float64); ok {
		return int(fuel)
	}
	return 1000 // Default
}

// Helper to get current and caller function names
func getFuncInfo(skip int) (current, caller string) {
	// skip=0 => this func, skip=1 => its caller, skip=2 => caller's caller, etc.
	if pc, _, _, ok := runtime.Caller(skip); ok {
		current = runtime.FuncForPC(pc).Name()
	}
	if pc, _, _, ok := runtime.Caller(skip + 1); ok {
		caller = runtime.FuncForPC(pc).Name()
	}
	return
}

// Helper methods for retry tracking (add these to SagaCoordinator struct)
func (s *SagaCoordinator) getRetryCount(key string) int {
	// Simple in-memory tracking - you might want to use Redis or database
	s.retryMutex.RLock()
	defer s.retryMutex.RUnlock()

	if s.retryCounters == nil {
		return 0
	}
	return s.retryCounters[key]
}

func (s *SagaCoordinator) incrementRetryCount(key string) {
	s.retryMutex.Lock()
	defer s.retryMutex.Unlock()

	if s.retryCounters == nil {
		s.retryCounters = make(map[string]int)
	}
	s.retryCounters[key]++
}

func (s *SagaCoordinator) cleanupExpiredAwaitedRequests() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		repo := NewStateRepository(s.db, s.logger)
		count, err := repo.CleanupExpiredAwaitedRequests(ctx)

		if err != nil {
			s.logger.Error("Failed to cleanup expired awaited requests", zap.Error(err))
		} else if count > 0 {
			s.logger.Info("Cleaned up expired awaited requests",
				zap.Int("count", count),
			)
		}

		cancel()
	}
}

// IsOptimisticLockError checks if an error is due to optimistic locking failure
func IsOptimisticLockError(err error) bool {
	if err == nil {
		return false
	}
	// Check for PostgreSQL unique violation or version mismatch
	return strings.Contains(err.Error(), "version mismatch") ||
		strings.Contains(err.Error(), "optimistic lock")
}

func (r *StateRepository) MarkAwaitedRequestComplete(ctx context.Context, requestID string) error {
	query := `
        UPDATE awaited_requests
        SET status = 'processed',
            processed_at = NOW()
        WHERE request_id = $1
    `
	_, err := r.db.ExecContext(ctx, query, requestID)
	return err
}

// ResetAwaitedRequestForRetry resets a claimed-but-not-processed request back to waiting
// This allows recovery when a pod claims a request but fails before processing it
func (r *StateRepository) ResetAwaitedRequestForRetry(ctx context.Context, requestID string) error {
	query := `
		UPDATE awaited_requests
		SET status = 'waiting',
			processing_started_at = NULL,
			processing_pod = NULL
		WHERE request_id = $1
		  AND processed_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, requestID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Reset awaited request for retry",
		zap.String("request_id", requestID),
		zap.Int64("rows_affected", rowsAffected))

	return nil
}

// AwaitedRequestExists checks if an awaited request exists (regardless of status)
// Used to distinguish "not inserted yet" from "already claimed"
func (r *StateRepository) AwaitedRequestExists(ctx context.Context, requestID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM awaited_requests WHERE request_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, requestID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetAwaitedRequestStatus returns the status and processed_at for a request
// Used to determine if a "claimed" request was actually processed successfully
func (r *StateRepository) GetAwaitedRequestStatus(ctx context.Context, requestID string) (status string, processedAt *time.Time, err error) {
	query := `SELECT status, processed_at FROM awaited_requests WHERE request_id = $1`

	var statusVal string
	var processedAtVal sql.NullTime

	err = r.db.QueryRowContext(ctx, query, requestID).Scan(&statusVal, &processedAtVal)
	if err == sql.ErrNoRows {
		return "", nil, nil // Not found
	}
	if err != nil {
		return "", nil, err
	}

	if processedAtVal.Valid {
		return statusVal, &processedAtVal.Time, nil
	}
	return statusVal, nil, nil
}

// resolveIterationNextStep determines the correct next_step for a loop substep
// Priority:
// 1. If substep.NextStep defined AND is a valid substep name → prefix with iteration
// 2. If substep.NextStep defined but NOT a substep → leave as-is (external step)
// 3. If substep.NextStep is empty → terminal step, go to next iteration start or loop_complete
func resolveIterationNextStep(
	originalNextStep string,
	loopName string,
	iterIdx int,
	totalIterations int,
	firstSubstepInOrder string,
	validSubstepSet map[string]bool,
	logger *zap.Logger,
) string {

	if originalNextStep != "" {
		// Check if it references another substep in this loop
		if validSubstepSet[originalNextStep] {
			// It's a reference to another substep - prefix with iteration
			return fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx, originalNextStep)
		}

		// It references something outside the loop - leave as-is
		// This could be an external step or a special value
		logger.Debug("NextStep references external step",
			zap.String("next_step", originalNextStep),
		)
		return originalNextStep
	}

	// NextStep is empty - this is a terminal step for the iteration
	// Go to next iteration's first step, or loop_complete if last iteration
	if iterIdx < totalIterations-1 {
		// Go to first substep of next iteration
		return fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx+1, firstSubstepInOrder)
	}

	// Last iteration - go to loop complete
	return fmt.Sprintf("%s_complete", loopName)
}

// prefixConfigStepReferences updates step references AND data field references in config with iteration prefix
// Handles:
// - Step refs: then_step, else_step, fallback_step, error_step, on_success, on_failure
// - Data refs: content_from, context_from, data_from, source_field, etc.
func prefixConfigStepReferences(config map[string]interface{}, loopName string, iterIdx int, validSubstepSet map[string]bool, substepOutputFields map[string]bool) {
	if config == nil {
		return
	}

	// List of config keys that contain step references
	stepRefKeys := []string{"then_step", "else_step", "fallback_step", "error_step", "on_success", "on_failure"}

	for _, key := range stepRefKeys {
		if val, ok := config[key].(string); ok && val != "" {
			// Only prefix if it references a valid substep
			if validSubstepSet[val] {
				config[key] = fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx, val)
			}
		}
	}

	// List of config keys that contain data field references
	// These might reference substep output fields like "generated_content.result"
	// IMPORTANT: Any config key that references step outputs must be listed here
	dataRefKeys := []string{
		"content_from",
		"context_from",
		"data_from",
		"source_field",
		"input_from",
		"result_from",
		"content_field", // Used by assemble_page, git_commit
		"commit_from",   // Used by update_page_status
	}

	for _, key := range dataRefKeys {
		if val, ok := config[key].(string); ok && val != "" {
			prefixedVal := prefixDataReference(val, loopName, iterIdx, substepOutputFields)
			if prefixedVal != val {
				config[key] = prefixedVal
			}
		}
	}

	// Also handle input_mapping which contains nested data references
	if inputMapping, ok := config["input_mapping"].(map[string]interface{}); ok {
		for key, val := range inputMapping {
			if valStr, ok := val.(string); ok && valStr != "" {
				prefixedVal := prefixDataReference(valStr, loopName, iterIdx, substepOutputFields)
				if prefixedVal != valStr {
					inputMapping[key] = prefixedVal
				}
			}
		}
	}

	for _, key := range dataRefKeys {
		if val, ok := config[key].(string); ok && val != "" {
			prefixedVal := prefixDataReference(val, loopName, iterIdx, substepOutputFields)
			if prefixedVal != val {
				config[key] = prefixedVal
			}
		}
	}
}

// prefixDataReference prefixes a data reference if it starts with a substep output field
// e.g., "generated_content.result" → "generated_content_0.result" (if generated_content is a substep output)
func prefixDataReference(ref string, loopName string, iterIdx int, substepOutputFields map[string]bool) string {
	if ref == "" {
		return ref
	}

	// Split on "." to get the first part (the field name)
	parts := strings.SplitN(ref, ".", 2)
	fieldName := parts[0]

	// Check if this field name is a substep output field
	if substepOutputFields[fieldName] {
		// Prefix the field name with iteration index
		prefixedField := fmt.Sprintf("%s_%d", fieldName, iterIdx)
		if len(parts) > 1 {
			return prefixedField + "." + parts[1]
		}
		return prefixedField
	}

	// Not a substep output, return as-is
	return ref
}

// deepCloneConfig creates a deep copy of a config map
func deepCloneConfig(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return make(map[string]interface{})
	}

	clone := make(map[string]interface{})
	for k, v := range config {
		switch val := v.(type) {
		case map[string]interface{}:
			clone[k] = deepCloneConfig(val)
		case []interface{}:
			clone[k] = deepCloneSlice(val)
		default:
			clone[k] = v
		}
	}
	return clone
}

// deepCloneSlice creates a deep copy of a slice
func deepCloneSlice(slice []interface{}) []interface{} {
	if slice == nil {
		return nil
	}
	clone := make([]interface{}, len(slice))
	for i, v := range slice {
		switch val := v.(type) {
		case map[string]interface{}:
			clone[i] = deepCloneConfig(val)
		case []interface{}:
			clone[i] = deepCloneSlice(val)
		default:
			clone[i] = v
		}
	}
	return clone
}

// checkCircuitBreaker checks for runaway execution and returns an error if detected
func (s *SagaCoordinator) checkCircuitBreaker(state *OrchestrationState, logger *zap.Logger) error {
	// Increment counter
	state.StepExecutionCount++

	// Check absolute limit
	if state.StepExecutionCount > MaxStepExecutions {
		logger.Error("CIRCUIT BREAKER: Step execution limit exceeded",
			zap.Int("step_count", state.StepExecutionCount),
			zap.Int("limit", MaxStepExecutions),
			zap.String("current_step", state.CurrentStep),
		)
		return fmt.Errorf("CIRCUIT_BREAKER: exceeded %d step executions at step '%s' - possible infinite loop",
			MaxStepExecutions, state.CurrentStep)
	}

	// Track recent steps for cycle detection
	if state.RecentSteps == nil {
		state.RecentSteps = make([]string, 0, RecentStepsWindow)
	}

	state.RecentSteps = append(state.RecentSteps, state.CurrentStep)
	if len(state.RecentSteps) > RecentStepsWindow {
		state.RecentSteps = state.RecentSteps[1:]
	}

	// Check for cycles once we have enough history
	if len(state.RecentSteps) >= RecentStepsWindow {
		if cycleLen := detectStepCycle(state.RecentSteps); cycleLen > 0 {
			logger.Error("CIRCUIT BREAKER: Infinite loop detected",
				zap.Int("cycle_length", cycleLen),
				zap.Strings("recent_steps", state.RecentSteps),
				zap.String("current_step", state.CurrentStep),
				zap.Int("total_steps", state.StepExecutionCount),
			)
			return fmt.Errorf("CIRCUIT_BREAKER: infinite loop detected (cycle of %d steps) at step '%s'",
				cycleLen, state.CurrentStep)
		}
	}

	// Log warning at 50% of limit
	if state.StepExecutionCount == MaxStepExecutions/2 {
		logger.Warn("Circuit breaker warning: 50% of step limit reached",
			zap.Int("step_count", state.StepExecutionCount),
			zap.Int("limit", MaxStepExecutions),
		)
	}

	return nil
}

// detectStepCycle checks if the recent steps form a repeating cycle
func detectStepCycle(steps []string) int {
	n := len(steps)
	if n < 6 {
		return 0
	}

	// Check for cycles of length 2, 3, 4, 5 (repeated at least twice)
	for cycleLen := 2; cycleLen <= 5; cycleLen++ {
		if n < cycleLen*2 {
			continue
		}

		// Check if the last cycleLen*2 steps form a repeating pattern
		isCycle := true
		for i := 0; i < cycleLen; i++ {
			if steps[n-1-i] != steps[n-1-i-cycleLen] {
				isCycle = false
				break
			}
		}

		if isCycle {
			return cycleLen
		}
	}

	return 0
}
