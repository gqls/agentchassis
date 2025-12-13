// Loop Action Implementation
// This action dynamically expands substeps into the workflow for each iteration

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// LoopAction executes substeps for each item in a collection
// It dynamically injects steps into the workflow plan for async execution
func LoopAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "loop"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("Starting loop action")

	// 1. Extract configuration
	config := params.StepConfig.Config

	iterateOverPath, ok := config["iterate_over"].(string)
	if !ok || iterateOverPath == "" {
		return nil, fmt.Errorf("loop action requires 'iterate_over' field")
	}

	loopVar, ok := config["loop_var"].(string)
	if !ok || loopVar == "" {
		loopVar = "loop_item" // default
	}

	maxIterations := 20 // default safety limit
	if max, ok := config["max_iterations"].(float64); ok {
		maxIterations = int(max)
	}

	substepsConfig, ok := config["substeps"].(map[string]interface{})
	if !ok || len(substepsConfig) == 0 {
		return nil, fmt.Errorf("loop action requires 'substeps' field")
	}

	logger.Info("Loop configuration",
		zap.String("iterate_over", iterateOverPath),
		zap.String("loop_var", loopVar),
		zap.Int("max_iterations", maxIterations),
		zap.Int("substep_count", len(substepsConfig)),
	)

	// 2. Get the collection to iterate over
	collection, err := getNestedValueForLoop(params.CollectedData, iterateOverPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection at '%s': %w", iterateOverPath, err)
	}

	items, ok := collection.([]interface{})
	if !ok {
		return nil, fmt.Errorf("iterate_over field '%s' is not an array", iterateOverPath)
	}

	if len(items) == 0 {
		logger.Warn("Collection is empty, skipping loop")
		return map[string]interface{}{
			"iterations": 0,
			"results":    []interface{}{},
		}, nil
	}

	if len(items) > maxIterations {
		logger.Warn("Collection exceeds max_iterations, truncating",
			zap.Int("collection_size", len(items)),
			zap.Int("max_iterations", maxIterations),
		)
		items = items[:maxIterations]
	}

	logger.Info("Starting loop expansion",
		zap.Int("iterations", len(items)),
	)

	// 3. Parse substeps into Step structs
	substeps, substepOrder, err := parseSubsteps(substepsConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse substeps: %w", err)
	}

	// 4. Get workflow plan from state (we need to inject steps)
	// This is passed through params.State or params.WorkflowPlan
	// For now, we'll return a special marker that the coordinator recognizes
	// and handles the injection there

	loopName := params.ExecutionContext.StepName

	// Build the expansion data
	expansion := map[string]interface{}{
		"loop_action":      true,
		"loop_name":        loopName,
		"items":            items,
		"loop_var":         loopVar,
		"substeps":         substeps,
		"substep_order":    substepOrder,
		"next_step":        params.StepConfig.NextStep,
		"output_field":     params.StepConfig.OutputField,
		"total_iterations": len(items),
	}

	logger.Info("Loop expansion prepared",
		zap.Int("iterations", len(items)),
		zap.Int("substeps_per_iteration", len(substeps)),
		zap.Int("total_steps_to_inject", len(items)*len(substeps)),
	)

	return expansion, nil
}

// LoopCompleteAction aggregates results from loop iterations
func LoopCompleteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "loop_complete"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("Starting loop completion")

	// Get loop metadata
	loopMetadata, ok := params.CollectedData["loop_metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("loop_metadata not found in CollectedData")
	}

	totalIterations, ok := loopMetadata["total_iterations"].(int)
	if !ok {
		if f, ok := loopMetadata["total_iterations"].(float64); ok {
			totalIterations = int(f)
		} else {
			return nil, fmt.Errorf("total_iterations not found in loop_metadata")
		}
	}

	loopName, _ := loopMetadata["loop_name"].(string)

	logger.Info("Aggregating loop results",
		zap.String("loop_name", loopName),
		zap.Int("total_iterations", totalIterations),
	)

	// Collect results from iteration_results in metadata
	// (populated by substeps that store results)
	iterationResults, ok := loopMetadata["iteration_results"].([]interface{})
	if !ok {
		iterationResults = []interface{}{}
	}

	// If iteration_results is empty, try to collect from output fields
	if len(iterationResults) == 0 {
		logger.Info("Collecting results from iteration output fields")

		// Try to find output fields from each iteration
		// This is a fallback if substeps didn't populate iteration_results
		for i := 0; i < totalIterations; i++ {
			// Look for common output field patterns
			resultKey := fmt.Sprintf("%s_iter_%d_result", loopName, i)
			if result, exists := params.CollectedData[resultKey]; exists {
				iterationResults = append(iterationResults, result)
			}
		}
	}

	logger.Info("Loop completion finished",
		zap.Int("results_collected", len(iterationResults)),
	)

	// Return aggregated results
	return map[string]interface{}{
		"iterations": totalIterations,
		"results":    iterationResults,
		"loop_name":  loopName,
	}, nil
}

// parseSubsteps converts substep config into Step structs
func parseSubsteps(substepsConfig map[string]interface{}, logger *zap.Logger) (map[string]models.Step, []string, error) {
	substeps := make(map[string]models.Step)
	order := []string{}

	for substepName, substepData := range substepsConfig {
		stepMap, ok := substepData.(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("substep '%s' is not a valid step definition", substepName)
		}

		step := models.Step{
			Action:      getStringValue(stepMap, "action"),
			Description: getStringValue(stepMap, "description"),
			NextStep:    getStringValue(stepMap, "next_step"),
			OutputField: getStringValue(stepMap, "output_field"),
			Topic:       getStringValue(stepMap, "topic"),
		}

		// Get config if present
		if config, ok := stepMap["config"].(map[string]interface{}); ok {
			step.Config = config
		} else {
			step.Config = make(map[string]interface{})
		}

		substeps[substepName] = step
		order = append(order, substepName)
	}

	logger.Info("Parsed substeps",
		zap.Int("count", len(substeps)),
		zap.Strings("order", order),
	)

	return substeps, order, nil
}

// getNestedValueForLoop safely traverses a map to find a value at a path
func getNestedValueForLoop(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found at position %d in path '%s'", part, i, path)
			}
			current = val
		case string:
			// Try to parse as JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err != nil {
				return nil, fmt.Errorf("cannot navigate into string at '%s'", part)
			}
			val, exists := parsed[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found in parsed JSON at position %d", part, i)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot navigate into type %T at '%s'", current, part)
		}
	}

	return current, nil
}

// getStringValue safely gets a string from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
