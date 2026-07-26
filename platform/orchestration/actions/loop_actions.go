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

	// Check for allow_missing config - if true, missing collection returns empty result
	allowMissing := false
	if am, ok := config["allow_missing"].(bool); ok {
		allowMissing = am
	}

	// continue_on_error: failed iterations are skipped rather than
	// failing the entire workflow
	continueOnError := false
	if coe, ok := config["continue_on_error"].(bool); ok {
		continueOnError = coe
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
		// Check if this is a "not found" error vs other errors
		errStr := err.Error()
		isMissingData := strings.Contains(errStr, "not found") ||
			strings.Contains(errStr, "key") ||
			strings.Contains(errStr, "does not exist")

		if isMissingData {
			// Log detailed diagnostic info
			logger.Warn("Collection data not found for loop - possible state persistence issue",
				zap.String("iterate_over", iterateOverPath),
				zap.Error(err),
				zap.Strings("available_keys", getCollectedDataKeys(params.CollectedData)),
			)

			// Check if upstream step failed
			upstreamFailed, failReason := checkUpstreamFailure(params.CollectedData, iterateOverPath, logger)

			if upstreamFailed || allowMissing {
				// Return graceful skip result
				skipReason := "collection data not found"
				if upstreamFailed {
					skipReason = fmt.Sprintf("upstream failure: %s", failReason)
				}

				logger.Warn("Skipping loop due to missing collection",
					zap.String("reason", skipReason),
					zap.Bool("upstream_failed", upstreamFailed),
					zap.Bool("allow_missing", allowMissing),
				)

				return map[string]interface{}{
					"loop_action":  true,
					"iterations":   0,
					"results":      []interface{}{},
					"skipped":      true,
					"skip_reason":  skipReason,
					"missing_path": iterateOverPath,
				}, nil
			}
		}

		// Return original error for non-missing-data errors
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
			"loop_action": false, // Explicitly mark as processed loop
			"iterations":  0,
			"results":     []interface{}{},
			"skipped":     true,
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
		"loop_action":       true,
		"loop_name":         loopName,
		"items":             items,
		"loop_var":          loopVar,
		"substeps":          substeps,
		"substep_order":     substepOrder,
		"next_step":         params.StepConfig.NextStep,
		"output_field":      params.StepConfig.OutputField,
		"total_iterations":  len(items),
		"continue_on_error": continueOnError,
	}

	logger.Info("Loop expansion prepared",
		zap.Int("iterations", len(items)),
		zap.Int("substeps_per_iteration", len(substeps)),
		zap.Int("total_steps_to_inject", len(items)*len(substeps)),
	)

	return expansion, nil
}

// checkUpstreamFailure checks if an upstream step that should have produced the data failed
func checkUpstreamFailure(collectedData map[string]interface{}, path string, logger *zap.Logger) (bool, string) {
	// path is like "section_components.components"
	// Check if section_components itself exists but has error/failure indicators

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return false, ""
	}

	topKey := parts[0]

	// Check if the top-level key exists
	if _, exists := collectedData[topKey]; !exists {
		// The step that should have produced this data either:
		// 1. Was never executed
		// 2. Failed and its result wasn't stored
		// 3. Succeeded but state wasn't persisted (race condition)

		// Look for any error indicators in the state
		for key, val := range collectedData {
			if valMap, ok := val.(map[string]interface{}); ok {
				if status, ok := valMap["status"].(string); ok && status == "failed" {
					logger.Info("Found failed step that may have prevented data creation",
						zap.String("failed_step", key),
						zap.String("missing_key", topKey))
					return true, fmt.Sprintf("step '%s' failed", key)
				}
				if _, hasError := valMap["error"]; hasError {
					logger.Info("Found step with error that may have prevented data creation",
						zap.String("error_step", key),
						zap.String("missing_key", topKey))
					return true, fmt.Sprintf("step '%s' had error", key)
				}
			}
		}

		// No explicit failure found, but data is still missing
		// This is likely a race condition / state persistence issue
		logger.Warn("Data missing without explicit upstream failure - possible race condition",
			zap.String("missing_key", topKey))
		return true, fmt.Sprintf("data '%s' missing (possible state persistence race)", topKey)
	}

	// Top-level key exists, check if it has failure indicators
	if topData, ok := collectedData[topKey].(map[string]interface{}); ok {
		if status, ok := topData["status"].(string); ok && status == "failed" {
			return true, fmt.Sprintf("step '%s' has failed status", topKey)
		}
		if errMsg, ok := topData["error"].(string); ok && errMsg != "" {
			return true, fmt.Sprintf("step '%s' has error: %s", topKey, errMsg)
		}
	}

	return false, ""
}

