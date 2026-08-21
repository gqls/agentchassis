// FILE: platform/orchestration/loop_error_handling.go
//
// Helpers for continue_on_error in loop iterations.
//
// When a loop step has continue_on_error: true, iteration failures
// are recorded as error results and the loop advances to the next
// iteration instead of failing the entire workflow.
//
// Used by: continueExecution, handleUnrecoverableError, handleRequestTimeout

package orchestration

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// loopStepPattern matches "{loopName}_iter_{N}_{substepName}"
var loopStepPattern = regexp.MustCompile(`^(.+)_iter_(\d+)_(.+)$`)

// parsedLoopStep holds the parsed parts of a loop iteration step name
type parsedLoopStep struct {
	LoopName    string
	IterIdx     int
	SubstepName string
}

// parseLoopStepName extracts loopName, iteration index, and substep from a step name.
// Returns nil if the step name doesn't match the loop iteration pattern.
func parseLoopStepName(stepName string) *parsedLoopStep {
	matches := loopStepPattern.FindStringSubmatch(stepName)
	if matches == nil {
		return nil
	}
	iterIdx, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil
	}
	return &parsedLoopStep{
		LoopName:    matches[1],
		IterIdx:     iterIdx,
		SubstepName: matches[3],
	}
}

// isLoopIterationStep checks whether the current step is part of a loop iteration
// by looking for loop_iteration in the step's config (injected during expansion).
func isLoopIterationStep(state *OrchestrationState) bool {
	step, exists := state.WorkflowPlan.Steps[state.CurrentStep]
	if !exists {
		return false
	}
	_, has := step.Config["loop_iteration"]
	return has
}

// shouldContinueLoopOnError checks whether a failed loop iteration step should
// skip to the next iteration rather than failing the workflow.
//
// Returns true if:
// 1. The current step is a loop iteration step
// 2. The loop's continue_on_error flag is true
func shouldContinueLoopOnError(state *OrchestrationState, logger *zap.Logger) bool {
	step, exists := state.WorkflowPlan.Steps[state.CurrentStep]
	if !exists {
		return false
	}

	// Check if this is a loop iteration step
	_, hasLoopIter := step.Config["loop_iteration"]
	if !hasLoopIter {
		return false
	}

	// Check the flag on this step's config (propagated during expansion)
	if cont, ok := step.Config["continue_on_error"].(bool); ok && cont {
		logger.Info("continue_on_error is true for this loop iteration step",
			zap.String("step", state.CurrentStep))
		return true
	}

	return false
}

// skipToNextLoopIteration records the error for this iteration and advances
// the orchestration to the next iteration's first step (or loop_complete if
// this was the last iteration).
//
// It modifies state in place and persists to the database.
//
// Instead of relying on the shared "loop_metadata" key (which gets
// overwritten when a second loop expands in the same orchestration),
// we derive total_iterations and first_substep from the workflow plan.
// The injected steps and the {loop}_complete step are always present.
//
// PERSISTENCE (bugs_open/343 P2, 2026-08-21). The mutations below are applied to
// a FRESHLY LOADED state inside an optimistic-lock retry loop, not to the caller's
// copy with a single unretried write. It used to be one bare repo.UpdateState: an
// optimistic-lock failure there returned an error and LOST the advance — and on
// the async path it lost the awaited-map delete with it, while the table row had
// already been marked terminal, so the reply's redelivery was then eaten by the
// processed_at duplicate guard (coordinator.go DUPLICATE_SKIPPED). A lost
// continuation, on exactly the error path that carried all 31 of 343's observed
// terminal outcomes.
//
// Every sibling advance on this coordinator already retries — the park 10
// attempts, handleCompleteResponse 15, the progress loop 5. This one did not, and
// that asymmetry is the whole of the defect.
//
// ⚠ Do NOT "simplify" this to repo.UpdateStateWithRetry. That helper retries by
// reloading and doing *state = *reloaded, then re-issuing the UPDATE with the
// reloaded, UNMUTATED state — so the caller's mutations are silently discarded and
// a no-op is persisted. It is not a retrying version of "save my changes".
//
// terminalRequestID, when non-empty, is an awaited request whose row the caller
// has already resolved; its map entry is re-deleted on every attempt so a reload
// cannot resurrect it. Empty from the synchronous caller, which has no awaited
// request in play.
func (s *SagaCoordinator) skipToNextLoopIteration(
	ctx context.Context,
	state *OrchestrationState,
	errorMsg string,
	logger *zap.Logger,
) error {
	return s.skipToNextLoopIterationWithAwaited(ctx, state, "", errorMsg, logger)
}

