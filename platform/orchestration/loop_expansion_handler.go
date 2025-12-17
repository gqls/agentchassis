// Loop Expansion Handler for Coordinator
// This handles the loop action result and injects steps into the workflow

package orchestration

import (
	"fmt"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// handleLoopExpansion injects loop substeps into the workflow plan
// Called when an action returns loop_action: true
func (s *SagaCoordinator) handleLoopExpansion(
	state *OrchestrationState,
	loopResult map[string]interface{},
	logger *zap.Logger,
) error {

	logger.Info("Handling loop expansion")

	// Extract loop metadata
	loopName, _ := loopResult["loop_name"].(string)
	items, _ := loopResult["items"].([]interface{})
	loopVar, _ := loopResult["loop_var"].(string)
	nextStep, _ := loopResult["next_step"].(string)
	outputField, _ := loopResult["output_field"].(string)

	substepsMap, _ := loopResult["substeps"].(map[string]models.Step)
	substepOrder, _ := loopResult["substep_order"].([]string)

	if len(items) == 0 {
		logger.Info("No items to iterate, skipping to next step")
		state.CurrentStep = nextStep
		state.CollectedData[outputField] = []interface{}{}
		return nil
	}

	logger.Info("Expanding loop",
		zap.String("loop_name", loopName),
		zap.Int("iterations", len(items)),
		zap.Int("substeps_per_iteration", len(substepOrder)),
	)

	// Initialize loop metadata in CollectedData
	loopMetadata := map[string]interface{}{
		"loop_name":         loopName,
		"total_iterations":  len(items),
		"current_iteration": 0,
		"iteration_results": []interface{}{},
	}
	state.CollectedData["loop_metadata"] = loopMetadata

	// Inject steps for each iteration
	firstStepName := ""

	for iterIdx, item := range items {
		logger.Debug("Creating iteration steps",
			zap.Int("iteration", iterIdx),
		)

		// Create steps for this iteration
		for substepIdx, substepName := range substepOrder {
			substep := substepsMap[substepName]

			// Generate unique step name: loopname_iter_N_substepname
			injectedStepName := fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx, substepName)

			// Clone the step
			// Clone the step
			injectedStep := models.Step{
				Action:      substep.Action,
				Description: fmt.Sprintf("[Iteration %d] %s", iterIdx, substep.Description),
				OutputField: makeIterationOutputField(substep.OutputField, iterIdx), // ← FIX: Make unique
				Topic:       substep.Topic,
				Config:      cloneConfig(substep.Config),
			}

			// Determine next step
			if substepIdx < len(substepOrder)-1 {
				// Next substep in same iteration
				nextSubstepName := substepOrder[substepIdx+1]
				injectedStep.NextStep = fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx, nextSubstepName)
			} else if iterIdx < len(items)-1 {
				// First substep of next iteration
				nextIterFirstSubstep := substepOrder[0]
				injectedStep.NextStep = fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx+1, nextIterFirstSubstep)
			} else {
				// Last substep of last iteration -> complete step
				injectedStep.NextStep = fmt.Sprintf("%s_complete", loopName)
			}

			// Add iteration context to config
			if injectedStep.Config == nil {
				injectedStep.Config = make(map[string]interface{})
			}
			injectedStep.Config["loop_iteration"] = iterIdx
			injectedStep.Config["loop_item_index"] = iterIdx

			// Store the current item in a field accessible by substeps
			// We'll set this in CollectedData before executing
			injectedStep.Config["loop_var_name"] = loopVar

			// Track first and last step names
			if iterIdx == 0 && substepIdx == 0 {
				firstStepName = injectedStepName
			}
			//lastStepName = injectedStep.NextStep

			// Inject into workflow plan
			state.WorkflowPlan.Steps[injectedStepName] = injectedStep

			logger.Debug("Injected step",
				zap.String("step_name", injectedStepName),
				zap.String("action", injectedStep.Action),
				zap.String("next_step", injectedStep.NextStep),
			)
		}

		// Store item in a predictable location for this iteration
		// We'll set it in CollectedData when we reach each iteration
		itemKey := fmt.Sprintf("%s_item_%d", loopName, iterIdx)
		state.CollectedData[itemKey] = item
	}

	// Create completion step that aggregates results
	completeStepName := fmt.Sprintf("%s_complete", loopName)
	state.WorkflowPlan.Steps[completeStepName] = models.Step{
		Action:      "loop_complete",
		Description: fmt.Sprintf("Aggregate results from %s", loopName),
		NextStep:    nextStep,
		OutputField: outputField,
		Config: map[string]interface{}{
			"loop_name":        loopName,
			"total_iterations": len(items),
		},
	}

	logger.Info("Loop expansion complete",
		zap.String("first_step", firstStepName),
		zap.String("complete_step", completeStepName),
		zap.String("final_next_step", nextStep),
		zap.Int("total_steps_injected", len(items)*len(substepOrder)+1),
	)

	// Set current step to first injected step
	state.CurrentStep = firstStepName

	return nil
}

// cloneConfig deep copies a config map
func cloneConfig(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return make(map[string]interface{})
	}

	clone := make(map[string]interface{})
	for k, v := range config {
		clone[k] = v
	}
	return clone
}

// makeIterationOutputField makes output fields unique per iteration
// Example: "page_html" + 0 → "page_html_0"
func makeIterationOutputField(outputField string, iterIdx int) string {
	if outputField == "" {
		return ""
	}
	return fmt.Sprintf("%s_%d", outputField, iterIdx)
}

// setLoopVariable sets the current loop variable in CollectedData before executing a loop substep
func (s *SagaCoordinator) setLoopVariable(
	state *OrchestrationState,
	stepConfig models.Step,
	logger *zap.Logger,
) {
	// Check if this is a loop iteration step
	loopIteration, hasIteration := stepConfig.Config["loop_iteration"]
	loopVarName, hasVarName := stepConfig.Config["loop_var_name"].(string)

	if !hasIteration || !hasVarName {
		return // Not a loop step
	}

	iterIdx, ok := loopIteration.(int)
	if !ok {
		iterIdx = int(loopIteration.(float64))
	}

	// Get loop metadata
	loopMetadata, ok := state.CollectedData["loop_metadata"].(map[string]interface{})
	if !ok {
		logger.Warn("Loop metadata not found in CollectedData")
		return
	}

	loopName, _ := loopMetadata["loop_name"].(string)

	// Get the item for this iteration
	itemKey := fmt.Sprintf("%s_item_%d", loopName, iterIdx)
	item, exists := state.CollectedData[itemKey]
	if !exists {
		logger.Warn("Loop item not found",
			zap.String("item_key", itemKey),
		)
		return
	}

	// Set the loop variable
	state.CollectedData[loopVarName] = item

	// Update current iteration in metadata
	loopMetadata["current_iteration"] = iterIdx

	logger.Debug("Set loop variable",
		zap.String("loop_var", loopVarName),
		zap.Int("iteration", iterIdx),
	)
}
