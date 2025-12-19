// Loop Action Implementation
// This action dynamically expands substeps into the workflow for each iteration

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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

	// PRESCRIPTIVE: Collect from known output fields
	iterationResults := make([]interface{}, 0, totalIterations)

	for i := 0; i < totalIterations; i++ {
		// KNOWN PATH: page_html_0, page_html_1, etc.
		outputKey := fmt.Sprintf("page_html_%d", i)

		stepResult, exists := params.CollectedData[outputKey]
		if !exists {
			logger.Error("Output field missing for iteration",
				zap.String("expected_key", outputKey),
				zap.Int("iteration", i))
			continue
		}

		// Extract HTML from the stored result
		html := extractHTMLFromResult(stepResult, logger)
		if html == "" {
			logger.Error("No HTML in result",
				zap.String("output_key", outputKey),
				zap.Int("iteration", i))
			continue
		}

		// Get page name from loop item
		itemKey := fmt.Sprintf("%s_item_%d", loopName, i)
		item := params.CollectedData[itemKey]
		pageName := extractPageNameFromItem(item)

		if pageName == "" {
			pageName = fmt.Sprintf("page_%d", i)
		}

		// Build result
		result := map[string]interface{}{
			"name":      pageName,
			"page_html": html,
			"iteration": i,
		}
		iterationResults = append(iterationResults, result)

		logger.Info("Collected iteration result",
			zap.String("page_name", pageName),
			zap.Int("html_length", len(html)),
			zap.Int("iteration", i))
	}

	logger.Info("Loop completion finished",
		zap.Int("results_collected", len(iterationResults)),
	)

	return map[string]interface{}{
		"iterations": totalIterations,
		"results":    iterationResults,
	}, nil
}

// extractHTMLFromResult extracts HTML from a step result
// Html-developer stores result as map with "final_html" field
func extractHTMLFromResult(result interface{}, logger *zap.Logger) string {
	// Result should be a map
	m, ok := result.(map[string]interface{})
	if !ok {
		logger.Error("Result is not a map",
			zap.String("type", fmt.Sprintf("%T", result)))
		return ""
	}

	// PRESCRIPTIVE: Html-developer returns final_html
	html, ok := m["final_html"].(string)
	if !ok || html == "" {
		logger.Error("final_html field missing or empty",
			zap.Strings("available_keys", datahelpers.GetMapKeys(m)))
		return ""
	}

	return html
}

// extractPageNameFromItem extracts page name from loop item
func extractPageNameFromItem(item interface{}) string {
	// Loop items are strings (page names)
	if name, ok := item.(string); ok {
		return name
	}

	// Or maps with "name" field
	if m, ok := item.(map[string]interface{}); ok {
		if name, ok := m["name"].(string); ok {
			return name
		}
	}

	return ""
}

// parseSubsteps converts substep config into Step structs
func parseSubsteps(substepsConfig map[string]interface{}, logger *zap.Logger) (map[string]models.Step, []string, error) {
	substeps := make(map[string]models.Step)

	// First pass: parse all substeps
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

		if config, ok := stepMap["config"].(map[string]interface{}); ok {
			step.Config = config
		} else {
			step.Config = make(map[string]interface{})
		}

		substeps[substepName] = step
	}

	// Second pass: build correct order by following next_step links
	order := buildSubstepOrder(substeps, logger)

	logger.Info("Parsed substeps",
		zap.Int("count", len(substeps)),
		zap.Strings("order", order),
	)

	return substeps, order, nil
}

// Build order by following next_step links
func buildSubstepOrder(substeps map[string]models.Step, logger *zap.Logger) []string {
	// Find the first step (one with no incoming next_step references)
	hasIncoming := make(map[string]bool)
	for _, step := range substeps {
		if step.NextStep != "" {
			hasIncoming[step.NextStep] = true
		}
	}

	var firstStep string
	for name := range substeps {
		if !hasIncoming[name] {
			firstStep = name
			break
		}
	}

	if firstStep == "" {
		// Fallback: alphabetical order
		logger.Warn("Could not determine substep order, using alphabetical")
		order := make([]string, 0, len(substeps))
		for name := range substeps {
			order = append(order, name)
		}
		sort.Strings(order)
		return order
	}

	// Follow next_step chain
	order := []string{firstStep}
	visited := make(map[string]bool)
	visited[firstStep] = true

	current := firstStep
	for {
		step := substeps[current]
		if step.NextStep == "" {
			break // End of chain
		}

		// next_step might reference a step outside substeps (like "complete")
		if _, exists := substeps[step.NextStep]; !exists {
			break
		}

		if visited[step.NextStep] {
			logger.Warn("Circular reference detected in substeps",
				zap.String("current", current),
				zap.String("next", step.NextStep))
			break
		}

		order = append(order, step.NextStep)
		visited[step.NextStep] = true
		current = step.NextStep
	}

	// Add any unvisited steps (shouldn't happen but be safe)
	for name := range substeps {
		if !visited[name] {
			order = append(order, name)
			logger.Warn("Substep not in chain, appending",
				zap.String("substep", name))
		}
	}

	return order
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

// extractHTMLFromStepData extracts HTML from a step's stored data
func extractHTMLFromStepData(stepData interface{}, logger *zap.Logger) string {
	if m, ok := stepData.(map[string]interface{}); ok {
		// Try common field names
		fields := []string{"final_html", "html", "page_html", "result", "output"}
		for _, field := range fields {
			if html, ok := m[field].(string); ok && html != "" {
				return html
			}
		}
	}

	// If it's a string, return directly
	if html, ok := stepData.(string); ok {
		return html
	}

	return ""
}

// extractPageName extracts page name from loop item
func extractPageName(item interface{}) string {
	// If item is a string, use it directly
	if name, ok := item.(string); ok {
		return name
	}

	// If item is a map, look for name fields
	if m, ok := item.(map[string]interface{}); ok {
		if name, ok := m["name"].(string); ok {
			return name
		}
		if name, ok := m["page"].(string); ok {
			return name
		}
	}

	return ""
}