func (s *SagaCoordinator) skipToNextLoopIterationWithAwaited(
	ctx context.Context,
	state *OrchestrationState,
	terminalRequestID string,
	errorMsg string,
	logger *zap.Logger,
) error {
	parsed := parseLoopStepName(state.CurrentStep)
	if parsed == nil {
		return fmt.Errorf("cannot parse loop step name: %s", state.CurrentStep)
	}

	// Derive total_iterations from the {loop}_complete step's config
	// This is always present and specific to this loop (not shared).
	completeStepName := fmt.Sprintf("%s_complete", parsed.LoopName)
	totalIterations := 0
	if completeStep, ok := state.WorkflowPlan.Steps[completeStepName]; ok {
		if ti, ok := completeStep.Config["total_iterations"].(float64); ok {
			totalIterations = int(ti)
		} else if ti, ok := completeStep.Config["total_iterations"].(int); ok {
			totalIterations = ti
		}
	}
	if totalIterations == 0 {
		return fmt.Errorf("cannot determine total_iterations for loop %s", parsed.LoopName)
	}

	// Derive first_substep by looking for iter_0 steps in the workflow plan.
	// The first substep is the one in {loop}_iter_0_{substep} that has the
	// lowest order (i.e. appears as the first step created). We find it by
	// checking which iter_0 step is NOT referenced as a next_step by another
	// iter_0 step — or more simply, we check what iter_{N+1} would start with
	// by looking at iter_0's entry point.
	firstSubstep := findFirstSubstep(state.WorkflowPlan.Steps, parsed.LoopName)
	if firstSubstep == "" {
		return fmt.Errorf("cannot determine first_substep for loop %s", parsed.LoopName)
	}

	// Determine next step
	failedStep := state.CurrentStep
	var nextStep string
	if parsed.IterIdx < totalIterations-1 {
		// Advance to next iteration's first substep
		nextStep = fmt.Sprintf("%s_iter_%d_%s", parsed.LoopName, parsed.IterIdx+1, firstSubstep)
	} else {
		// Last iteration — go to loop_complete
		nextStep = completeStepName
	}

	iterErrorKey := fmt.Sprintf("%s_iter_%d_error", parsed.LoopName, parsed.IterIdx)
	errorCountKey := fmt.Sprintf("%s_error_count", parsed.LoopName)

	// applySkip writes this skip onto whichever copy of the state it is handed.
	// Kept as one closure so the retry loop below and the caller's in-memory copy
	// cannot drift: every attempt applies exactly the same mutations to a state
	// freshly loaded from the database.
	applySkip := func(target *OrchestrationState) int {
		if target.CollectedData == nil {
			target.CollectedData = make(map[string]interface{})
		}

		// Record the error as this iteration's result
		target.CollectedData[iterErrorKey] = map[string]interface{}{
			"status":    "error",
			"error":     errorMsg,
			"step":      failedStep,
			"iteration": parsed.IterIdx,
			"skipped":   true,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		// Track error count per-loop (keyed by loop name, not shared). Counted off
		// the TARGET, so a reload picks up a sibling's increment rather than
		// overwriting it with a number derived from a stale copy.
		errorCount := 0
		if ec, ok := target.CollectedData[errorCountKey].(float64); ok {
			errorCount = int(ec)
		} else if ec, ok := target.CollectedData[errorCountKey].(int); ok {
			errorCount = ec
		}
		target.CollectedData[errorCountKey] = errorCount + 1

		// The awaited request the caller has already resolved must not come back
		// from the database copy.
		if terminalRequestID != "" && target.AwaitedRequests != nil {
			delete(target.AwaitedRequests, terminalRequestID)
		}

		target.CurrentStep = nextStep
		target.Status = StatusExecutingStep
		target.LastActivity = time.Now()

		return errorCount + 1
	}

	totalErrors := applySkip(state)

	logger.Warn("Skipping failed loop iteration, advancing to next",
		zap.String("failed_step", failedStep),
		zap.String("error", errorMsg),
		zap.Int("failed_iteration", parsed.IterIdx),
		zap.Int("total_iterations", totalIterations),
		zap.String("next_step", nextStep),
		zap.Int("total_errors", totalErrors))

	// Persist, retrying optimistic-lock failures by reloading and RE-APPLYING.
	// See the function comment: a single unretried write here loses the advance,
	// and on the async path the awaited-map delete with it.
	repo := NewStateRepository(s.db, s.logger)
	maxRetries := 10
	baseDelay := 50 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := repo.UpdateState(ctx, state)
		if err == nil {
			if attempt > 1 {
				logger.Info("Loop skip persisted after retry",
					zap.Int("attempts", attempt),
					zap.String("orchestration_id", state.OrchestrationID),
					zap.String("next_step", nextStep))
			}
			// The advance is durable: NOW retire the awaited row. Same ordering as
			// handleCompleteResponse, and the reason it is inside the loop rather
			// than after continueExecution — the row must be retired the moment the
			// advance is safe, not once the whole next iteration has run.
			if terminalRequestID != "" {
				if markErr := repo.MarkAwaitedRequestComplete(ctx, terminalRequestID); markErr != nil {
					logger.Warn("Failed to mark awaited request complete during loop skip",
						zap.String("request_id", terminalRequestID),
						zap.Error(markErr))
					// Deliberately not fatal: the state is saved, which is the half
					// that cannot be recovered from a redelivery.
				}
			}
			break
		}

		if !IsOptimisticLockError(err) {
			return fmt.Errorf("failed to persist state after loop skip: %w", err)
		}
		if attempt >= maxRetries {
			return fmt.Errorf("failed to persist state after loop skip after %d attempts: %w", attempt, err)
		}

		delay := backoffWithJitter(baseDelay, attempt)
		logger.Warn("Optimistic lock failure persisting loop skip, reloading and retrying",
			zap.Int("attempt", attempt),
			zap.Duration("backoff", delay),
			zap.String("orchestration_id", state.OrchestrationID))
		time.Sleep(delay)

		freshState, loadErr := repo.GetState(ctx, state.OrchestrationID)
		if loadErr != nil {
			return fmt.Errorf("failed to reload state for loop skip retry: %w", loadErr)
		}
		applySkip(freshState)
		// Adopt the reloaded copy wholesale: the caller goes on to build a
		// continuation context from this pointer, and it must describe the row
		// that was actually written.
		*state = *freshState
	}

	// Continue execution from the next step
	execCtx := s.createContinuationContext(state)
	return s.continueExecution(ctx, state, execCtx)
}