// Use substep_output_fields from the _complete step config (set by Part A of
// this fix in handleLoopExpansion) as the primary lookup. Fall back to legacy
// HTML patterns for orchestrations that were expanded before the Part A fix
// was deployed.
func LoopCompleteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "loop_complete"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("Starting loop completion")

	// DIAGNOSTIC: Log ALL available keys
	allKeys := make([]string, 0, len(params.CollectedData))
	for k := range params.CollectedData {
		allKeys = append(allKeys, k)
	}
	logger.Info("LoopComplete: Available CollectedData keys",
		zap.Strings("all_keys", allKeys))

	// Get loop metadata — try step config first (set by handleLoopExpansion),
	// fall back to shared loop_metadata in CollectedData.
	var totalIterations int
	var loopName string

	// The _complete step always has loop_name and total_iterations in its config
	// (set by handleLoopExpansion). These are per-loop so they're reliable even
	// when multiple loops exist in the same orchestration.
	if ln, ok := params.StepConfig.Config["loop_name"].(string); ok {
		loopName = ln
	}
	if ti, ok := params.StepConfig.Config["total_iterations"].(float64); ok {
		totalIterations = int(ti)
	} else if ti, ok := params.StepConfig.Config["total_iterations"].(int); ok {
		totalIterations = ti
	}

	// Fallback to shared loop_metadata (backward compat, pre-expansion fix)
	if totalIterations == 0 {
		if loopMetadata, ok := params.CollectedData["loop_metadata"].(map[string]interface{}); ok {
			if f, ok := loopMetadata["total_iterations"].(float64); ok {
				totalIterations = int(f)
			} else if i, ok := loopMetadata["total_iterations"].(int); ok {
				totalIterations = i
			}
			if loopName == "" {
				loopName, _ = loopMetadata["loop_name"].(string)
			}
		}
	}

	if totalIterations == 0 {
		return nil, fmt.Errorf("total_iterations not found in step config or loop_metadata")
	}

	logger.Info("Aggregating loop results",
		zap.String("loop_name", loopName),
		zap.Int("total_iterations", totalIterations),
	)

	// === PRIMARY PATH: Use substep_output_fields from config ===
	// These are set by the Part A fix in handleLoopExpansion.
	// Each field becomes {field}_{N} in CollectedData after makeIterationOutputField.
	var substepFields []string
	if fields, ok := params.StepConfig.Config["substep_output_fields"].([]interface{}); ok {
		for _, f := range fields {
			if s, ok := f.(string); ok && s != "" {
				substepFields = append(substepFields, s)
			}
		}
	}

	// Also check output_field_base from config (legacy mechanism, still useful)
	outputFieldBase := ""
	if base, ok := params.StepConfig.Config["output_field_base"].(string); ok && base != "" {
		outputFieldBase = base
	}

	iterationResults := make([]interface{}, 0, totalIterations)

	for i := 0; i < totalIterations; i++ {
		result := map[string]interface{}{
			"iteration": i,
		}
		foundAny := false

		// --- Strategy 1: Collect ALL substep outputs for this iteration ---
		if len(substepFields) > 0 {
			for _, field := range substepFields {
				key := fmt.Sprintf("%s_%d", field, i)
				if val, exists := params.CollectedData[key]; exists {
					result[field] = val
					foundAny = true

					// If this substep output contains HTML, also set page_html
					// for backward compatibility with deployer-agent and other consumers.
					if html := extractHTMLFromResult(val, logger); html != "" {
						result["page_html"] = html
						result["rendered_html"] = html
						result["source_key"] = key
					}
				}
			}

			if foundAny {
				logger.Debug("Collected iteration via substep_output_fields",
					zap.Int("iteration", i),
					zap.Int("fields_found", countNonMeta(result)))
			}
		}

		// --- Strategy 2: Legacy HTML patterns (backward compat) ---
		// Only used when substep_output_fields didn't find anything.
		// This covers old orchestrations expanded before Part A was deployed.
		if !foundAny {
			legacyPatterns := buildLegacyPatterns(outputFieldBase, loopName)

			for _, pattern := range legacyPatterns {
				var outputKey string
				if strings.Contains(pattern, "%d") {
					outputKey = fmt.Sprintf(pattern, i)
				} else {
					outputKey = fmt.Sprintf("%s_%d", pattern, i)
				}

				if val, exists := params.CollectedData[outputKey]; exists {
					html := extractHTMLFromResult(val, logger)
					if html != "" {
						result["page_html"] = html
						result["rendered_html"] = html
						result["source_key"] = outputKey
						foundAny = true

						// Also include component info if available
						if resultMap, ok := val.(map[string]interface{}); ok {
							for _, metaKey := range []string{"component_id", "component_name", "component_function"} {
								if v, ok := resultMap[metaKey]; ok {
									result[metaKey] = v
								}
							}
						}
						break
					}
				}
			}
		}

		// --- Strategy 3: Generic fallback — scan ALL keys matching this iteration ---
		// Collects any key with _{i} suffix, regardless of whether it contains HTML.
		if !foundAny {
			iterSuffix := fmt.Sprintf("_%d", i)
			iterPrefix := fmt.Sprintf("%s_iter_%d_", loopName, i)

			for k, v := range params.CollectedData {
				if strings.HasSuffix(k, iterSuffix) || strings.HasPrefix(k, iterPrefix) {
					// Derive the base field name by stripping the suffix
					baseName := k
					if strings.HasSuffix(k, iterSuffix) {
						baseName = strings.TrimSuffix(k, iterSuffix)
					} else if strings.HasPrefix(k, iterPrefix) {
						baseName = strings.TrimPrefix(k, iterPrefix)
					}
					result[baseName] = v
					foundAny = true

					// Check for HTML in this result too
					if html := extractHTMLFromResult(v, logger); html != "" && result["page_html"] == nil {
						result["page_html"] = html
						result["rendered_html"] = html
						result["source_key"] = k
					}
				}
			}

			if foundAny {
				logger.Info("LoopComplete: Found results via generic fallback scan",
					zap.Int("iteration", i),
					zap.Int("fields_found", countNonMeta(result)))
			}
		}

		// Check for error results (from continue_on_error iterations)
		errorKey := fmt.Sprintf("%s_iter_%d_error", loopName, i)
		if errResult, exists := params.CollectedData[errorKey]; exists {
			result["error"] = errResult
			result["status"] = "error"
			foundAny = true
		}

		if !foundAny {
			logger.Warn("No results found for iteration",
				zap.Int("iteration", i))
			result["status"] = "missing"
		}

		// Get page/section name from loop item
		itemKey := fmt.Sprintf("%s_item_%d", loopName, i)
		item := params.CollectedData[itemKey]
		pageName := extractPageNameFromItem(item)
		if pageName == "" {
			pageName = fmt.Sprintf("item_%d", i)
		}
		result["name"] = pageName

		// Also include the original item for reference
		if item != nil {
			result["original_item"] = item
		}

		iterationResults = append(iterationResults, result)

		logger.Info("Collected iteration result",
			zap.String("name", pageName),
			zap.Int("iteration", i),
			zap.Bool("has_html", result["page_html"] != nil),
			zap.String("status", fmt.Sprintf("%v", result["status"])))
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

// buildLegacyPatterns returns the hardcoded HTML-specific key patterns.
// Used as fallback for orchestrations expanded before the substep_output_fields
// fix was deployed.
func buildLegacyPatterns(outputFieldBase string, loopName string) []string {
	patterns := []string{}

	if outputFieldBase != "" {
		patterns = append(patterns, outputFieldBase)
	}

	patterns = append(patterns,
		"section_output",
		"page_html",
		"page_result",
		"rendered_html",
		"result",
		"generated_content",
	)

	if loopName != "" {
		patterns = append(patterns,
			loopName+"_iter_%d_render_from_template",
			loopName+"_iter_%d_render_section",
			loopName+"_iter_%d_section_output",
			loopName+"_iter_%d_generated_content",
			loopName+"_iter_%d_complete_page",
			loopName+"_iter_%d_call_rerender",
		)
	}

	return patterns
}

// countNonMeta counts result fields that aren't iteration/name metadata
func countNonMeta(result map[string]interface{}) int {
	count := 0
	for k := range result {
		if k != "iteration" && k != "name" && k != "status" {
			count++
		}
	}
	return count
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
			// Same omission as convertToWorkflowPlan carried into loop substeps:
			// without this a substep's error_step never reaches the expanded plan
			// (bugs_open/086). Expansion prefixes it per iteration.
			ErrorStep:   getStringValue(stepMap, "error_step"),
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
