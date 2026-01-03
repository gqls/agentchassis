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

	// 1. Extract configuration - supports both naming conventions
	// Workflow definitions use: items_field, item_variable, sub_workflow
	// Original loop action used: iterate_over, loop_var, substeps
	config := params.StepConfig.Config

	// Get iterate_over path (supports both 'iterate_over' and 'items_field')
	iterateOverPath, ok := config["iterate_over"].(string)
	if !ok || iterateOverPath == "" {
		// Fallback to items_field (used by workflow definitions)
		iterateOverPath, ok = config["items_field"].(string)
		if !ok || iterateOverPath == "" {
			return nil, fmt.Errorf("loop action requires 'iterate_over' or 'items_field' field")
		}
	}

	// Get loop variable name (supports both 'loop_var' and 'item_variable')
	loopVar, ok := config["loop_var"].(string)
	if !ok || loopVar == "" {
		// Fallback to item_variable (used by workflow definitions)
		loopVar, ok = config["item_variable"].(string)
		if !ok || loopVar == "" {
			loopVar = "loop_item" // default
		}
	}

	maxIterations := 20 // default safety limit
	if max, ok := config["max_iterations"].(float64); ok {
		maxIterations = int(max)
	}

	// Get substeps (supports both 'substeps' and 'sub_workflow.steps')
	var substepsConfig map[string]interface{}
	var startStep string

	substepsConfig, ok = config["substeps"].(map[string]interface{})
	if !ok || len(substepsConfig) == 0 {
		// Fallback to sub_workflow structure (used by workflow definitions)
		if subWorkflow, swOk := config["sub_workflow"].(map[string]interface{}); swOk {
			if steps, stepsOk := subWorkflow["steps"].(map[string]interface{}); stepsOk {
				substepsConfig = steps
			}
			if ss, ssOk := subWorkflow["start_step"].(string); ssOk {
				startStep = ss
			}
		}

		if len(substepsConfig) == 0 {
			return nil, fmt.Errorf("loop action requires 'substeps' or 'sub_workflow.steps' field")
		}
	}

	logger.Info("Loop configuration",
		zap.String("iterate_over", iterateOverPath),
		zap.String("loop_var", loopVar),
		zap.Int("max_iterations", maxIterations),
		zap.Int("substep_count", len(substepsConfig)),
		zap.String("start_step", startStep),
	)

	// 2. Get the collection to iterate over
	collection, err := getNestedValueForLoop(params.CollectedData, iterateOverPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection at '%s': %w", iterateOverPath, err)
	}

	// Handle different array types - Go's type system requires explicit handling
	var items []interface{}
	switch v := collection.(type) {
	case []interface{}:
		items = v
	case []string:
		// Convert []string to []interface{}
		items = make([]interface{}, len(v))
		for i, s := range v {
			items[i] = s
		}
		logger.Info("Converted []string to []interface{}", zap.Int("count", len(items)))
	case []map[string]interface{}:
		// Convert []map[string]interface{} to []interface{}
		items = make([]interface{}, len(v))
		for i, m := range v {
			items[i] = m
		}
		logger.Info("Converted []map to []interface{}", zap.Int("count", len(items)))
	default:
		return nil, fmt.Errorf("iterate_over field '%s' is not an array (got %T)", iterateOverPath, collection)
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
	substeps, substepOrder, err := parseSubsteps(substepsConfig, startStep, logger)
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
// ==============================================================================
// ISSUE 27: LoopCompleteAction hardcoded to look for page_html_N keys
// ==============================================================================
//
// PROBLEM: LoopCompleteAction looks for "page_html_0", "page_html_1", etc.
// But the workflow uses output_field: "section_output", storing as
// "section_output_0", "section_output_1", etc.
//
// FIX: Try multiple patterns or read output_field from config
//
// File: platform/orchestration/actions/loop_actions.go
// Replace the key lookup section in LoopCompleteAction:
// ==============================================================================

// In LoopCompleteAction, replace the hardcoded key lookup with this:

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

	// Get the output field pattern from config, or use defaults
	outputFieldBase := "page_html"
	if base, ok := params.StepConfig.Config["output_field_base"].(string); ok && base != "" {
		outputFieldBase = base
	}

	// IMPROVED: Try multiple patterns to find iteration results
	keyPatterns := []string{
		outputFieldBase,  // From config (e.g., "page_html")
		"section_output", // page-content-writer pattern
		"page_html",      // Original pattern
		"rendered_html",  // Alternative pattern
		"result",         // Generic pattern
		loopName + "_iter_%d_render_from_template", // Full step name pattern
		loopName + "_iter_%d_render_section",       // Full step name pattern for LLM path
	}

	// Collect results from known output fields
	iterationResults := make([]interface{}, 0, totalIterations)

	for i := 0; i < totalIterations; i++ {
		var stepResult interface{}
		var foundKey string

		// Try each pattern
		for _, pattern := range keyPatterns {
			var outputKey string
			if strings.Contains(pattern, "%d") {
				outputKey = fmt.Sprintf(pattern, i)
			} else {
				outputKey = fmt.Sprintf("%s_%d", pattern, i)
			}

			if result, exists := params.CollectedData[outputKey]; exists {
				stepResult = result
				foundKey = outputKey
				break
			}
		}

		// Also try the full step name patterns for rendered output
		if stepResult == nil {
			// Try render_from_template result
			templateKey := fmt.Sprintf("%s_iter_%d_render_from_template", loopName, i)
			if result, exists := params.CollectedData[templateKey]; exists {
				stepResult = result
				foundKey = templateKey
			}
		}

		if stepResult == nil {
			// Try render_section result (LLM path)
			sectionKey := fmt.Sprintf("%s_iter_%d_render_section", loopName, i)
			if result, exists := params.CollectedData[sectionKey]; exists {
				stepResult = result
				foundKey = sectionKey
			}
		}

		if stepResult == nil {
			logger.Error("Output field missing for iteration",
				zap.Int("iteration", i),
				zap.Strings("tried_patterns", keyPatterns))
			continue
		}

		logger.Debug("Found iteration result",
			zap.Int("iteration", i),
			zap.String("key", foundKey))

		// Extract HTML from the stored result
		html := extractHTMLFromResult(stepResult, logger)
		if html == "" {
			logger.Warn("No HTML in result",
				zap.String("output_key", foundKey),
				zap.Int("iteration", i))
			// Don't skip - collect even empty results to maintain order
		}

		// Get page/section name from loop item
		itemKey := fmt.Sprintf("%s_item_%d", loopName, i)
		item := params.CollectedData[itemKey]
		pageName := extractPageNameFromItem(item)

		if pageName == "" {
			pageName = fmt.Sprintf("section_%d", i)
		}

		// Build result
		result := map[string]interface{}{
			"name":          pageName,
			"page_html":     html,
			"rendered_html": html,
			"iteration":     i,
			"source_key":    foundKey,
		}

		// Also include component info if available
		if resultMap, ok := stepResult.(map[string]interface{}); ok {
			if compID, ok := resultMap["component_id"]; ok {
				result["component_id"] = compID
			}
			if compName, ok := resultMap["component_name"]; ok {
				result["component_name"] = compName
			}
			if compFunc, ok := resultMap["component_function"]; ok {
				result["component_function"] = compFunc
			}
		}

		iterationResults = append(iterationResults, result)

		logger.Info("Collected iteration result",
			zap.String("name", pageName),
			zap.Int("html_length", len(html)),
			zap.Int("iteration", i),
			zap.String("source", foundKey))
	}

	logger.Info("Loop completion finished",
		zap.Int("results_collected", len(iterationResults)),
		zap.Int("total_iterations", totalIterations),
	)

	return map[string]interface{}{
		"iterations": totalIterations,
		"results":    iterationResults,
		"count":      len(iterationResults),
	}, nil
}

