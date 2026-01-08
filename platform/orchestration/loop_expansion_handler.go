// Loop Expansion Handler for Coordinator
// This handles the loop action result and injects steps into the workflow

package orchestration

import (
	"fmt"
	"strings"

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

	// Build a set of valid substep names for quick lookup
	validSubstepSet := make(map[string]bool)
	for name := range substepsMap {
		validSubstepSet[name] = true
	}

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

	// Identify the first substep (from substepOrder[0])
	firstSubstepInOrder := ""
	if len(substepOrder) > 0 {
		firstSubstepInOrder = substepOrder[0]
	}

	for iterIdx, item := range items {
		logger.Debug("Creating iteration steps",
			zap.Int("iteration", iterIdx),
		)

		// Create steps for this iteration
		for _, substepName := range substepOrder {
			substep := substepsMap[substepName]

			// Generate unique step name: loopname_iter_N_substepname
			injectedStepName := fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx, substepName)

			// Deep clone the config to avoid mutating the original
			clonedConfig := deepCloneConfig(substep.Config)

			// Clone the step
			injectedStep := models.Step{
				Action:      substep.Action,
				Description: fmt.Sprintf("[Iteration %d] %s", iterIdx, substep.Description),
				OutputField: makeIterationOutputField(substep.OutputField, iterIdx),
				Topic:       substep.Topic,
				Config:      clonedConfig,
			}

			// =====================================================
			// Properly resolve NextStep based on workflow definition
			// =====================================================
			injectedStep.NextStep = resolveIterationNextStep(
				substep.NextStep,
				loopName,
				iterIdx,
				len(items),
				firstSubstepInOrder,
				validSubstepSet,
				logger,
			)

			// =====================================================
			// Also prefix config fields that reference other steps
			// This handles conditionals (then_step, else_step) and other step refs
			// =====================================================
			prefixConfigStepReferences(injectedStep.Config, loopName, iterIdx, validSubstepSet)

			// Add iteration context to config
			if injectedStep.Config == nil {
				injectedStep.Config = make(map[string]interface{})
			}
			injectedStep.Config["loop_iteration"] = iterIdx
			injectedStep.Config["loop_item_index"] = iterIdx
			injectedStep.Config["loop_var_name"] = loopVar

			// Track first step name
			if iterIdx == 0 && substepName == substepOrder[0] {
				firstStepName = injectedStepName
			}

			// Inject into workflow plan
			state.WorkflowPlan.Steps[injectedStepName] = injectedStep

			logger.Debug("Injected step",
				zap.String("step_name", injectedStepName),
				zap.String("action", injectedStep.Action),
				zap.String("original_next_step", substep.NextStep),
				zap.String("resolved_next_step", injectedStep.NextStep),
			)
		}

		// Store item in a predictable location for this iteration
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

// makeIterationOutputField makes output fields unique per iteration
// Example: "page_html" + 0 â†’ "page_html_0"
func makeIterationOutputField(outputField string, iterIdx int) string {
	if outputField == "" {
		return ""
	}
	return fmt.Sprintf("%s_%d", outputField, iterIdx)
}

// setLoopVariable sets the current loop variable in CollectedData before executing a loop substep
// It also propagates outputs from previous substeps in the same iteration to their original field names
func (s *SagaCoordinator) setLoopVariable(
	state *OrchestrationState,
	stepConfig models.Step,
	logger *zap.Logger,
) {
	// Check if this is a loop iteration step
	loopIteration, hasIteration := stepConfig.Config["loop_iteration"]
	loopVarName, hasVarName := stepConfig.Config["loop_var_name"].(string)

	// ISSUE 33 FIX: Add logging for early return cases
	if !hasIteration {
		logger.Debug("setLoopVariable: not a loop step (no loop_iteration in config)",
			zap.String("step", state.CurrentStep),
			zap.Any("config_keys", getConfigKeys(stepConfig.Config)),
		)
		return
	}
	if !hasVarName {
		logger.Warn("setLoopVariable: loop_var_name missing or not a string",
			zap.String("step", state.CurrentStep),
			zap.Any("loop_var_name_raw", stepConfig.Config["loop_var_name"]),
		)
		return
	}

	iterIdx, ok := loopIteration.(int)
	if !ok {
		if f, ok := loopIteration.(float64); ok {
			iterIdx = int(f)
		} else {
			logger.Error("setLoopVariable: loop_iteration is not int or float64",
				zap.String("step", state.CurrentStep),
				zap.Any("loop_iteration", loopIteration),
			)
			return
		}
	}

	// Get loop metadata
	loopMetadata, ok := state.CollectedData["loop_metadata"].(map[string]interface{})
	if !ok {
		logger.Warn("setLoopVariable: loop_metadata not found in CollectedData")
		return
	}

	loopName, _ := loopMetadata["loop_name"].(string)

	// Get the item for this iteration
	itemKey := fmt.Sprintf("%s_item_%d", loopName, iterIdx)
	item, exists := state.CollectedData[itemKey]
	if !exists {
		logger.Error("setLoopVariable: loop item not found",
			zap.String("item_key", itemKey),
			zap.Int("iteration", iterIdx),
			zap.Strings("available_keys", getTopLevelKeys(state.CollectedData)),
		)
		return
	}

	// Set the loop variable
	state.CollectedData[loopVarName] = item

	// Update current iteration in metadata
	loopMetadata["current_iteration"] = iterIdx

	// ISSUE 33 FIX: Log at INFO level so we can confirm this is working
	logger.Info("setLoopVariable: set loop variable successfully",
		zap.String("loop_var", loopVarName),
		zap.Int("iteration", iterIdx),
		zap.String("item_key", itemKey),
	)

	// Also store the item key so conditional can use it as fallback
	state.CollectedData["__current_loop_item_key__"] = itemKey

	// Propagate previous substep outputs from iteration-suffixed names to original names
	propagateIterationOutputs(state, iterIdx, logger)
}

// Helper to get config keys for logging
func getConfigKeys(config map[string]interface{}) []string {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	return keys
}

// Helper to get top-level keys for logging
func getTopLevelKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// propagateIterationOutputs copies iteration-suffixed output fields to their original names
// Example: page_content_0 -> page_content, page_html_0 -> page_html
// This allows substeps within a loop iteration to find data from previous substeps
func propagateIterationOutputs(state *OrchestrationState, iterIdx int, logger *zap.Logger) {
	suffix := fmt.Sprintf("_%d", iterIdx)

	// List of common output field base names that need propagation
	// These are the output_field names from substep configs before iteration suffix is added
	// Note: The auto-propagation scan below catches others, but explicitly listing
	// ensures these are propagated first (before the check for "alreadySet")
	commonOutputFields := []string{
		// Page building outputs
		"page_content",
		"page_html",
		"reviewed_content",
		"assembled_page",
		"page_deployed",

		// Content outputs
		"content_result",
		"html_result",

		// Site-level outputs
		"site_content",
		"site_architecture",

		// Research outputs
		"search_results",
		"scrape_results",
		"research_content",
		"synthesis",
	}

	propagatedCount := 0
	for _, baseName := range commonOutputFields {
		iterationKey := baseName + suffix // e.g., "page_content_0"

		if data, exists := state.CollectedData[iterationKey]; exists {
			// Copy to original field name for this iteration's scope
			state.CollectedData[baseName] = data
			propagatedCount++
			logger.Debug("Propagated iteration output to original field name",
				zap.String("from", iterationKey),
				zap.String("to", baseName),
				zap.Int("iteration", iterIdx),
			)
		}
	}

	// Also scan for any other iteration-suffixed keys and propagate them
	// This catches output fields not in the hardcoded list
	for key, value := range state.CollectedData {
		if strings.HasSuffix(key, suffix) {
			baseName := strings.TrimSuffix(key, suffix)
			// Don't overwrite if we already set it (avoid overwriting with wrong iteration data)
			if _, alreadySet := state.CollectedData[baseName]; !alreadySet {
				// Only propagate if baseName looks like an output field (not internal keys)
				if !strings.HasPrefix(baseName, "__") && !strings.Contains(baseName, "_item_") && !strings.Contains(baseName, "_iter_") {
					state.CollectedData[baseName] = value
					propagatedCount++
					logger.Debug("Auto-propagated iteration output",
						zap.String("from", key),
						zap.String("to", baseName),
						zap.Int("iteration", iterIdx),
					)
				}
			}
		}
	}

	if propagatedCount > 0 {
		logger.Info("propagateIterationOutputs completed",
			zap.Int("iteration", iterIdx),
			zap.Int("fields_propagated", propagatedCount),
		)
	}
}

// prefixStepReferencesInConfig updates step references in config to use iteration prefix
// This handles then_step, else_step in conditionals, and any other step references
func prefixStepReferencesInConfig(config map[string]interface{}, loopName string, iterIdx int, validSubsteps []string) {
	if config == nil {
		return
	}

	// Build a set of valid substep names for quick lookup
	validSubstepSet := make(map[string]bool)
	for _, name := range validSubsteps {
		validSubstepSet[name] = true
	}

	// Fields that contain step references
	stepRefFields := []string{"then_step", "else_step", "error_step", "fallback_step", "retry_step"}

	for _, field := range stepRefFields {
		if stepName, ok := config[field].(string); ok && stepName != "" {
			// Only prefix if it's a reference to a substep in our loop
			if validSubstepSet[stepName] {
				prefixedName := fmt.Sprintf("%s_iter_%d_%s", loopName, iterIdx, stepName)
				config[field] = prefixedName
			}
			// If it's not a valid substep, leave it alone (might be external step)
		}
	}

	// Also check for nested configs (like in some complex actions)
	for _, value := range config {
		if nestedConfig, ok := value.(map[string]interface{}); ok {
			prefixStepReferencesInConfig(nestedConfig, loopName, iterIdx, validSubsteps)
		}
	}
}

// ensure deep copy (if not already doing so)
func cloneConfig(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}

	dst := make(map[string]interface{})
	for k, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[k] = cloneConfig(val) // Deep clone nested maps
		case []interface{}:
			dst[k] = cloneSlice(val) // Deep clone slices
		default:
			dst[k] = v // Primitives are copied by value
		}
	}
	return dst
}

func cloneSlice(src []interface{}) []interface{} {
	if src == nil {
		return nil
	}
	dst := make([]interface{}, len(src))
	for i, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[i] = cloneConfig(val)
		case []interface{}:
			dst[i] = cloneSlice(val)
		default:
			dst[i] = v
		}
	}
	return dst
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func isMap(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}
