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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	perrors "github.com/gqls/agentchassis/platform/errors"
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
	ErrWaitingForResponse   = errors.New("orchestration is waiting for responses")
	ErrVersionMismatch      = errors.New("optimistic lock failure: version mismatch")
	ErrLoopExpansionHandled = errors.New("loop expansion handled: outer continueExecution must not continue")
	// ErrTakeoverLost: the row this caller judged stale from its SNAPSHOT was no
	// longer stale on the FRESH read inside the claim — another actor resumed it,
	// or it moved on of its own accord. The caller returns having executed
	// nothing (bugs_open/329).
	ErrTakeoverLost = errors.New("stale-orchestration takeover lost: the row is being driven")
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

	// Ownership: TAKE OVER, never discard (bugs_open/075).
	//
	// This used to return nil on a pod-name mismatch, which under at-least-once
	// consume commits the offset — the response was gone for good. Two facts make
	// that unconditional loss rather than a safety check: processing_node is
	// stamped once at row creation and never refreshed (so a dead pod's name
	// never matches any living consumer again), and a response is delivered to
	// exactly one member of whichever consumer group holds it — this pod. Since
	// F2 gave expired requests a durable retry driver, the strand also stopped
	// being silent: the step was re-executed every ~3 minutes for ever, with real
	// external side effects (a GitHub commit) per cycle.
	//
	// Single-processor safety does not come from this stamp. It comes from
	// ClaimAwaitedRequest above — exactly one actor claims any given request —
	// and from UpdateStateWithVersion's optimistic lock. So we re-stamp the row
	// to ourselves for the audit trail and process the response either way.
	if state.ProcessingNode != "" && state.ProcessingNode != s.podName {
		previousPod := state.ProcessingNode
		won, takeErr := repo.TakeOverOrchestration(ctx, state.OrchestrationID, s.podName, previousPod)
		switch {
		case takeErr != nil:
			contextLogger.Error("ORCHESTRATION_TAKEOVER_FAILED: applying the response anyway rather than losing it",
				zap.String("orchestration_id", state.OrchestrationID),
				zap.String("previous_pod", previousPod),
				zap.String("my_pod", s.podName),
				zap.Error(takeErr))
		case won:
			contextLogger.Warn("ORCHESTRATION_TAKEN_OVER: driving an orchestration stamped to another pod",
				zap.String("orchestration_id", state.OrchestrationID),
				zap.String("previous_pod", previousPod),
				zap.String("my_pod", s.podName))
		default:
			contextLogger.Warn("ORCHESTRATION_TAKEOVER_RACED: another pod took the handover first; applying this response anyway",
				zap.String("orchestration_id", state.OrchestrationID),
				zap.String("previous_pod", previousPod),
				zap.String("my_pod", s.podName))
		}
		state.ProcessingNode = s.podName
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
		return s.handleRecoverableError(ctx, state, requestID, execCtx, response, awaitedReq)
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

				// Request was claimed but NOT processed (processed_at is nil).
				// Two different causes share that shape, and only one may be reset:
				//   - the claimer died between claim and process → reset and re-claim
				//   - a duplicate delivery of this response raced a LIVE claim made
				//     milliseconds ago → resetting steals the claim and the step's
				//     side effects run twice (chassis_replica_scaling CS-1; the
				//     hazard 075 deliberately left unfixed)
				// The staleness predicate lives inside the UPDATE so the decision
				// is atomic with the reset. A live 'processing' claim spans only
				// parse → state save → mark complete (handleCompleteResponse marks
				// processed BEFORE continueExecution), so 2 minutes is far beyond
				// any live window. 'retrying'/'expired' still reset immediately —
				// the F2 late-response path depends on that.
				reset, resetErr := repo.ResetStaleAwaitedRequestForRetry(ctx, requestID, claimRecoveryStaleness)
				if resetErr != nil {
					contextLogger.Error("Failed to reset awaited request for retry",
						zap.String("request_id", requestID),
						zap.Error(resetErr))
					// Continue to next retry attempt anyway
					continue
				}
				if !reset {
					contextLogger.Info("CLAIM_RECOVERY_STALENESS_HELD: live claim in progress, treating as duplicate delivery",
						zap.String("request_id", requestID),
						zap.String("current_status", status),
						zap.String("my_pod", podName))
					return nil, nil
				}
				contextLogger.Warn("CLAIM_RECOVERY: request was claimed but not processed, reset for retry",
					zap.String("request_id", requestID),
					zap.String("current_status", status),
					zap.String("my_pod", podName))
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
    RETURNING ` + awaitedRequestColumns + `
`

	record, err := scanAwaitedRequestRow(r.db.QueryRowContext(ctx, query, requestID, claimerPodName).Scan)
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

	// MISROUTED DELIVERY (bugs_open/129, defence in depth). We are on the REQUEST
	// path — ExecuteWorkflow, not ProcessResponse — so an inbound request whose
	// request_id this orchestration is ITSELF awaiting cannot be a resumption of
	// this orchestration: it is the awaited work, delivered against the waiter's
	// row instead of the worker's. Every branch below would decline it
	// (AWAITING_RESPONSES returns ErrWaitingForResponse, FAILED returns nil) and
	// the caller would log "ProcessMessage completed successfully" while nothing
	// ran and nobody replied — six minutes of silence, then the parent times out.
	//
	// The upstream cause was the retry path reconstructing the message with this
	// orchestration's own id; that is fixed in handleRecoverableError. This check
	// is what keeps the failure loud and attributable if any other sender ever
	// gets the identity wrong: it can no longer be reported as success.
	if execCtx.RequestID != "" {
		if awaited, isAwaited := state.AwaitedRequests[execCtx.RequestID]; isAwaited {
			s.logger.Error("MISROUTED_REQUEST: an inbound request was delivered against the orchestration that is awaiting it — refusing to report success",
				zap.String("orchestration_id", state.OrchestrationID),
				zap.String("request_id", execCtx.RequestID),
				zap.String("awaited_step_name", awaited.StepName),
				zap.String("target_agent_type", awaited.TargetAgentType),
				zap.Int("retry_version", execCtx.RetryVersion),
				zap.String("state_status", string(state.Status)))
			return fmt.Errorf("misrouted request %s: delivered with orchestration_id %s, which is the orchestration awaiting it (step %q)",
				execCtx.RequestID, state.OrchestrationID, awaited.StepName)
		}
	}

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
		// The SNAPSHOT saying "stuck" is a reason to TRY a takeover, not to perform
		// one. takeOverStaleOrchestration re-judges the FRESH row inside the
		// version-CAS and resumes only if it wins the claim (bugs_open/329). This
		// used to be ClearExecutingStep + reload + continueExecution — three
		// separate reads, none of which re-tested the predicate.
		if state.CurrentlyExecuting != nil && time.Since(state.LastActivity) > StuckOrchestrationTimeout {
			return s.takeOverStaleOrchestration(ctx, repo, state, execCtx, StatusExecutingStep)
		}

		s.logger.Info("Orchestration is actively executing")
		return nil

	case StatusRunning:
		// RUNNING is the inter-step GAP, not a resting state. The EXECUTING_STEP
		// takeover just above sets it (inside its claim, since bugs_open/329 —
		// it was ClearExecutingStep before) immediately before continueExecution's
		// loop re-marks the step as executing, so a healthy row occupies RUNNING
		// for milliseconds. A message arriving inside that window belongs to a pod
		// that is already resuming this orchestration, and resuming it here too
		// would double-execute the step — so the safe default is to leave it alone
		// and take over only once the row is demonstrably stale, mirroring the
		// StatusExecutingStep arm above.
		//
		// Before bugs_closed/294 there was no case for this status at all, so it
		// fell through to the default arm and returned "unknown orchestration
		// status: RUNNING" — the one in-process path that could have rescued a
		// stalled row rejected it instead. The reaper's invariant arm (migration
		// 465) now bounds such rows at 4h whatever happens; this recovers them in
		// seconds when a message does arrive, rather than erroring.
		// Claimed, not assumed (bugs_open/329). This arm previously wrote NOTHING
		// at all before resuming, so two arrivals seconds apart both resumed.
		if time.Since(state.LastActivity) > StuckOrchestrationTimeout {
			return s.takeOverStaleOrchestration(ctx, repo, state, execCtx, StatusRunning)
		}

		s.logger.Info("Orchestration is between steps (RUNNING) - another process is resuming it",
			zap.String("current_step", state.CurrentStep))
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

// takeOverStaleOrchestration claims a row an arm judged stale from its SNAPSHOT and
// then resumes it — or returns having done NOTHING, when another actor holds it.
//
// nil on a lost claim, not an error: the message was for an orchestration somebody
// else is driving, which is exactly the disposition of the arms' own non-stale
// branch. Returning an error here would turn a normal outcome into a retry.
//
// STALE_TAKEOVER_CLAIMED / STALE_TAKEOVER_LOST are the field meters. Before this
// there was no durable needle at all, which is why "have the arms ever fired?" could
// only be answered inside a pod's live log window (bugs_open/329).
func (s *SagaCoordinator) takeOverStaleOrchestration(ctx context.Context, repo *StateRepository,
	snapshot *OrchestrationState, execCtx *types.ExecutionContext, from OrchestrationStatus) error {

	claimed, err := repo.ClaimStaleOrchestration(ctx, snapshot.OrchestrationID, from, StuckOrchestrationTimeout, s.podName)
	if errors.Is(err, ErrTakeoverLost) {
		s.logger.Warn("STALE_TAKEOVER_LOST: stale on the snapshot, driven on the fresh row — not resuming",
			zap.String("orchestration_id", snapshot.OrchestrationID),
			zap.String("snapshot_status", string(from)),
			zap.Duration("snapshot_idle", time.Since(snapshot.LastActivity)))
		return nil
	}
	if err != nil {
		return fmt.Errorf("stale takeover claim for %s: %w", snapshot.OrchestrationID, err)
	}

	s.logger.Warn("STALE_TAKEOVER_CLAIMED: resuming an orchestration idle past the stuck threshold",
		zap.String("orchestration_id", claimed.OrchestrationID),
		zap.String("from_status", string(from)),
		zap.String("current_step", claimed.CurrentStep),
		zap.Int("claimed_version", claimed.Version))

	return s.continueExecution(ctx, claimed, execCtx)
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

	// Fail orchestrations that have been running far longer than their timeout.
	// After a pod restart, the CronJob may have cleaned up job.* Kafka topics.
	// Resuming these orchestrations just wastes cycles on "Unknown Topic" errors.
	// Uses the workflow's own timeout (3x multiplier for safety) or a 60-minute
	// fallback for workflows with no timeout configured.
	if !state.CreatedAt.IsZero() {
		age := time.Since(state.CreatedAt)
		var maxAge time.Duration
		if state.WorkflowPlan.TimeoutSeconds > 0 {
			maxAge = time.Duration(state.WorkflowPlan.TimeoutSeconds) * time.Second * 3
		} else {
			maxAge = 60 * time.Minute // fallback for workflows without timeout_seconds
		}
		if age > maxAge {
			s.logger.Warn("Orchestration exceeded max age — likely stale after pod restart",
				zap.String("orchestration_id", state.OrchestrationID),
				zap.String("owner_agent_type", state.OwnerAgentType),
				zap.Duration("age", age),
				zap.Duration("max_age", maxAge),
				zap.Int("workflow_timeout_seconds", state.WorkflowPlan.TimeoutSeconds),
				zap.String("current_step", state.CurrentStep))
			return s.failWorkflow(ctx, state,
				fmt.Sprintf("Orchestration stale — running for %s (workflow timeout: %ds, max age: %s)",
					age.Round(time.Second), state.WorkflowPlan.TimeoutSeconds, maxAge))
		}
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
			// Loop expansion recursed into continueExecution and took ownership.
			// The outer for-loop must not continue or transition — return cleanly.
			if errors.Is(err, ErrLoopExpansionHandled) {
				l.Info("Loop expansion handled — outer continueExecution exiting",
					zap.String("loop_step", state.CurrentStep),
				)
				return nil
			}
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
// defaultLocalActionTimeout bounds how long a LOCAL action may execute.
//
// OWNER DIRECTIVE 2026-08-03: every step gets a limit, and the default must be
// "very generous so we never cut off something still running". This constant is
// therefore sized from the LONGEST LEGITIMATE ACTION EVER OBSERVED, not from a
// percentile — a p99 is the wrong statistic when the requirement is "never cut
// off real work", because the whole risk lives in the tail it discards.
//
// Measured over orchestration_state_audit, 2026-07-31..08-02, 91,210 local
// (non-call_) step executions:
//
//	all non-call_ steps      p95 6s     p99 22s    max 1391s
//	local LLM (execute_llm_prompt, council seats)  p95 68s  p99 103s  max 281s
//	med_scrape_prices        p95 1372s  p99 1391s  max 1391s   <- the ceiling
//
// The longest legitimate local action in the fleet is `med_scrape_prices` at
// ~1391s (23 minutes). 7200s is ~5.2x that — an action would have to run five
// times longer than anything ever recorded before this could touch it, while
// still halving the 4-hour hang that motivated the change (bugs_closed/169).
//
// An earlier revision of this file used 600s, chosen from the p99.9. That was
// WRONG and would have killed med_scrape_prices on every run: the percentile was
// computed over a population dominated by sub-second steps, so the one action
// that legitimately takes twenty minutes vanished into the tail. Recorded here
// because the mistake is easy to repeat — see WRONG_CALLS.md 2026-08-03.
const defaultLocalActionTimeout = 7200 * time.Second

// localActionSlowFraction is the share of the deadline an action may consume
// before it is reported, even when it SUCCEEDS.
//
// This is what keeps the constant above honest. A generous default is only safe
// while nothing is quietly creeping toward it; without this the first evidence
// that 7200s had become too tight would be a working pipeline being cut off.
// At 0.5 an action that takes over an hour is named in the logs while still
// completing normally — early warning instead of a post-mortem.
const localActionSlowFraction = 0.5

// localActionTimeoutKey is a NEW step-config key, deliberately NOT `timeout_seconds`.
//
// Measured before choosing (bugs_open/169): of the live steps carrying
// `timeout_seconds`, 53 of 64 are `call_agent`, and most of the rest are waiting
// semantics too (`await_approval` carries 86400, `request_human_input`,
// `dispatch_thunder_*`). That key means "how long to wait for something EXTERNAL",
// not "how long this action may execute". Reusing it would make one shared word
// mean two different things depending on the step it sits on — the defect class
// RFC 006 was decided on. So the two stay separate.
const localActionTimeoutKey = "local_action_timeout_seconds"

// localActionContext derives the deadline a local action executes under.
//
// Returns the step's override when `local_action_timeout_seconds` is present, and
// the generous default otherwise. A value <= 0 DISABLES the bound for that step —
// the deliberate escape hatch for an action that genuinely must run unbounded,
// so that discovering such a case is a config change rather than a rebuild.
//
// The bound is a context deadline and NOT "run it in a goroutine and abandon it".
// Abandoning would regain control even from an action that ignores ctx, but it
// leaks the goroutine, and a late write from an abandoned action into a step the
// coordinator has already failed is a worse hazard than the hang. The plausible
// blockers here — database/sql, the Kafka producer's ProduceWithValidation(ctx,…),
// the K8s client — are all ctx-aware, so a deadline actually cuts them.
// disableLocalActionTimeoutEnv is the fleet-wide kill switch: set
// DISABLE_LOCAL_ACTION_TIMEOUT=true and every local action runs unbounded again,
// exactly as before this change.
//
// It exists because this is a behaviour change to EVERY action, chosen from a
// measured distribution, and the failure mode of a too-tight bound is fleet-wide
// breakage — worse than the rare hang it fixes. A per-step override cannot help if
// the default turns out wrong across many steps at 03:00; this can, by config,
// without a rebuild or a roll.
//
// On the 2026-08-02 owner ruling (RFC_010) that "new authority on a shared seam
// ships as an OPT-IN FIELD with the unsafe default OFF": that ruling governs a seam
// that GRANTS authority on an assumption about callers. This one REMOVES authority
// (an action may no longer run forever) and is licensed by measurement over all
// 96,047 recorded step executions, not by an assumption. Defaulting it OFF would
// reproduce precisely the inert-by-omission defect RFC 006 was decided on — a
// protection nobody enables protects nobody. So the safe default is ON, and the
// escape hatches are explicit instead.
const disableLocalActionTimeoutEnv = "DISABLE_LOCAL_ACTION_TIMEOUT"

// stepName is passed EXPLICITLY rather than read from step.Name, because on the live
// coordinator path models.Step.Name is empty — proven in production 2026-08-03 by
// inducing this deadline on endpoint-health-checker, whose error read `on step ""`
// while the orchestration's current_step was `check_health`. The unit fixtures set
// Name and so could never have caught it. state.CurrentStep is the reliable name.
func localActionContext(ctx context.Context, step models.Step, stepName string, logger *zap.Logger) (context.Context, context.CancelFunc) {
	if os.Getenv(disableLocalActionTimeoutEnv) == "true" {
		logger.Warn("local action deadlines are DISABLED fleet-wide by env switch — actions can park an orchestration indefinitely",
			zap.String("env", disableLocalActionTimeoutEnv),
			zap.String("step", stepName),
			zap.String("action", step.Action))
		return ctx, func() {}
	}

	timeout := defaultLocalActionTimeout
	if raw, ok := step.Config[localActionTimeoutKey]; ok {
		seconds, parsed := datahelpers.ToFloat64(raw)
		if !parsed {
			logger.Warn("local action timeout override is not a number — using the default",
				zap.String("step", stepName),
				zap.String("action", step.Action),
				zap.Any("value", raw),
				zap.Duration("default", defaultLocalActionTimeout))
		} else if seconds <= 0 {
			// Explicitly unbounded. Logged at Warn every time, because an
			// unbounded action is exactly what bugs_open/169 is about and it
			// should never be quietly true of a step.
			logger.Warn("local action is explicitly UNBOUNDED for this step — it can park the orchestration indefinitely",
				zap.String("step", stepName),
				zap.String("action", step.Action),
				zap.String("config_key", localActionTimeoutKey))
			return ctx, func() {}
		} else {
			timeout = time.Duration(seconds) * time.Second
		}
	}

	// Never EXTEND an inherited deadline: if the caller is already on a shorter
	// clock, honour it. context.WithTimeout does this itself, but stating it
	// keeps the intent readable — this bounds, it does not grant time.
	return context.WithTimeout(ctx, timeout)
}

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
		zap.String("action", step.Action),
		zap.String("step", step.Name),
	)

	// 6. Execute the action, under a deadline.
	//
	// bugs_open/169 part A: nothing in continueExecution -> executeStep ->
	// executeLocalAction -> executeAction bounded this call, so an action blocking
	// on a network call parked its orchestration at EXECUTING_STEP and held the
	// build-dispatch-loop's site_work_items row in `claimed`. Measured over the
	// 3-day audit window: 6,951 spawn-step executions, p99 24s, and exactly ONE
	// above 300s — at 14,475s. The distribution is bimodal with nothing between
	// seconds and hours, which is what makes a generous bound safe.
	actionCtx, cancelAction := localActionContext(ctx, step, state.CurrentStep, contextLogger)
	actionStarted := time.Now()
	result, err := executeAction(actionCtx, handler, params, contextLogger)
	cancelAction()

	// Name the timeout where it happened. A bare "context deadline exceeded"
	// surfacing three layers up is exactly the kind of error that sends the next
	// reader looking in the wrong place — the whole point of 169 part A was that
	// nothing said which step was stuck.
	if err != nil && errors.Is(actionCtx.Err(), context.DeadlineExceeded) {
		elapsed := time.Since(actionStarted).Round(time.Millisecond)
		if elapsed >= time.Second {
			elapsed = elapsed.Round(time.Second)
		}
		stepName := state.CurrentStep
		if stepName == "" {
			stepName = step.Name
		}
		contextLogger.Error("Local action exceeded its deadline and was cancelled",
			zap.String("action", step.Action),
			zap.String("step", stepName),
			zap.Duration("elapsed", elapsed),
			zap.String("override_key", localActionTimeoutKey),
			zap.Error(err))
		err = fmt.Errorf("local action %q on step %q exceeded its deadline after %s "+
			"(set %q on the step to change it, or <=0 to disable): %w",
			step.Action, stepName, elapsed, localActionTimeoutKey, err)
	}

	// Early warning that the generous default is becoming tight. Fires on a
	// SUCCESSFUL action, so the first sign that something is creeping toward the
	// deadline is a log line rather than a cut-off pipeline.
	if err == nil {
		if deadline, ok := actionCtx.Deadline(); ok {
			budget := time.Until(deadline) + time.Since(actionStarted)
			if elapsed := time.Since(actionStarted); budget > 0 &&
				float64(elapsed) > localActionSlowFraction*float64(budget) {
				contextLogger.Warn("Local action used most of its deadline — it succeeded, but the budget is becoming tight",
					zap.String("action", step.Action),
					zap.String("step", state.CurrentStep),
					zap.Duration("elapsed", elapsed.Round(time.Second)),
					zap.Duration("budget", budget.Round(time.Second)),
					zap.String("override_key", localActionTimeoutKey))
			}
		}
	}

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

			// Continue workflow with first iteration step.
			// The recursive continueExecution drives the loop forward from here.
			// We return ErrLoopExpansionHandled so the outer continueExecution
			// for-loop exits cleanly and does not fall through to process_items.NextStep.
			contextLogger.Info("Loop expanded, continuing to first iteration",
				zap.String("next_step", state.CurrentStep),
			)
			if err := s.continueExecution(ctx, state, execCtx); err != nil {
				return err
			}
			return ErrLoopExpansionHandled
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

	// RunAgentType — rung 1 of the RSH-009 ladder, and the sibling of the Sender
	// backfill above. A step resumed after an await rebuilds execCtx from the
	// RESPONSE message's headers, which never carry it, so every consumer of the
	// ladder drops to rung 2 for the rest of the run. state.OwnerAgentType is
	// determineOwnerAgentType's own durable output from run start, so it is
	// strictly the right source (RFC_019 §7; bugs_open/093 shape: one call site
	// was guarded, its sibling was not). Accepted edge: if resolution at run start
	// bottomed out at the "generic" filler, this propagates that — the ladder's own
	// durable answer, and no worse than the Sender-based read it replaces.
	if execCtx.RunAgentType == "" && state.OwnerAgentType != "" {
		execCtx.RunAgentType = state.OwnerAgentType
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

	// Guarantee step.Config is a non-nil (possibly empty) map before
	// it reaches any action. Step.Config has `omitempty` in JSON tags
	// (pkg/models/contracts.go), so a workflow step without a "config"
	// key unmarshals to Config == nil. Reading from a nil map is
	// fine, but writing panics. Initialising here is belt-and-braces:
	// actions shouldn't mutate config (see render_css_from_spec_action
	// for the pattern of resolving values to local vars), but if any
	// do, they won't panic.
	if step.Config == nil {
		step.Config = make(map[string]interface{})
	}

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
		AgentType:       state.OwnerAgentType,
		CurrentStep:     state.CurrentStep,
		// The whole plan, not just this step: an action that tolerates a partial
		// result has to know whether anything downstream guards it
		// (bugs_open/076, findTruncationAwareConsumer).
		WorkflowSteps: state.WorkflowPlan.Steps,
	}
}

// Execute the action handler.
//
// A deferred recover() converts any panic from inside the handler into
// an error. Without this, a panic kills the processing goroutine and
// leaves the orchestration stuck in EXECUTING_STEP with no log trail
// past the panic — recovery via the reaper only kicks in after 30+
// minutes. Converting to error lets the existing handleActionError +
// error_step routing machinery fail the orchestration cleanly and
// release any associated work items.
//
// The full stack trace is logged at Error level so the root cause is
// visible in the pod logs even though the error that propagates back
// to the caller is a single-line summary.
func executeAction(ctx context.Context, handler actions.ActionFunc, params actions.ActionParams, logger *zap.Logger) (result interface{}, err error) {
	logger.Info("Calling action handler",
		zap.String("action", params.StepConfig.Action))

	defer func() {
		if r := recover(); r != nil {
			// Capture a bounded-size stack trace. 64KiB is plenty for
			// typical panics and keeps the log line size sane.
			stackBuf := make([]byte, 64*1024)
			n := runtime.Stack(stackBuf, false)
			stack := string(stackBuf[:n])

			logger.Error("PANIC recovered in action handler",
				zap.String("action", params.StepConfig.Action),
				zap.String("step_name", params.StepConfig.Name),
				zap.Any("panic_value", r),
				zap.String("stack", stack),
			)

			// Replace any (nil, nil) return from a panicking handler
			// with a real error so the caller routes to the failure
			// path. %v on the panic value covers both string panics
			// and runtime.Error panics (like the "assignment to entry
			// in nil map" that prompted this).
			result = nil
			err = fmt.Errorf("action %q panicked: %v", params.StepConfig.Action, r)
		}
	}()

	result, err = handler(ctx, params)
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

	// Set status to awaiting (must happen before persist). Capture the prior
	// status first: the failure path below must restore it, not blank it.
	priorStatus := state.Status
	state.Status = StatusAwaitingResponses

	// PERSIST STATE BEFORE TABLE INSERT to prevent race condition
	// The JSONB must be persisted before the table insert, otherwise:
	// 1. Response arrives after table insert but before JSONB persist
	// 2. Response handler loads state with empty AwaitedRequests
	// 3. Workflow incorrectly thinks all responses are in
	ctx := context.Background()
	repo := NewStateRepository(coordinator.db, logger)
	outcome, err := persistAwaitingStateWithRetry(ctx, state, repo, logger)
	if err != nil {
		logger.Error("Failed to persist state before table insert",
			zap.String("request_id", requestID),
			zap.Error(err))
		// Remove from in-memory state since we failed
		delete(state.AwaitedRequests, requestID)
		// Restore the prior status rather than blanking it. "" is not a member of
		// the status vocabulary, and this state IS persisted afterwards:
		// returning false sends control back to continueExecution, which — seeing
		// a status that is not StatusAwaitingResponses — falls through to
		// saveStepResultWithRetry. A blank status therefore reached the database.
		state.Status = priorStatus
		return false
	}

	if outcome == parkSkippedReplyArrived {
		// The reply to THIS request beat the park, so the response consumer has
		// already applied it and owns the continuation. Nothing was persisted, so
		// there is nothing to insert and nothing to time out.
		//
		// bug 343 (silent post-abandonment freeze): this path used to fall through and insert a table row and
		// arm a timeout for a request the state row knows nothing about — the
		// orphan row that made the map and the table disagree with no reconciler
		// anywhere. Drop the in-memory entry so this state cannot re-assert it.
		//
		// Still `true`: the executor must stop driving this step either way, and
		// the consumer's already-advanced state is the authoritative one.
		delete(state.AwaitedRequests, requestID)
		logger.Info("Reply beat the park - consumer owns the continuation, no awaited row inserted",
			zap.String("request_id", requestID),
			zap.String("orchestration_id", state.OrchestrationID))
		return true
	}

	// NOW insert into database table (JSONB is already persisted)
	err = repo.InsertAwaitedRequest(ctx, awaitedReq)
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

// parkOutcome says what a park actually DID, so no caller can read "no error" as
// "the awaited entry is persisted".
//
// bug 343: persistAwaitingStateWithRetry's arrival check returned a bare nil
// on a hit without persisting anything, and processAwaitResponse then inserted the
// table row and armed a timeout for a request whose entry exists in no map. Making
// the outcome part of the signature makes "returned success without persisting"
// unrepresentable — the compiler forces every caller to say which case it is in.
// It is only meaningful when the error is nil.
type parkOutcome int

const (
	// parkPersisted: the awaited entry is durably in the state row. Insert the
	// table row and arm the timeout.
	parkPersisted parkOutcome = iota
	// parkSkippedReplyArrived: a reply to THIS request landed while we were
	// parking, so the response consumer already owns the continuation. Nothing was
	// persisted and nothing should be: no table row, no timeout.
	parkSkippedReplyArrived
)

// persistAwaitingStateWithRetry saves state with awaited request before table insert
// This prevents the race condition where responses arrive before JSONB is persisted
func persistAwaitingStateWithRetry(ctx context.Context, state *OrchestrationState, repo *StateRepository, logger *zap.Logger) (parkOutcome, error) {
	maxRetries := 10
	baseDelay := 50 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Load fresh state
		freshState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return parkPersisted, fmt.Errorf("failed to load state: %w", err)
		}

		// Has a reply to one of the requests we are parking ALREADY arrived?
		//
		// The question is about IDENTITY, not presence. A marker under the step
		// name says only "some reply was recorded here once"; on a loop step that
		// is true of every iteration after the first. bug 343: keying on the
		// step name alone made every re-registration of an answered step name read
		// as an arrival, so the park returned success without persisting and the
		// orchestration froze holding no waiting awaited request.
		//
		// The race is real, but only where the row is pre-registered BEFORE the
		// send, so the consumer can claim a fast ack while the park is still
		// persisting. Five callers do that today via preRegisterAwaitedRequest:
		// actions/spawn_actions.go:115, actions/dispatch_actions.go:229 and the
		// three thunder_prepare_*_dispatch.go actions. Everywhere else a marker
		// under an awaited step name is stale by construction.
		for reqID := range state.AwaitedRequests {
			existingData, exists := freshState.CollectedData[state.AwaitedRequests[reqID].StepName].(map[string]interface{})
			if !exists {
				continue
			}
			if _, hasResponse := existingData[awaitedResponseMarker]; !hasResponse {
				continue
			}

			recordedID, hasID := existingData[awaitedResponseIDMarker].(string)
			switch {
			case hasID && recordedID == reqID:
				// A reply to THIS request beat the park. The consumer has already
				// applied it and owns the continuation.
				logger.Info("Response already arrived during state persist - continuing",
					zap.String("request_id", reqID))
				return parkSkippedReplyArrived, nil
			case hasID:
				// Stale residue: an EARLIER request answered under this same step
				// name. Park normally — this is the bug 343 mechanism, and
				// skipping here is what stranded the orchestration.
				logger.Info("Stale response marker under this step name belongs to an earlier request - parking normally",
					zap.String("request_id", reqID),
					zap.String("recorded_request_id", recordedID),
					zap.String("step_name", state.AwaitedRequests[reqID].StepName))
			default:
				// LEGACY (2026-08-21): a marker written by an image that predates
				// awaitedResponseIDMarker carries no id, so identity is
				// unrecoverable and today's behaviour — treat as arrived — is the
				// safe reading during the mixed-fleet window after the roll.
				//
				// Do NOT "fix" this branch to park: an old pod's genuine arrival
				// would then be double-driven. Delete the branch outright once no
				// pre-roll pod and no pre-roll orchestration can still be live.
				logger.Info("Response already arrived during state persist - continuing (legacy marker, no request id)",
					zap.String("request_id", reqID))
				return parkSkippedReplyArrived, nil
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

		// Carry the dispatching step's own work across. Additive — see the
		// function's comment for why this is not a merge in either direction.
		if carried := carryCollectedDataOntoFreshState(freshState, state, logger); len(carried) > 0 {
			logger.Info("Carried the dispatching step's collected_data onto the parked state",
				zap.String("orchestration_id", state.OrchestrationID),
				zap.Strings("carried_keys", carried),
				zap.Int("carried_count", len(carried)))
		}

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
			return parkPersisted, nil
		}

		// Check if optimistic lock error
		if !IsOptimisticLockError(err) {
			return parkPersisted, fmt.Errorf("failed to persist awaiting state: %w", err)
		}

		if attempt >= maxRetries {
			return parkPersisted, fmt.Errorf("failed to persist awaiting state after %d attempts: %w", attempt, err)
		}

		// Backoff
		delay := backoffWithJitter(baseDelay, attempt)
		logger.Warn("Optimistic lock failure persisting awaiting state, retrying",
			zap.Int("attempt", attempt),
			zap.Duration("backoff", delay))
		time.Sleep(delay)
	}

	return parkPersisted, nil
}

// awaitedResponseMarker is the sub-key applyResponseToState writes when an
// adapter reply lands, and the key persistAwaitingStateWithRetry's arrival check
// reads to decide a reply beat the park. It is a protocol signal, never data.
const awaitedResponseMarker = "response"

// awaitedResponseIDMarker is the sibling key recording WHICH awaited request the
// recorded reply answered. Written wherever applyResponseToState records a reply;
// read only by persistAwaitingStateWithRetry's arrival check, which needs identity,
// not mere presence.
//
// bug 343: the arrival check used to key on step NAME alone, so a step name
// that had already been answered once — every iteration N+1 re-registration of a
// loop's call_handler — read as "the reply beat the park" for a request that had
// never been sent. The park then returned success WITHOUT persisting, the caller
// read nil as success, and the orchestration sat in EXECUTING_STEP holding no
// waiting awaited request: a row in the table, nothing in the map, and nothing
// routine to notice. Recording the id is what lets the check tell a genuine
// beat-the-park arrival from stale residue under the same step name.
//
// It is a protocol signal, never data — like awaitedResponseMarker it is stripped
// on the carry path (withoutResponseMarker) so an action's own result can never
// forge it.
const awaitedResponseIDMarker = "response_request_id"

// setAwaitedResponseID records which awaited request a reply answered, on the
// container applyResponseToState is about to store under the step name.
//
// Nil-guarded and no-ops on an empty id: awaitedReq is legitimately nil on the
// dynamic-step path, and an absent id is exactly what the arrival check's legacy
// branch is written to tolerate — never write an empty string, which would read as
// a real id belonging to no request.
func setAwaitedResponseID(container map[string]interface{}, awaitedReq *AwaitedRequest) {
	if container == nil || awaitedReq == nil || awaitedReq.RequestID == "" {
		return
	}
	container[awaitedResponseIDMarker] = awaitedReq.RequestID
}

// carryCollectedDataOntoFreshState copies the dispatching step's in-memory
// CollectedData onto the freshly-loaded state, ADDITIVELY: a key already present
// on the fresh copy is left untouched. That direction is the whole safety
// argument — nothing here can overwrite a concurrent writer, and nothing here can
// overwrite a reply that landed while we were parking.
//
// Without this, parking discards everything the action computed. The park reloads
// the row from the DB and copies only AwaitedRequests/Status/LastActivity onto it,
// so storeActionResult's own step-name and output_field writes — and any sibling
// keys the action wrote — never reached the DB at all. The reply-time merge then
// "preserved" a map that had never held the action's work, and the persisted
// record ended up holding exactly the reply and nothing else. The status said
// complete and nothing recorded a loss.
//
// bugs_open/236 — mechanism confirmed 2026-08-14 (witnessed on two parked rows;
// ordering verified at processActionResult, which calls storeActionResult before
// processAwaitResponse with the same state). RFC_012 question (a), owner ruling
// 2026-08-15: fix additively, at the park path.
//
// Returns the carried keys, sorted, for the caller's log line and for tests.
func carryCollectedDataOntoFreshState(freshState, state *OrchestrationState, logger *zap.Logger) []string {
	if freshState == nil || state == nil || len(state.CollectedData) == 0 {
		return nil
	}
	if freshState.CollectedData == nil {
		freshState.CollectedData = make(map[string]interface{}, len(state.CollectedData))
	}

	// The steps we are about to park on. Carrying a key spelled "response" — or,
	// since bug 343, its "response_request_id" sibling — under one of these
	// would forge the signal the arrival check reads, and a forged arrival is
	// indistinguishable from a real one. The id is the MORE dangerous of the two
	// to carry: an id equal to the request being parked forges precisely the
	// identity the check now keys on.
	awaitedSteps := make(map[string]struct{}, len(state.AwaitedRequests))
	for _, req := range state.AwaitedRequests {
		if req != nil && req.StepName != "" {
			awaitedSteps[req.StepName] = struct{}{}
		}
	}

	carried := make([]string, 0, len(state.CollectedData))
	for key, value := range state.CollectedData {
		if _, taken := freshState.CollectedData[key]; taken {
			continue
		}
		if _, awaited := awaitedSteps[key]; awaited {
			value = withoutResponseMarker(key, value, logger)
		}
		freshState.CollectedData[key] = value
		carried = append(carried, key)
	}
	sort.Strings(carried)
	return carried
}

// withoutResponseMarker returns value with BOTH response-protocol markers removed
// (awaitedResponseMarker and awaitedResponseIDMarker), copying rather than
// mutating because the map handed in is still the live in-memory one the caller
// goes on using.
//
// No action returns such a key today — measured 2026-08-15 across the 15 files
// that return await_response: every "response" spelling in them is either a
// reader of an arrived reply or a nested topics map, never a top-level result
// key. This guards the next action, not a live case.
func withoutResponseMarker(stepName string, value interface{}, logger *zap.Logger) interface{} {
	asMap, ok := value.(map[string]interface{})
	if !ok {
		return value
	}
	_, hasMarker := asMap[awaitedResponseMarker]
	_, hasIDMarker := asMap[awaitedResponseIDMarker]
	if !hasMarker && !hasIDMarker {
		return value
	}

	stripped := make(map[string]interface{}, len(asMap))
	for k, v := range asMap {
		if k == awaitedResponseMarker || k == awaitedResponseIDMarker {
			continue
		}
		stripped[k] = v
	}
	if logger != nil {
		logger.Warn("Dropped a response-protocol marker from an awaited step's own result before parking - it would be indistinguishable from an arrived reply",
			zap.String("step_name", stepName),
			zap.Bool("had_response", hasMarker),
			zap.Bool("had_response_request_id", hasIDMarker))
	}
	return stripped
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
		RequestPayload:  extractRetryPayload(result),
	}
}

// extractRetryPayload lifts the message the action actually produced out of its
// result, so a timeout can REPLAY it rather than reconstruct one from the
// awaiting orchestration's state (bugs_open/129). Absent is a valid state and is
// handled at retry time — the coordinator refuses to retry rather than emitting
// a message carrying its own orchestration_id.
func extractRetryPayload(result map[string]interface{}) json.RawMessage {
	raw, ok := result[types.RetryPayloadKey]
	if !ok || raw == nil {
		return nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	return json.RawMessage(encoded)
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

// reconcileAllDoneAgainstTable cross-checks a map-derived "all responses are in"
// against the awaited_requests table, and returns the allDone the caller should
// act on.
//
// Detection is UNCONDITIONAL: a disagreement is logged loudly and recorded in
// agent_error_log whatever the flag says, because the whole defect class
// bug 343 sits in is invisible today — the advance decision reads the JSONB
// map alone and the table is never consulted, so a divergence becomes a wrong
// decision with nothing written down anywhere.
//
// ENFORCEMENT — adopting the table's rows and re-parking — happens only under
// WorkflowPlan.AwaitReconcileEnforce. With the flag off this function cannot
// change the outcome: it returns allDone unchanged. See the field's comment for
// the two hazards that keep the unsafe side off by default.
//
// Best-effort in the failure direction: a query error warns and lets the caller
// proceed exactly as today. A cross-check must never become a new way to fail.
func (s *SagaCoordinator) reconcileAllDoneAgainstTable(ctx context.Context, repo *StateRepository, freshState *OrchestrationState, requestID string, allDone bool) bool {
	if !allDone || repo == nil || freshState == nil {
		return allDone
	}

	outstanding, err := repo.OutstandingAwaitedRequests(ctx, freshState.OrchestrationID, requestID)
	if err != nil {
		s.logger.Warn("Await reconcile cross-check failed - proceeding on the in-memory map alone",
			zap.String("orchestration_id", freshState.OrchestrationID),
			zap.Error(err))
		return allDone
	}
	if len(outstanding) == 0 {
		return allDone
	}

	// The map says done; the table disagrees.
	ids := make([]string, 0, len(outstanding))
	statuses := make([]string, 0, len(outstanding))
	for _, req := range outstanding {
		if req == nil {
			continue
		}
		ids = append(ids, req.RequestID)
		statuses = append(statuses, req.Status)
	}

	enforce := freshState.WorkflowPlan.AwaitReconcileEnforce
	s.logger.Error("AWAIT_DIVERGENCE_DETECTED: the awaited map says all responses are in, the table still shows outstanding rows",
		zap.String("orchestration_id", freshState.OrchestrationID),
		zap.String("current_step", freshState.CurrentStep),
		zap.String("completing_request_id", requestID),
		zap.Int("outstanding_count", len(outstanding)),
		zap.Strings("outstanding_request_ids", ids),
		zap.Strings("outstanding_statuses", statuses),
		zap.Bool("enforced", enforce))

	s.logAgentError(ctx, AgentErrorEntry{
		OrchestrationID: freshState.OrchestrationID,
		StepName:        freshState.CurrentStep,
		Action:          "await_reconcile",
		ErrorMessage: fmt.Sprintf("await divergence: map says all done, awaited_requests still shows %d outstanding (%v)",
			len(outstanding), ids),
		ErrorCode: "await_divergence",
		Severity:  "error",
		Context: map[string]interface{}{
			"outstanding_request_ids": ids,
			"outstanding_statuses":    statuses,
			"completing_request_id":   requestID,
			"enforced":                enforce,
			"bug":                     "bug 343",
		},
	})

	if !enforce {
		// Detection only. The decision stays byte-identical to today.
		return allDone
	}

	// Adopt the table's view: re-park on the rows it still shows outstanding
	// rather than advancing past work this orchestration still owes. No timeout
	// goroutines are armed here — the retry driver already owns expired rows, so
	// an adopted request that never answers ends at the retry ladder and then the
	// fail path: recoverable-but-slow, never silent.
	if freshState.AwaitedRequests == nil {
		freshState.AwaitedRequests = make(map[string]*AwaitedRequest, len(outstanding))
	}
	for _, req := range outstanding {
		if req == nil || req.RequestID == "" {
			continue
		}
		freshState.AwaitedRequests[req.RequestID] = req
	}
	freshState.Status = StatusAwaitingResponses
	freshState.LastActivity = time.Now()
	s.logger.Warn("await_reconcile_enforce is on - adopted the table's outstanding requests instead of advancing",
		zap.String("orchestration_id", freshState.OrchestrationID),
		zap.Strings("adopted_request_ids", ids))

	return false
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

	// Adapter-reported conditions (e.g. the image adapter's unrouted-kind
	// guard) become durable agent_error_log rows here — the one point every
	// complete response crosses, whichever workflow sent it. Best-effort,
	// once per response (outside the optimistic-lock retry loop below).
	// See agent_error_log.go for the contract (bugs_open/011 §4 residual).
	s.persistReportedConditions(ctx, state, stepName, response, normalisedData)

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
			// The map says the orchestration owes nothing more. Ask the TABLE the
			// same question before acting on it — this is the one moment the two
			// representations' disagreement becomes a wrong decision, and until
			// bug 343 nothing ever compared them.
			allDone = s.reconcileAllDoneAgainstTable(ctx, repo, freshState, requestID, allDone)
		}
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
		// Record WHICH request this answered. This branch wrote no arrival marker
		// at all before bug 343, so a reply that beat the park on an
		// output_mapping step was invisible to the check — and output_mapping IS
		// live on call_agent await paths (107_image_build_handler.sql:589
		// call_variant_gen, :1119 call_imagery_gen). One write covers stepName and
		// outputField below: they share this map reference.
		setAwaitedResponseID(mappedResult, awaitedReq)

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

		existingData[awaitedResponseMarker] = normalisedData
		existingData["response_received_at"] = time.Now().Format(time.RFC3339)
		existingData["response_status"] = "complete"
		setAwaitedResponseID(existingData, awaitedReq)

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
			outputData[awaitedResponseMarker] = normalisedData
			outputData["response_received_at"] = time.Now().Format(time.RFC3339)
			outputData["response_status"] = "complete"
			setAwaitedResponseID(outputData, awaitedReq)
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

	// Handle spawn_agent without existing data.
	//
	// Unreachable at HEAD: the isAgentResponse test above already claims every
	// stepExists && spawn_agent case and returns. The id marker is written here
	// anyway so the branch cannot come back to life marker-blind — the shape of
	// hole bug 343 found in the output_mapping branch.
	if stepExists && step.Action == "spawn_agent" {
		spawnData := s.extractSpawnData(normalisedData, step)
		setAwaitedResponseID(spawnData, awaitedReq)
		state.CollectedData[stepName] = spawnData
		if step.OutputField != "" {
			state.CollectedData[step.OutputField] = spawnData
		}
		return
	}

	// Default: store response directly (for non-agent actions). Adapter and HITL
	// await paths land here and wrote no arrival marker before bug 343.
	setAwaitedResponseID(normalisedData, awaitedReq)
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

// maxStepRetries caps how many times one step may be retried inside a single
// orchestration. The message path enforces it through awaited_requests.retry_version;
// the adapter-action path (which mints a fresh request per attempt) enforces it
// through execution_metadata.retry_count — see nextAdapterRetryAttempt.
const maxStepRetries = 3

// nextAdapterRetryAttempt returns the attempt number a re-execution of stepName
// would be, and whether the cap has already been reached (bugs_open/075).
//
// It takes the HIGHER of the durable per-step counter and the request's own
// retry_version, so neither counter can be walked round: retry_version leads on
// the message path (the same request is resent and incremented), the map leads
// on the adapter path (each attempt is a new request pinned at 0).
//
// The defect this replaces: `RetryCount[step] = awaited.RetryVersion + 1` — an
// assignment, not an increment. Against a fresh adapter request (rv always 0)
// it wrote 1 on every cycle, so no accumulation ever happened and a genuinely
// failing adapter step retried for ever.
func nextAdapterRetryAttempt(retryCounts map[string]int, stepName string, retryVersion int) (attempt int, capped bool) {
	attempts := retryCounts[stepName]
	if retryVersion > attempts {
		attempts = retryVersion
	}
	if attempts >= maxStepRetries {
		return attempts, true
	}
	return attempts + 1, false
}

// handleRecoverableError handles errors that can be retried. awaited is the
// row the CALLER claimed — the claim's RETURNING carries the current DB row,
// request_payload included, and the claim is exclusive, so no re-read can be
// fresher. This function used to re-read the row here instead, but the
// response path's claim sets status='processing', which the re-read predicate
// ('waiting','retrying') excludes — so every cross-pod response-driven retry
// fell to an in-memory fallback whose request_payload is never populated
// (json:"-", bugs_closed/129) and died at RETRY_PAYLOAD_UNAVAILABLE
// milliseconds after bumping retry_version (bugs_open/216).
func (s *SagaCoordinator) handleRecoverableError(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response types.ResponseMessage, awaited *AwaitedRequest) error {
	s.logger.Warn("Recoverable error received",
		zap.String("request_id", requestID),
		zap.Int("retry_version_from_execCtx", execCtx.RetryVersion),
	)

	repo := NewStateRepository(s.db, s.logger)
	if awaited == nil {
		return fmt.Errorf("no awaited request found for %s", requestID)
	}

	s.logger.Info("Retry decided on the claimed awaited row passed through from the claim",
		zap.String("request_id", requestID),
		zap.Int("retry_version_from_claim", awaited.RetryVersion),
		zap.Bool("payload_present", len(awaited.RequestPayload) > 0),
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
		// Cap the re-executions BEFORE doing any of them (bugs_open/075). The
		// retry_version check at the top of this function cannot see an adapter
		// loop: re-executing the step mints a BRAND NEW awaited request and
		// createAwaitedRequest hardcodes RetryVersion: 0, so every cycle read 0
		// and the >= 3 cap was unreachable. The durable counter is
		// execution_metadata.retry_count, persisted with the state below, so it
		// survives the pod death F2's retry driver exists to recover from.
		if state.ExecutionMetadata.RetryCount == nil {
			state.ExecutionMetadata.RetryCount = make(map[string]int)
		}
		attempt, capped := nextAdapterRetryAttempt(state.ExecutionMetadata.RetryCount, awaited.StepName, awaited.RetryVersion)
		if capped {
			s.logger.Error("ADAPTER_RETRY_CAP_REACHED: step re-executed the maximum number of times, failing instead of looping",
				zap.String("request_id", requestID),
				zap.String("step_name", awaited.StepName),
				zap.Int("attempts", attempt),
				zap.Int("max_step_retries", maxStepRetries))
			// Release the claim terminally first, so no downstream failure can
			// leave the row parked in 'retrying' — same order as the exhaustion
			// path in retryExpiredAwaitedRequest. Guarded on 'retrying' in SQL,
			// so it is a no-op when we arrived here holding a 'processing' claim.
			if failErr := repo.MarkAwaitedRequestFailed(ctx, requestID); failErr != nil {
				s.logger.Warn("Failed to mark awaited request failed at the adapter retry cap",
					zap.String("request_id", requestID), zap.Error(failErr))
			}
			return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
		}

		s.logger.Info("Re-executing adapter action for retry",
			zap.String("step_name", awaited.StepName),
			zap.String("request_topic", awaited.RequestsTopic),
			zap.Int("retry_attempt", attempt))

		// Remove the failed awaited request
		delete(state.AwaitedRequests, requestID)

		// Mark it processed in DB so it doesn't linger
		if err := repo.MarkAwaitedRequestComplete(ctx, requestID); err != nil {
			s.logger.Warn("Failed to mark awaited request complete for re-execution",
				zap.String("request_id", requestID),
				zap.Error(err))
		}

		// Track retry count in execution metadata. INCREMENT (via the attempt
		// computed above), never assign: the old line here was
		// `= awaited.RetryVersion + 1`, which on a fresh rv=0 request wrote 1
		// every cycle and made the cap unreachable (bugs_open/075).
		state.ExecutionMetadata.RetryCount[awaited.StepName] = attempt

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
			zap.Int("retry_attempt", attempt))

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

	// A retry waits the window the STEP DECLARED, not a recomputed one
	// (bugs_open/029). See retryWindow — the block that stood here capped every
	// retry at 5 minutes, and dropped to 3 minutes for any step declaring more
	// than 30, so the longer a step declared the LESS it was given.
	newTimeout := retryWindow(state, awaited, s.logger)

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

	// A RETRY IS A REPLAY OF THE ORIGINAL REQUEST — never a reconstruction
	// (bugs_open/129). This block used to synthesise a fresh RequestMessage from
	// the AWAITING orchestration's own state:
	//
	//     OrchestrationID: state.OrchestrationID   // the PARENT's id
	//     Action:          "execute"               // never the original action
	//     Body:            {"is_retry": true, …}   // the payload, gone
	//
	// which is how the spawned child came to load the PARENT's row, find it at
	// AWAITING_RESPONSES, decline the work and log "ProcessMessage completed
	// successfully" while never replying. Measured on the live database
	// 2026-07-28: all 430 retried awaited_requests of the previous 14 days went
	// out this way, and 294 of them exhausted the retry budget.
	//
	// The identity, action and body now all come from the message that was
	// actually sent, and the Kafka headers are regenerated from that same message
	// so the two cannot disagree.
	payload, err := types.DecodeRetryPayload(awaited.RequestPayload)
	if err != nil {
		// No recorded payload: the only messages we could build here are ones we
		// know to be wrong. Fail the request with a named, greppable error rather
		// than emitting one that will be silently swallowed by whatever receives
		// it. Every sender of an awaited request is expected to record its
		// payload — this counts the ones that do not.
		s.logger.Error("RETRY_PAYLOAD_UNAVAILABLE: refusing to synthesise a retry (would carry this orchestration's own id)",
			zap.String("request_id", requestID),
			zap.String("step_name", awaited.StepName),
			zap.String("target_agent_type", awaited.TargetAgentType),
			zap.String("requests_topic", awaited.RequestsTopic),
			zap.Error(err))
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	}

	retryHeaders, retryBytes, replayed, err := payload.ReplayRequest(awaited.RetryVersion, int(newTimeout.Seconds()))
	if err != nil {
		s.logger.Error("RETRY_PAYLOAD_UNDECODABLE: stored request payload could not be replayed",
			zap.String("request_id", requestID),
			zap.String("step_name", awaited.StepName),
			zap.Error(err))
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	}

	// The invariant, asserted rather than assumed: a request this orchestration
	// sends OUT must never carry this orchestration's own id, or the receiver
	// resolves our row instead of its own and declines the work. Cheap, and it
	// holds against any future sender that records a payload incorrectly.
	if replayed.Headers.OrchestrationID == state.OrchestrationID {
		s.logger.Error("RETRY_SELF_ADDRESSED: replayed request carries the awaiting orchestration's own id — refusing to send",
			zap.String("request_id", requestID),
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("step_name", awaited.StepName))
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	}

	s.logger.Info("Replaying original request to target agent requests topic",
		zap.String("request_id", requestID),
		zap.String("topic", payload.Topic),
		zap.String("child_orchestration_id", replayed.Headers.OrchestrationID),
		zap.String("action", replayed.Headers.Action),
		zap.Int("retry_version", awaited.RetryVersion),
		zap.String("target_agent_id", awaited.TargetAgentID))

	// Send the retry to the topic the original went to, not a re-derived one.
	err = s.producer.Produce(ctx, payload.Topic, retryHeaders, []byte(payload.Key), retryBytes)
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
			// Resolve error_step from the step-level field first, then config fallback
			resolvedErrorStep := step.ErrorStep
			if resolvedErrorStep == "" {
				if cfgErrorStep, ok := step.Config["error_step"].(string); ok {
					resolvedErrorStep = cfgErrorStep
				}
			}
			if resolvedErrorStep != "" {
				// Clean up the awaited request before routing to error_step
				delete(state.AwaitedRequests, requestID)
				repo := NewStateRepository(s.db, s.logger)
				if markErr := repo.MarkAwaitedRequestComplete(ctx, requestID); markErr != nil {
					s.logger.Warn("Failed to mark awaited request complete for error_step routing",
						zap.String("request_id", requestID), zap.Error(markErr))
				}
				return s.routeToErrorStep(ctx, state, failedStepName, resolvedErrorStep, errorMsg)
			}
		}
	}

	return s.failWorkflow(ctx, state, fmt.Sprintf("Request %s failed: %s", requestID, errorMsg))

}

// handleRequestTimeout is the FAST PATH for awaited-request timeouts: an
// in-process timer armed at request-send time, which dies with its pod —
// bugs_open/003 root cause 3. The per-minute ticker in
// cleanupExpiredAwaitedRequests is the durable path that survives restarts.
// Both funnel through an atomic status->'retrying' claim, so exactly one
// actor drives any given expiry.
func (s *SagaCoordinator) handleRequestTimeout(ctx context.Context, orchestrationID, requestID string, timeoutAt time.Time) {
	time.Sleep(time.Until(timeoutAt))

	repo := NewStateRepository(s.db, s.logger)
	awaited, err := repo.ClaimAwaitedRequestForRetry(ctx, requestID, s.podName)
	if err != nil {
		s.logger.Error("TIMEOUT_FAST_PATH_CLAIM_FAILED (retry ticker will recover)",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}
	if awaited == nil {
		return // answered, cancelled, or another actor holds the claim
	}

	s.retryExpiredAwaitedRequest(ctx, awaited)
}

// retryExpiredAwaitedRequest drives one timed-out awaited request through
// retry-or-fail (bugs_open/003 F2). PRECONDITION: the caller holds the row in
// status='retrying' (ClaimAwaitedRequestForRetry or
// ClaimExpiredAwaitedRequestsForRetry). Every exit moves the row out of
// 'retrying' — the resend sets 'waiting' (UpdateAwaitedRequestRetry, inside
// handleRecoverableError), exhaustion sets 'error', release sets 'processed'
// — except the state-load failure, which deliberately leaves the claim for
// the ticker's stale-'retrying' reclaim (>5 min).
func (s *SagaCoordinator) retryExpiredAwaitedRequest(ctx context.Context, awaited *AwaitedRequest) {
	requestID := awaited.RequestID
	orchestrationID := awaited.OrchestrationID
	repo := NewStateRepository(s.db, s.logger)

	state, err := repo.GetState(ctx, orchestrationID)
	if err != nil || state == nil {
		s.logger.Error("RETRY_DRIVER_STATE_LOAD_FAILED (claim left for ticker reclaim)",
			zap.String("request_id", requestID),
			zap.String("orchestration_id", orchestrationID),
			zap.Error(err))
		return
	}

	// The orchestration may have moved on (completed, failed, reaped) between
	// expiry and claim — release the claim rather than resurrect the request.
	if state.Status != StatusAwaitingResponses {
		s.logger.Info("RETRY_DRIVER_RELEASED: orchestration no longer awaiting",
			zap.String("request_id", requestID),
			zap.String("orchestration_id", orchestrationID),
			zap.String("orchestration_status", string(state.Status)))
		if markErr := repo.MarkAwaitedRequestComplete(ctx, requestID); markErr != nil {
			s.logger.Warn("Failed to release claimed awaited request",
				zap.String("request_id", requestID), zap.Error(markErr))
		}
		return
	}

	// Retry budget exhausted: terminal release FIRST, so no downstream failure
	// can leave the row parked in 'retrying', then route the failure.
	if awaited.RetryVersion >= 3 {
		s.logger.Error("Max retries exceeded",
			zap.String("request_id", requestID),
			zap.Int("retry_version", awaited.RetryVersion))
		if failErr := repo.MarkAwaitedRequestFailed(ctx, requestID); failErr != nil {
			s.logger.Warn("Failed to mark awaited request failed",
				zap.String("request_id", requestID), zap.Error(failErr))
		}
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
		return
	}

	// Check if still waiting (per the orchestration's own state)
	if _, exists := state.AwaitedRequests[requestID]; exists {
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
							"timeout_after": time.Since(awaited.TimeoutAt).String(),
							"retry_count":   awaited.RetryVersion,
						},
					},
				},
			}

			s.handleRecoverableError(ctx, state, requestID, execCtx, timeoutResponse, awaited)
		} else {
			s.routeToErrorStepOrFail(ctx, state, awaited.StepName, fmt.Sprintf("Request %s timed out after %d retries", requestID, awaited.RetryVersion))
		}
	} else {
		// The response was already applied to the orchestration's state —
		// release the claim so the row cannot wedge in 'retrying'.
		s.logger.Info("RETRY_DRIVER_RELEASED: request no longer in awaited set",
			zap.String("request_id", requestID))
		if markErr := repo.MarkAwaitedRequestComplete(ctx, requestID); markErr != nil {
			s.logger.Warn("Failed to release claimed awaited request",
				zap.String("request_id", requestID), zap.Error(markErr))
		}
	}
}

// Helper methods
func (s *SagaCoordinator) determineOwnerAgentType(execCtx *types.ExecutionContext) string {
	// The context's own answer: the RESOLVED real agent type set by the processor,
	// falling back to the dispatch-path sender (bugs_open/060). Both rungs now
	// live on the context itself so the actions-package error-log door climbs the
	// SAME ladder instead of its own shorter one — see
	// types.ExecutionContext.ResolvedAgentType, which also records why the two
	// rungs below stay here rather than moving with them.
	if agentType := execCtx.ResolvedAgentType(); agentType != "" {
		return agentType
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
			// Check step-level first (parallel to NextStep) — preferred location
			if step.ErrorStep != "" {
				return s.routeToErrorStep(ctx, state, failedStepName, step.ErrorStep, errorMsg)
			}
			// Fallback to config-level for backward compatibility
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

	// bugs_open/243: `__step_error` above is a SINGLE key and it is OVERWRITTEN on
	// every routed failure, so a workflow with two failing steps keeps only the last
	// and no downstream step can ask "did MY predecessor fail?". The council gate
	// needs exactly that question answered — see step_error_record.go for the
	// contract, the cap and why the writer lives in a function it can be tested in.
	recordStepError(state.CollectedData, failedStepName, errorMsg, time.Now())

	// Log to agent_error_log for persistent error tracking
	entry := s.buildErrorEntry(state, failedStepName, errorMsg)
	entry.Severity = "error" // routed to error_step, not fatal
	s.logAgentError(ctx, entry)

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

// ownIdentity is this orchestration's own identity as the SENDER of a reply —
// the same shape the coordinator already stamps onto execution contexts at four
// other sites. Its AgentType feeds the sender_agent_type header, which
// ValidateOutgoingMessage requires on every non-error outgoing message
// (bugs_open/274: this and the step name below were never set here, so every
// completed child workflow failed producer-side validation and was reported to
// its parent as FAILED).
func (s *SagaCoordinator) ownIdentity(state *OrchestrationState) types.AgentIdentity {
	return types.AgentIdentity{
		AgentType:    state.OwnerAgentType,
		AgentID:      state.OwnerAgentID,
		PodName:      s.podName,
		AgentVersion: os.Getenv("AGENT_VERSION"),
		Role:         state.OwnerAgentRole,
	}
}

// parentReplyStepName recovers the PARENT's spawning step name — the step this
// reply is in response to. It was recorded into CollectedData when the child's
// state was created (BuildCollectedData), in the same breath as the
// __parent_responses_topic__/__reply_to_request_id__ keys the notify functions
// already trust, and nothing overwrites it afterwards. After a DB round-trip
// __execution_context__ is a plain map, but a freshly-built state can still hold
// the typed struct, so both shapes are handled.
func parentReplyStepName(state *OrchestrationState) string {
	if ec, ok := state.CollectedData["__execution_context__"]; ok {
		switch v := ec.(type) {
		case *types.ExecutionContext:
			if v.StepName != "" {
				return v.StepName
			}
		case map[string]interface{}:
			if name, _ := v["step_name"].(string); name != "" {
				return name
			}
		}
	}
	if wr, ok := state.CollectedData["__work_request__"].(map[string]interface{}); ok {
		if name, _ := wr["step_name"].(string); name != "" {
			return name
		}
	}
	return ""
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

	// Build result from CollectedData - extract workflow outputs.
	// If the result cannot be delivered (exceeds the bus size cap), the workflow
	// ran its steps but did NOT hand a result back — that is not a success. Surface
	// it as a failure so the parent's error_step / needs-review handling fires,
	// instead of a silent "completed" stub that loses the output.
	resultData, err := s.extractWorkflowResultWithSizeLimit(state)
	if err != nil {
		s.logger.Error("notifyParentOfSuccess: result undeliverable — notifying parent of FAILURE instead",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.Error(err))
		s.notifyParentOfFailure(ctx, state, err.Error())
		return
	}

	// bugs_open/274: Sender and InResponseToStepName are two of the five headers
	// ValidateOutgoingMessage requires, and this literal shipped without them
	// from 2026-01-11. Harmless while the produce was unvalidated; when 158's
	// fix put this site on the validated path (2026-08-03) the reply became
	// deterministically undeliverable and every completed child workflow was
	// reported to its parent as FAILED (~16,869 rows in 12 days).
	inResponseToStepName := parentReplyStepName(state)
	if inResponseToStepName == "" {
		// Validation will refuse the reply and the failure arm below will fire —
		// exactly the pre-fix behaviour, but named, so it cannot be mistaken for
		// a transport problem.
		s.logger.Error("notifyParentOfSuccess: no parent step name recoverable from collected_data — the success reply cannot pass validation (bugs_open/274)",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("correlation_id", state.CorrelationID))
	}

	successResponse := types.ResponseMessage{
		Headers: types.ResponseHeaders{
			InResponseToRequestID: replyToRequestID,
			InResponseToStepName:  inResponseToStepName,
			Status:                "complete",
			IsComplete:            true,
			MessageType:           "response",
			//MessageID:             uuid.New().String(),
			TimeSent:        time.Now(),
			OrchestrationID: state.OrchestrationID,
			CorrelationID:   state.CorrelationID,
			ClientID:        state.ClientID,
			Sender:          s.ownIdentity(state),
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

	// bugs_open/158 item 1, owner ruling 2026-08-03 (option b — the plumbing
	// sites first, because every agent inherits them).
	//
	// This was a log-and-carry-on: if the produce failed, the parent was never
	// told the child had finished and simply waited out its timeout, with the
	// cause visible only in THIS pod's logs. The rule (016b §9, bugs_closed/062)
	// is that a reply which cannot be delivered must become a deliverable error.
	//
	// degrade is nil deliberately: extractWorkflowResultWithSizeLimit above has
	// already bounded the payload, so there is nothing left here to shrink — an
	// oversize refusal at this point means the bound itself is wrong, and a
	// second guaranteed-to-fail produce would only hide that.
	//
	// The answer to "could not tell the parent it succeeded" is the one the
	// size-limit path above already chose for the same predicament: TELL THE
	// PARENT IT FAILED. A workflow whose result never reached its parent did not
	// succeed from the parent's point of view, and notifyParentOfFailure routes
	// into error_step / needs-review handling instead of leaving a silent gap.
	// bugs_open/040: WithRetry opts this call — and only this call — into the
	// bounded produce retry. DeliverReply deliberately does not retry transients
	// by default ("the caller's existing retry path stays in charge"); on this
	// path that caller is failWorkflow, which has none.
	outcome, err := kafka.DeliverReply(ctx, s.producer, s.logger,
		parentTopic, successResponse.Headers.ToMap(), []byte(replyToRequestID), responseBytes, nil,
		kafka.WithRetry(kafka.DefaultReplyRetryPolicy))
	if outcome.Answered() {
		s.logger.Info("Successfully notified parent of workflow completion",
			zap.String("parent_topic", parentTopic),
			zap.String("reply_to_request_id", replyToRequestID),
			zap.String("delivery_outcome", outcome.String()))
		return
	}
	s.logger.Error("Could not notify parent of success — notifying parent of FAILURE instead",
		zap.Error(err),
		zap.String("delivery_outcome", outcome.String()),
		zap.String("parent_topic", parentTopic),
		zap.String("orchestration_id", state.OrchestrationID))
	s.notifyParentOfFailure(ctx, state,
		fmt.Sprintf("workflow completed but its result could not be delivered to the parent (%s): %v", outcome, err))
}

// extractWorkflowResult builds the result payload to send back to parent
// respects output_fields config from complete_workflow step
func (s *SagaCoordinator) extractWorkflowResult(state *OrchestrationState) map[string]interface{} {
	result := make(map[string]interface{})

	// Resolve the result contract from the complete step's config. Centralised,
	// loud and testable — see resolveResultSpec (result_spec.go). This replaces
	// the old inline "if output_fields else dump" branch that silently ignored
	// the singular output_field and the output mapping.
	var completeConfig map[string]interface{}
	if state.WorkflowPlan.Steps != nil {
		if completeStep, ok := state.WorkflowPlan.Steps["complete"]; ok {
			completeConfig = completeStep.Config
		}
	}

	spec := resolveResultSpec(completeConfig, s.logger)

	switch spec.Mode {
	case ResultModeFields:
		// Each named field, nested under its own key (historic plural behaviour).
		for _, fieldName := range spec.Fields {
			if value := datahelpers.ExtractNestedField(state.CollectedData, fieldName); value != nil {
				result[fieldName] = datahelpers.ExtractStepData(value)
			}
		}

	case ResultModeFlatten:
		// The single named field's CONTENTS become the body — the flat contract
		// page-build-handler / page-rebuild / site-work-orchestrator read.
		value := datahelpers.ExtractNestedField(state.CollectedData, spec.From)
		unwrapped := datahelpers.ExtractStepData(value)
		switch v := unwrapped.(type) {
		case map[string]interface{}:
			for k, val := range v {
				result[k] = val
			}
		case nil:
			s.logger.Warn("extractWorkflowResult: result_from field not found in collected data",
				zap.String("field", spec.From))
		default:
			// Non-map single field: nothing to flatten — preserve under its name.
			result[spec.From] = unwrapped
			s.logger.Warn("extractWorkflowResult: result_from field is not a map; stored under its name",
				zap.String("field", spec.From))
		}

	case ResultModeMapping:
		// Build the body from explicit target<-source.path pairs.
		for target, sourcePath := range spec.Mapping {
			if value := datahelpers.ExtractNestedField(state.CollectedData, sourcePath); value != nil {
				result[target] = value
			}
		}

	default: // ResultModeFallback
		s.fallbackDumpInto(result, state)
	}

	// Completion metadata — only if absent, so a flattened/mapped field whose
	// payload happens to carry these keys cannot be clobbered.
	setIfAbsent(result, "orchestration_id", state.OrchestrationID)
	setIfAbsent(result, "completed_steps", state.ExecutionMetadata.CompletedSteps)
	setIfAbsent(result, "completed_at", time.Now().Format(time.RFC3339))

	if resultBytes, err := json.Marshal(result); err == nil {
		s.logger.Info("extractWorkflowResult: result built",
			zap.String("mode", spec.Mode.String()),
			zap.Int("size_bytes", len(resultBytes)),
			zap.Int("field_count", len(result)))
	}

	return result
}

const MaxResultSizeBytes = 900000 // Leave headroom below 1MB Kafka limit

// extractWorkflowResultWithSizeLimit builds the result and enforces the Kafka
// delivery cap. It returns an ERROR (not a stub) when the result cannot be
// delivered: a workflow that ran its steps but cannot hand its result back has
// not succeeded from the parent's point of view, and must never report success.
//
// It deliberately does NOT silently truncate to fit — a truncated page_html or
// sections_metadata delivered as "success" is a corrupt result, worse than a
// clean failure. The fix for an oversize result is a result contract on the
// complete step (result_from for a single object, multiple_output_fields for
// selected fields) so the workflow returns only what the parent needs instead of
// its whole working state. Oversize is now a loud, actionable signal, not a
// silent stub.
func (s *SagaCoordinator) extractWorkflowResultWithSizeLimit(state *OrchestrationState) (map[string]interface{}, error) {
	result := s.extractWorkflowResult(state)

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow result for delivery: %w", err)
	}

	if len(resultBytes) <= MaxResultSizeBytes {
		return result, nil
	}

	return nil, s.oversizeResultError(state, result, len(resultBytes))
}

// oversizeResultError logs the per-field size breakdown and returns an actionable
// error naming the largest field and the fix. The breakdown goes to the log
// (Error level, so it surfaces); the message carries the headline numbers + the
// remedy so the failure is self-explanatory wherever it lands.
func (s *SagaCoordinator) oversizeResultError(state *OrchestrationState, result map[string]interface{}, size int) error {
	fieldSizes := make(map[string]int, len(result))
	largestField, largestSize := "", 0
	for k, v := range result {
		b, err := json.Marshal(v)
		if err != nil {
			continue
		}
		fieldSizes[k] = len(b)
		if len(b) > largestSize {
			largestField, largestSize = k, len(b)
		}
	}

	s.logger.Error("extractWorkflowResult: result exceeds delivery cap — failing instead of stubbing a success",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("owner_agent_type", state.OwnerAgentType),
		zap.Int("result_size_bytes", size),
		zap.Int("max_result_size_bytes", MaxResultSizeBytes),
		zap.Int("field_count", len(result)),
		zap.String("largest_field", largestField),
		zap.Int("largest_field_bytes", largestSize),
		zap.Any("field_sizes", fieldSizes),
	)

	return fmt.Errorf(
		"workflow result %d bytes exceeds the %d-byte delivery cap (largest field %q=%d bytes); "+
			"declare a result contract on the complete step — result_from for a single object, or "+
			"multiple_output_fields for selected fields — so the workflow returns only what the parent "+
			"needs instead of its whole working state",
		size, MaxResultSizeBytes, largestField, largestSize,
	)
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

	// Log to agent_error_log for persistent error tracking
	entry := s.buildErrorEntry(state, state.CurrentStep, errorMsg)
	entry.Severity = "fatal" // workflow failed entirely
	s.logAgentError(ctx, entry)

	// bugs_open/217: this sender used to hardcode error_unrecoverable, making
	// every child-orchestration failure terminal at the parent — and on the
	// orchestrated path this response claims the awaited request FIRST, so no
	// later, better-classified response could win. Classify before stamping,
	// through the same sequenced seam as the processor senders (bugs_open/207).
	// Only errorMsg survives to here — callers stringified long ago — so the
	// typed DomainError branches cannot fire and the prose needles decide,
	// which is exactly the fallback they exist for. RetryDisposition, never
	// MatchedTransientFailure bare: this prose can carry both a permanent and
	// a transient needle, and permanent must be asked first.
	recoverable, matched := perrors.RetryDisposition(errors.New(errorMsg))
	status := "error_unrecoverable"
	if recoverable {
		status = "error_recoverable"
	}
	s.logger.Info("retry disposition decided at the child-orchestration failure sender by the sequenced shared classifier",
		zap.String("disposition", status),
		zap.String("disposition_matched", matched),
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("correlation_id", state.CorrelationID))

	failureResponse := types.ResponseMessage{
		Headers: types.ResponseHeaders{
			InResponseToRequestID: replyToRequestID,
			// bugs_open/274: the same envelope fields the success arm was
			// missing, plus ClientID, which this literal never set — the reply
			// only passed the parent's incoming validation via the is_error
			// bypass. This produce is unvalidated (is_error is always let
			// through), so these are truthfulness, not deliverability.
			InResponseToStepName: parentReplyStepName(state),
			Status:               status,
			IsError:              true,
			MessageType:          "response",
			//MessageID:             uuid.New().String(),
			TimeSent:        time.Now(),
			OrchestrationID: state.OrchestrationID,
			CorrelationID:   state.CorrelationID,
			ClientID:        state.ClientID,
			Sender:          s.ownIdentity(state),
		},
		Body: types.ResponseBody{
			Success: false,
			Error: &types.ErrorInfo{
				// The code and message stay verbatim whatever the disposition:
				// their consumers (agenterrors' code mapper, the 090 verdict
				// recovery runbook) key on CHILD_ORCHESTRATION_FAILED.
				Code:        "CHILD_ORCHESTRATION_FAILED",
				Message:     errorMsg,
				Recoverable: recoverable,
			},
		},
	}

	responseBytes, _ := json.Marshal(failureResponse)
	// bugs_open/040: opted into the bounded produce retry, closing the asymmetry
	// with notifyParentOfSuccess — which has gone through the shared reply seam
	// since bugs_open/133 while this one stayed a bare fire-once log-and-drop. If
	// this send fails the parent learns nothing and waits out its own timeout, so
	// it is exactly the case worth resending. The last-resort log stays: there is
	// no one left to tell, but it now retried first.
	if err := kafka.ProduceWithRetry(ctx, s.producer, s.logger, kafka.DefaultReplyRetryPolicy,
		parentTopic, failureResponse.Headers.ToMap(), []byte(replyToRequestID), responseBytes); err != nil {
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

	// bugs_open/354: a run that ends at a terminal its author declared to be an
	// ERROR terminal, having suffered a routed step failure, used to be recorded
	// with error NULL — indistinguishable from a run that executed every step.
	// Record the failure on the row. The status stays COMPLETED and the parent is
	// still told success: changing either is authority on a shared seam and
	// belongs to 354's architecture RFC (bugs_closed/344 deferred exactly this
	// twice; RFC_023 is a hard guardian veto on scope for re-typing COMPLETED on
	// this table). See error_route_completion.go for why the discriminator is a
	// declaration and not a structural rule — three structural rules were tried
	// and measured, the best reaching 36% of the real population.
	if errMsg, endedOnErrorTerminal := errorRouteTermination(state); endedOnErrorTerminal {
		state.Error = errMsg
		state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
			PodName:   s.podName,
			StepName:  state.CurrentStep,
			Action:    "completed_on_error_route",
			Timestamp: time.Now(),
			Details:   errMsg,
		})
		s.logger.Warn("WORKFLOW_COMPLETION: run ended at a declared error terminal — recording the failure on the row (bugs_open/354)",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("terminal_step", state.CurrentStep),
			zap.String("owner_agent_type", state.OwnerAgentType),
			zap.String("error", errMsg))
	}

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

// retryWindow returns how long a REPLAYED request is waited for.
//
// The step's DECLARED timeout is authoritative (bugs_open/029). The block this
// replaced derived the window from the row instead —
//
//	originalDuration := awaited.TimeoutAt.Sub(awaited.SentAt)
//	if originalDuration > 30*time.Minute { newTimeout = 3 * time.Minute }
//	if newTimeout > 5*time.Minute       { newTimeout = 5 * time.Minute }
//
// — which had three defects, all measured live on 2026-08-18:
//
//  1. It INVERTED the declaration. A step asking for more than 30 minutes was
//     given THREE; everything else was capped at FIVE. The longer you declared,
//     the less you got. 33 live steps across 25 agent types declare more than
//     300s (600 … 86400, the largest being a human-approval step that would have
//     been given three minutes for a person to answer).
//  2. The row is a poisoned source anyway: UpdateAwaitedRequestRetry resets
//     sent_at on every retry, so TimeoutAt-SentAt reports the PREVIOUS retry's
//     window, not the declared one. From retry 2 onward it is self-referential.
//  3. The truncation manufactured premature exhaustion. build-dispatch-loop's
//     COMPLETED runs exceed 5 minutes 25.5% of the time but its declared 15
//     minutes only 5.9%; page-build-handler, 17.6% against 0.5%. So the cap
//     turned a retry that would usually have succeeded into one that usually
//     failed — at BOTH levels of the call tree, since the loop's own
//     iter_N_call_handler awaits are truncated the same way and abandon real
//     page work when they exhaust.
//
// Loop-expanded steps resolve here: the stored plan carries the suffixed keys
// (process_item_iter_1_call_handler et al) with their config intact, which is
// exactly the population the bug was measured on. ConvertStepTimeout is required
// because the expanded steps carry config.timeout_seconds with the `timeout`
// field empty.
//
// Falls back to the row only when the plan no longer carries the step, and never
// below the system default — a shorter-than-default retry is the defect above in
// miniature.
func retryWindow(state *OrchestrationState, awaited *AwaitedRequest, logger *zap.Logger) time.Duration {
	systemDefault := time.Duration(datahelpers.DefaultRequestTimeout) * time.Second

	if state != nil && awaited != nil {
		if step, ok := state.WorkflowPlan.Steps[awaited.StepName]; ok {
			// Go through getTimeout, the SAME helper that sets the initial
			// awaited window at the two registration sites, so the declared
			// timeout is read one way rather than two (council `reuse_agent`,
			// corr 7c92389a). ConvertStepTimeout first because the stored plan
			// carries config.timeout_seconds with `timeout` empty — including
			// on loop-expanded steps, verified on a live row.
			datahelpers.ConvertStepTimeout(&step, logger)
			if d := getTimeout(step); d > 0 {
				return d
			}
		}
	}

	if awaited != nil {
		if d := awaited.TimeoutAt.Sub(awaited.SentAt); d > systemDefault {
			return d
		}
	}
	return systemDefault
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

		// bugs_open/003 F2: the DURABLE retry driver. The cleanup above has
		// just marked timed-out 'waiting' rows 'expired'; claim them (plus any
		// 'retrying' rows whose claiming pod died) and drive each through the
		// same body the in-process timers use. This is what makes per-workflow
		// timeouts survive pod restarts: the timers die with their pod, these
		// rows do not, and any surviving chassis pod becomes the rescuer.
		claimed, claimErr := repo.ClaimExpiredAwaitedRequestsForRetry(ctx, s.podName, 25)
		if claimErr != nil {
			s.logger.Error("RETRY_TICKER_CLAIM_FAILED", zap.Error(claimErr))
		}
		for _, awaited := range claimed {
			s.logger.Info("RETRY_TICKER_CLAIMED expired awaited request",
				zap.String("request_id", awaited.RequestID),
				zap.String("orchestration_id", awaited.OrchestrationID),
				zap.String("step_name", awaited.StepName),
				zap.Int("retry_version", awaited.RetryVersion))
			s.retryExpiredAwaitedRequest(ctx, awaited)
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

// claimRecoveryStaleness is how old a 'processing' claim must be before
// CLAIM_RECOVERY may steal it back to 'waiting'. The live claim window covers
// only parse → optimistic-lock state save → MarkAwaitedRequestComplete
// (handleCompleteResponse stamps processed_at BEFORE continueExecution runs),
// i.e. milliseconds to low seconds even under version contention — so two
// minutes distinguishes "claimer died mid-processing" from "claimer is live
// and this is a duplicate delivery" with a wide margin on both sides.
const claimRecoveryStaleness = 2 * time.Minute

// ResetStaleAwaitedRequestForRetry resets a claimed-but-not-processed request
// back to 'waiting' so it can be re-claimed — UNLESS the row is a fresh
// 'processing' claim, which means a live actor holds it and the caller is
// looking at a duplicate delivery, not a crashed claimer. Reports whether the
// reset happened. The freshness check is part of the UPDATE's WHERE clause,
// never a separate read: a read-then-reset pair reintroduces exactly the race
// this guard exists to close. Non-'processing' states (waiting, retrying,
// expired, cancelled, error) reset unconditionally, as they always have —
// the F2 retry driver's late-response window relies on 'retrying' doing so.
func (r *StateRepository) ResetStaleAwaitedRequestForRetry(ctx context.Context, requestID string, staleness time.Duration) (bool, error) {
	query := `
		UPDATE awaited_requests
		SET status = 'waiting',
			processing_started_at = NULL,
			processing_pod = NULL
		WHERE request_id = $1
		  AND processed_at IS NULL
		  AND (status <> 'processing'
		       OR processing_started_at IS NULL
		       OR processing_started_at < NOW() - make_interval(secs => $2))
	`
	result, err := r.db.ExecContext(ctx, query, requestID, staleness.Seconds())
	if err != nil {
		return false, err
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Reset awaited request for retry",
		zap.String("request_id", requestID),
		zap.Int64("rows_affected", rowsAffected))

	return rowsAffected > 0, nil
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
		"content_field",           // Used by assemble_page, git_commit
		"page_component_id_field", // Used by update_page_status
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

	// Generic pass (bugs_open/287 §9a: "stop enumerating"). ANY other top-level config
	// entry whose value is a bare data reference gets the same treatment as the
	// allow-listed keys above. The allow-list was a promise every future action had to
	// know to keep, and complete_work_item's "result" did not — that is how the
	// spawn-record bug shipped: `"result": "handler_result"` was never rewritten to the
	// iteration-suffixed key, so the reference could only resolve through the resolver's
	// whole-tree search, which prefers a same-named key anywhere in the tree
	// (RESOLVER_MAPPING_BYPASSED, RFC_029 §10.6).
	//
	// prefixDataReference already rewrites only when the first dotted segment is a
	// sibling output_field, so the added gate here is the SHAPE of the value: a bare
	// identifier or dotted path. Expressions and prose — conditional `condition` strings;
	// one live census row is an OR carrying TWO references, which a whole-string rewrite
	// would half-edit — are excluded by construction. Step-name keys are excluded because
	// their values name steps, not data. input_mapping is handled above. A `!`-suffixed
	// strict key (RFC_029 §9 D3) is covered automatically, since this pass keys on the
	// value, not the key. NOTE: top-level string values only — `input_fields` ARRAYS hold
	// Strategy-1 field NAMES, not references, and must never be rewritten (pinned by
	// loop_config_reference_suffixing_test.go).
	//
	// Census before shipping (2026-08-17, live agent_definitions; query in
	// docs024_key_docs_latest/bugfix_287_spawn_record/RUNBOOK): 22 sites / 7 agents match
	// the first-segment rule; every one is a read-reference — zero literals, zero
	// write-destinations.
	stepRefKeySet := make(map[string]bool, len(stepRefKeys))
	for _, k := range stepRefKeys {
		stepRefKeySet[k] = true
	}
	for key, raw := range config {
		if stepRefKeySet[key] || key == "input_mapping" {
			continue
		}
		val, ok := raw.(string)
		if !ok || val == "" || !referenceShapedConfigValue.MatchString(val) {
			continue
		}
		if prefixedVal := prefixDataReference(val, loopName, iterIdx, substepOutputFields); prefixedVal != val {
			config[key] = prefixedVal
		}
	}
}

// referenceShapedConfigValue matches a bare data reference: an identifier optionally
// followed by dotted segments. Deliberately narrow — anything with spaces, operators,
// brackets or quotes is an expression or prose and is left alone by the generic
// suffixing pass above.
var referenceShapedConfigValue = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z0-9_]+)*$`)

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