// findFirstSubstep scans the workflow plan for {loopName}_iter_0_* steps
// and determines which substep name comes first. It does this by finding
// the iter_0 step that no other iter_0 step points to as its next_step
// (i.e. the entry point of the loop body).
func findFirstSubstep(steps map[string]models.Step, loopName string) string {
	prefix := fmt.Sprintf("%s_iter_0_", loopName)

	// Collect all iter_0 substep names
	var substepNames []string
	for stepName := range steps {
		if strings.HasPrefix(stepName, prefix) {
			substepName := strings.TrimPrefix(stepName, prefix)
			substepNames = append(substepNames, substepName)
		}
	}

	if len(substepNames) == 0 {
		return ""
	}
	if len(substepNames) == 1 {
		return substepNames[0]
	}

	// Find which substep is NOT a next_step target of another iter_0 step
	// That's the entry point.
	referencedAsNext := make(map[string]bool)
	for stepName, step := range steps {
		if strings.HasPrefix(stepName, prefix) && step.NextStep != "" {
			// step.NextStep is already prefixed like "{loop}_iter_0_{substep}"
			if strings.HasPrefix(step.NextStep, prefix) {
				target := strings.TrimPrefix(step.NextStep, prefix)
				referencedAsNext[target] = true
			}
		}
	}

	for _, name := range substepNames {
		if !referencedAsNext[name] {
			return name
		}
	}

	// Fallback: just return the first one found
	return substepNames[0]
}

// skipToNextLoopIterationForAsync is the same as skipToNextLoopIteration but
// also cleans up the awaited request that failed. Used by handleUnrecoverableError
// and handleRequestTimeout.
//
// ORDERING (bugs_open/343 P2, 2026-08-21): the awaited row is marked terminal
// only AFTER the advance is durably persisted — the same rule handleCompleteResponse
// calls "the key fix". It used to be marked FIRST. When the persist then failed,
// the advance and the map delete were both lost while the row was already
// terminal, so the reply's redelivery was eaten by the processed_at duplicate
// guard and the continuation was gone for good. Marking last means a persist
// failure leaves the row claimable and the redelivery still able to drive it.
//
// On the timeout path (retryExpiredAwaitedRequest) the row is already 'error'
// before this function is entered, so the ordering there is unchanged by
// construction — this only moves the response path's mark.
func (s *SagaCoordinator) skipToNextLoopIterationForAsync(
	ctx context.Context,
	state *OrchestrationState,
	requestID string,
	errorMsg string,
	logger *zap.Logger,
) error {
	// The advance re-deletes the map entry on every persist attempt so a reload
	// cannot resurrect the request we are resolving, and marks the row terminal
	// the moment — and only if — that advance is durable.
	if err := s.skipToNextLoopIterationWithAwaited(ctx, state, requestID, errorMsg, logger); err != nil {
		logger.Warn("Loop skip did not persist - the awaited row is left claimable so the redelivery can still drive it",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}
	return nil
}