// extractHTMLFromResult extracts HTML from various result formats
func extractHTMLFromResult(stepResult interface{}, logger *zap.Logger) string {
	if stepResult == nil {
		return ""
	}

	// Direct string
	if html, ok := stepResult.(string); ok {
		return html
	}

	// Map with various HTML field names
	if resultMap, ok := stepResult.(map[string]interface{}); ok {
		// Try common field names in order of preference
		htmlFields := []string{
			"rendered_html",
			"html",
			"final_html",
			"page_html",
			"content",
			"result",
			"output",
		}

		for _, field := range htmlFields {
			if html, ok := resultMap[field].(string); ok && html != "" {
				logger.Debug("Extracted HTML from field", zap.String("field", field))
				return html
			}
		}
	}

	return ""
}

// extractPageNameFromItem extracts the page/section name from a loop item
func extractPageNameFromItem(item interface{}) string {
	if item == nil {
		return ""
	}

	if itemMap, ok := item.(map[string]interface{}); ok {
		// Try various name fields
		nameFields := []string{"name", "page_name", "function", "display_name", "title"}
		for _, field := range nameFields {
			if name, ok := itemMap[field].(string); ok && name != "" {
				return name
			}
		}
	}

	return ""
}

// parseSubsteps converts substep config into Step structs
func parseSubsteps(substepsConfig map[string]interface{}, startStep string, logger *zap.Logger) (map[string]models.Step, []string, error) {
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
	order := buildSubstepOrder(substeps, startStep, logger)

	logger.Info("Parsed substeps",
		zap.Int("count", len(substeps)),
		zap.Strings("order", order),
	)

	return substeps, order, nil
}

// Build order by following next_step links, using explicit start_step if provided
func buildSubstepOrder(substeps map[string]models.Step, startStep string, logger *zap.Logger) []string {
	var firstStep string

	// Use explicit start_step if provided (from sub_workflow.start_step)
	if startStep != "" {
		if _, exists := substeps[startStep]; exists {
			firstStep = startStep
		} else {
			logger.Warn("Specified start_step not found in substeps, will auto-detect",
				zap.String("start_step", startStep))
		}
	}

	// If no explicit start or it wasn't found, find first step by looking for one with no incoming references
	if firstStep == "" {
		hasIncoming := make(map[string]bool)
		for _, step := range substeps {
			if step.NextStep != "" {
				hasIncoming[step.NextStep] = true
			}
		}

		for name := range substeps {
			if !hasIncoming[name] {
				firstStep = name
				break
			}
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
