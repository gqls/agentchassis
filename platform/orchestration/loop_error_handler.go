// FILE: platform/orchestration/loop_error_handler.go
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
	"time"

	"go.uber.org/zap"
)

// The flow is:
//   workflow config: "continue_on_error": true
//     → LoopAction reads it from params.StepConfig.Config
//     → LoopAction passes it in the expansion return map
//     → handleLoopExpansion reads it from loopResult
//     → handleLoopExpansion stores it in loop_metadata + per-step config
//     → shouldContinueLoopOnError reads it from per-step config

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
func (s *SagaCoordinator) skipToNextLoopIteration(
	ctx context.Context,
	state *OrchestrationState,
	errorMsg string,
	logger *zap.Logger,
) error {
	parsed := parseLoopStepName(state.CurrentStep)
	if parsed == nil {
		return fmt.Errorf("cannot parse loop step name: %s", state.CurrentStep)
	}

	// Get loop metadata for total_iterations and first_substep
	loopMeta, _ := state.CollectedData["loop_metadata"].(map[string]interface{})
	if loopMeta == nil {
		return fmt.Errorf("loop_metadata not found in collected data")
	}

	totalIterations := 0
	if ti, ok := loopMeta["total_iterations"].(float64); ok {
		totalIterations = int(ti)
	} else if ti, ok := loopMeta["total_iterations"].(int); ok {
		totalIterations = ti
	}

	firstSubstep, _ := loopMeta["first_substep"].(string)
	if firstSubstep == "" {
		return fmt.Errorf("first_substep not found in loop_metadata")
	}

	// Record the error as this iteration's result
	iterErrorKey := fmt.Sprintf("%s_iter_%d_error", parsed.LoopName, parsed.IterIdx)
	state.CollectedData[iterErrorKey] = map[string]interface{}{
		"status":    "error",
		"error":     errorMsg,
		"step":      state.CurrentStep,
		"iteration": parsed.IterIdx,
		"skipped":   true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Track error count in loop metadata
	errorCount := 0
	if ec, ok := loopMeta["error_count"].(float64); ok {
		errorCount = int(ec)
	} else if ec, ok := loopMeta["error_count"].(int); ok {
		errorCount = ec
	}
	loopMeta["error_count"] = errorCount + 1
	state.CollectedData["loop_metadata"] = loopMeta

	// Determine next step
	var nextStep string
	if parsed.IterIdx < totalIterations-1 {
		// Advance to next iteration's first substep
		nextStep = fmt.Sprintf("%s_iter_%d_%s", parsed.LoopName, parsed.IterIdx+1, firstSubstep)
	} else {
		// Last iteration — go to loop_complete
		nextStep = fmt.Sprintf("%s_complete", parsed.LoopName)
	}

	logger.Warn("Skipping failed loop iteration, advancing to next",
		zap.String("failed_step", state.CurrentStep),
		zap.String("error", errorMsg),
		zap.Int("failed_iteration", parsed.IterIdx),
		zap.Int("total_iterations", totalIterations),
		zap.String("next_step", nextStep),
		zap.Int("total_errors", errorCount+1))

	// Update state
	state.CurrentStep = nextStep
	state.Status = StatusExecutingStep
	state.LastActivity = time.Now()

	// Persist
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to persist state after loop skip: %w", err)
	}

	// Continue execution from the next step
	execCtx := s.createContinuationContext(state)
	return s.continueExecution(ctx, state, execCtx)
}

// skipToNextLoopIterationForAsync is the same as skipToNextLoopIteration but
// also cleans up the awaited request that failed. Used by handleUnrecoverableError
// and handleRequestTimeout.
func (s *SagaCoordinator) skipToNextLoopIterationForAsync(
	ctx context.Context,
	state *OrchestrationState,
	requestID string,
	errorMsg string,
	logger *zap.Logger,
) error {
	// Clean up the failed awaited request
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.MarkAwaitedRequestComplete(ctx, requestID); err != nil {
		logger.Warn("Failed to mark awaited request complete during loop skip",
			zap.String("request_id", requestID),
			zap.Error(err))
	}
	delete(state.AwaitedRequests, requestID)

	return s.skipToNextLoopIteration(ctx, state, errorMsg, logger)
}
